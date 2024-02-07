package mshelpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/mshelpers"
)

var _ = Describe("multisite helpers", func() {
	Context("Validate Instance", func() {
		When("the targetInfo argument doesn't indicate a valid directory with saved target info", func() {
			It("returns an expected error", func() {
				err := mshelpers.ValidateInstance("badDirectory", "foo", "foo")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("target store badDirectory is not a directory"))
			})
		})
	})
})
