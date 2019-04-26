// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package plugin_test

import (
	"bytes"
	"errors"
	"io"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/pluginfakes"
)

var _ = Describe("Plugin Commands", func() {
	var (
		fakeMigrator *pluginfakes.FakeMigrator
		fakeFinder   *pluginfakes.FakeBindingFinder
		logOutput    *bytes.Buffer
	)

	const migrateUsage = `Usage: cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> (destination-service-plan (DEPRECATED, favor -p) | -p <p.mysql-plan-type> | -s <destination-service-instance>)`
	const findUsage = `Usage: cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`

	BeforeEach(func() {
		fakeMigrator = new(pluginfakes.FakeMigrator)
		fakeFinder = new(pluginfakes.FakeBindingFinder)

		logOutput = &bytes.Buffer{}

		w := io.MultiWriter(GinkgoWriter, logOutput)
		log.SetOutput(w)
	})

	Context("Migrate", func() {
		When("a plan name is specified", func() {
			It("migrates data from a source service instance to a newly created instance", func() {
				args := []string{
					"-p", "some-plan", "some-donor",
				}
				Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

				By("logging a message that we don't migrate triggers, routines and events", func() {
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

				Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(BeZero())
			})

			Context("when creating a service instance fails", func() {
				BeforeEach(func() {
					fakeMigrator.CreateAndConfigureServiceInstanceReturns(errors.New("some-cf-error"))
				})

				It("returns an error and attempts to delete the new service instance", func() {
					args := []string{"some-donor", "-p", "some-plan"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(1))
				})

				It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
					args := []string{
						"some-donor", "-p", "some-plan", "--no-cleanup",
					}

					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError("error creating service instance: some-cf-error. Not cleaning up service some-donor-new"))
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})
			})

			Context("when migrating data fails", func() {
				BeforeEach(func() {
					fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
				})

				It("returns an error and attempts to delete the new service instance", func() {
					args := []string{"some-donor", "-p", "some-plan"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError(MatchRegexp("error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.Cleanup).To(BeTrue())
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(1))
				})

				It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
					args := []string{
						"some-donor", "-p", "some-plan", "--no-cleanup",
					}

					err := plugin.Migrate(fakeMigrator, args)

					Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-donor-new"))
					Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.Cleanup).To(BeFalse())
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})
			})

			Context("when renaming the service instances fail", func() {
				BeforeEach(func() {
					fakeMigrator.RenameServiceInstancesReturns(errors.New("some-cf-error"))
				})

				It("returns an error and doesn't clean up regardless of --no-cleanup flag", func() {
					args := []string{"some-donor", "-p", "some-plan"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError("some-cf-error"))
					Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.Cleanup).To(BeTrue())
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})
			})
		})

		When("a service name is specified", func() {
			It("migrates data from a source service instance to an existing service instance", func() {
				args := []string{
					"some-donor", "-s", "some-service",
				}
				Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

				By("logging a message that we don't migrate triggers, routines and events", func() {
					Expect(logOutput.String()).To(ContainSubstring(`Warning: The mysql-tools migrate command will not migrate any triggers, routines or events`))
				})

				By("checking that donor and recipient exist", func() {
					Expect(fakeMigrator.CheckServiceExistsCallCount()).
						To(Equal(2), `Expected to call CheckServiceExists twice (once for the donor and once for the recipient) `)
					Expect(fakeMigrator.CheckServiceExistsArgsForCall(0)).
						To(Equal("some-donor"))
				})

				By("configuring the existing service instance", func() {
					Expect(fakeMigrator.CreateAndConfigureServiceInstanceCallCount()).
						To(BeZero(), `Expected not to have called CreateAndConfigureServiceInstance`)
					Expect(fakeMigrator.ConfigureServiceInstanceCallCount()).
						To(Equal(1), `Expected to have called ConfigureServiceInstance exactly once`)

					serviceInstanceName := fakeMigrator.ConfigureServiceInstanceArgsForCall(0)
					Expect(serviceInstanceName).To(Equal("some-service"))
				})

				By("migrating data from the donor to the recipient", func() {
					Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.DonorInstanceName).To(Equal("some-donor"))
					Expect(opts.RecipientInstanceName).To(Equal("some-service"))
					Expect(opts.Cleanup).To(BeFalse())
					Expect(opts.SkipTLSValidation).To(BeFalse())
				})

				// Never rename or delete existing service instances
				Expect(fakeMigrator.RenameServiceInstancesCallCount()).To(BeZero())
				Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(BeZero())
			})

			Context("when configuring a service instance fails", func() {
				BeforeEach(func() {
					fakeMigrator.ConfigureServiceInstanceReturns(errors.New("some-cf-error"))
				})

				It("returns an error and doesn't clean up the existing service instance", func() {
					args := []string{"some-donor", "-s", "some-service",}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError(MatchRegexp("error configuring service instance: some-cf-error. Not cleaning up service some-service")))
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})

				It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
					args := []string{"some-donor", "-s", "some-service", "--no-cleanup",}

					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(MatchError("error configuring service instance: some-cf-error. Not cleaning up service some-service"))
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})
			})

			Context("when migrating data fails", func() {
				BeforeEach(func() {
					fakeMigrator.MigrateDataReturns(errors.New("some-cf-error"))
				})

				It("returns an error and doesn't to delete the new service instance", func() {
					args := []string{"some-donor", "-s", "some-service"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
					Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-service"))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.Cleanup).To(BeFalse())
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})

				It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
					args := []string{
						"some-donor", "-s", "some-service", "--no-cleanup",
					}

					err := plugin.Migrate(fakeMigrator, args)

					Expect(err).To(MatchError("error migrating data: some-cf-error. Not cleaning up service some-service"))
					Expect(fakeMigrator.MigrateDataCallCount()).To(Equal(1))
					opts := fakeMigrator.MigrateDataArgsForCall(0)
					Expect(opts.Cleanup).To(BeFalse())
					Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(Equal(0))
				})
			})

			Context("when recipient doesn't exist", func() {
				BeforeEach(func() {
					fakeMigrator.CheckServiceExistsReturnsOnCall(1, errors.New("recipient-does-not-exist"))
				})
				It("returns an error", func() {
					args := []string{"some-donor", "-s", "some-service"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(fakeMigrator.CheckServiceExistsCallCount()).To(Equal(2), `Expected CheckServiceExists to be called exactly twice`)
					Expect(err).To(MatchError("recipient-does-not-exist"))
				})
			})
		})

		When("a positional plan name is specified", func(){
			It("migrates data from a source service instance to a newly created instance", func() {
				args := []string{"some-donor", "some-plan"}

				Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

				By("logging a message that we don't migrate triggers, routines and events", func() {
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

				Expect(fakeMigrator.DeleteServiceInstanceOnErrorCallCount()).To(BeZero())
			})

			Context("and plan flag is also specified", func(){
				It("returns an error", func(){
					args := []string{"some-donor", "some-plan", "-p", "some-plan"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(migrateUsage + "\n\nYou must specify only one plan name"))
				})
			})

			Context("and a service flag is also specified", func(){
				It("returns an error", func() {
					args := []string{"some-donor", "some-plan", "-s", "some-service"}
					err := plugin.Migrate(fakeMigrator, args)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(migrateUsage + "\n\nYou must specify either the plan name OR the service name"))
				})
			})
		})

		When("both service name and plan name are specified", func() {
			It("returns an error", func() {
				args := []string{
					"some-donor", "-p", "some-plan", "-s", "some-service",
				}
				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(migrateUsage + "\n\nYou must specify either the plan name OR the service name"))
			})
		})

		When("neither service name nor plan name are specified", func() {
			It("returns an error", func() {
				args := []string{
					"some-donor",
				}
				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(migrateUsage + "\n\nYou must specify either the plan name OR the service name"))
			})
		})

		When("skip-tls-validation is specified", func() {
			It("Requests that the data be migrated insecurely", func() {
				args := []string{
					"--skip-tls-validation",
					"some-donor", "-p", "some-plan",
				}
				Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

				opts := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(opts.SkipTLSValidation).To(
					BeTrue(),
					`Expected MigrateOptions to have SkipTLSValidation set to true, but it was false`)
			})
		})

		It("returns an error if the donor service instance does not exist", func() {
			fakeMigrator.CheckServiceExistsReturns(errors.New("some-donor does not exist"))

			args := []string{"some-donor", "-p", "some-plan"}

			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("some-donor does not exist"))
		})

		It("returns an error if not enough args are passed", func() {
			args := []string{"just-a-source"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nYou must specify either the plan name OR the service name"))
		})

		It("returns an error if too many args are passed", func() {
			args := []string{"source", "plan-type", "extra-arg"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(migrateUsage + "\n\nunexpected arguments: extra-arg"))
		})

		It("returns an error if an invalid flag is passed", func() {
			args := []string{"source", "plan-type", "--invalid-flag"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nunknown flag `invalid-flag'"))
		})
	})

	Context("FindBindings", func() {
		It("returns an error if not enough args are passed", func() {
			args := []string{}
			err := plugin.FindBindings(fakeFinder, args)
			Expect(err).To(MatchError(findUsage + "\n\nthe required argument `<mysql-v1-service-name>` was not provided"))
		})

		It("returns an error if too many args are passed", func() {
			args := []string{"p.mysql", "somethingelse"}
			err := plugin.FindBindings(fakeFinder, args)
			Expect(err).To(MatchError(findUsage + "\n\nunexpected arguments: somethingelse"))
		})

		It("returns an error if an invalid flag is passed", func() {
			args := []string{"p.mysql", "--invalid-flag"}
			err := plugin.FindBindings(fakeFinder, args)
			Expect(err).To(MatchError(findUsage + "\n\nunknown flag `invalid-flag'"))
		})

		Context("When find binding runs successfully", func() {
			It("succeeds", func() {
				args := []string{"p.mysql"}
				err := plugin.FindBindings(fakeFinder, args)
				Expect(err).To(Not(HaveOccurred()))
				Expect(fakeFinder.FindBindingsCallCount()).To(Equal(1))
				Expect(fakeFinder.FindBindingsArgsForCall(0)).To(Equal("p.mysql"))
			})
		})

		Context("When find binding returns an error", func() {
			It("fails", func() {
				args := []string{"p.mysql"}
				fakeFinder.FindBindingsReturns([]find_bindings.Binding{}, errors.New("some-error"))
				err := plugin.FindBindings(fakeFinder, args)
				Expect(err).To(MatchError("some-error"))
				Expect(fakeFinder.FindBindingsCallCount()).To(Equal(1))
				Expect(fakeFinder.FindBindingsArgsForCall(0)).To(Equal("p.mysql"))
			})
		})
	})
})
