package contract_tests

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cf-test-helpers/v2/config"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
)

func TestFoundation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Contract Tests", Label("contract"))
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
