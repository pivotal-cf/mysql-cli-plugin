package specs

import (
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var (
	TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
	Config    *config.Config
)

var _ = BeforeSuite(func() {
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
