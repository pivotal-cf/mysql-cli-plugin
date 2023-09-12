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

package acceptance

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cf-test-helpers/v2/generator"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

var _ = Describe("Migrate Integration Tests v2", func() {
	var (
		appDomain          string
		sinatraAppName     string
		destInstance       string
		destPlan           string
		sourceInstance     string
		sourceInstanceGUID string
		destInstanceGUID   string
	)

	BeforeEach(func() {
		destPlan = os.Getenv("RECIPIENT_PLAN_NAME")
	})

	Context("when a valid donor service instance exists", func() {
		BeforeEach(func() {
			appDomain = Config.AppsDomain
			sinatraAppName = generator.PrefixedRandomName("MYSQL", "SINATRA")
			test_helpers.PushApp(sinatraAppName, "../assets/sinatra-app")

			sourceInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_SOURCE")
			test_helpers.CreateService(os.Getenv("V2_DONOR_SERVICE_NAME"), os.Getenv("V2_DONOR_PLAN_NAME"), sourceInstance)
			destInstance = sourceInstance + "-new"

			test_helpers.WaitForService(sourceInstance, `[Ss]tatus:\s+create succeeded`)
			sourceInstanceGUID = test_helpers.InstanceUUID(sourceInstance)
		})

		AfterEach(func() {
			if sinatraAppName != "" {
				test_helpers.DeleteApp(sinatraAppName)
			}
			oldInstance := sourceInstance + "-old"
			test_helpers.DeleteService(destInstance)
			test_helpers.DeleteService(sourceInstance)
			test_helpers.DeleteService(oldInstance)
			test_helpers.WaitForService(destInstance, "Service instance .* not found")
			test_helpers.WaitForService(sourceInstance, "Service instance .* not found")
			test_helpers.WaitForService(oldInstance, "Service instance .* not found")
		})

		It("migrates data from donor to recipient", func() {
			var (
				sinatraAppURI string
				allDb         map[string]string
			)

			By("Binding an app to the source instance", func() {
				test_helpers.BindAppToService(sinatraAppName, sourceInstance)
				test_helpers.StartApp(sinatraAppName)
			})

			By("Writing data to the source instance", func() {
				sinatraAppURI = sinatraAppName + "." + appDomain

				createdDb := test_helpers.CreateDb(true, sinatraAppURI)
				allDb = test_helpers.ReadDb(true, sinatraAppURI)

				createdDbName := createdDb["db"]
				createdDbValue := createdDb["value"]
				Expect(createdDbName).ToNot(BeEmpty())
				Expect(createdDbValue).ToNot(BeEmpty())

				Expect(allDb[createdDbName]).To(Equal(createdDbValue))
				test_helpers.UnbindAppFromService(sinatraAppName, sourceInstance)
			})

			By("Migrating data using the migrate command", func() {
				cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destPlan)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, migrationTimeout, "1s").Should(gexec.Exit(0))
			})

			By("Verifying the destination service was renamed to the source's name", func() {
				destInstanceGUID = test_helpers.InstanceUUID(sourceInstance)
				Expect(destInstanceGUID).NotTo(Equal(sourceInstanceGUID))
			})

			By("Binding the app to the newly created destination instance and reading back data", func() {
				test_helpers.BindAppToService(sinatraAppName, sourceInstance)
				test_helpers.ExecuteCfCmd("restage", sinatraAppName)

				dbAfterMigrate := test_helpers.ReadDb(true, sinatraAppURI)

				Expect(allDb).To(Equal(dbAfterMigrate))
			})
		})

		Context("DB names with the containing substrings of the filtered grep will be migrated", func() {
			var testDbName string

			BeforeEach(func() {
				testDbName = "sysblah"
				createTestDb(sourceInstance, testDbName)
			})

			It("transfers all DBs", func() {
				cmd := exec.Command("cf", "mysql-tools", "migrate", sourceInstance, destPlan)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, migrationTimeout, "1s").Should(gexec.Exit(0))

				dbExists, err := dbExists(sourceInstance, testDbName)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbExists).To(BeTrue())
			})
		})
	})
})

func createTestDb(sourceInstance, dbName string) {
	var err error

	appName := generator.PrefixedRandomName("MYSQL", "MIGRATION_TEST")
	sourceServiceKey := generator.PrefixedRandomName("MYSQL", "SERVICE_KEY")

	test_helpers.PushApp(appName, "../assets/spring-music")
	test_helpers.BindAppToService(appName, sourceInstance)
	defer func() {
		test_helpers.DeleteApp(appName)
		Expect(test_helpers.AssertAppIsDeleted(appName)).ToNot(HaveOccurred())
	}()

	test_helpers.StartApp(appName)

	serviceKeyCreds := test_helpers.GetServiceKey(sourceInstance, sourceServiceKey)
	defer test_helpers.DeleteServiceKey(sourceInstance, sourceServiceKey)

	db, closeTunnel := test_helpers.OpenDatabaseTunnelToApp(appName, serviceKeyCreds)
	defer closeTunnel()
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	Expect(err).NotTo(HaveOccurred())
}

func dbExists(sourceInstance, dbName string) (bool, error) {
	appName := generator.PrefixedRandomName("MYSQL", "INVALID_MIGRATION")
	sourceServiceKey := generator.PrefixedRandomName("MYSQL", "SERVICE_KEY")

	test_helpers.PushApp(appName, "../assets/spring-music")
	test_helpers.BindAppToService(appName, sourceInstance)
	defer func() {
		test_helpers.DeleteApp(appName)
		Expect(test_helpers.AssertAppIsDeleted(appName)).ToNot(HaveOccurred())
	}()

	test_helpers.StartApp(appName)

	serviceKeyCreds := test_helpers.GetServiceKey(sourceInstance, sourceServiceKey)
	defer test_helpers.DeleteServiceKey(sourceInstance, sourceServiceKey)

	db, closeTunnel := test_helpers.OpenDatabaseTunnelToApp(appName, serviceKeyCreds)
	defer closeTunnel()
	defer db.Close()

	result, err := db.Query(fmt.Sprintf("SHOW DATABASES LIKE '%s'", dbName))
	if err != nil {
		return false, err
	}

	return result.Next(), nil
}
