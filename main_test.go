// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package main_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	migrateUsage = `cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>
`
	findBindingUsage = `cf mysql-tools find-bindings [-h] <mysql-v1-service-name>
`
	longUsage = `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools find-bindings [-h] <mysql-v1-service-name>
   cf mysql-tools version
`
)

var _ = Describe("MysqlCliPlugin", func() {
	BeforeEach(func() {
		format.TruncatedDiff = false
	})

	It("displays long usage string with no arguments", func() {
		cmd := exec.Command("cf", "mysql-tools")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(string(session.Err.Contents())).To(Equal(longUsage))
	})

	It("Displays long usage string when -h flag is passed to base command", func() {
		cmd := exec.Command("cf", "mysql-tools", "-h")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(longUsage))
	})

	It("Displays long usage string when -h flag is passed to migrate command", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate", "-h")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(longUsage))
	})

	It("Displays long usage string when -h flag is passed to find-binding command", func() {
		cmd := exec.Command("cf", "mysql-tools", "find-bindings", "-h")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(ContainSubstring(longUsage))
	})

	It("migrate requires exactly 4 arguments", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(string(session.Err.Contents())).To(Equal(
			"Usage: " + migrateUsage +
				"\nthe required arguments `<source-service-instance>` and `<p.mysql-plan-type>` were not provided\n"))
	})

	It("find-binding requires exactly 1 arguments", func() {
		cmd := exec.Command("cf", "mysql-tools", "find-bindings")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(string(session.Err.Contents())).To(Equal(
			"Usage: " + findBindingUsage +
				"\nthe required argument `<mysql-v1-service-name>` was not provided\n"))
	})

	It("reports an error when given an unknown subcommand", func() {
		cmd := exec.Command("cf", "mysql-tools", "invalid")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(1))

		Expect(session.Err).To(gbytes.Say(`unknown command 'invalid'`))
	})

	It("shows plugin version", func() {
		if _, ok := os.LookupEnv("BEING_RUN_ON_CI"); !ok {
			Skip("Version check disabled since BEING_RUN_ON_CI is not set")
		}

		cmd := exec.Command("cf", "mysql-tools", "version")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "60s", "1s").Should(gexec.Exit(0))

		Expect(session).To(gbytes.Say(`\d+\.\d+\.\d+(-[\w.]+)? \(\w+\)`)) // Allows for versions like 0.1.0 (abcde) and  0.1.0-build.23 (b9ff4d2)
	})
})
