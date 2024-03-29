package multisite

import (
	"os"
	"testing"
	"time"

	"github.com/cloudfoundry/cf-test-helpers/v2/config"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/internal/testing/helpers"
)

var (
	leaderConfig      config.Config
	followerConfig    config.Config
	leaderTestSetup   *workflowhelpers.ReproducibleTestSuiteSetup
	followerTestSetup *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestMultisite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multisite System Tests", Label("multisite"))
}

var _ = BeforeSuite(func() {
	requiredEnvs := []string{
		"DC1_CONFIG",
		"DC2_CONFIG",
		"LEADER_PLAN_NAME",
		"FOLLOWER_PLAN_NAME",
		"SERVICE_NAME",
	}
	helpers.CheckForRequiredEnvVars(requiredEnvs)

	Expect(config.Load(os.Getenv("DC1_CONFIG"), &leaderConfig)).To(Succeed())
	leaderTestSetup = workflowhelpers.NewTestSuiteSetup(&leaderConfig)
	leaderTestSetup.Setup()
	leaderConfig.UseExistingOrganization = true
	leaderConfig.ExistingOrganization = leaderTestSetup.GetOrganizationName()
	leaderConfig.UseExistingSpace = true
	leaderConfig.ExistingSpace = leaderTestSetup.TestSpace.SpaceName()
	leaderConfig.UseExistingUser = true
	DeferCleanup(func() {
		workflowhelpers.AsUser(leaderTestSetup.AdminUserContext(), time.Hour, func() {
			leaderTestSetup.Teardown()
		})
	})

	Expect(config.Load(os.Getenv("DC2_CONFIG"), &followerConfig)).To(Succeed())
	followerTestSetup = workflowhelpers.NewTestSuiteSetup(&followerConfig)
	followerTestSetup.Setup()
	followerConfig.UseExistingOrganization = true
	followerConfig.ExistingOrganization = followerTestSetup.GetOrganizationName()
	followerConfig.UseExistingSpace = true
	followerConfig.ExistingSpace = followerTestSetup.TestSpace.SpaceName()
	followerConfig.UseExistingUser = true
	DeferCleanup(func() {
		workflowhelpers.AsUser(followerTestSetup.AdminUserContext(), time.Hour, func() {
			followerTestSetup.Teardown()
		})
	})
})
