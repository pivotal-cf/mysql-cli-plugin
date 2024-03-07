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

package multisite

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
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

var _ = Describe("Multisite Setup Integration Tests", Ordered, Label("multisite"), func() {
	// TODO: Let's name these differently.
	//       "leader" and "follower" are fluid roles.  Let's just call these db0 / db1 or similar to reduce cognitive load
	//       db0 may be a leader or follower at any given moment, but the name doesn't need to denote the role
	var (
		leaderInstanceName   string
		followerInstanceName string
		leaderPlanName       string
		followerPlanName     string
	)

	BeforeAll(func() {
		leaderPlanName = os.Getenv("LEADER_PLAN_NAME")
		followerPlanName = os.Getenv("FOLLOWER_PLAN_NAME")
		serviceName := os.Getenv("SERVICE_NAME")
		By("initiating leader instance creation")
		workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
			leaderInstanceName = generator.PrefixedRandomName("MYSQL", "MS_LEADER")
			test_helpers.CreateService(serviceName, leaderPlanName, leaderInstanceName, "-c", `{"enable_external_access": true}`)
		})

		By("initiating follower instance creation")
		workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
			followerInstanceName = generator.PrefixedRandomName("MYSQL", "MS_FOLLOWER")
			test_helpers.CreateService(serviceName, followerPlanName, followerInstanceName, "-c", `{"enable_external_access": true}`)
		})

		By("waiting for leader instance creation")
		leaderTestSetup.Setup()
		workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.WaitForService(leaderInstanceName, `[Ss]tatus:\s+create succeeded`)
		})

		By("waiting for follower instance creation")
		workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.WaitForService(followerInstanceName, `[Ss]tatus:\s+create succeeded`)
		})
	})

	AfterAll(func() {
		workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.DeleteService(leaderInstanceName)
		})

		workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.DeleteService(followerInstanceName)
		})

		workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.WaitForService(leaderInstanceName, "Service instance .* not found")
		})

		workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
			test_helpers.WaitForService(followerInstanceName, "Service instance .* not found")
		})
	})

	It("can target multi-foundations through the targeting subcommands", func() {
		By("logging into the leader foundation and saving a target config")
		exitCode := RunCfInFoundation(leaderTestSetup.RegularUserContext(),
			"mysql-tools", "save-target", leaderFoundationHandle)
		Expect(exitCode).To(Equal(0), `Failed to save target in leader foundation!`)

		By("logging into the follower foundation and saving a target config")
		exitCode = RunCfInFoundation(followerTestSetup.RegularUserContext(),
			"mysql-tools", "save-target", followerFoundationHandle)
		Expect(exitCode).To(Equal(0), `Failed to save target in follower foundation!`)

	})

	It("configure the initial replication channel between two instances in two foundations", func() {
		session := cf.Cf("mysql-tools", "setup-replication",
			fmt.Sprintf("--primary-target=%s", leaderFoundationHandle),
			fmt.Sprintf("--primary-instance=%s", leaderInstanceName),
			fmt.Sprintf("--secondary-target=%s", followerFoundationHandle),
			fmt.Sprintf("--secondary-instance=%s", followerInstanceName),
		)

		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[leader-foundation\] Checking whether instance '%s' exists`, leaderInstanceName))
		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[follower-foundation\] Checking whether instance '%s' exists`, followerInstanceName))
		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[follower-foundation\] Retrieving information for secondary instance '%s'`, followerInstanceName))
		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[leader-foundation\] Registering secondary instance information on primary instance '%s'`, leaderInstanceName))
		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[leader-foundation\] Retrieving replication configuration from primary instance '%s'`, leaderInstanceName))
		Eventually(session.Out, "15m", "10s").Should(
			gbytes.Say(`\[follower-foundation\] Updating secondary instance '%s' with replication configuration`, followerInstanceName))

		Eventually(session, "15m", "10s").Should(gexec.Exit(0))

		if leaderPlanName == "multisite-80" {

			By("Validating the primary is not configured as a follower", func() {
				workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(leaderInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_follower_is_follower")
					Expect(isFollowerMetric).To(Equal("0"))
				})

			})
		} else if leaderPlanName == "ha-80" {
			By("Validating the primary is configured as HA", func() {

				workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(leaderInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_galera_wsrep_ready")
					Expect(isFollowerMetric).To(Equal("1"))
				})

				workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(leaderInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_galera_wsrep_cluster_size")
					Expect(isFollowerMetric).To(Equal("3"))
				})
			})
		}

		By("Validating the secondary is configured as a follower", func() {
			workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
				leaderGUID := test_helpers.InstanceUUID(followerInstanceName)
				isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_follower_is_follower")
				Expect(isFollowerMetric).To(Equal("1"))
			})

		})
	})

	It("can switch the replication roles between two foundations", func() {
		session := cf.Cf("mysql-tools", "switchover",
			fmt.Sprintf("--primary-target=%s", leaderFoundationHandle),
			fmt.Sprintf("--primary-instance=%s", leaderInstanceName),
			fmt.Sprintf("--secondary-target=%s", followerFoundationHandle),
			fmt.Sprintf("--secondary-instance=%s", followerInstanceName),
			"--force",
		)

		Eventually(session.Out, "1h", "10s").Should(
			gbytes.Say(`Successfully switched replication roles`))

		Eventually(session, "10m", "10s").Should(gexec.Exit(0))

		if leaderPlanName == "multisite-80" {

			By("Validating the new primary is not configured as a follower", func() {
				workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(followerInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_follower_is_follower")
					Expect(isFollowerMetric).To(Equal("0"))
				})

			})

		} else if leaderPlanName == "ha-80" {
			By("Validating the new primary is configured as HA", func() {

				workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(followerInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_galera_wsrep_ready")
					Expect(isFollowerMetric).To(Equal("1"))
				})

				workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
					leaderGUID := test_helpers.InstanceUUID(followerInstanceName)
					isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_galera_wsrep_cluster_size")
					Expect(isFollowerMetric).To(Equal("3"))
				})
			})
		}

		By("Validating the new secondary is configured as a follower", func() {
			workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
				leaderGUID := test_helpers.InstanceUUID(leaderInstanceName)
				isFollowerMetric := getMetricValue(leaderGUID, "_p_mysql_follower_is_follower")
				Expect(isFollowerMetric).To(Equal("1"))
			})

		})

		By("Validating the plans are switched as expected", func() {

			workflowhelpers.AsUser(leaderTestSetup.RegularUserContext(), 10*time.Minute, func() {
				planName := test_helpers.InstancePlanName(leaderInstanceName)
				Expect(planName).To(Equal(followerPlanName))
			})

			workflowhelpers.AsUser(followerTestSetup.RegularUserContext(), 10*time.Minute, func() {
				planName := test_helpers.InstancePlanName(followerInstanceName)
				Expect(planName).To(Equal(leaderPlanName))
			})
		})

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

func RunCfInFoundation(ctx workflowhelpers.UserContext, args ...string) (exitCode int) {
	ctx.Login()
	ctx.TargetSpace()
	defer ctx.Logout()

	return RunCf(args...)
}

func RunCf(args ...string) (exitcode int) {
	return cf.Cf(args...).Wait().ExitCode()
}
