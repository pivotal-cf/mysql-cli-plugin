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

package specs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

var _ = Describe("Migrate Integration Tests", func() {
	var (
		appDomain      string
		appName        string
		destInstance   string
		sourceInstance string
		serviceKey     = "tls-key"
	)

	BeforeEach(func() {
		appDomain = os.Getenv("APP_DOMAIN")

		sourceInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_SOURCE")
		test_helpers.CreateService(os.Getenv("DONOR_SERVICE_NAME"), os.Getenv("DONOR_PLAN_NAME"), sourceInstance)
		destInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_DEST")
		test_helpers.CreateService(os.Getenv("RECIPIENT_SERVICE_NAME"), os.Getenv("RECIPIENT_PLAN_NAME"), destInstance)

		test_helpers.WaitForService(sourceInstance, `[Ss]tatus:\s+create succeeded`)
		test_helpers.WaitForService(destInstance, `[Ss]tatus:\s+create succeeded`)
	})

	AfterEach(func() {
		if appName != "" {
			test_helpers.DeleteApp(appName)
		}

		test_helpers.DeleteServiceKey(destInstance, serviceKey)
		test_helpers.DeleteService(destInstance)
		test_helpers.DeleteService(sourceInstance)
		test_helpers.WaitForService(destInstance, fmt.Sprintf("Service instance %s not found", destInstance))
		test_helpers.WaitForService(sourceInstance, fmt.Sprintf("Service instance %s not found", sourceInstance))
	})

	It("migrates data from donor to recipient", func() {
		var (
			readValue  string
			appURI     string
			albumID    string
			writeValue string
		)

		By("Binding an app to the source instance", func() {
			appName = generator.PrefixedRandomName("MYSQL", "APP")
			test_helpers.PushApp(appName, "assets/spring-music")

			test_helpers.BindAppToService(appName, sourceInstance)
			test_helpers.StartApp(appName)
		})

		By("Writing data to the source instance", func() {
			appURI = appName + "." + appDomain
			test_helpers.CheckAppInfo(true, appURI, sourceInstance)

			writeValue = "DM Greatest Hits"
			albumID = test_helpers.WriteData(true, appURI, writeValue)
			readValue = test_helpers.ReadData(true, appURI, albumID)
			Expect(readValue).To(Equal(writeValue))

			test_helpers.UnbindAppFromService(appName, sourceInstance)
		})

		By("Migrating data using the migrate command", func() {
			cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destInstance)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "5m", "1s").Should(gexec.Exit(0))
		})

		By("Binding the app to the destination instance and reading back data", func() {
			test_helpers.BindAppToService(appName, destInstance)
			test_helpers.ExecuteCfCmd("restage", appName)

			readValue = test_helpers.ReadData(true, appURI, albumID)
			Expect(readValue).To(Equal(writeValue))
		})
	})

	Context("when migrating data to a TLS enabled service-instance", func() {
		BeforeEach(func() {
			test_helpers.CreateServiceKey(destInstance, serviceKey)
			key := test_helpers.GetServiceKey(destInstance, serviceKey)
			test_helpers.DeleteServiceKey(destInstance, serviceKey)

			test_helpers.ExecuteCfCmd("update-service", destInstance, "-c", fmt.Sprintf(`{ "enable_tls": [%q] }`, key.Hostname))
			test_helpers.WaitForService(destInstance, `[Ss]tatus:\s+update succeeded`)
		})

		It("migrates the data successfully", func() {
			cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destInstance)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "5m", "1s").Should(gexec.Exit(0))
		})
	})
})
