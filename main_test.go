package main_test

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MysqlV2CliPlugin", func() {

	var (
		donorService   string
		restoreService string
	)

	BeforeEach(func() {
		donorService = generator.PrefixedRandomName("MYSQL", "DED")
		test_helpers.CreateService(os.Getenv("SERVICE_NAME"), os.Getenv("PLAN_NAME"), donorService)

		restoreService = generator.PrefixedRandomName("MYSQL", "DED")
		test_helpers.CreateService(os.Getenv("SERVICE_NAME"), os.Getenv("PLAN_NAME"), restoreService)

		test_helpers.WaitForService(donorService, "status:    create succeeded")
		test_helpers.WaitForService(restoreService, "status:    create succeeded")
	})

	AfterEach(func() {
		test_helpers.DeleteService(donorService)
		test_helpers.DeleteService(restoreService)
		test_helpers.WaitForService(donorService, fmt.Sprintf("Service instance %s not found", donorService))
		test_helpers.WaitForService(restoreService, fmt.Sprintf("Service instance %s not found", restoreService))
	})

	It("migrates the contents of one database to another", func() {
		By("writing some information to the donor instance")
		donorDeploymentName := test_helpers.GetDeploymentName(donorService)
		restoreDeploymentName := test_helpers.GetDeploymentName(restoreService)
		writeStmt := `CREATE TABLE service_instance_db.ketchup(num INT PRIMARY KEY); INSERT INTO service_instance_db.ketchup values(1),(2),(3);`
		test_helpers.ExecuteMysqlQueryAsAdmin(donorDeploymentName, "0", writeStmt)

		By("running the plugin")
		test_helpers.ExecuteCfCmd("mysql-migrate", donorService, restoreService)

		By("seeing the data on the restore service")
		dataOnFollower := test_helpers.ExecuteMysqlQueryAsAdmin(restoreDeploymentName, "0", "SELECT num FROM service_instance_db.ketchup")
		Expect(dataOnFollower).To(ContainSubstring("1"))
		Expect(dataOnFollower).To(ContainSubstring("2"))
		Expect(dataOnFollower).To(ContainSubstring("3"))
	})
})
