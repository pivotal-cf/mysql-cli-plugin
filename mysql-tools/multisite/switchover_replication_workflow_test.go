package multisite_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	fakes2 "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/fakes"
)

var _ = Describe("SwitchoverReplication", func() {
	var (
		workflow        multisite.Workflow
		operations      []string
		fakeFoundation1 *fakes2.FakeFoundation
		fakeFoundation2 *fakes2.FakeFoundation
	)

	BeforeEach(func() {
		operations = nil
		fakeFoundation1 = &fakes2.FakeFoundation{FoundationName: "foundation1", Operations: &operations}
		fakeFoundation2 = &fakes2.FakeFoundation{FoundationName: "foundation2", Operations: &operations}
		logger := &fakes2.FakeLogger{Operations: &operations}

		workflow = multisite.NewWorkflow(fakeFoundation1, fakeFoundation2, logger)

		// HappyPath
		// Use random UUID suffixes to verify implementation isn't hardcoding any values
		fakeFoundation1.CreateHostInfoKeyResult.Key = "foundation1-host-info." + uuid.NewString()
		fakeFoundation2.CreateCredentialsKeyResult.Key = "foundation2-cred-info." + uuid.NewString()
	})

	It("can be configured with a real logger", func() {
		realLogger := log.New(GinkgoWriter, "[prefix]", log.LstdFlags)
		var buffer bytes.Buffer
		GinkgoWriter.TeeTo(&buffer)
		workflow = multisite.NewWorkflow(fakeFoundation1, fakeFoundation2, realLogger)

		_ = workflow.SwitchoverReplication("primary", "secondary")
		Expect(buffer.String()).To(ContainSubstring(`Checking whether instance 'primary' exists`),
			`Expected workflow to log messages to a real logger, but it did not.`)
	})

	It("works with single-node plans", func() {
		fakeFoundation1.PlanExistsResult.PlanExists = true
		fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"

		fakeFoundation2.PlanExistsResult.PlanExists = true
		fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"

		err := workflow.SwitchoverReplication("db0", "db1")
		Expect(err).NotTo(HaveOccurred())

		Expect(operations).To(Equal([]string{
			`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
			`foundation1.InstanceExists("db0")`,
			`foundation1.InstancePlanName("db0")`,
			`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
			`foundation2.PlanExists("single-node-plan")`,

			`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
			`foundation2.InstanceExists("db1")`,
			`foundation2.InstancePlanName("db1")`,
			`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
			`foundation1.PlanExists("single-node-plan")`,

			`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
			`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,

			`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
			`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,

			`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
			`foundation1.CreateHostInfoKey("db0")`,

			`logger.Printf("[foundation2] Registering secondary instance information on new primary instance 'db1'")`,
			fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q, <nil>)`, fakeFoundation1.CreateHostInfoKeyResult.Key),

			`logger.Printf("[foundation2] Retrieving replication configuration from new primary instance 'db1'")`,
			`foundation2.CreateCredentialsKey("db1")`,
			`logger.Printf("[foundation1] Updating new secondary instance 'db0' with replication configuration")`,
			fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q, <nil>)`, fakeFoundation2.CreateCredentialsKeyResult.Key),
			`logger.Printf("Successfully switched replication roles. primary = [foundation2] db1, secondary = [foundation1] db0")`,
		}))
	})

	It("works with HA leader and single node follower plans", func() {
		followerPlanName := "single-node-plan"
		leaderPlanName := "HA-plan"

		fakeFoundation2.PlanExistsResult.PlanExists = true
		fakeFoundation2.InstancePlanNameResult.PlanName = followerPlanName

		fakeFoundation1.PlanExistsResult.PlanExists = true
		fakeFoundation1.InstancePlanNameResult.PlanName = leaderPlanName

		err := workflow.SwitchoverReplication("db0", "db1")
		Expect(err).NotTo(HaveOccurred())

		expectedOperations := []string{
			`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
			`foundation1.InstanceExists("db0")`,

			`foundation1.InstancePlanName("db0")`,
			fmt.Sprintf(`logger.Printf("[foundation2] Checking whether plan '%s' exists")`, leaderPlanName),
			fmt.Sprintf(`foundation2.PlanExists("%s")`, leaderPlanName),

			`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
			`foundation2.InstanceExists("db1")`,

			`foundation2.InstancePlanName("db1")`,
			fmt.Sprintf(`logger.Printf("[foundation1] Checking whether plan '%s' exists")`, followerPlanName),
			fmt.Sprintf(`foundation1.PlanExists("%s")`, followerPlanName),

			`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
			fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", %v)`, followerPlanName),

			`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
			fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", %v)`, leaderPlanName),

			`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
			`foundation1.CreateHostInfoKey("db0")`,

			`logger.Printf("[foundation2] Registering secondary instance information on new primary instance 'db1'")`,
			fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q, <nil>)`, fakeFoundation1.CreateHostInfoKeyResult.Key),

			`logger.Printf("[foundation2] Retrieving replication configuration from new primary instance 'db1'")`,
			`foundation2.CreateCredentialsKey("db1")`,
			`logger.Printf("[foundation1] Updating new secondary instance 'db0' with replication configuration")`,
			fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q, <nil>)`, fakeFoundation2.CreateCredentialsKeyResult.Key),
			`logger.Printf("Successfully switched replication roles. primary = [foundation2] db1, secondary = [foundation1] db0")`,
		}

		for idx, value := range expectedOperations {
			By(fmt.Sprintf("expecting step %d: %s", idx, value))
			Expect(value).To(Equal(operations[idx]))
		}
		Expect(len(operations)).To(Equal(len(expectedOperations)))
	})

	When("the primary instance does not exist", func() {
		It("returns an error", func() {
			fakeFoundation1.InstanceExistsResult.Err = fmt.Errorf("primary instance does not exist error")

			err := workflow.SwitchoverReplication("primaryInstance", "secondaryInstance")
			Expect(err).To(MatchError(`primary instance does not exist error`))
		})

		It("does not execute the rest of the workflow", func() {
			fakeFoundation1.InstanceExistsResult.Err = fmt.Errorf("primary instance does not exist")

			_ = workflow.SwitchoverReplication("primaryInstance", "secondaryInstance")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'primaryInstance' exists")`,
				`foundation1.InstanceExists("primaryInstance")`,
			}))
		})
	})

	When("the secondary instance does not exist", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstanceExistsResult.Err = fmt.Errorf("secondary instance does not exist error")
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError(`secondary instance does not exist error`))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
			}))
		})
	})

	When("Demoting the original primary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true
			fakeFoundation1.UpdateServiceResult.Err = fmt.Errorf("some demotion error")
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("some demotion error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,
			}))
		})
	})

	When("Promoting the original secondary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true
			fakeFoundation2.UpdateServiceResult.Err = fmt.Errorf("some promotion error")
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("some promotion error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,

				`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
				`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,
			}))
		})
	})

	When("creating a host-info key on the original primary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true
			fakeFoundation1.CreateHostInfoKeyResult.Err = fmt.Errorf("[db0] create host info key error")
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("[db0] create host info key error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,
				`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
				`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,
				`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
				`foundation1.CreateHostInfoKey("db0")`,
			}))
		})
	})

	When("registering the new follower on the new primary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true
			fakeFoundation2.UpdateServiceResult.ErrFunc = func(instanceName, arbitraryParams string) error {
				if strings.Contains(arbitraryParams, `foundation1-host-info`) {
					return fmt.Errorf("some registration error on primary instance")
				}
				return nil
			}
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("some registration error on primary instance"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,
				`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
				`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,
				`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
				`foundation1.CreateHostInfoKey("db0")`,
				`logger.Printf("[foundation2] Registering secondary instance information on new primary instance 'db1'")`,
				fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q, <nil>)`, fakeFoundation1.CreateHostInfoKeyResult.Key),
			}))
		})
	})

	When("retrieving replication credentials from the new primary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true
			fakeFoundation2.CreateCredentialsKeyResult.Err = fmt.Errorf("create credentials service key error")
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("create credentials service key error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,
				`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
				`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,
				`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
				`foundation1.CreateHostInfoKey("db0")`,
				`logger.Printf("[foundation2] Registering secondary instance information on new primary instance 'db1'")`,
				fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q, <nil>)`, fakeFoundation1.CreateHostInfoKeyResult.Key),
				`logger.Printf("[foundation2] Retrieving replication configuration from new primary instance 'db1'")`,
				`foundation2.CreateCredentialsKey("db1")`,
			}))
		})
	})

	When("updating the new secondary instance with replication credentials fails ", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation1.PlanExistsResult.PlanExists = true

			fakeFoundation1.UpdateServiceResult.ErrFunc = func(instanceName, arbitraryParams string) error {
				if strings.Contains(arbitraryParams, "foundation2-cred-info") {
					return fmt.Errorf("some update-service w/ replication credentials error")
				}
				return nil
			}
		})

		It("returns an error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(MatchError("some update-service w/ replication credentials error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SwitchoverReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`foundation1.InstancePlanName("db0")`,
				`logger.Printf("[foundation2] Checking whether plan 'single-node-plan' exists")`,
				`foundation2.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`foundation2.InstancePlanName("db1")`,
				`logger.Printf("[foundation1] Checking whether plan 'single-node-plan' exists")`,
				`foundation1.PlanExists("single-node-plan")`,

				`logger.Printf("[foundation1] Demoting primary instance 'db0'")`,
				`foundation1.UpdateServiceAndWait("db0", "{ \"initiate-failover\": \"make-leader-read-only\" }", single-node-plan)`,
				`logger.Printf("[foundation2] Promoting secondary instance 'db1'")`,
				`foundation2.UpdateServiceAndWait("db1", "{ \"initiate-failover\": \"promote-follower-to-leader\" }", single-node-plan)`,
				`logger.Printf("[foundation1] Retrieving information for new secondary instance 'db0'")`,
				`foundation1.CreateHostInfoKey("db0")`,
				`logger.Printf("[foundation2] Registering secondary instance information on new primary instance 'db1'")`,
				fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q, <nil>)`, fakeFoundation1.CreateHostInfoKeyResult.Key),
				`logger.Printf("[foundation2] Retrieving replication configuration from new primary instance 'db1'")`,
				`foundation2.CreateCredentialsKey("db1")`,
				`logger.Printf("[foundation1] Updating new secondary instance 'db0' with replication configuration")`,
				fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q, <nil>)`, fakeFoundation2.CreateCredentialsKeyResult.Key),
			}))
		})
	})

	When("an unexpected error occurs while checking the current leader plan", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.Err = fmt.Errorf("unexpected CF error")
		})
		It("returns the error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("unexpected CF error"))
		})
	})

	When("an unexpected error occurs when fetching the primary plan name", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.Err = fmt.Errorf("unexpected CF error")
		})
		It("returns the error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")

			Expect(err).To(MatchError(`unexpected CF error`))
		})
	})

	When("an error occurs while checking the primary plan exists in the secondary foundation", func() {
		var planName = "plan-name"
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = planName
			fakeFoundation2.PlanExistsResult.Err = fmt.Errorf("unexpected error checking plans")
		})
		It("returns a descriptive error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`unexpected error checking plans`))
		})
	})

	When("an unexpected error occurs while fetching the follower's plan name", func() {
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = "single-node-plan"
			fakeFoundation2.PlanExistsResult.PlanExists = true
			fakeFoundation2.InstancePlanNameResult.Err = fmt.Errorf("unexpected CF error")
		})
		It("returns a descriptive error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")

			Expect(err).To(MatchError(`unexpected CF error`))
		})
	})

	When("an error occurs while checking the secondary plan exists in the primary foundation", func() {
		var planName = "plan-name"
		BeforeEach(func() {
			fakeFoundation1.InstancePlanNameResult.PlanName = planName
			fakeFoundation2.InstancePlanNameResult.PlanName = planName
			fakeFoundation2.PlanExistsResult.Err = fmt.Errorf("unexpected CF error")
		})
		It("returns a descriptive error", func() {
			err := workflow.SwitchoverReplication("db0", "db1")

			Expect(err).To(MatchError(`unexpected CF error`))
		})
	})
})
