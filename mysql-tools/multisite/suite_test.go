package multisite_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkflows(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multi-Site Unit Test Suite")
}
