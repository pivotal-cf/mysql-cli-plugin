package service_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/service"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/service/servicefakes"
)

var _ = Describe("ServiceInstance", func() {

	var (
		instanceName    = "some-instance"
		fakeCliRunner   *servicefakes.FakeCfCommandRunner
		serviceInstance *service.ServiceInstance
	)

	BeforeEach(func() {
		fakeCliRunner = new(servicefakes.FakeCfCommandRunner)
		serviceInstance = service.NewServiceInstance(fakeCliRunner, instanceName)
	})

	Context("ServiceInfo", func() {

		BeforeEach(func() {
			fakeCliRunner.CliCommandStub = func(s ...string) ([]string, error) {
				switch s[0] {
				case "create-service-key":
					return nil, nil
				case "service-key":
					return []string{
						"Getting key key for service instance tls as admin...",
						"",
						"{",
						`"hostname": "some-hostname",`,
						`"name": "some-db-name",`,
						`"password": "some-password",`,
						`"username": "some-username"`,
						"}",
					}, nil

				default:
					return nil, fmt.Errorf("Passed in an unexpected command: %s", s[0])
				}
			}
		})

		It("returns a valid service key", func() {
			var (
				createServiceKeyCall = []string{"create-service-key", instanceName, service.ServiceKeyName}
				getServiceKeyCall    = []string{"service-key", instanceName, service.ServiceKeyName}
				expectedInfo         = service.ServiceInfo{
					Hostname: "some-hostname",
					Username: "some-username",
					Password: "some-password",
					DBName:   "some-db-name",
				}
			)

			info, err := serviceInstance.ServiceInfo()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCliRunner.CliCommandCallCount()).To(Equal(2))
			Expect(fakeCliRunner.CliCommandArgsForCall(0)).To(Equal(createServiceKeyCall))
			Expect(fakeCliRunner.CliCommandArgsForCall(1)).To(Equal(getServiceKeyCall))

			Expect(info).To(Equal(expectedInfo))
		})

		Context("when create-service-key fails", func() {
			BeforeEach(func() {
				fakeCliRunner.CliCommandStub = func(s ...string) ([]string, error) {
					switch s[0] {
					case "create-service-key":
						return nil, fmt.Errorf("some-error")
					default:
						return nil, fmt.Errorf("Passed in an unexpected command: %s", s[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := serviceInstance.ServiceInfo()
				Expect(err).To(MatchError("failed to create-service-key: some-error"))
			})
		})

		Context("when service-key fails", func() {
			BeforeEach(func() {
				fakeCliRunner.CliCommandStub = func(s ...string) ([]string, error) {
					switch s[0] {
					case "create-service-key":
						return nil, nil
					case "service-key":
						return nil, fmt.Errorf("some-error")
					default:
						return nil, fmt.Errorf("Passed in an unexpected command: %s", s[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := serviceInstance.ServiceInfo()
				Expect(err).To(MatchError("failed to create-service-key: some-error"))
			})
		})
	})

	Context("Cleanup", func() {
		It("deletes the associated service key", func() {
			serviceInstance.Cleanup()
			Expect(fakeCliRunner.CliCommandCallCount()).To(Equal(1))
		})
	})
})
