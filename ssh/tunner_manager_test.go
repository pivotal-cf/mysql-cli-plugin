package ssh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/service"
	"github.com/pivotal-cf/mysql-cli-plugin/ssh"
	"github.com/pivotal-cf/mysql-cli-plugin/ssh/sshfakes"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var _ = Describe("TunnelManager", func() {
	var (
		tmpDir              string
		tunnelManager       *ssh.TunnelManager
		fakeCfCommandRunner *sshfakes.FakeCfCommandRunner
		fakeDB              *sshfakes.FakeDB
		servicesInfo        []*service.ServiceInfo
		appDir              = func() string {
			return filepath.Join(tmpDir, "static-app")
		}
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "mysql-plugin")
		Expect(err).NotTo(HaveOccurred())

		fakeCfCommandRunner = new(sshfakes.FakeCfCommandRunner)
		fakeDB = new(sshfakes.FakeDB)
		tunnelManager = ssh.NewTunnerManager(fakeCfCommandRunner, fakeDB, tmpDir, 3*time.Second)

		servicesInfo = []*service.ServiceInfo{
			{Hostname: "10.0.0.1"},
			{Hostname: "10.0.0.2"},
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Start", func() {
		It("creates pushes up an app exposes ssh ports to service instances", func() {
			Expect(tunnelManager.Start(servicesInfo...)).To(Succeed())

			Expect(filepath.Join(appDir(), "Staticfile")).To(BeAnExistingFile())
			Expect(filepath.Join(appDir(), "index.html")).To(BeAnExistingFile())

			args := fakeCfCommandRunner.CliCommandArgsForCall(0)
			Expect(args).To(Equal([]string{"push", "static-app", "--random-route", "-b", "staticfile_buildpack", "-p", appDir()}))

			args = fakeCfCommandRunner.CliCommandWithoutTerminalOutputArgsForCall(0)
			Expect(strings.Join(args, " ")).To(MatchRegexp(`ssh static-app -N -L \d+:10.0.0.1:3306 -L \d+:10.0.0.2:3306`))

			serviceInfo, port := fakeDB.PingArgsForCall(0)
			Expect(serviceInfo).To(BeEquivalentTo(servicesInfo[0]))
			Expect(port).NotTo(BeZero())
			Expect(servicesInfo[0].LocalSSHPort).To(Equal(port))

			serviceInfo, port = fakeDB.PingArgsForCall(1)
			Expect(serviceInfo).To(Equal(servicesInfo[1]))
			Expect(port).NotTo(BeZero())
			Expect(servicesInfo[1].LocalSSHPort).To(Equal(port))
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
				err := tunnelManager.Start(servicesInfo...)
				Expect(err).To(MatchError("failed to push application: some-error"))
			})
		})

		Context("when waiting for ssh requires multiple attempts", func() {
			BeforeEach(func() {
				fakeDB.PingReturnsOnCall(0, errors.New("some-error"))
				fakeDB.PingReturnsOnCall(1, errors.New("some-error"))
			})

			It("eventually succeeds", func() {
				Expect(tunnelManager.Start(servicesInfo...)).To(Succeed())
			})
		})

		Context("when waiting for ssh doesn't succeed before timeout", func() {
			BeforeEach(func() {
				fakeDB.PingReturns(errors.New("some-error"))
			})

			It("returns an error", func() {
				err := tunnelManager.Start(servicesInfo...)
				Expect(err).To(MatchError("timeout"))
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
