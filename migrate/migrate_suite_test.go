package migrate_test

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMigrate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migrate Suite")
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
})