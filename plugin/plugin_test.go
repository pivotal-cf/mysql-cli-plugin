package plugin_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin/pluginfakes"
)

var _ = Describe("Plugin Commands", func() {
	var (
		fakeMigrator *pluginfakes.FakeMigrator
	)

	BeforeEach(func() {
		fakeMigrator = new(pluginfakes.FakeMigrator)
	})

	Context("Migrate", func() {
		BeforeEach(func() {
		})

		It("migrates data from a source service instance to a newly created instance", func() {
			args := []string{
				"mysql-tools", "migrate", "some-donor", "--create", "some-plan",
			}
			Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

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
				migratedDonorName, migratedRecipientname := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(migratedDonorName).To(Equal("some-donor"))
				Expect(migratedRecipientname).To(Equal("some-donor-new"))
			})

			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(BeZero())
		})

		It("returns an error if the donor service instance does not exist", func() {
			fakeMigrator.CheckServiceExistsReturns(errors.New("some-donor does not exist"))

			args := []string{"mysql-tools", "migrate", "some-donor", "--create", "some-plan"}

			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError("some-donor does not exist"))
		})

		It("returns an error if an incorrect number of args are passed", func() {
			args := []string{"mysql-tools", "migrate", "just-a-source"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError("Usage: cf mysql-tools migrate <v1-service-instance> --create <v2-plan>"))
		})

		It("returns an error if creating a service instance fails", func() {
			fakeMigrator.CreateAndConfigureServiceInstanceReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "migrate", "some-donor", "--create", "some-plan"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError("some-cf-error"))
		})

		It("returns an error and attempts to delete the new service instance if migrating data fails", func() {
			fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "migrate", "some-donor", "--create", "some-plan"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(MatchRegexp("Error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
			Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
		})
	})

	Context("Replace", func() {
		It("migrates data from an existing source and destination service and renames the destination to source", func() {
			args := []string{
				"mysql-tools", "replace", "some-donor", "some-recipient",
			}
			Expect(plugin.Replace(fakeMigrator, args)).To(Succeed())

			By("checking that donor and recipient instances exists", func() {
				Expect(fakeMigrator.CheckServiceExistsCallCount()).
					To(Equal(2))
				Expect(fakeMigrator.CheckServiceExistsArgsForCall(0)).
					To(Equal("some-donor"))
				Expect(fakeMigrator.CheckServiceExistsArgsForCall(1)).
					To(Equal("some-recipient"))
			})

			By("migrating data from the donor to the recipient", func() {
				Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
				migratedDonorName, migratedRecipientname := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(migratedDonorName).To(Equal("some-donor"))
				Expect(migratedRecipientname).To(Equal("some-recipient"))
			})

			By("renaming the recipient instance to the donor instance", func() {
				Expect(fakeMigrator.RenameServiceInstancesCallCount()).
					To(Equal(1))
				renamedDonorInstance, renamedRecipientInstance := fakeMigrator.RenameServiceInstancesArgsForCall(0)
				Expect(renamedDonorInstance).To(Equal("some-donor"))
				Expect(renamedRecipientInstance).To(Equal("some-recipient"))
			})
		})

		It("returns an error if an incorrect number of args are passed", func() {
			args := []string{"mysql-tools", "replace", "source", "dest", "extra-dest-not-allowed"}
			err := plugin.Replace(fakeMigrator, args)
			Expect(err).To(MatchError("Usage: cf mysql-tools replace <v1-service-instance> <v2-service-instance>"))
		})

		It("returns an error if migrating data fails", func() {
			fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "replace", "some-donor", "some-recipient"}
			err := plugin.Replace(fakeMigrator, args)
			Expect(err).To(MatchError("some-cf-error"))
		})

		It("returns an error if renaming instances fails", func() {
			fakeMigrator.RenameServiceInstancesReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "replace", "some-donor", "some-recipient"}
			err := plugin.Replace(fakeMigrator, args)
			Expect(err).To(MatchError("some-cf-error"))
		})
	})
})
