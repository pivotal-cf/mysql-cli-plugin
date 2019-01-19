package plugin_errors_test

import (
	. "github.com/onsi/ginkgo"
)

// TODO: is this needed?	

var _ = Describe("Plugin Errors", func() {
	//
	//	const usage = `NAME:
	//   mysql-tools - Plugin to migrate mysql instances
	//
	//USAGE:
	//   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
	//   cf mysql-tools version`
	//
	//	Describe("Run", func() {
	//
	//		Context("usage", func() {
	//
	//			Context("when no commands are passed", func() {
	//				It("displays usage instructions", func() {
	//					args := []string{
	//						"mysql-tools",
	//					}
	//
	//					mysqlPlugin := new(command.MySQLPlugin)
	//					fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//					mysqlPlugin.Run(fakeCliConnection, args)
	//					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
	//				})
	//			})
	//
	//			Context("when -h is passed", func() {
	//				It("adds the usage to the Err method", func() {
	//					args := []string{
	//						"mysql-tools", "-h",
	//					}
	//
	//					mysqlPlugin := new(command.MySQLPlugin)
	//					fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//					mysqlPlugin.Run(fakeCliConnection, args)
	//					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
	//				})
	//				Context("when the -h is in the middle of arguments", func() {
	//
	//					It("adds the usage to the Err method", func() {
	//						args := []string{
	//							"mysql-tools", "foo", "-h", "bar",
	//						}
	//
	//						mysqlPlugin := new(command.MySQLPlugin)
	//						fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//						mysqlPlugin.Run(fakeCliConnection, args)
	//						Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
	//					})
	//				})
	//			})
	//
	//			Context("when the plugin does not pass us the name of the command", func() {
	//				It("adds an error which can be accessed from Err()", func() {
	//					args := []string{}
	//
	//					mysqlPlugin := new(command.MySQLPlugin)
	//					fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//					mysqlPlugin.Run(fakeCliConnection, args)
	//
	//					Expect(mysqlPlugin.Err().Error()).To(ContainSubstring("Error: plugin did not receive the expected input from the CLI"))
	//				})
	//			})
	//
	//			Context("when the plugin  passes  the 'CLI-MESSAGE-UNINSTALL'", func() {
	//
	//				It("does not err or print an error message when uninstalling", func() {
	//					args := []string{
	//						"CLI-MESSAGE-UNINSTALL",
	//					}
	//
	//					mysqlPlugin := new(command.MySQLPlugin)
	//					fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//					mysqlPlugin.Run(fakeCliConnection, args)
	//
	//					Expect(mysqlPlugin.Err()).To(BeNil())
	//				})
	//			})
	//		})
	//
	//		Context("version", func() {
	//			var (
	//				logOutput *bytes.Buffer
	//			)
	//			BeforeEach(func() {
	//				logOutput = &bytes.Buffer{}
	//				log.SetOutput(logOutput)
	//			})
	//
	//			It("outputs the version to the user", func() {
	//				args := []string{
	//					"mysql-tools", "version",
	//				}
	//
	//				mysqlPlugin := new(command.MySQLPlugin)
	//				fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//				mysqlPlugin.Run(fakeCliConnection, args)
	//				Expect(logOutput.String()).To(ContainSubstring(`built from source (unknown)`))
	//			})
	//		})
	//
	//		Context("unknown command", func() {
	//			It("adds the usage to the Err method", func() {
	//				args := []string{
	//					"mysql-tools", "foo",
	//				}
	//
	//				mysqlPlugin := new(command.MySQLPlugin)
	//				fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//				mysqlPlugin.Run(fakeCliConnection, args)
	//				Expect(mysqlPlugin.Err().Error()).To(ContainSubstring("unknown command: \"foo\""))
	//				Expect(mysqlPlugin.Err().Error()).To(ContainSubstring(usage))
	//			})
	//		})
	//
	//		FContext("migrate", func() {
	//			FIt("calls the migrate command", func(){
	//				//migrateCmd := &commandfakes.FakeCommand{}
	//				args := []string{
	//					"mysql-tools", "migrate", "arg1", "arg2",
	//				}
	//
	//				mysqlPlugin := new(command.MySQLPlugin)
	//				fakeCliConnection := &cliPluginFakes.FakeCliConnection{}
	//				mysqlPlugin.Run(fakeCliConnection, args)
	//				Expect(migrateCmd.RunCallCount()).To(Equal(1))
	//				Expect(migrateCmd.RunArgsForCall(0)).To(Equal([]string{"arg1", "arg2"}))
	//			})
	//		})
	//	})
})
