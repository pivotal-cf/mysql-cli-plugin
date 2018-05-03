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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
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

var _ = Describe("Replace Integration Tests", func() {
	var (
		albumID    string
		appDomain  string
		appName    string
		appURI     string
		readValue  string
		writeValue string
	)

	var (
		destInstance          string
		sourceInstance        string
		renamedSourceInstance string
	)

	BeforeEach(func() {
		appDomain = os.Getenv("APP_DOMAIN")

		sourceInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_SOURCE")
		renamedSourceInstance = sourceInstance + "-old"
		test_helpers.CreateService(os.Getenv("DONOR_SERVICE_NAME"), os.Getenv("DONOR_PLAN_NAME"), sourceInstance)

		destInstance = generator.PrefixedRandomName("MYSQL", "MIGRATE_DEST")
		test_helpers.CreateService(os.Getenv("RECIPIENT_SERVICE_NAME"), os.Getenv("RECIPIENT_PLAN_NAME"), destInstance)

		test_helpers.WaitForService(sourceInstance, `[Ss]tatus:\s+create succeeded`)
		test_helpers.WaitForService(destInstance, `[Ss]tatus:\s+create succeeded`)

		writeValue = "DM Greatest Hits"
		appName = generator.PrefixedRandomName("MYSQL", "APP")
		appURI = appName + "." + appDomain

		test_helpers.PushApp(appName, "assets/spring-music")
		test_helpers.BindAppToService(appName, sourceInstance)
		test_helpers.StartApp(appName)

		test_helpers.CheckAppInfo(true, appURI, sourceInstance)

		albumID = test_helpers.WriteData(true, appURI, writeValue)
		readValue = test_helpers.ReadData(true, appURI, albumID)
		Expect(readValue).To(Equal(writeValue))

		test_helpers.UnbindAppFromService(appName, sourceInstance)
	})

	AfterEach(func() {
		if appName != "" {
			test_helpers.DeleteApp(appName)
		}

		test_helpers.DeleteService(sourceInstance)
		test_helpers.DeleteService(renamedSourceInstance)

		test_helpers.WaitForService(sourceInstance, fmt.Sprintf("Service instance %s not found", sourceInstance))
		test_helpers.WaitForService(renamedSourceInstance, fmt.Sprintf("Service instance %s not found", renamedSourceInstance))
	})

	It("Migrates data from donor to recipient, then renames the instances", func() {
		var (
			sourceInstanceGUID string
			destInstanceGUID   string
		)
		By("Obtaining the GUIDs for the original source and destination instance", func() {
			sourceInstanceGUID = test_helpers.InstanceUUID(sourceInstance)
			destInstanceGUID = test_helpers.InstanceUUID(destInstance)
		})

		By("Migrating data by using the replace command", func() {
			cmd := exec.Command("cf", "mysql-tools", "replace", sourceInstance, destInstance)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "10m", "1s").Should(gexec.Exit(0))
		})

		By("Verifying the destination service was renamed to the source's name", func() {
			renamedSourceInstanceGUID := test_helpers.InstanceUUID(sourceInstance)
			Expect(renamedSourceInstanceGUID).To(Equal(destInstanceGUID))
		})

		By("Binding the app to the renamed destination instance and reading back data", func() {
			test_helpers.BindAppToService(appName, sourceInstance)
			test_helpers.ExecuteCfCmd("restage", appName)

			readValue = test_helpers.ReadData(true, appURI, albumID)
			Expect(readValue).To(Equal(writeValue))
		})

		By("Verifying that the credhub reference in the binding only contains the destination service's GUID", func() {
			appGUID := strings.TrimSpace(test_helpers.ExecuteCfCmd("app", appName, "--guid"))

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
	})
})
