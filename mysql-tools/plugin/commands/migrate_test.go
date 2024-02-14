package commands_test

import (
	"bytes"
	"errors"
	"io"
	"log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("Migrate", func() {
	var (
		fakeMigrator *fakes.FakeMigrator
		logOutput    *bytes.Buffer
	)

	const (
		migrateUsage = `cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>`
	)

	BeforeEach(func() {
		fakeMigrator = new(fakes.FakeMigrator)
		logOutput = &bytes.Buffer{}
		w := io.MultiWriter(GinkgoWriter, logOutput)
		log.SetOutput(w)
	})

	It("migrates data from a source service instance to a newly created instance", func() {
		args := []string{
			"some-donor", "some-plan",
		}
		Expect(commands.Migrate(args, fakeMigrator)).To(Succeed())

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
			Expect(fakeMigrator.CreateServiceInstanceCallCount()).
				To(Equal(1))

			createdServicePlan, createdServiceInstanceName := fakeMigrator.CreateServiceInstanceArgsForCall(0)
			Expect(createdServicePlan).To(Equal("some-plan"))
			Expect(createdServiceInstanceName).
				To(Equal("some-donor-new"))
		})

		By("migrating data from the donor to the recipient", func() {
			Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
			opts := fakeMigrator.MigrateDataArgsForCall(0)
			Expect(opts.DonorInstanceName).To(Equal("some-donor"))
			Expect(opts.RecipientInstanceName).To(Equal("some-donor-new"))
			Expect(opts.Cleanup).To(BeTrue())
			Expect(opts.SkipTLSValidation).To(BeFalse())
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

	Context("when skip-tls-validation is specified", func() {
		It("Requests that the data be migrated insecurely", func() {
			args := []string{
				"--skip-tls-validation",
				"some-donor", "some-plan",
			}
			Expect(commands.Migrate(args, fakeMigrator)).To(Succeed())

			opts := fakeMigrator.MigrateDataArgsForCall(0)
			Expect(opts.SkipTLSValidation).To(
				BeTrue(),
				`Expected MigrateOptions to have SkipTLSValidation set to true, but it was false`)
		})
	})

	It("returns an error if the donor service instance does not exist", func() {
		fakeMigrator.CheckServiceExistsReturns(errors.New("some-donor does not exist"))

		args := []string{"some-donor", "some-plan"}

		err := commands.Migrate(args, fakeMigrator)
		Expect(err).To(MatchError("some-donor does not exist"))
	})

	It("returns an error if not enough args are passed", func() {
		args := []string{"just-a-source"}
		err := commands.Migrate(args, fakeMigrator)
		Expect(err).To(MatchError("Usage: " + migrateUsage + "\n\nthe required argument `<p.mysql-plan-type>` was not provided"))
	})

	It("returns an error if too many args are passed", func() {
		args := []string{"source", "plan-type", "extra-arg"}
		err := commands.Migrate(args, fakeMigrator)
		Expect(err).To(MatchError("Usage: " + migrateUsage + "\n\nunexpected arguments: extra-arg"))
	})

	It("returns an error if an invalid flag is passed", func() {
		args := []string{"source", "plan-type", "--invalid-flag"}
		err := commands.Migrate(args, fakeMigrator)
		Expect(err).To(MatchError("Usage: " + migrateUsage + "\n\nunknown flag `invalid-flag'"))
	})

	Context("when creating a service instance fails", func() {
		BeforeEach(func() {
			fakeMigrator.CreateServiceInstanceReturns(errors.New("some-cf-error"))
		})

		It("returns an error and attempts to delete the new service instance", func() {
			args := []string{"some-donor", "some-plan"}
			err := commands.Migrate(args, fakeMigrator)
			Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
		})

		It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
			args := []string{
				"some-donor", "some-plan", "--no-cleanup",
			}

			err := commands.Migrate(args, fakeMigrator)
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
			err := commands.Migrate(args, fakeMigrator)
			Expect(err).To(MatchError(MatchRegexp("error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
			opts := fakeMigrator.MigrateDataArgsForCall(0)
			Expect(opts.Cleanup).To(BeTrue())
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
		})

		It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
			args := []string{
				"some-donor", "some-plan", "--no-cleanup",
			}

			err := commands.Migrate(args, fakeMigrator)

			Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-donor-new"))
			Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
			opts := fakeMigrator.MigrateDataArgsForCall(0)
			Expect(opts.Cleanup).To(BeFalse())
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
		})
	})

	Context("when renaming the service instances fail", func() {
		BeforeEach(func() {
			fakeMigrator.RenameServiceInstancesReturns(errors.New("some-cf-error"))
		})

		It("returns an error and doesn't clean up regardless of --no-cleanup flag", func() {
			args := []string{"some-donor", "some-plan"}
			err := commands.Migrate(args, fakeMigrator)
			Expect(err).To(MatchError("some-cf-error"))
			Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
			opts := fakeMigrator.MigrateDataArgsForCall(0)
			Expect(opts.Cleanup).To(BeTrue())
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
		})
	})

})
