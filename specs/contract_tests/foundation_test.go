package contract_tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/config"
	"github.com/cloudfoundry/cf-test-helpers/v2/generator"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/foundation"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

func TestFoundation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Foundation Test Suite")
}

var (
	Config    *config.Config
	TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
)

var _ = BeforeSuite(func() {
	Config = config.LoadConfig()
	TestSetup = workflowhelpers.NewTestSuiteSetup(Config)
	TestSetup.Setup()
})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})

var _ = Describe("Foundation", Ordered, func() {
	var api foundation.Handler

	var serviceInstanceName string

	BeforeAll(func() {
		cfHomeDir := os.Getenv("CF_HOME")

		api = foundation.New("foundation-name", cfHomeDir)

		serviceInstanceName = generator.PrefixedRandomName("plugin", "contract-test")
		test_helpers.CreateService(
			"p.mysql",
			os.Getenv("SINGLE_NODE_PLAN_NAME"),
			serviceInstanceName,
			// Register a fake "follower" w/ this instance so creating a credentials key will work without
			// an extra update-service
			"-c", "./fixtures/sample-host-info-key.json",
		)
		test_helpers.WaitForService(serviceInstanceName, `[Ss]tatus:\s+create succeeded`)
	})

	Context("UpdateServiceAndWait", func() {
		It("updates a service instance with arbitrary params", func() {
			err := api.UpdateServiceAndWait(serviceInstanceName, `{}`)
			Expect(err).NotTo(HaveOccurred())

			session := cf.Cf("service", serviceInstanceName).Wait("15m")
			Expect(session.Out).To(gbytes.Say("update succeeded"))
		})

		It("updates a service instance with arbitrary params", func() {
			err := api.UpdateServiceAndWait(serviceInstanceName, `{ "invalid-arbitrary-params": "value"}`)
			Expect(err).To(MatchError(HavePrefix("cf update-service failed: exit status 1")))
		})
	})

	Context("InstanceExists", func() {
		It("succeeds when the instance exists", func() {
			err := api.InstanceExists(serviceInstanceName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when the instance does not exist", func() {
			err := api.InstanceExists("does-not-exist-instance-name")
			Expect(err).To(MatchError(`instance 'does-not-exist-instance-name' does not exist`, serviceInstanceName))
		})

		It("fails when some other error occurs", func() {
			api.CfHomeDir, _ = os.MkdirTemp("", "no_creds_cf_home_")
			defer func() {
				api.CfHomeDir = os.Getenv("CF_HOME")
			}()

			err := api.InstanceExists("some-instance")
			Expect(err).To(MatchError(MatchRegexp(`(?s)error when checking whether instance exists: cf service failed: exit status 1\noutput:\n.*FAILED`)))
		})
	})

	Context("CreateHostInfoKey", func() {
		It("succeeds when the instance exists", func() {
			key, err := api.CreateHostInfoKey(serviceInstanceName)
			Expect(err).NotTo(HaveOccurred())

			var value map[string]any
			Expect(json.Unmarshal([]byte(key), &value)).To(Succeed())

			Expect(value).To(gstruct.MatchAllKeys(gstruct.Keys{
				"replication": gstruct.MatchAllKeys(gstruct.Keys{
					"role": Equal("leader"),
					"peer-info": gstruct.MatchAllKeys(gstruct.Keys{
						"uuid":          MatchRegexp(`[a-f0-9-]{36}`),
						"hostname":      MatchRegexp(`[a-f0-9-]{36}\.mysql\.service\.internal`),
						"ip":            MatchRegexp(`\d+\.\d+\.\d+\.\d+`),
						"system_domain": Not(BeEmpty()),
					}),
				}),
			}))
		})

		It("fails when some other error occurs", func() {
			_, err := api.CreateHostInfoKey("does-not-exist")
			Expect(err).To(MatchError(MatchRegexp(`(?ms)cf create-service-key failed: exit status 1\noutput:\n.*FAILED`)))
		})
	})

	Context("CreateCredentialsKey", func() {
		It("succeeds when the instance exists", func() {
			key, err := api.CreateCredentialsKey(serviceInstanceName)
			Expect(err).NotTo(HaveOccurred())

			var value map[string]any
			Expect(json.Unmarshal([]byte(key), &value)).To(Succeed())

			Expect(value).To(gstruct.MatchAllKeys(gstruct.Keys{
				"replication": gstruct.MatchAllKeys(gstruct.Keys{
					"role": Equal("follower"),
					"credentials": SatisfyAll(
						HaveKeyWithValue("username", Not(BeEmpty())),
						HaveKeyWithValue("password", Not(BeEmpty())),
					),
					"peer-info": gstruct.MatchKeys(gstruct.IgnoreExtras, gstruct.Keys{
						"uuid":     MatchRegexp(`[a-f0-9-]{36}`),
						"hostname": Not(BeEmpty()),
						"ports": SatisfyAll(
							HaveKey("mysql"),
							HaveKey("agent"),
							HaveKey("backup"),
						),
					}),
				}),
			}))
		})

		It("fails when some other error occurs", func() {
			_, err := api.CreateCredentialsKey("does-not-exist")
			Expect(err).To(MatchError(MatchRegexp(`(?ms)cf create-service-key failed: exit status 1\noutput:\n.*FAILED`)))
		})
	})
})
