package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("MysqlCliPlugin", func() {
	It("migrates data given the right number of args", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate", "test-v1-donor", "test-v2-recipient")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "5m", "1s").Should(gexec.Exit(0))
		cmd = exec.Command("cf", "app", "migrate-app")
		output, _ := cmd.CombinedOutput()
		Expect(string(output)).To(ContainSubstring("App migrate-app not found"))
	})

	It("migrates data to a tls database given the right number of args", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate", "test-v1-donor", "test-v2-tls-recipient")
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

	It("shows a mysql version", func() {
		cmd := exec.Command("cf", "mysql-tools", "version")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(session).To(gbytes.Say(`\d\.\d\.\d\s\(.*\)`))
	})
})
