package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("SetupReplication", func() {
	longFlagArgs := []string{
		"--primary-target=primary-target-name",
		"--primary-instance=primary-instance-name",
		"--secondary-target=secondary-target-name",
		"--secondary-instance=secondary-instance-name"}

	BeforeEach(func() {
	})
	It("returns an error if called with too many arguments", func() {
		err := commands.SetupReplication(append(longFlagArgs, "extra_arg"), &fakes.FakeMultisiteConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("Usage: " + commands.SetupReplicationUsage + "\n\nunexpected arguments: extra_arg"))
	})

	It("returns an error if called with too few arguments", func() {
		args := []string{
			"--primary-target=primary-target-name",
			"--primary-instance=primary-instance-name"}
		err := commands.SetupReplication(args, &fakes.FakeMultisiteConfig{})
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("Usage: " + commands.SetupReplicationUsage + "\n\nthe required flags `-S, --secondary-target' and `-s, --secondary-instance' were not specified"))
	})
})
