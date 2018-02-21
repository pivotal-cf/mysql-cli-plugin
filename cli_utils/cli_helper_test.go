package cli_utils_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils/cli_utilsfakes"
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
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(Equal([]string{"delete-service-key", "some-instance", "service-key", "-f"}))
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
			Expect(cfCommandRunnerFake.CliCommandArgsForCall(0)).To(Equal([]string{"service-key", "some-instance", "service-key"}))
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
		var (
			donorServiceKey     cli_utils.ServiceKey
			recipientServiceKey cli_utils.ServiceKey
			tunnels             []cli_utils.Tunnel
			cfCommandRunnerFake *cli_utilsfakes.FakeCfCommandRunner
		)
		BeforeEach(func() {
			cfCommandRunnerFake = &cli_utilsfakes.FakeCfCommandRunner{}

			donorServiceKey = cli_utils.ServiceKey{Hostname: "10.0.0.1"}
			recipientServiceKey = cli_utils.ServiceKey{Hostname: "10.0.0.2"}
			tunnels = []cli_utils.Tunnel{
				{
					ServiceKey: donorServiceKey,
				},
				{
					ServiceKey: recipientServiceKey,
				},
			}

		})
		It("calls the correct cf ssh commands", func() {
			tunnelManager, err := cli_utils.NewTunnelManager(cfCommandRunnerFake, tunnels)
			tunnelManager.AppName = "my-cool-app"
			Expect(err).NotTo(HaveOccurred())

			err = tunnelManager.CreateSSHTunnel()
			Expect(err).NotTo(HaveOccurred())

			Expect(tunnels[0].Port).NotTo(BeZero())
			Expect(tunnels[1].Port).NotTo(BeZero())
			Expect(cfCommandRunnerFake.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal([]string{
				"ssh",
				"my-cool-app",
				"-N",
				"-L",
				fmt.Sprintf("%d:10.0.0.1:3306", tunnels[0].Port),
				"-L",
				fmt.Sprintf("%d:10.0.0.2:3306", tunnels[1].Port),
			}))
		})

		It("errors when the cf ssh fails", func() {

			tunnelManager, err := cli_utils.NewTunnelManager(cfCommandRunnerFake, tunnels)
			tunnelManager.AppName = "my-cool-app"
			Expect(err).NotTo(HaveOccurred())
			tunnelManager.CreateSSHTunnel()

			cfCommandRunnerFake.CliCommandWithoutTerminalOutputReturns(nil, errors.New("some-error"))
			err = tunnelManager.CreateSSHTunnel()
			Expect(err).To(MatchError("Failed to open ssh tunnel to app my-cool-app: some-error"))
		})

		It("returns no error when the cf ssh gives an EOF", func() {

			tunnelManager, err := cli_utils.NewTunnelManager(cfCommandRunnerFake, tunnels)
			tunnelManager.AppName = "my-cool-app"
			Expect(err).NotTo(HaveOccurred())
			tunnelManager.CreateSSHTunnel()

			cfCommandRunnerFake.CliCommandWithoutTerminalOutputReturns(nil, errors.New("Error: EOF"))
			Expect(tunnelManager.CreateSSHTunnel()).To(Succeed())
		})
	})

	Context("WaitForTunnel", func() {
		It("waits for select 1 to succeed", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			tunnelManager := &cli_utils.TunnelManager{
				Tunnels: []cli_utils.Tunnel{
					{DB: db}, {DB: db},
				},
			}

			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))
			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))
			mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("1"))
			mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("1"))

			Expect(tunnelManager.WaitForTunnel(5 * time.Second)).To(Succeed())
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		It("returns a error if it doesn't succeed before timeout", func() {
			db, mock, err := sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			tunnelManager := &cli_utils.TunnelManager{
				Tunnels: []cli_utils.Tunnel{
					{DB: db}, {DB: db},
				},
			}

			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))
			mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("some-error"))

			Expect(tunnelManager.WaitForTunnel(2 * time.Second)).To(MatchError("Timeout"))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

	})
})
