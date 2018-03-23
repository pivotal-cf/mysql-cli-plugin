package specs

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smoke Tests Suite")
}

var _ = BeforeSuite(func() {
	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())
})