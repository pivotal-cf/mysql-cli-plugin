package integration

import (
	cfplugin "code.cloudfoundry.org/cli/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin"
	"github.com/pivotal-cf/mysql-cli-plugin/plugin/pluginfakes"
	"io"
	"os/exec"
	"strings"
	"time"
)

var _ = Describe("Migrate Command", func() {

	var (
		p               *plugin.Mysql
		stdout          *gbytes.Buffer
		fakeConnWrapper *pluginfakes.FakeConnectionWrapper
		fakeExiter      *pluginfakes.FakeExiter

		expectCleanupToHaveOccurred = func() {
			Expect(fakeConnWrapper.CleanupCallCount()).To(Equal(1))
		}
	)

	BeforeEach(func() {
		//Fill in donor data
		command := exec.Command("mysql",
			"-h", "127.0.0.1",
			"-u", "root",
			"-P", donorPort,
			"-e",
			"DROP DATABASE IF EXISTS donor_db; CREATE DATABASE donor_db; CREATE TABLE donor_db.ketchup(num INT PRIMARY KEY); INSERT INTO donor_db.ketchup values(1);",
		)
		command.Env = []string{"MYSQL_PWD=" + donorPassword}
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))

		//Fill in recipient data
		command = exec.Command("mysql",
			"-h", "127.0.0.1",
			"-u", "root",
			"-P", recipientPort,
			"-p"+recipientPassword,
			"-e",
			"DROP DATABASE IF EXISTS recipient_db; CREATE DATABASE recipient_db;",
		)
		output, err = command.CombinedOutput()
		command.Env = []string{"MYSQL_PWD=" + recipientPassword}
		Expect(err).NotTo(HaveOccurred(), string(output))

		fakeConnWrapper = &pluginfakes.FakeConnectionWrapper{}
		fakeExiter = &pluginfakes.FakeExiter{}
		stdout = gbytes.NewBuffer()
		p = plugin.New(fakeConnWrapper, fakeExiter, io.MultiWriter(GinkgoWriter, stdout), time.Millisecond)
	})

	Context("when the user is not a space developer", func() {

		BeforeEach(func() {
			fakeConnWrapper.IsSpaceDeveloperReturns(false, nil)
		})

		It("returns an error message", func() {
			p.Run(nil, []string{"cf", "migrate", donorServiceName, recipientServiceName})

			Eventually(stdout).Should(gbytes.Say("You must have the 'Space Developer' privilege to use the 'cf mysql migrate' command"))
			Expect(fakeExiter.ExitArgsForCall(0)).To(Equal(1))
			expectCleanupToHaveOccurred()
		})
	})

	Context("when the user is space developer", func() {

		BeforeEach(func() {
			fakeConnWrapper.IsSpaceDeveloperReturns(true, nil)
		})

		Context("when migration task runs from donor to recipient instance", func() {
			BeforeEach(func() {
				fakeConnWrapper.ExecuteMigrateTaskStub = func(connection cfplugin.CliConnection, appName, sourceServiceName, destinationServiceName string) (state string, err error) {
					command := exec.Command("docker", "exec", "migrate-app", "/home/vcap/app/migrate", sourceServiceName, destinationServiceName)
					output, err := command.CombinedOutput()
					Expect(err).NotTo(HaveOccurred(), string(output))
					return "SUCCEEDED", nil
				}
			})

			It("migrates data from the donor to recipient database", func() {
				p.Run(nil, []string{"cf", "migrate", donorServiceName, recipientServiceName})

				Expect(fakeExiter.ExitArgsForCall(0)).To(BeZero())

				//get output
				command := exec.Command("mysql",
					"-h", "127.0.0.1",
					"-u", "root",
					"-P", recipientPort,
					"-sse",
					"select num from recipient_db.ketchup",
				)
				command.Env = []string{"MYSQL_PWD=" + recipientPassword}
				output, err := command.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
				Expect(strings.TrimSpace(string(output))).To(Equal("1"))

				expectCleanupToHaveOccurred()
			})
		})

		//TODO extend with TLS test

		Context("when migration fails", func() {
			BeforeEach(func() {
				fakeConnWrapper.ExecuteMigrateTaskStub = func(connection cfplugin.CliConnection, appName, sourceServiceName, destinationServiceName string) (state string, err error) {
					return "FAILED", nil
				}
			})

			It("shows log output", func() {
				p.Run(nil, []string{"cf", "migrate", "some-bad-service-name", "some-bad-service-name"})

				Eventually(stdout).Should(gbytes.Say("Migration failed. Fetching log output..."))
				Expect(fakeConnWrapper.ShowRecentLogsCallCount()).To(Equal(1))
				Expect(fakeExiter.ExitArgsForCall(0)).To(Equal(1))

				expectCleanupToHaveOccurred()
			})
		})
	})
})
