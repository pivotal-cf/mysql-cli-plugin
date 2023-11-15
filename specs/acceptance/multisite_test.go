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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cf-test-helpers/v2/config"
	"github.com/cloudfoundry/cf-test-helpers/v2/generator"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

const leaderFoundationHandle = "leader-foundation"
const followerFoundationHandle = "follower-foundation"

var _ = Describe("Multisite Setup Integration Tests", Serial, Label("multisite"), func() {

	var leaderTestSetup, followerTestSetup *workflowhelpers.ReproducibleTestSuiteSetup
	var leaderConfig, followerConfig config.Config
	var leaderInstanceName, followerInstanceName string

	BeforeEach(func() {
		By("logging into the leader foundation")
		Expect(config.Load(os.Getenv("DC1_CONFIG"), &leaderConfig)).To(Succeed())
		leaderTestSetup = workflowhelpers.NewTestSuiteSetup(&leaderConfig)
		leaderTestSetup.Setup()
		leaderConfig.UseExistingOrganization = true
		leaderConfig.ExistingOrganization = leaderTestSetup.GetOrganizationName()
		leaderConfig.UseExistingSpace = true
		leaderConfig.ExistingSpace = leaderTestSetup.TestSpace.SpaceName()
		leaderConfig.UseExistingUser = true

		By("saving the leader foundation config")
		cmd := exec.Command("cf", "mysql-tools", "save-target", leaderFoundationHandle)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		By("logging into the follower foundation")
		Expect(config.Load(os.Getenv("DC2_CONFIG"), &followerConfig)).To(Succeed())
		followerTestSetup = workflowhelpers.NewTestSuiteSetup(&followerConfig)
		followerTestSetup.Setup()
		followerConfig.UseExistingOrganization = true
		followerConfig.ExistingOrganization = followerTestSetup.GetOrganizationName()
		followerConfig.UseExistingSpace = true
		followerConfig.ExistingSpace = followerTestSetup.TestSpace.SpaceName()
		followerConfig.UseExistingUser = true

		By("saving the follower foundation config")
		cmd = exec.Command("cf", "mysql-tools", "save-target", followerFoundationHandle)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})

	Context("valid service instances and configs", func() {
		BeforeEach(func() {
			By("initiating leader instance creation")
			leaderTestSetup.Setup()
			leaderInstanceName = generator.PrefixedRandomName("MYSQL", "MS_LEADER")
			test_helpers.CreateService(os.Getenv("SERVICE_NAME"), os.Getenv("SINGLE_NODE_PLAN_NAME"), leaderInstanceName, "-c", `{"enable_external_access": true}`)

			By("initiating follower instance creation")
			followerTestSetup.Setup()
			followerInstanceName = generator.PrefixedRandomName("MYSQL", "MS_FOLLOWER")
			test_helpers.CreateService(os.Getenv("SERVICE_NAME"), os.Getenv("SINGLE_NODE_PLAN_NAME"), followerInstanceName, "-c", `{"enable_external_access": true}`)

			By("waiting for leader instance creation")
			leaderTestSetup.Setup()
			test_helpers.WaitForService(leaderInstanceName, `[Ss]tatus:\s+create succeeded`)

			By("waiting for follower instance creation")
			followerTestSetup.Setup()
			test_helpers.WaitForService(followerInstanceName, `[Ss]tatus:\s+create succeeded`)

		})
		It("Can setup multisite replication between two foundations", func() {
			cmd := exec.Command("cf", "mysql-tools", "setup-replication",
				leaderFoundationHandle, leaderInstanceName,
				followerFoundationHandle, followerInstanceName)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Validating the primary instance: '%s'.\n", leaderInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Validating the secondary instance: '%s'.\n", followerInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Creating a 'host-info' service-key: 'MSHostInfo-.*' on the secondary instance: '%s'.\n", followerInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Getting the 'host-info' service-key from the secondary instance: '%s'.\n", followerInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Updating the primary with the secondary's 'host-info' service-key: 'MSHostInfo-.*'.\n"))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Creating a 'credentials' service-key: 'MSCredInfo-.*' on the primary instance: '%s'.\n", leaderInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Getting the 'credentials' service-key from the primary instance. '%s'.\n", leaderInstanceName))
			Eventually(session.Out, "15m", "10s").Should(
				gbytes.Say("Updating the secondary instance with the primary's 'credentials' service-key: 'MSCredInfo-.*'.\n"))

			Eventually(session, "15m", "10s").Should(gexec.Exit(0))
			Expect(err).NotTo(HaveOccurred())

			// Validate the secondary is configured as a follower
			followerTestSetup.Setup()
			followerGUID := test_helpers.InstanceUUID(followerInstanceName)
			isFollowerMetric := getMetricValue(followerGUID, "_p_mysql_follower_is_follower")
			Expect(isFollowerMetric).To(Equal("1"))

			// Validate the primary is not configured as a follower
			leaderTestSetup.Setup()
			leaderGUID := test_helpers.InstanceUUID(leaderInstanceName)
			isFollowerMetric = getMetricValue(leaderGUID, "_p_mysql_follower_is_follower")
			Expect(isFollowerMetric).To(Equal("0"))
		})
	})

	AfterEach(func() {
		if followerTestSetup != nil {
			followerTestSetup.Setup()
			test_helpers.DeleteService(followerInstanceName)
			test_helpers.WaitForService(followerInstanceName, "Service instance .* not found")
			followerTestSetup.Teardown()
		}
		if leaderTestSetup != nil {
			leaderTestSetup.Setup()
			test_helpers.DeleteService(leaderInstanceName)
			test_helpers.WaitForService(leaderInstanceName, "Service instance .* not found")
			leaderTestSetup.Teardown()
		}
	})
})

func getMetricValue(instanceGuid, metric string) string {
	type logCacheResult struct {
		Data struct {
			Result []struct {
				Value []any `json:"value"`
			} `json:"result"`
		}
	}
	var logMetric logCacheResult

	GinkgoHelper()
	metricQuery := fmt.Sprintf(`%s{source_id="%s",deployment="service-instance_%s"}`,
		metric, instanceGuid, instanceGuid)

	metricJson := test_helpers.ExecuteCfCmd("query", metricQuery)
	err := json.Unmarshal([]byte(metricJson), &logMetric)
	Expect(err).NotTo(HaveOccurred())

	Expect(len(logMetric.Data.Result)).Should(BeNumerically(">", 0))
	values := logMetric.Data.Result[0].Value
	Expect(len(values)).Should(Equal(2)) // values[0] is the timestamp

	return values[1].(string)
}
