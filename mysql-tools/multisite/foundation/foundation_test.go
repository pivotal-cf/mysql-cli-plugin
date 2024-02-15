package foundation_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/foundation"
)

func TestFoundation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Foundation Unit Tests")
}

var _ = Describe("Foundation", Label("unit"), func() {
	var subject foundation.Handler

	BeforeEach(func() {
		subject = foundation.New("some-name", "some-cf-home")
	})

	It("reports the foundation identifier it was constructed with", func() {
		Expect(subject.ID()).To(Equal("some-name"))
	})

	Context("UpdateServiceAndWait", func() {
		It("runs cf update-service with arbitrary parameters and waits for the operation to complete", func() {
			var capturedArgs []string
			subject.CF = func(cfHomeDir string, args ...string) (string, error) {
				capturedArgs = args

				return "OK", nil
			}

			err := subject.UpdateServiceAndWait("some-instance", `{ "some-param": "value" }`)
			Expect(err).NotTo(HaveOccurred())

			Expect(capturedArgs).To(Equal([]string{
				"update-service",
				"some-instance",
				"-c",
				`{ "some-param": "value" }`,
				"--wait",
			}))
		})

		When("update-service fails", func() {
			BeforeEach(func() {
				subject.CF = func(cfHomeDir string, args ...string) (string, error) {
					if args[0] == "update-service" {
						return "error", fmt.Errorf("some update-service error")
					}
					return "OK", nil
				}
			})

			It("returns an error", func() {
				err := subject.UpdateServiceAndWait("some-instance", `{ "some-param": "value" }`)
				Expect(err).To(MatchError(`some update-service error`))
			})
		})
	})
	Context("CreateHostInfoKey", func() {
		BeforeEach(func() {
			subject.CF = func(_ string, args ...string) (string, error) {
				Expect(args).ToNot(BeEmpty())

				switch args[0] {
				case "create-service-key":
					return "OK", nil
				case "service-key":
					contents, err := os.ReadFile("fixtures/sample-host-info-key.json")
					return string(contents), err
				default:
					panic("unsupported cf command: " + args[0])
				}
			}
		})

		It("creates a service key", func() {
			key, err := subject.CreateHostInfoKey("some-instance")
			Expect(err).NotTo(HaveOccurred())

			Expect(key).To(MatchJSON(`{
              "replication": {
                "peer-info": {
                  "hostname": "4b9b01ba-47a5-4bf3-98ee-10533af31959.mysql.service.internal",
                  "ip": "10.0.16.11",
                  "system_domain": "bali.dedicated-mysql.cf-app.com",
                  "uuid": "4b9b01ba-47a5-4bf3-98ee-10533af31959"
                },
                "role": "leader"
              }
            }`))
		})

		When("creating a service key fails", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					return "", fmt.Errorf("some cf create-service-key error")
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateHostInfoKey("some-instance")
				Expect(err).To(MatchError(`failed to create service key: some cf create-service-key error`))

			})
		})

		When("retrieving a service key fails", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					switch args[0] {
					case "create-service-key":
						return "OK", nil
					case "service-key":
						return "", fmt.Errorf("some cf service-key retrieval error")
					default:
						panic("unsupported cf command: " + args[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateHostInfoKey("some-instance")
				Expect(err).To(MatchError(MatchRegexp(`failed to retrieve service-key 'host-info-\d+' on instance 'some-instance': some cf service-key retrieval error`)))
			})
		})

		When("cf service-key output does not include valid json", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					switch args[0] {
					case "create-service-key":
						return "OK", nil
					case "service-key":
						return "foo bar bar...\n# this command succeed and should output valid json, but it did not", nil
					default:
						panic("unsupported cf command: " + args[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateHostInfoKey("some-instance")
				Expect(err).To(MatchError(ContainSubstring(`failed to parse host-info service key: invalid character`)))
			})
		})
	})

	Context("CreateCredentialsKey", func() {
		BeforeEach(func() {
			subject.CF = func(_ string, args ...string) (string, error) {
				Expect(args).ToNot(BeEmpty())

				switch args[0] {
				case "create-service-key":
					return "OK", nil
				case "service-key":
					contents, err := os.ReadFile("fixtures/sample-credentials-key.json")
					return string(contents), err
				default:
					panic("unsupported cf command: " + args[0])
				}
			}
		})

		It("creates a service key", func() {
			key, err := subject.CreateCredentialsKey("some-instance")
			Expect(err).NotTo(HaveOccurred())

			Expect(key).To(MatchJSON(`{
              "replication": {
                "credentials": {
                  "password": "replication-password",
                  "username": "replication-user"
                },
                "peer-info": {
                  "hostname": "tcp.sample.foundation.tld",
                  "ip": "1.2.3.4",
                  "ports": {
                    "agent": 1301,
                    "backup": 1102,
                    "mysql": 1037
                  },
                  "uuid": "47de60e7-c8d3-4dff-b1be-09bec561286e"
                },
                "role": "follower"
              }
            }`))
		})

		When("creating a service key fails", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					return "", fmt.Errorf("some cf create-service-key error")
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateCredentialsKey("some-instance")
				Expect(err).To(MatchError(`failed to create service key: some cf create-service-key error`))
			})
		})

		When("retrieving a service key fails", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					switch args[0] {
					case "create-service-key":
						return "OK", nil
					case "service-key":
						return "", fmt.Errorf("some cf service-key retrieval error")
					default:
						panic("unsupported cf command: " + args[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateCredentialsKey("some-instance")
				Expect(err).To(MatchError(MatchRegexp(`failed to retrieve service-key 'credentials-\d+' on instance 'some-instance': some cf service-key retrieval error`)))
			})
		})

		When("cf service-key output does not include valid json", func() {
			BeforeEach(func() {
				subject.CF = func(_ string, args ...string) (string, error) {
					switch args[0] {
					case "create-service-key":
						return "OK", nil
					case "service-key":
						return "foo bar bar...\n# this command succeed and should output valid json, but it did not", nil
					default:
						panic("unsupported cf command: " + args[0])
					}
				}
			})

			It("returns an error", func() {
				_, err := subject.CreateCredentialsKey("some-instance")
				Expect(err).To(MatchError(ContainSubstring(`failed to parse host-info service key: invalid character`)))
			})
		})
	})

	Context("InstanceExists", func() {
		It("succeeds if an instance exists", func() {
			subject.CF = func(cfHomeDir string, args ...string) (string, error) {
				return "", nil
			}

			err := subject.InstanceExists("some-instance")
			Expect(err).NotTo(HaveOccurred())
		})

		When("the instance does not exist", func() {
			BeforeEach(func() {
				subject.CF = func(cfHomeDir string, args ...string) (string, error) {
					return "some output\nService instance 'some-other-instance' not found", nil
				}
			})

			It("returns a helpful error", func() {
				err := subject.InstanceExists("some-other-instance")
				Expect(err).To(MatchError(`instance 'some-other-instance' does not exist`))
			})
		})

		When("cf service fails to determine if the instance doesn't exist for some other reason", func() {
			BeforeEach(func() {
				subject.CF = func(cfHomeDir string, args ...string) (string, error) {
					return "E.g. TCP connection timed out connecting to CF $api endpoint", fmt.Errorf("some tcp timeout message")
				}
			})

			It("still returns an error", func() {
				err := subject.InstanceExists("some-other-instance")
				Expect(err).To(MatchError(`error when checking whether instance exists: some tcp timeout message`))
			})
		})
	})
})
