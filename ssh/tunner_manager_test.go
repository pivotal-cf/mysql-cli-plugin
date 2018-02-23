package ssh_test

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/service"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/ssh"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/ssh/sshfakes"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("TunnelManager", func() {
	var (
		tmpDir                string
		tunnelManager         *ssh.TunnelManager
		fakeCfCommandRunner   *sshfakes.FakeCfCommandRunner
		fakeDatabaseConnector *sshfakes.FakeDatabaseConnector
		servicesInfo          []*service.ServiceInfo
		appDir                = func() string {
			return filepath.Join(tmpDir, "static-app")
		}
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "mysql-plugin")
		Expect(err).NotTo(HaveOccurred())

		fakeCfCommandRunner = new(sshfakes.FakeCfCommandRunner)
		fakeDatabaseConnector = new(sshfakes.FakeDatabaseConnector)
		tunnelManager = ssh.NewTunnerManager(fakeCfCommandRunner, tmpDir)

		servicesInfo = []*service.ServiceInfo{
			{Hostname: "10.0.0.1"},
			{Hostname: "10.0.0.2"},
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Start", func() {
		BeforeEach(func() {
			fakeCfCommandRunner.CliCommandStub = func(args ...string) ([]string, error) {
				switch args[0] {
				case "push":
					Expect(args).To(Equal([]string{"push", "static-app", "--random-route", "-b", "staticfile_buildpack", "-p", appDir()}))
					return nil, nil
				default:
					Fail("unexpected call to CliCommand")
					return nil, errors.New("")
				}
			}

			fakeCfCommandRunner.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				Expect(strings.Join(args, " ")).To(MatchRegexp(`ssh static-app -N -L \d+:10.0.0.1:3306 -L \d+:10.0.0.2:3306`))
				return nil, nil
			}
		})

		It("creates pushes up an app exposes ssh ports to service instances", func() {
			Expect(tunnelManager.Start(servicesInfo)).To(Succeed())

			Expect(filepath.Join(appDir(), "Staticfile")).To(BeAnExistingFile())
			Expect(filepath.Join(appDir(), "index.html")).To(BeAnExistingFile())

			Expect(fakeCfCommandRunner.CliCommandCallCount()).To(Equal(1))
			Expect(fakeCfCommandRunner.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
		})

		Context("when pushing an application fails", func() {
			BeforeEach(func() {
				fakeCfCommandRunner.CliCommandStub = func(args ...string) ([]string, error) {
					switch args[0] {
					case "push":
						return nil, errors.New("some-error")
					default:
						Fail("unexpected call to CliCommand")
						return nil, errors.New("")
					}
				}
			})

			It("returns an error", func() {
				err := tunnelManager.Start(servicesInfo)
				Expect(err).To(MatchError("failed to push application: some-error"))
			})
		})
	})

	Describe("Close", func() {
		It("deletes the pushed application", func() {
			tunnelManager.Close()
			Expect(fakeCfCommandRunner.CliCommandCallCount()).To(Equal(1))
		})
	})
})
