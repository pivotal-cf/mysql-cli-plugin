package mshelpers_test

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFindBindings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MSHelper Suite")
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
})
