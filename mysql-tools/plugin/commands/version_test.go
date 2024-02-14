package commands_test

import (
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/version"
)

var _ = Describe("Version", func() {
	It("emits the version", func() {
		rd, wr, err := os.Pipe()
		Expect(err).NotTo(HaveOccurred())
		defer wr.Close()
		defer rd.Close()

		os.Stdout = wr

		err = commands.Version()
		Expect(err).NotTo(HaveOccurred())
		Expect(wr.Close()).To(Succeed())

		output, err := io.ReadAll(rd)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(output)).To(Equal(fmt.Sprintf("%s (%s)\n", version.Version, version.GitSHA)))
	})
})
