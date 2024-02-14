package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("SetupReplication", func() {
	var fakeMultiSite *fakes.FakeMultiSite

	longFlagArgs := []string{
		"--primary-target=primary-target-name",
		"--primary-instance=primary-instance-name",
		"--secondary-target=secondary-target-name",
		"--secondary-instance=secondary-instance-name"}

	BeforeEach(func() {
		fakeMultiSite = new(fakes.FakeMultiSite)

	})
	It("returns an error if called with too many arguments", func() {
		err := commands.SetupReplication(append(longFlagArgs, "extra_arg"), fakeMultiSite)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("Usage: " + commands.SetupReplicationUsage + "\n\nunexpected arguments: extra_arg"))
	})

	It("returns an error if called with too few arguments", func() {
		args := []string{
			"--primary-target=primary-target-name",
			"--primary-instance=primary-instance-name"}
		err := commands.SetupReplication(args, fakeMultiSite)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("Usage: " + commands.SetupReplicationUsage + "\n\nthe required flags `-S, --secondary-target' and `-s, --secondary-instance' were not specified"))
	})

	It("returns an error if SetupReplication returns an error", func() {
		fakeMultiSite.SetupReplicationReturns(errors.New("Low-level error message"))
		err := commands.SetupReplication(longFlagArgs, fakeMultiSite)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("replication setup error: Low-level error message")))
	})

	It("passes arguments with long flags to SetupReplication", func() {
		err := commands.SetupReplication(longFlagArgs, fakeMultiSite)
		Expect(err).NotTo(HaveOccurred())

		rcv1, rcv2, rcv3, rcv4 := fakeMultiSite.SetupReplicationArgsForCall(0)
		Expect(rcv1).To(Equal("primary-target-name"))
		Expect(rcv2).To(Equal("primary-instance-name"))
		Expect(rcv3).To(Equal("secondary-target-name"))
		Expect(rcv4).To(Equal("secondary-instance-name"))
	})

	It("passes arguments with short flags to SetupReplication", func() {
		shortFlagArgs := []string{
			"-P=primary-target-name",
			"-p=primary-instance-name",
			"-S=secondary-target-name",
			"-s=secondary-instance-name"}
		err := commands.SetupReplication(shortFlagArgs, fakeMultiSite)
		Expect(err).NotTo(HaveOccurred())

		rcv1, rcv2, rcv3, rcv4 := fakeMultiSite.SetupReplicationArgsForCall(0)
		Expect(rcv1).To(Equal("primary-target-name"))
		Expect(rcv2).To(Equal("primary-instance-name"))
		Expect(rcv3).To(Equal("secondary-target-name"))
		Expect(rcv4).To(Equal("secondary-instance-name"))
	})

})
