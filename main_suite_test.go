package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMysqlCLIPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestMysqlCLIPlugin Suite")
}

var _ = BeforeSuite(func() {
	cmd := exec.Command("go", "build", "-ldflags=-X 'main.version=0.0.1'", ".")
	Expect(cmd.Run()).To(Succeed())

	cmd = exec.Command("cf", "install-plugin", "-f", "mysql-cli-plugin")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, "60s").Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	cmd := exec.Command("cf", "uninstall-plugin", "MysqlTools")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, "60s").Should(gexec.Exit())
})