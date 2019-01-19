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

package command_test

import (
	"bytes"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command"
	"log"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command/commandfakes"
)

// TODO: What goes here?
var _ = Describe("Plugin Commands", func() {

	const usage = `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools version`

	Describe("Run", func() {
		var (
			cmdRouter       *command.MySQCmdLRouter
			fakeVersionFunc *commandfakes.FakeCommandRunner
			fakeMigrateFunc *commandfakes.FakeCommandRunner
			logOutput       *bytes.Buffer
			mysqlPlugin     *command.MySQLPlugin
		)

		BeforeEach(func() {
			routes := make(map[string]interface{ command.CommandRunner })
			fakeVersionFunc = &commandfakes.FakeCommandRunner{}
			fakeMigrateFunc = &commandfakes.FakeCommandRunner{}
			routes["version"] = fakeVersionFunc
			routes["migrate"] = fakeMigrateFunc
			cmdRouter = &command.MySQCmdLRouter{
				Routes: routes,
			}

		})
		Context("usage", func() {

			Context("when no commands are passed", func() {
				It("displays usage instructions", func() {
					args := []string{
						"mysql-tools",
					}

					mysqlPlugin := new(command.MySQLPlugin)
					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					mysqlPlugin.Run(fakeCliConnection, args)
					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
				})
			})

			Context("when -h is passed", func() {
				It("adds the usage to the Err method", func() {
					args := []string{
						"mysql-tools", "-h",
					}

					mysqlPlugin := new(command.MySQLPlugin)
					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					mysqlPlugin.Run(fakeCliConnection, args)
					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
				})
				Context("when the -h is in the middle of arguments", func() {

					It("adds the usage to the Err method", func() {
						args := []string{
							"mysql-tools", "foo", "-h", "bar",
						}

						mysqlPlugin := new(command.MySQLPlugin)
						fakeCliConnection := &pluginfakes.FakeCliConnection{}
						mysqlPlugin.Run(fakeCliConnection, args)
						Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
					})
				})
			})

			Context("when the plugin does not pass us the name of the command", func() {
				It("adds an error which can be accessed from Err()", func() {
					args := []string{}

					mysqlPlugin := new(command.MySQLPlugin)
					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					mysqlPlugin.Run(fakeCliConnection, args)

					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring("Error: plugin did not receive the expected input from the CLI"))
				})
			})

			Context("when the plugin  passes  the 'CLI-MESSAGE-UNINSTALL'", func() {

				It("does not err or print an error message when uninstalling", func() {
					args := []string{
						"CLI-MESSAGE-UNINSTALL",
					}

					mysqlPlugin := new(command.MySQLPlugin)
					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					mysqlPlugin.Run(fakeCliConnection, args)

					Expect(mysqlPlugin.Err()).To(BeNil())
				})
			})
		})

		Context("version", func() {

			BeforeEach(func() {
				logOutput = &bytes.Buffer{}
				log.SetOutput(logOutput)
				mysqlPlugin = &command.MySQLPlugin{
					CommandRouter: cmdRouter,
				}
			})

			It("outputs the version to the user", func() {
				args := []string{
					"mysql-tools", "version",
				}

				//mysqlPlugin := new(command.MySQLPlugin)
				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				mysqlPlugin.Run(fakeCliConnection, args)
				Expect(fakeVersionFunc.ExecuteCallCount()).To(Equal(1))
			})
		})

		Context("unknown command", func() {
			BeforeEach(func() {
				logOutput = &bytes.Buffer{}
				log.SetOutput(logOutput)
				mysqlPlugin = &command.MySQLPlugin{
					CommandRouter: cmdRouter,
				}
			})

			It("adds the usage to the Err method", func() {
				args := []string{
					"mysql-tools", "foo",
				}

				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				mysqlPlugin.Run(fakeCliConnection, args)
				Expect(mysqlPlugin.Err().Error()).To(ContainSubstring("unknown command: \"foo\""))
				Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
			})
		})

		Context("migrate", func() {
			BeforeEach(func() {
				logOutput = &bytes.Buffer{}
				log.SetOutput(logOutput)
				mysqlPlugin = &command.MySQLPlugin{
					CommandRouter: cmdRouter,
				}
			})

			It("calls the migrate command", func() {
				//migrateCmd := &commandfakes.FakeCommand{}
				args := []string{
					"mysql-tools", "migrate", "arg1", "arg2",
				}

				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				mysqlPlugin.Run(fakeCliConnection, args)
				Expect(fakeMigrateFunc.ExecuteCallCount()).To(Equal(1))

				//Expect(migrateCmd.RunCallCount()).To(Equal(1))
				//Expect(migrateCmd.RunArgsForCall(0)).To(Equal([]string{"arg1", "arg2"}))
			})
		})
	})
})
