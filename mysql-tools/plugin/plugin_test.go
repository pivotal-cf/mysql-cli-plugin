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

	const migrateUsage = `Usage: cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> -p <p.mysql-plan-type>`
	const findUsage = `Usage: cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`

	BeforeEach(func() {
		fakeMigrator = new(pluginfakes.FakeMigrator)
		fakeFinder = new(pluginfakes.FakeBindingFinder)

		logOutput = &bytes.Buffer{}

		w := io.MultiWriter(GinkgoWriter, logOutput)
		log.SetOutput(w)
	})

	Context("Migrate", func() {
		It("migrates data from a source service instance to a newly created instance", func() {
			args := []string{
				"some-donor", "-p", "some-plan",
			}
			Expect(plugin.Migrate(fakeMigrator, args)).To(Succeed())

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
			Expect(err).To(MatchError("some-donor does not exist"))
		})

		It("returns an error if not enough args are passed", func() {
			args := []string{"just-a-source"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nthe required flag `-p, --plan' was not specified"))
		})

		It("returns an error if too many args are passed", func() {
			args := []string{"source", "-p", "plan-type", "extra-arg"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nunexpected arguments: extra-arg"))
		})

		It("returns an error if an invalid flag is passed", func() {
			args := []string{"source", "plan-type", "--invalid-flag"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nunknown flag `invalid-flag'"))
		})

		Context("when creating a service instance fails", func() {
			BeforeEach(func() {
				fakeMigrator.CreateServiceInstanceReturns(errors.New("some-cf-error"))
			})

			It("returns an error and attempts to delete the new service instance", func() {
				args := []string{"some-donor", "-p", "some-plan"}
				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				args := []string{
					"some-donor", "-p", "some-plan", "--no-cleanup",
				}

				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(MatchError("error creating service instance: some-cf-error. Not cleaning up service some-donor-new"))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
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
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
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
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
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
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(0))
			})
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
