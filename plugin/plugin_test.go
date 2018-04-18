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
