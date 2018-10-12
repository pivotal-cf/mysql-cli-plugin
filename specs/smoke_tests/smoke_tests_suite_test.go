package smoke_tests_test

import (
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

func TestSmokeTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmokeTests Suite")
}

var (
	TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
	Config    *config.Config
)

var _ = BeforeSuite(func() {
	test_helpers.CheckForRequiredEnvVars([]string{
		"DONOR_PLAN_NAME",
		"DONOR_SERVICE_NAME",
		"RECIPIENT_PLAN_NAME",
		"RECIPIENT_SERVICE_NAME",
	})

	Config = config.LoadConfig()

	TestSetup = workflowhelpers.NewTestSuiteSetup(Config)
	TestSetup.Setup()

	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())
})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})
