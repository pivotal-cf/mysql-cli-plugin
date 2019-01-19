package command_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command/commandfakes"
	"github.com/pkg/errors"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/cf"
	"log"
)

var _ = Describe("Plugin Commands", func() {
	var (
		fakeMigrator      *commandfakes.FakeMigrator
		logOutput         *bytes.Buffer
		me                *command.MigratorExecutor
		donorInstanceName string
		destPlan          string
		cleanup           bool
		fakeCliConnection *pluginfakes.FakeCliConnection
	)
	const usage = `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools version`

	BeforeEach(func() {
		fakeMigrator = new(commandfakes.FakeMigrator)
		me = &command.MigratorExecutor{
			Migrator: fakeMigrator,
		}
		logOutput = &bytes.Buffer{}
		log.SetOutput(logOutput)
		cleanup = true
	})
	Context("Migrate", func() {
		It("migrates data from a source service instance to a newly created instance", func() {
			donorInstanceName = "some-donor"
			destPlan = "some-plan"
			Expect(me.Migrate(donorInstanceName, destPlan, cleanup)).To(Succeed())

			By("log a message that we don't migrate triggers, routines and events", func() {
				Expect(logOutput.String()).To(ContainSubstring(`Warning: The mysql-tools migrate command will not migrate any triggers, routines or events`))
			})

			By("checking that donor exists", func() {
				Expect(fakeMigrator.CheckServiceExistsCallCount()).
					To(Equal(1))
				Expect(fakeMigrator.CheckServiceExistsArgsForCall(0)).
					To(Equal("some-donor"))
			})

			By("creating and configuring a new service instance", func() {
				Expect(fakeMigrator.CreateAndConfigureServiceInstanceCallCount()).
					To(Equal(1))

				createdServicePlan, createdServiceInstanceName := fakeMigrator.CreateAndConfigureServiceInstanceArgsForCall(0)
				Expect(createdServicePlan).To(Equal("some-plan"))
				Expect(createdServiceInstanceName).
					To(Equal("some-donor-new"))
			})

			By("migrating data from the donor to the recipient", func() {
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				migratedDonorName, migratedRecipientname, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(migratedDonorName).To(Equal("some-donor"))
				Expect(migratedRecipientname).To(Equal("some-donor-new"))
				Expect(cleanup).To(BeTrue())
			})

			By("renaming the service instances", func() {
				Expect(fakeMigrator.RenameServiceInstancesCallCount()).
					To(Equal(1))
				donorInstance, recipientName := fakeMigrator.RenameServiceInstancesArgsForCall(0)
				Expect(donorInstance).To(Equal("some-donor"))
				Expect(recipientName).To(Equal("some-donor-new"))
			})

			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(BeZero())
		})

		It("returns an error if the donor service instance does not exist", func() {
			fakeMigrator.CheckServiceExistsReturns(errors.New("some-donor does not exist"))
			donorInstanceName = "some-donor"
			destPlan = "some-plan"
			err := me.Migrate(donorInstanceName, destPlan, cleanup)
			Expect(err).To(MatchError("some-donor does not exist"))
		})

		Context("when creating a service instance fails", func() {
			BeforeEach(func() {
				fakeMigrator.CreateAndConfigureServiceInstanceReturns(errors.New("some-cf-error"))
				donorInstanceName = "some-donor"
				destPlan = "some-plan"
				cleanup = true
			})

			It("returns an error and attempts to delete the new service instance", func() {
				err := me.Migrate(donorInstanceName, destPlan, cleanup)
				Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				cleanup = false

				err := me.Migrate(donorInstanceName, destPlan, cleanup)
				Expect(err).To(MatchError("error creating service instance: some-cf-error. Not cleaning up service some-donor-new"))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})

		Context("when migrating data fails", func() {
			BeforeEach(func() {
				fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
				donorInstanceName = "some-donor"
				destPlan = "some-plan"
			})

			It("returns an error and attempts to delete the new service instance", func() {
				err := me.Migrate(donorInstanceName, destPlan, cleanup)
				Expect(err).To(MatchError(MatchRegexp("error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeTrue())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				cleanup = false

				err := me.Migrate(donorInstanceName, destPlan, cleanup)

				Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-donor-new"))
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeFalse())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})

		Context("when renaming the service instances fail", func() {
			BeforeEach(func() {
				fakeMigrator.RenameServiceInstancesReturns(errors.New("some-cf-error"))
			})

			It("returns an error and doesn't clean up regardless of --no-cleanup flag", func() {
				donorInstanceName = "some-donor"
				destPlan = "some-plan"
				err := me.Migrate(donorInstanceName, destPlan, cleanup)
				Expect(err).To(MatchError("some-cf-error"))
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeTrue())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})
	})
	Context("Execute", func() {
		var client *cf.Client
		//var cclient

		BeforeEach(func() {
			fakeCliConnection = &pluginfakes.FakeCliConnection{}

			client = cf.NewClient(fakeCliConnection)
		})

		It("returns an error if not enough args are passed", func() {
			args := []string{"source"}
			donorInstanceName = "some-donor"
			err := me.Execute(client, args)
			Expect(err).To(MatchError("the required argument `<p.mysql-plan-type>` was not provided"))
		})

		It("returns an error if too many args are passed", func() {
			args := []string{"source", "plan-type", "extra-arg"}
			err := me.Execute(client, args)
			Expect(err).To(MatchError("unexpected arguments: extra-arg"))
		})

		It("returns an error if an invalid flag is passed", func() {
			args := []string{"source", "plan-type", "--invalid-flag"}
			err := me.Execute(client, args)
			Expect(err).To(MatchError("unknown flag `invalid-flag'"))
		})

		It("returns doesn't clean up when the --no-cleanup flag is passed", func() {
			args := []string{
				"some-donor", "some-plan", "--no-cleanup",
			}

			err := me.Execute(client, args)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
		})

		Context("when creating a service instance fails", func() {
			BeforeEach(func() {
				fakeMigrator.CreateAndConfigureServiceInstanceReturns(errors.New("some-cf-error"))
			})

			It("returns an error", func() {
				args := []string{"some-donor", "some-plan"}
				err := me.Execute(client, args)
				Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				args := []string{
					"some-donor", "some-plan", "--no-cleanup",
				}

				err := me.Execute(client, args)
				Expect(err).To(MatchError("error creating service instance: some-cf-error. Not cleaning up service some-donor-new"))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})

		Context("when migrating data fails", func() {
			BeforeEach(func() {
				fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
			})

			It("returns an error and attempts to delete the new service instance", func() {
				args := []string{"some-donor", "some-plan"}
				err := me.Execute(client, args)
				Expect(err).To(MatchError(ContainSubstring("error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeTrue())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				args := []string{
					"some-donor", "some-plan", "--no-cleanup",
				}

				err := me.Execute(client, args)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-donor-new"))
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeFalse())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})

		Context("when renaming the service instances fail", func() {
			BeforeEach(func() {
				fakeMigrator.RenameServiceInstancesReturns(errors.New("some-cf-error"))
			})

			It("returns an error and doesn't clean up regardless of --no-cleanup flag", func() {
				donorInstanceName = "some-donor"
				destPlan = "some-plan"
				err := me.Migrate(donorInstanceName, destPlan, cleanup)
				Expect(err).To(MatchError("some-cf-error"))
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				_, _, cleanup := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(cleanup).To(BeTrue())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
		})
	})
})
