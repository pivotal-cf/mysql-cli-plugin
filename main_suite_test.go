package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMysqlV2CliPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MysqlV2CliPlugin Suite")
}

var _ = BeforeSuite(func() {
	binaryPath, err := gexec.Build("github.com/pivotal-cf/mysql-v2-cli-plugin")
	Expect(err).NotTo(HaveOccurred())

	command := exec.Command("cf", "install-plugin", binaryPath, "-f")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, "1m", "1s").Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	command := exec.Command("cf", "uninstall-plugin", "MysqlMigrate")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
})
