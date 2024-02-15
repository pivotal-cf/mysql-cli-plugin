package multisite_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	fakes2 "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/fakes"
)

var _ = Describe("SetupReplication", func() {
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
	})

	BeforeEach(func() {
		// HappyPath
		// Use random UUID suffixes to verify implementation isn't hardcoding any values
		fakeFoundation2.CreateHostInfoKeyResult.Key = "foundation2-host-info." + uuid.NewString()
		fakeFoundation1.CreateCredentialsKeyResult.Key = "foundation1-cred-info." + uuid.NewString()
	})

	It("can be configured with a real logger", func() {
		realLogger := log.New(GinkgoWriter, "[prefix]", log.LstdFlags)
		var buffer bytes.Buffer
		GinkgoWriter.TeeTo(&buffer)
		workflow = multisite.NewWorkflow(fakeFoundation1, fakeFoundation2, realLogger)

		_ = workflow.SetupReplication("primary", "secondary")
		Expect(buffer.String()).To(ContainSubstring(`Checking whether instance 'primary' exists`),
			`Expected workflow to log messages to a real logger, but it did not.`)
	})

	It("works", func() {
		err := workflow.SetupReplication("primaryInstance", "secondaryInstance")
		Expect(err).NotTo(HaveOccurred())

		Expect(operations).To(Equal([]string{
			`logger.Printf("[foundation1] Checking whether instance 'primaryInstance' exists")`,
			`foundation1.InstanceExists("primaryInstance")`,
			`logger.Printf("[foundation2] Checking whether instance 'secondaryInstance' exists")`,
			`foundation2.InstanceExists("secondaryInstance")`,
			`logger.Printf("[foundation2] Retrieving information for secondary instance 'secondaryInstance'")`,
			`foundation2.CreateHostInfoKey("secondaryInstance")`,
			`logger.Printf("[foundation1] Registering secondary instance information on primary instance 'primaryInstance'")`,
			fmt.Sprintf(`foundation1.UpdateServiceAndWait("primaryInstance", %q)`, fakeFoundation2.CreateHostInfoKeyResult.Key),
			`logger.Printf("[foundation1] Retrieving replication configuration from primary instance 'primaryInstance'")`,
			`foundation1.CreateCredentialsKey("primaryInstance")`,
			`logger.Printf("[foundation2] Updating secondary instance 'secondaryInstance' with replication configuration")`,
			fmt.Sprintf(`foundation2.UpdateServiceAndWait("secondaryInstance", %q)`, fakeFoundation1.CreateCredentialsKeyResult.Key),
			`logger.Printf("Successfully configured replication")`,
		}))
	})

	When("the primary instance does not exist", func() {
		It("returns an error", func() {
			fakeFoundation1.InstanceExistsResult.Err = fmt.Errorf("primary instance does not exist error")

			err := workflow.SetupReplication("primaryInstance", "secondaryInstance")
			Expect(err).To(MatchError(`primary instance does not exist error`))
		})

		It("does not execute the rest of the workflow", func() {
			fakeFoundation1.InstanceExistsResult.Err = fmt.Errorf("primary instance does not exist")

			_ = workflow.SetupReplication("primaryInstance", "secondaryInstance")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'primaryInstance' exists")`,
				`foundation1.InstanceExists("primaryInstance")`,
			}))
		})
	})

	When("the secondary instance does not exist", func() {
		BeforeEach(func() {
			fakeFoundation2.InstanceExistsResult.Err = fmt.Errorf("secondary instance does not exist error")
		})

		It("returns an error", func() {
			err := workflow.SetupReplication("db0", "db1")
			Expect(err).To(MatchError(`secondary instance does not exist error`))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SetupReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
			}))
		})
	})

	When("creating a host-info key fails", func() {
		BeforeEach(func() {
			fakeFoundation2.CreateHostInfoKeyResult.Err = fmt.Errorf("create host info key error")
		})

		It("returns an error", func() {
			err := workflow.SetupReplication("db0", "db1")
			Expect(err).To(MatchError("create host info key error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SetupReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`logger.Printf("[foundation2] Retrieving information for secondary instance 'db1'")`,
				`foundation2.CreateHostInfoKey("db1")`,
			}))
		})
	})

	When("updating the primary instance fails", func() {
		BeforeEach(func() {
			fakeFoundation1.UpdateServiceResult.Err = fmt.Errorf("update primary instance error")
		})

		It("returns an error", func() {
			err := workflow.SetupReplication("db0", "db1")
			Expect(err).To(MatchError("update primary instance error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SetupReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`logger.Printf("[foundation2] Retrieving information for secondary instance 'db1'")`,
				`foundation2.CreateHostInfoKey("db1")`,
				`logger.Printf("[foundation1] Registering secondary instance information on primary instance 'db0'")`,
				fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q)`, fakeFoundation2.CreateHostInfoKeyResult.Key),
			}))
		})
	})

	When("creating replication credentials on the primary fails", func() {
		BeforeEach(func() {
			fakeFoundation1.CreateCredentialsKeyResult.Err = fmt.Errorf("create credentials service key error")
		})

		It("returns an error", func() {
			err := workflow.SetupReplication("db0", "db1")
			Expect(err).To(MatchError("create credentials service key error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SetupReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`logger.Printf("[foundation2] Retrieving information for secondary instance 'db1'")`,
				`foundation2.CreateHostInfoKey("db1")`,
				`logger.Printf("[foundation1] Registering secondary instance information on primary instance 'db0'")`,
				fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q)`, fakeFoundation2.CreateHostInfoKeyResult.Key),
				`logger.Printf("[foundation1] Retrieving replication configuration from primary instance 'db0'")`,
				`foundation1.CreateCredentialsKey("db0")`,
			}))
		})
	})

	When("update the secondary instance fails ", func() {
		BeforeEach(func() {
			fakeFoundation2.UpdateServiceResult.Err = fmt.Errorf("update secondary service instance error")
		})

		It("returns an error", func() {
			err := workflow.SetupReplication("db0", "db1")
			Expect(err).To(MatchError("update secondary service instance error"))
		})

		It("does not execute the rest of the workflow", func() {
			_ = workflow.SetupReplication("db0", "db1")

			Expect(operations).To(Equal([]string{
				`logger.Printf("[foundation1] Checking whether instance 'db0' exists")`,
				`foundation1.InstanceExists("db0")`,
				`logger.Printf("[foundation2] Checking whether instance 'db1' exists")`,
				`foundation2.InstanceExists("db1")`,
				`logger.Printf("[foundation2] Retrieving information for secondary instance 'db1'")`,
				`foundation2.CreateHostInfoKey("db1")`,
				`logger.Printf("[foundation1] Registering secondary instance information on primary instance 'db0'")`,
				fmt.Sprintf(`foundation1.UpdateServiceAndWait("db0", %q)`, fakeFoundation2.CreateHostInfoKeyResult.Key),
				`logger.Printf("[foundation1] Retrieving replication configuration from primary instance 'db0'")`,
				`foundation1.CreateCredentialsKey("db0")`,
				`logger.Printf("[foundation2] Updating secondary instance 'db1' with replication configuration")`,
				fmt.Sprintf(`foundation2.UpdateServiceAndWait("db1", %q)`, fakeFoundation1.CreateCredentialsKeyResult.Key),
			}))
		})
	})
})
