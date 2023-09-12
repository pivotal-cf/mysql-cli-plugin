package find_bindings_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFindBindings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FindBindings Suite")
}
