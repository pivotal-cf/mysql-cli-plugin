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
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

type credential struct {
	Name        string            `json:"name"`
	Credentials map[string]string `json:"credentials"`
}

type envResult struct {
	Env struct {
		VCAPServices map[string][]credential `json:"VCAP_SERVICES"`
	} `json:"system_env_json"`
}

var _ = Describe("Migrate Integration Tests", func() {
	var (
		appDomain          string
		springAppName      string
		destInstance       string
		destPlan           string
		sourceInstance     string
		sourceInstanceGUID string
		destInstanceGUID   string
		serviceKey         = "tls-key"
	)

	BeforeEach(func() {
		destPlan = os.Getenv("RECIPIENT_PLAN_NAME")
	})

	It("fails on invalid donor service instance", func() {
		cmd := exec.Command("cf", "mysql-tools", "migrate", "fake-donor-service", destPlan)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "1m", "1s").Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("Service instance fake-donor-service not found"))
	})

	Context("when a valid donor service instance exists", func() {
		BeforeEach(func() {
			appDomain = os.Getenv("APP_DOMAIN")

			sourceInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_SOURCE")
			test_helpers.CreateService(os.Getenv("DONOR_SERVICE_NAME"), os.Getenv("DONOR_PLAN_NAME"), sourceInstance)
			destInstance = sourceInstance + "-new"

			test_helpers.WaitForService(sourceInstance, `[Ss]tatus:\s+create succeeded`)
			sourceInstanceGUID = test_helpers.InstanceUUID(sourceInstance)
		})

		AfterEach(func() {
			if springAppName != "" {
				test_helpers.DeleteApp(springAppName)
			}

			test_helpers.DeleteServiceKey(destInstance, serviceKey)
			test_helpers.DeleteService(destInstance)
			test_helpers.DeleteService(sourceInstance)
			test_helpers.WaitForService(destInstance, fmt.Sprintf("Service instance %s not found", destInstance))
			test_helpers.WaitForService(sourceInstance, fmt.Sprintf("Service instance %s not found", sourceInstance))
		})

		It("migrates data from donor to recipient", func() {
			var (
				readValue    string
				springAppURI string
				albumID      string
				writeValue   string
			)

			By("Binding an app to the source instance", func() {
				springAppName = generator.PrefixedRandomName("MYSQL", "APP")
				test_helpers.PushApp(springAppName, "assets/spring-music")

				test_helpers.BindAppToService(springAppName, sourceInstance)
				test_helpers.StartApp(springAppName)
			})

			By("Writing data to the source instance", func() {
				springAppURI = springAppName + "." + appDomain
				test_helpers.CheckAppInfo(true, springAppURI, sourceInstance)

				writeValue = "DM Greatest Hits"
				albumID = test_helpers.WriteData(true, springAppURI, writeValue)
				readValue = test_helpers.ReadData(true, springAppURI, albumID)

				Expect(readValue).To(Equal(writeValue))

				test_helpers.UnbindAppFromService(springAppName, sourceInstance)
			})

			By("Migrating data using the migrate command", func() {
				cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destPlan)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10m", "1s").Should(gexec.Exit(0))
			})

			By("Verifying the destination service was renamed to the source's name", func() {
				destInstanceGUID = test_helpers.InstanceUUID(sourceInstance)
				Expect(destInstanceGUID).NotTo(Equal(sourceInstanceGUID))
			})

			By("Binding the app to the newly created destination instance and reading back data", func() {
				test_helpers.BindAppToService(springAppName, sourceInstance)
				test_helpers.ExecuteCfCmd("restage", springAppName)

				readValue = test_helpers.ReadData(true, springAppURI, albumID)
				Expect(readValue).To(Equal(writeValue))
			})

			By("Verifying that the credhub reference in the binding only contains the destination service's GUID", func() {
				appGUID := strings.TrimSpace(test_helpers.ExecuteCfCmd("app", springAppName, "--guid"))

				envOutput := test_helpers.ExecuteCfCmd("curl", fmt.Sprintf("/v2/apps/%s/env", appGUID))

				var result envResult

				err := json.Unmarshal([]byte(envOutput), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Env.VCAPServices).To(HaveKey(os.Getenv("RECIPIENT_SERVICE_NAME")))
				mysqlServices := result.Env.VCAPServices[os.Getenv("RECIPIENT_SERVICE_NAME")]

				Expect(mysqlServices).To(HaveLen(1))
				Expect(mysqlServices[0].Credentials).To(HaveKey("credhub-ref"))
				Expect(mysqlServices[0].Credentials["credhub-ref"]).To(ContainSubstring(destInstanceGUID))
				Expect(mysqlServices[0].Credentials["credhub-ref"]).NotTo(ContainSubstring(sourceInstanceGUID))
			})

			By("Verifying TLS was enabled on the recipient instance", func() {
				test_helpers.CreateServiceKey(sourceInstance, "tls-check")
				serviceKey := test_helpers.GetServiceKey(sourceInstance, "tls-check")
				test_helpers.DeleteServiceKey(sourceInstance, "tls-check")

				Expect(serviceKey.TLS.Cert.CA).
					NotTo(BeEmpty(),
						"Expected recipient service instance to be TLS enabled, but it was not")
			})
		})

		Context("when the --no-cleanup flag is specified", func() {
			var (
				destinationGUID string
			)
			AfterEach(func() {
				srcGUID := test_helpers.InstanceUUID(sourceInstance)
				test_helpers.UnbindAllAppsFromService(srcGUID)
				test_helpers.UnbindAllAppsFromService(destinationGUID)
			})

			It("doesn't delete the migration app when the --no-cleanup flag is specified", func() {
				cmd := exec.Command("cf", "mysql-tools", "migrate", "--no-cleanup", sourceInstance, destPlan)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, "10m", "1s").Should(gexec.Exit(0))

				destinationGUID = test_helpers.InstanceUUID(sourceInstance)
				appGUIDs := test_helpers.BoundAppGUIDs(destinationGUID)
				Expect(appGUIDs).NotTo(BeEmpty())
			})
		})

		It("fails on invalid service plan", func() {
			cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, "fake-service-plan")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "1m", "1s").Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Could not find plan with name fake-service-plan"))
		})
	})

	Context("When migration fails", func() {
		BeforeEach(func() {
			sourceInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_SOURCE")
			test_helpers.CreateService(os.Getenv("DONOR_SERVICE_NAME"), os.Getenv("DONOR_PLAN_NAME"), sourceInstance)
			destInstance = sourceInstance + "-new"
			test_helpers.WaitForService(sourceInstance, `[Ss]tatus:\s+create succeeded`)

			createInvalidMigrationState(sourceInstance)
		})

		AfterEach(func() {
			test_helpers.DeleteService(sourceInstance)
			test_helpers.DeleteService(destInstance)
			test_helpers.WaitForService(sourceInstance, fmt.Sprintf("Service instance %s not found", sourceInstance))
			test_helpers.WaitForService(destInstance, fmt.Sprintf("Service instance %s not found", destInstance))
		})

		It("Deletes the recipient service instance", func() {
			cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destPlan)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "10m", "1s").Should(gexec.Exit(1))
			test_helpers.WaitForService(destInstance, fmt.Sprintf("Service instance %s not found", destInstance))
		})

		Context("When --no-cleanup flag is specified", func() {
			var (
				destinationGUID       string
				renamedSourceInstance string
			)
			AfterEach(func() {
				srcGUID := test_helpers.InstanceUUID(sourceInstance)
				test_helpers.UnbindAllAppsFromService(srcGUID)
				test_helpers.UnbindAllAppsFromService(destinationGUID)
				test_helpers.DeleteService(renamedSourceInstance)
				test_helpers.WaitForService(renamedSourceInstance, fmt.Sprintf("Service instance %s not found", renamedSourceInstance))
			})

			It("Does not delete the recipient service instance when the --no-cleanup flag is specified", func() {
				renamedSourceInstance = sourceInstance + "-old"
				cmd := exec.Command("cf", "mysql-tools", "migrate", "--no-cleanup", sourceInstance, destPlan)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, "10m", "1s").Should(gexec.Exit(1))
				test_helpers.WaitForService(destInstance, `[Ss]tatus:\s+update succeeded`)

				destinationGUID = test_helpers.InstanceUUID(destInstance)

				appGUIDs := test_helpers.BoundAppGUIDs(destinationGUID)
				Expect(appGUIDs).NotTo(BeEmpty())
			})
		})
	})
})

func createInvalidMigrationState(sourceInstance string) {
	appName := generator.PrefixedRandomName("MYSQL", "INVALID_MIGRATION")
	sourceServiceKey := generator.PrefixedRandomName("MYSQL", "SERVICE_KEY")

	test_helpers.PushApp(appName, "assets/spring-music")
	test_helpers.BindAppToService(appName, sourceInstance)
	defer func() {
		test_helpers.DeleteApp(appName)
		test_helpers.AssertAppIsDeleted(appName)
	}()

	test_helpers.StartApp(appName)

	serviceKeyCreds := test_helpers.GetServiceKey(sourceInstance, sourceServiceKey)
	defer test_helpers.DeleteServiceKey(sourceInstance, sourceServiceKey)

	closeTunnel := test_helpers.OpenDatabaseTunnelToApp(63308, appName, serviceKeyCreds)
	defer closeTunnel()

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:63308)/%s",
		serviceKeyCreds.Username,
		serviceKeyCreds.Password,
		serviceKeyCreds.Name,
	)
	db, err := sql.Open("mysql", dsn)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()

	_, err = db.Exec("CREATE TABLE migrate_fail (id VARCHAR(1))")
	Expect(err).NotTo(HaveOccurred())

	_, err = db.Exec("CREATE VIEW migrate_fail_view AS SELECT * FROM migrate_fail")
	Expect(err).NotTo(HaveOccurred())

	_, err = db.Exec("DROP TABLE migrate_fail")
	Expect(err).NotTo(HaveOccurred())
}
