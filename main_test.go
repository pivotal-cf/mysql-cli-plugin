package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("MysqlV2CliPlugin", func() {
	It("pushes an app given the right number of args", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate", "test-v1-donor", "test-v2-recipient")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "5m", "1s").Should(gexec.Exit(0))
	})

	It("requires exactly 4 arguments", func() {
		cmd := exec.Command("cf", "mysql-tools")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(session.Err).To(gbytes.Say(`Usage: cf mysql-tools migrate <v1-service-instance> <v2-service-instance>`))
	})

	It("reports an error when given an unknown subcommand", func() {
		cmd := exec.Command("cf", "mysql-tools", "invalid")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(session.Err).To(gbytes.Say(`Unknown command 'invalid'`))

	})
})