package specs

import (
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

var (
	TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
	Config    *config.Config
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var _ = BeforeSuite(func() {
	test_helpers.CheckForRequiredEnvVars([]string{
		"APP_DOMAIN",
		"DONOR_SERVICE_NAME",
		"DONOR_PLAN_NAME",
		"RECIPIENT_SERVICE_NAME",
		"RECIPIENT_PLAN_NAME",
	})

	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())

	Config = config.LoadConfig()

	TestSetup = workflowhelpers.NewTestSuiteSetup(Config)
	TestSetup.Setup()
})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})
