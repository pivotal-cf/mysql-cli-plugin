package cli_utils_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/cf/commands/servicekey"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils/cli_utilsfakes"

	"github.com/DATA-DOG/go-sqlmock"
)

var _ = Describe("CLI Helpers", func() {
	Context("PushApp", func() {
		It("calls the correct cf push commands", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			tmpDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).NotTo(HaveOccurred())
			err = cli_utils.PushApp(cfCommandRunnerFake, tmpDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(Equal([]string{"push", "static-app", "-p", filepath.Join(tmpDir, "static-app")}))
		})

		It("errors if the tmpDir does not exist", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			tmpDir := "invalid-dir"
			err := cli_utils.PushApp(cfCommandRunnerFake, tmpDir)
			Expect(err).To(MatchError("Failed to create app directory: mkdir invalid-dir/static-app: no such file or directory"))
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(0))
		})

		It("errors when the cf push fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			tmpDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).NotTo(HaveOccurred())
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-fun-error"))
			err = cli_utils.PushApp(cfCommandRunnerFake, tmpDir)
			Expect(err).To(MatchError("Failed to push app: some-fun-error"))
		})
	})

	Context("DeleteApp", func() {
		It("calls the correct cf delete commands", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			err := cli_utils.DeleteApp(cfCommandRunnerFake)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(ConsistOf("delete", "static-app", "-f"))
		})

		It("errors when the cf delete fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-error"))
			err := cli_utils.DeleteApp(cfCommandRunnerFake)
			Expect(err).To(MatchError("Failed to delete app: some-error"))
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(ConsistOf("delete", "static-app", "-f"))
		})
	})

	Context("CreateServiceKey", func() {
		It("calls the correct cf create-service-key commands", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			srcInstanceName := "some-instance"
			err := cli_utils.CreateServiceKey(cfCommandRunnerFake, srcInstanceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(ConsistOf("create-service-key", "some-instance", "service-key"))
		})

		It("errors when the cf create-service-key fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-error"))
			err := cli_utils.CreateServiceKey(cfCommandRunnerFake, "some-instance")
			Expect(err).To(MatchError("Failed to create-service-key: some-error"))
		})
	})

	Context("DeleteServiceKey", func() {
		It("calls the correct cf delete-service-key commands", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			srcInstanceName := "some-instance"
			err := cli_utils.DeleteServiceKey(cfCommandRunnerFake, srcInstanceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(ConsistOf("delete-service-key", "some-instance", "service-key", "f"))
		})

		It("errors when the cf delete-service-key fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-error"))
			err := cli_utils.DeleteServiceKey(cfCommandRunnerFake, "some-instance")
			Expect(err).To(MatchError("Failed to delete-service-key: some-error"))
		})
	})

	Context("GetServiceKey", func() {
		It("calls the correct cf service-key commands and returns a serviceKey", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			srcInstanceName := "some-instance"
			cliOutput := []string{
				"Getting key key for service instance tls as admin...",
				"",
				"{",
				`"hostname": "some-hostname",`,
				`"name": "some-db-name",`,
				`"password": "some-password",`,
				`"username": "some-username"`,
				"}",
			}
			cfCommandRunnerFake.CliCommandReturns(cliOutput, nil)
			serviceKey, err := cli_utils.GetServiceKey(cfCommandRunnerFake, srcInstanceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfCommandRunnerFake.CliCommandCallCount()).To(Equal(1))
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(ConsistOf("service-key", "some-instance", "service-key"))
			Expect(serviceKey.Hostname).To(Equal("some-hostname"))
			Expect(serviceKey.Username).To(Equal("some-username"))
			Expect(serviceKey.Password).To(Equal("some-password"))
			Expect(serviceKey.DBName).To(Equal("some-db-name"))
		})

		It("errors when the cf service-key fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-error"))
			_, err := cli_utils.GetServiceKey(cfCommandRunnerFake, "some-instance")
			Expect(err).To(MatchError("Failed to get service-key: some-error"))
		})

		It("errors when the cf service-key fails", func() {
			cliOutput := []string{"bad", "json"}
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(cliOutput, nil)
			_, err := cli_utils.GetServiceKey(cfCommandRunnerFake, "some-instance")
			Expect(err).To(MatchError("Failed to get service-key: unexpected end of JSON input"))
		})
	})

	Context("CreateSshTunnel", func() {
		It("calls the correct cf ssh commands", func() {
			serviceKey := servicekey.ServiceKey{}
			timeout := time.Second
			tunnelManager := NewTunnelManager(serviceKey, timeout)
			tunnelManager.CreateSshTunnel()
		})

		It("errors when the cf ssh fails", func() {
			cfCommandRunnerFake := &cli_utilsfakes.FakeCfCommandRunner{}
			cfCommandRunnerFake.CliCommandReturns(nil, errors.New("some-error"))
			err := cli_utils.CreateSshTunnel(cfCommandRunnerFake, "some-hostname")
			Expect(err).To(MatchError("Failed to open ssh tunnel for service host some-hostname: some-error"))
		})
	})

	Context("WaitForTunnel", func() {
		It("waits for select 1 to succeed", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))
			mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("1"))

			Expect(cli_utils.WaitForTunnel(db, 2001*time.Millisecond)).To(Succeed())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("errors if the select does not return before the timeout", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))

			Expect(cli_utils.WaitForTunnel(db, 1001*time.Millisecond)).NotTo(Succeed())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	})
})
