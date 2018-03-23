package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("MysqlCliPlugin", func() {
	It("requires a command", func() {
		cmd := exec.Command("cf", "mysql-tools")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(session.Err).To(gbytes.Say(`Please pass in a command \[migrate\|version\] to mysql-tools`))
	})

	It("requires exactly 4 arguments", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate")
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

	It("shows plugin version", func() {
		cmd := exec.Command("cf", "mysql-tools", "version")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(session).To(gbytes.Say(`\d+\.\d+\.\d+(-[\w.]+)? \(\w+\)`)) // Allows for versions like 0.1.0 (abcde) and  0.1.0-build.23 (b9ff4d2)
	})
})
