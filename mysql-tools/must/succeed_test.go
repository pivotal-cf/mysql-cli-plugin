package must_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/must"
)

func TestMust(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Must Test Suite")
}

var _ = Describe("Must", func() {
	successfulFunc := func() error { return nil }

	failingFunc := func() error { return fmt.Errorf("error") }

	When("no error is indicated", func() {
		It("Succeeds", func() {

			Expect(func() {
				must.Succeed(successfulFunc())
			}).ToNot(Panic())
		})
	})

	When("an error is indicated", func() {
		It("panics", func() {
			Expect(func() {
				must.Succeed(failingFunc())
			}).To(PanicWith("error"))
		})
	})
})

var _ = Describe("SucceedWithValue", func() {
	successfulFunc := func() (string, error) { return "value", nil }

	failingFunc := func() (string, error) { return "", fmt.Errorf("error") }

	When("no error is indicated", func() {
		It("returns a value", func() {
			Expect(must.SucceedWithValue(successfulFunc())).To(Equal("value"))
		})
	})

	When("an error is indicated", func() {
		It("panics", func() {
			Expect(func() {
				must.SucceedWithValue(failingFunc())
			}).To(PanicWith("error"))
		})
	})
})
