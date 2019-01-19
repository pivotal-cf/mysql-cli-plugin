package command_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command/commandfakes"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin_errors"
	"log"
)

var _ = Describe("Plugin Command Router", func() {
	var (
		//fakeMigrator *commandfakes.FakeMigrator
		logOutput *bytes.Buffer
		mcmdr     *command.MySQCmdLRouter
		//migratorExecutor command.MigratorExecutor
	)
	const usage = `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools version`
	const migrateUsage = `cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>`

	BeforeEach(func() {
		mcmdr = &command.MySQCmdLRouter{}
		logOutput = &bytes.Buffer{}
		log.SetOutput(logOutput)
	})

	Describe("Match", func() {
		Context("when no commands are passed", func() {
			It("displays usage instructions", func() {
				var command string
				args := []string{}

				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				err := mcmdr.Match(command, fakeCliConnection, args)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(plugin_errors.NewUsageError("unknown command: \"\"")))
			})
		})

		Context("version", func() {
			var (
				logOutput   *bytes.Buffer
				versionFunc *commandfakes.FakeCommandRunner
			)
			BeforeEach(func() {
				logOutput = &bytes.Buffer{}
				log.SetOutput(logOutput)

				versionFunc = &commandfakes.FakeCommandRunner{}
				mcmdr.Routes = make(map[string]interface{ command.CommandRunner })
				mcmdr.Routes["version"] = versionFunc
			})

			It("outputs the version to the user", func() {
				command := "version"
				args := []string{}

				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				err := mcmdr.Match(command, fakeCliConnection, args)
				Expect(err).NotTo(HaveOccurred())
				Expect(versionFunc.ExecuteCallCount()).To(Equal(1))
			})
		})

		Context("unknown command", func() {
			It("adds the usage to the Err method", func() {
				command := "foo"
				args := []string{}

				fakeCliConnection := &pluginfakes.FakeCliConnection{}
				err := mcmdr.Match(command, fakeCliConnection, args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown command: \"foo\""))
				Expect(err.Error()).To(ContainSubstring(usage))
			})
		})

		Context("migrate", func() {
			var migrateFunc *commandfakes.FakeCommandRunner

			BeforeEach(func() {
				migrateFunc = &commandfakes.FakeCommandRunner{}
				mcmdr.Routes = make(map[string]interface{ command.CommandRunner })
				mcmdr.Routes["migrate"] = migrateFunc
			})
			Context("calls the migrate command with the wrong args", func() {
				BeforeEach(func() {
					migrateFunc.ExecuteReturns(plugin_errors.NewCustomUsageError("", migrateUsage))
				})
				It("shows a custom usage", func() {
					command := "migrate"
					args := []string{
						"arg1",
					}

					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					err := mcmdr.Match(command, fakeCliConnection, args)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(migrateUsage))
				})
			})
			Context("calls the migrate command with the proper args", func() {
				It("runs the command without errors", func() {
					command := "migrate"
					args := []string{
						"arg1", "args2",
					}

					fakeCliConnection := &pluginfakes.FakeCliConnection{}
					err := mcmdr.Match(command, fakeCliConnection, args)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
