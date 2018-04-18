package plugin_test

import (
	"errors"

	"github.com/gobuffalo/packr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin/pluginfakes"
	"github.com/pivotal-cf/mysql-cli-plugin/unpack"
)

var _ = Describe("Plugin Commands", func() {
	var (
		fakeClient *pluginfakes.FakeCFClient
		unpacker   *unpack.Unpacker
	)

	BeforeEach(func() {
		fakeClient = new(pluginfakes.FakeCFClient)
		unpacker = unpack.NewUnpacker()
		unpacker.Box = packr.NewBox("../unpack/fixtures")
	})

	Context("Migrate", func() {
		BeforeEach(func() {
			fakeClient.ServiceExistsReturns(true)
		})

		It("migrates data from a source service instance to a newly created instance", func() {
			args := []string{
				"mysql-tools", "migrate", "some-donor", "--create", "some-plan",
			}
			Expect(plugin.Migrate(fakeClient, unpacker, args)).To(Succeed())
		})

		It("returns an error if an incorrect number of args are passed", func() {
			args := []string{"mysql-tools", "migrate", "just-a-source"}
			err := plugin.Migrate(fakeClient, unpacker, args)
			Expect(err).To(MatchError("Usage: cf mysql-tools migrate <v1-service-instance> --create <v2-plan>"))
		})

		It("returns an error if creating a service instance fails", func() {
			fakeClient.CreateServiceInstanceReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "migrate", "some-donor", "--create", "some-plan"}
			err := plugin.Migrate(fakeClient, unpacker, args)
			Expect(err).To(MatchError("some-cf-error"))
		})

		It("returns an error and attempts to delete the new service instance if migrating data fails", func() {
			fakeClient.PushAppReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "migrate", "some-donor", "--create", "some-plan"}
			err := plugin.Migrate(fakeClient, unpacker, args)
			Expect(err).To(MatchError(MatchRegexp("Error migrating data: failed to push application: some-cf-error. Attempting to clean up app .* and service some-donor-new")))
			Expect(fakeClient.DeleteAppCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteServiceInstanceCallCount()).To(Equal(1))
		})
	})

	Context("Replace", func() {
		BeforeEach(func() {
			fakeClient.ServiceExistsReturns(true)
		})

		It("migrates data from an existing source and destination service and renames the destination to source", func() {
			args := []string{
				"mysql-tools", "replace", "some-donor", "some-recipient",
			}
			Expect(plugin.Replace(fakeClient, unpacker, args)).To(Succeed())
		})

		It("returns an error if an incorrect number of args are passed", func() {
			args := []string{"mysql-tools", "replace", "source", "dest", "extra-dest-not-allowed"}
			err := plugin.Replace(fakeClient, unpacker, args)
			Expect(err).To(MatchError("Usage: cf mysql-tools replace <v1-service-instance> <v2-service-instance>"))
		})

		It("returns an error if migrating data fails", func() {
			fakeClient.PushAppReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "replace", "some-donor", "some-recipient"}
			err := plugin.Replace(fakeClient, unpacker, args)
			Expect(err).To(MatchError("failed to push application: some-cf-error"))
		})

		It("returns an error if renaming instances fails", func() {
			fakeClient.RenameServiceReturns(errors.New("some-cf-error"))
			args := []string{"mysql-tools", "replace", "some-donor", "some-recipient"}
			err := plugin.Replace(fakeClient, unpacker, args)
			Expect(err).To(MatchError("Error renaming service instance some-donor: some-cf-error"))
		})

	})
})
