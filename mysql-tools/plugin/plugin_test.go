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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/pluginfakes"
)

var _ = Describe("Plugin Commands", func() {
	var (
		fakeMigrator  *pluginfakes.FakeMigrator
		fakeFinder    *pluginfakes.FakeBindingFinder
		fakeMultiSite *pluginfakes.FakeMultiSite
		logOutput     *bytes.Buffer
	)

	const migrateUsage = `Usage: cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>`
	const findUsage = `Usage: cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`
	const saveTargetUsage = `Usage: cf mysql-tools save-target <target-name>`
	const removeTargetUsage = `Usage: cf mysql-tools remove-target <target-name>`
	const setupReplicationUsage = `Usage: cf mysql-tools find-bindings <primary-target-name> <secondary-target-name>`

	BeforeEach(func() {
		fakeMigrator = new(pluginfakes.FakeMigrator)
		fakeFinder = new(pluginfakes.FakeBindingFinder)
		fakeMultiSite = new(pluginfakes.FakeMultiSite)
		logOutput = &bytes.Buffer{}
		w := io.MultiWriter(GinkgoWriter, logOutput)
		log.SetOutput(w)
	})

	Context("Migrate", func() {
		It("migrates data from a source service instance to a newly created instance", func() {
			args := []string{
				"some-donor", "some-plan",
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
					"some-donor", "some-plan",
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

			args := []string{"some-donor", "some-plan"}

			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError("some-donor does not exist"))
		})

		It("returns an error if not enough args are passed", func() {
			args := []string{"just-a-source"}
			err := plugin.Migrate(fakeMigrator, args)
			Expect(err).To(MatchError(migrateUsage + "\n\nthe required argument `<p.mysql-plan-type>` was not provided"))
		})

		It("returns an error if too many args are passed", func() {
			args := []string{"source", "plan-type", "extra-arg"}
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
				args := []string{"some-donor", "some-plan"}
				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(MatchError(MatchRegexp("error creating service instance: some-cf-error. Attempting to clean up service some-donor-new")))
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				args := []string{
					"some-donor", "some-plan", "--no-cleanup",
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
				args := []string{"some-donor", "some-plan"}
				err := plugin.Migrate(fakeMigrator, args)
				Expect(err).To(MatchError(MatchRegexp("error migrating data: some-cf-error. Attempting to clean up service some-donor-new")))
				opts := fakeMigrator.MigrateDataArgsForCall(0)
				Expect(opts.Cleanup).To(BeTrue())
				Expect(fakeMigrator.CleanupOnErrorCallCount()).To(Equal(1))
			})

			It("returns an error and doesn't clean up when the --no-cleanup flag is passed", func() {
				args := []string{
					"some-donor", "some-plan", "--no-cleanup",
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
				args := []string{"some-donor", "some-plan"}
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

	Context("Multi Foundation Replication Setup", func() {
		const savedOrg = "SavedOrgName"
		const savedSpace = "SavedSpaceName"
		const savedEndpoint = "SavedAPIEndpoint"
		const savedTarget = "SavedTarget"
		returnedConfigSummary := multisite.ConfigCoreSubset{
			OrganizationFields: struct {
				Name string
			}{
				Name: savedOrg,
			},
			SpaceFields: struct {
				Name string
			}{
				Name: savedSpace,
			},
			Target: savedEndpoint,
			Name:   savedTarget,
		}
		Context("List Targets", func() {
			It("Prints a summary when successful", func() {
				fakeMultiSite.ListConfigsReturns([]*multisite.ConfigCoreSubset{&returnedConfigSummary}, nil)

				r, w, _ := os.Pipe()
				tmp := os.Stdout
				defer func() {
					os.Stdout = tmp
				}()

				os.Stdout = w
				err := plugin.ListTargets(fakeMultiSite)
				w.Close()
				Expect(err).NotTo(HaveOccurred())

				By("showing a summary of the saved config")
				stdout, _ := io.ReadAll(r)
				Expect(string(stdout)).To(ContainSubstring("Targets:"))
				Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
				Expect(string(stdout)).To(ContainSubstring(savedOrg))
				Expect(string(stdout)).To(ContainSubstring(savedOrg))
				Expect(string(stdout)).To(ContainSubstring(savedTarget))
				Expect(fakeMultiSite.ListConfigsCallCount()).To(Equal(1))
			})
			It("Prints a summary when there is an error", func() {
				fakeMultiSite.ListConfigsReturns([]*multisite.ConfigCoreSubset{&returnedConfigSummary}, errors.New("some-error"))

				r, w, _ := os.Pipe()
				tmp := os.Stdout
				defer func() {
					os.Stdout = tmp
				}()

				os.Stdout = w
				err := plugin.ListTargets(fakeMultiSite)
				w.Close()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("error listing multisite targets: some-error"))

				By("showing a summary of the saved config")
				stdout, _ := io.ReadAll(r)
				Expect(string(stdout)).To(ContainSubstring("Targets:"))
				Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
				Expect(string(stdout)).To(ContainSubstring(savedOrg))
				Expect(string(stdout)).To(ContainSubstring(savedOrg))
				Expect(string(stdout)).To(ContainSubstring(savedTarget))
				Expect(fakeMultiSite.ListConfigsCallCount()).To(Equal(1))
			})
			It("prints nothing when there are no configs", func() {
				fakeMultiSite.ListConfigsReturns([]*multisite.ConfigCoreSubset{}, nil)

				r, w, _ := os.Pipe()
				tmp := os.Stdout
				defer func() {
					os.Stdout = tmp
				}()

				os.Stdout = w
				err := plugin.ListTargets(fakeMultiSite)
				w.Close()
				Expect(err).NotTo(HaveOccurred())

				By("showing a summary of the saved config")
				stdout, _ := io.ReadAll(r)
				Expect(string(stdout)).To(ContainSubstring("Targets:"))
				Expect(fakeMultiSite.ListConfigsCallCount()).To(Equal(1))

			})
		})

		Context("Save Target", func() {
			BeforeEach(func() {
				fakeMultiSite.SaveConfigReturns(&returnedConfigSummary, nil)
			})

			When("CF_HOME is set to a directory", func() {
				var (
					testCFHome, originalCFHome string
					hadCFHome                  bool
					err                        error
				)
				BeforeEach(func() {
					testCFHome, err = os.MkdirTemp("", "test_CFHOME_")
					if err != nil {
						panic("Failed to create temp CF_HOME: " + err.Error())
					}
					err = os.Mkdir(filepath.Join(testCFHome, ".cf"), 0750)
					if err != nil {
						panic("Failed to create temp CF_HOME/.cf/: " + err.Error())
					}
					originalCFHome, hadCFHome = os.LookupEnv("CF_HOME")
					err = os.Setenv("CF_HOME", testCFHome)
					if err != nil {
						panic("Failed to set new test CF_HOME environment variable: " + err.Error())
					}
				})

				AfterEach(func() {
					if hadCFHome {
						err = os.Setenv("CF_HOME", originalCFHome)
						if err != nil {
							panic("Failed to restore original CF_HOME environment variable: " + err.Error())
						}
					}
					_ = os.RemoveAll(testCFHome)
				})
				When("CF_HOME contains a config file", func() {
					BeforeEach(func() {
						os.OpenFile(filepath.Join(testCFHome, ".cf", "config.json"), os.O_RDONLY|os.O_CREATE, 0666)
					})

					It("saves a target without error", func() {
						var err error
						args := []string{"targetName"}

						r, w, _ := os.Pipe()
						tmp := os.Stdout
						defer func() {
							os.Stdout = tmp
						}()

						os.Stdout = w
						err = plugin.SaveTarget(fakeMultiSite, args)
						w.Close()
						Expect(err).To(Not(HaveOccurred()))

						By("constructing the path to the saved cf config file")
						inputConfigFile, target := fakeMultiSite.SaveConfigArgsForCall(0)
						Expect(inputConfigFile).To(Equal(filepath.Join(testCFHome, ".cf", "config.json")))
						Expect(target).To(Equal("targetName"))

						By("showing a summary of the saved config")
						stdout, _ := io.ReadAll(r)
						Expect(string(stdout)).To(ContainSubstring("Success"))
						Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
						Expect(string(stdout)).To(ContainSubstring(savedOrg))
						Expect(string(stdout)).To(ContainSubstring(savedSpace))
					})
					It("surfaces any underlying multisite errors", func() {
						fakeMultiSite.SaveConfigReturns(nil, errors.New("low-level save error"))
						args := []string{"targetName"}

						err := plugin.SaveTarget(fakeMultiSite, args)
						Expect(err).To(MatchError("error saving target targetName: low-level save error"))
					})
				})

			})
			When("CF_HOME is set to an invalid directory with no config", func() {
				var (
					testCFHome, originalCFHome string
					hadCFHome                  bool
					err                        error
				)
				BeforeEach(func() {
					testCFHome = os.TempDir()
					originalCFHome, hadCFHome = os.LookupEnv("CF_HOME")
					err = os.Setenv("CF_HOME", testCFHome)
					if err != nil {
						panic("Failed to set new test CF_HOME environment variable: " + err.Error())
					}
				})
				AfterEach(func() {
					if hadCFHome {
						err = os.Setenv("CF_HOME", originalCFHome)
						if err != nil {
							panic("Failed to restore original CF_HOME environment variable: " + err.Error())
						}
					}
				})
				It("errors without ever calling Multisite", func() {
					args := []string{"targetName"}

					err := plugin.SaveTarget(fakeMultiSite, args)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
					Expect(fakeMultiSite.SaveConfigCallCount()).To(Equal(0))
				})
			})
			// When("no CF_HOME is set"), plugin uses default ${HOME}/.cf

			It("errors if not enough args are passed", func() {
				args := []string{}
				err := plugin.SaveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(saveTargetUsage + "\n\nthe required argument `<target-name>` was not provided"))
			})

			It("errors if too many args are passed", func() {
				args := []string{"targetName", "extra-arg"}
				err := plugin.SaveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(saveTargetUsage + "\n\nunexpected arguments: extra-arg"))
			})

			It("errors if an invalid flag is passed", func() {
				args := []string{"targetName", "--invalid-flag"}
				err := plugin.SaveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(saveTargetUsage + "\n\nunknown flag `invalid-flag'"))
			})
		})

		Context("Remove Target", func() {
			It("able to remove target config without an error", func() {
				args := []string{"targetName"}
				err := plugin.RemoveTarget(fakeMultiSite, args)
				Expect(err).To(Not(HaveOccurred()))
				Expect(fakeMultiSite.RemoveConfigCallCount()).To(Equal(1))

				target := fakeMultiSite.RemoveConfigArgsForCall(0)
				Expect(target).To(Equal("targetName"))
			})

			It("errors if the remove target fails", func() {
				fakeMultiSite.RemoveConfigReturns(errors.New("some-error"))
				args := []string{"targetName"}
				err := plugin.RemoveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError("error trying to remove the target config: some-error"))
			})

			It("errors if not enough args are passed", func() {
				args := []string{}
				err := plugin.RemoveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(removeTargetUsage + "\n\nthe required argument `<target-name>` was not provided"))
			})

			It("errors if too many args are passed", func() {
				args := []string{"targetName", "extra-arg"}
				err := plugin.RemoveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(removeTargetUsage + "\n\nunexpected arguments: extra-arg"))
			})

			It("errors if an invalid flag is passed", func() {
				args := []string{"targetName", "--invalid-flag"}
				err := plugin.RemoveTarget(fakeMultiSite, args)
				Expect(err).To(MatchError(removeTargetUsage + "\n\nunknown flag `invalid-flag'"))
			})
		})

		Context("Setup Replication", func() {

		})
	})

})
