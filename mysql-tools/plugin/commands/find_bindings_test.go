package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	findbindings "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("FindBindings", func() {
	const (
		findUsage = `cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`
	)

	var fakeFinder *fakes.FakeBindingFinder

	BeforeEach(func() {
		fakeFinder = new(fakes.FakeBindingFinder)
	})

	It("returns an error if not enough args are passed", func() {
		var args []string
		err := commands.FindBindings(args, fakeFinder)
		Expect(err).To(MatchError("Usage: " + findUsage + "\n\nthe required argument `<mysql-v1-service-name>` was not provided"))
	})

	It("returns an error if too many args are passed", func() {
		args := []string{"p.mysql", "somethingelse"}
		err := commands.FindBindings(args, fakeFinder)
		Expect(err).To(MatchError("Usage: " + findUsage + "\n\nunexpected arguments: somethingelse"))
	})

	It("returns an error if an invalid flag is passed", func() {
		args := []string{"p.mysql", "--invalid-flag"}
		err := commands.FindBindings(args, fakeFinder)
		Expect(err).To(MatchError("Usage: " + findUsage + "\n\nunknown flag `invalid-flag'"))
	})

	When("find binding runs successfully", func() {
		It("succeeds", func() {
			args := []string{"p.mysql"}
			err := commands.FindBindings(args, fakeFinder)
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeFinder.FindBindingsCallCount()).To(Equal(1))
			Expect(fakeFinder.FindBindingsArgsForCall(0)).To(Equal("p.mysql"))
		})
	})

	When("find binding returns an error", func() {
		It("fails", func() {
			args := []string{"p.mysql"}
			fakeFinder.FindBindingsReturns([]findbindings.Binding{}, errors.New("some-error"))
			err := commands.FindBindings(args, fakeFinder)
			Expect(err).To(MatchError("some-error"))
			Expect(fakeFinder.FindBindingsCallCount()).To(Equal(1))
			Expect(fakeFinder.FindBindingsArgsForCall(0)).To(Equal("p.mysql"))
		})
	})
})
