package commands_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("Switchover", func() {
	When("the user hits enter and provides no confirm", func() {
		It("prompts for confirmation and aborts", func() {
			var out bytes.Buffer
			var in = bytes.NewBufferString("\n")

			args := []string{
				"--primary-target=target-1",
				"--primary-instance=instance-1",
				"--secondary-target=target-2",
				"--secondary-instance=instance-2",
			}
			err := commands.SwitchoverReplication(args, nil, &out, in)
			Expect(err).NotTo(HaveOccurred())

			Expect(out.String()).To(ContainSubstring(`When successful, instance-1 will become secondary and instance-2 will become primary. Do you want to continue? [yN]:`))
			Expect(out.String()).To(ContainSubstring(`Operation cancelled`))
		})
	})

	When("the user provides no arguments", func() {
		It("returns an error", func() {
			var args []string

			err := commands.SwitchoverReplication(args, nil, nil, nil)

			Expect(err).To(MatchError(ContainSubstring("Usage: cf mysql-tools switchover [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ] [ --force | -f ]")))
			Expect(err).To(MatchError(ContainSubstring("the required flags `-P, --primary-target', `-S, --secondary-target', `-p, --primary-instance' and `-s, --secondary-instance' were not specified")))
		})
	})

	When("the user an extra invalid argument", func() {
		It("returns an error", func() {
			args := []string{
				"--primary-target=target-1",
				"--primary-instance=instance-1",
				"--secondary-target=target-2",
				"--secondary-instance=instance-2",
				"extra-argument",
			}

			err := commands.SwitchoverReplication(args, nil, nil, nil)

			Expect(err).To(MatchError(ContainSubstring("Usage: cf mysql-tools switchover [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ] [ --force | -f ]")))
			Expect(err).To(MatchError(ContainSubstring("unexpected arguments: extra-argument")))
		})
	})

	When("the user explicit enters negative confirmation", func() {
		It("prompts for confirmation and aborts", func() {
			var out bytes.Buffer
			var in = bytes.NewBufferString("n\n")

			args := []string{
				"--primary-target=target-1",
				"--primary-instance=instance-1",
				"--secondary-target=target-2",
				"--secondary-instance=instance-2",
			}
			err := commands.SwitchoverReplication(args, nil, &out, in)
			Expect(err).NotTo(HaveOccurred())

			Expect(out.String()).To(ContainSubstring(`When successful, instance-1 will become secondary and instance-2 will become primary. Do you want to continue? [yN]:`))
			Expect(out.String()).To(ContainSubstring(`Operation cancelled`))
		})
	})

	When("the user explicit enters positive confirmation", func() {
		It("attempts switchover operations", func() {
			var out bytes.Buffer
			var in = bytes.NewBufferString("y\n")

			cfg := new(fakes.FakeMultisiteConfig)
			cfg.ConfigDirReturns("/some/invalid/path")

			args := []string{
				"--primary-target=primary-target-name",
				"--primary-instance=primary-instance-name",
				"--secondary-target=secondary-target-name",
				"--secondary-instance=secondary-instance-name",
			}

			err := commands.SwitchoverReplication(args, cfg, &out, in)
			Expect(err).To(MatchError(ContainSubstring(`error when checking whether instance exists`)))
		})
	})

	When("forcing the operation to continue", func() {
		It("does not prompt for confirmation when using the long force option (--force)", func() {
			cfg := new(fakes.FakeMultisiteConfig)
			cfg.ConfigDirReturns("/some/invalid/path")

			args := []string{
				"--primary-target=primary-target-name",
				"--primary-instance=primary-instance-name",
				"--secondary-target=secondary-target-name",
				"--secondary-instance=secondary-instance-name",
				"--force",
			}

			err := commands.SwitchoverReplication(args, cfg, nil, nil)
			Expect(err).To(MatchError(ContainSubstring(`error when checking whether instance exists`)))
		})

		It("does not prompt for confirmation when using the short force option (-f)", func() {
			cfg := new(fakes.FakeMultisiteConfig)
			cfg.ConfigDirReturns("/some/invalid/path")

			args := []string{
				"--primary-target=primary-target-name",
				"--primary-instance=primary-instance-name",
				"--secondary-target=secondary-target-name",
				"--secondary-instance=secondary-instance-name",
				"-f",
			}

			err := commands.SwitchoverReplication(args, cfg, nil, nil)
			Expect(err).To(MatchError(ContainSubstring(`error when checking whether instance exists`)))
		})
	})

	When("the user provides no input to the prompt", func() {
		It("provide a useful error", func() {
			cfg := new(fakes.FakeMultisiteConfig)
			cfg.ConfigDirReturns("/some/invalid/path")

			args := []string{
				"--primary-target=primary-target-name",
				"--primary-instance=primary-instance-name",
				"--secondary-target=secondary-target-name",
				"--secondary-instance=secondary-instance-name",
			}

			out := &bytes.Buffer{}
			in := &bytes.Buffer{}

			err := commands.SwitchoverReplication(args, cfg, out, in)

			Expect(err).NotTo(HaveOccurred())
			Expect(out.String()).To(ContainSubstring(`Operation cancelled`))
		})
	})
})
