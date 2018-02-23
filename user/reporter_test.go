package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/pkg/errors"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/user/userfakes"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/user"
)

var _ = Describe("Reporter IsSpaceDeveloper", func() {
	var (
		userReporter *user.Reporter
		fakeCliConnection *userfakes.FakeCliConnection
	)

	BeforeEach(func() {
		fakeCliConnection = new(userfakes.FakeCliConnection)
		userReporter = user.NewReporter(fakeCliConnection)

		fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{
			plugin_models.OrganizationFields{Name: "some-org-name"},
		}, nil)
		fakeCliConnection.GetCurrentSpaceReturns(plugin_models.Space{
			plugin_models.SpaceFields{Name: "some-space-name"},
		}, nil)
		fakeCliConnection.UserGuidReturns("abc-some-user-guid", nil)
	})

	It("passes the correct arguments to GetSpaceUsers", func() {
		userReporter.IsSpaceDeveloper()

		org, space := fakeCliConnection.GetSpaceUsersArgsForCall(0)
		Expect(org).To(Equal("some-org-name"))
		Expect(space).To(Equal("some-space-name"))
	})

	Context("user is a space developer", func() {
		BeforeEach(func() {
			fakeCliConnection.GetSpaceUsersReturns([]plugin_models.GetSpaceUsers_Model{
				{Guid: "abc-some-user-guid", Roles: []string{"RoleSpaceDeveloper"}},
				{Guid: "def-some-incorrect-user", Roles: []string{"RoleSpaceAuditor"}},
			}, nil)
		})

		It("returns true", func() {
			Expect(userReporter.IsSpaceDeveloper()).To(BeTrue())
		})
	})

	Context("user is not a space developer", func() {
		BeforeEach(func() {
			fakeCliConnection.GetSpaceUsersReturns([]plugin_models.GetSpaceUsers_Model{
				{Guid: "abc-some-user-guid", Roles: []string{"RoleSpaceAuditor"}},
				{Guid: "def-some-incorrect-user", Roles: []string{"RoleSpaceAuditor"}},
			}, nil)
		})

		It("returns false", func() {
			Expect(userReporter.IsSpaceDeveloper()).To(BeFalse())
		})
	})

	Context("getting the current org returns an error", func() {
		BeforeEach(func() {
			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{}, errors.New("some-error"))
		})

		It("returns the error", func() {
			_, err := userReporter.IsSpaceDeveloper()
			Expect(err).To(MatchError("some-error"))
		})
	})

	Context("getting the current space returns an error", func() {
		BeforeEach(func() {
			fakeCliConnection.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("some-error"))
		})

		It("returns the error", func() {
			_, err := userReporter.IsSpaceDeveloper()
			Expect(err).To(MatchError("some-error"))
		})
	})

	Context("getting the user guid returns an error", func() {
		BeforeEach(func() {
			fakeCliConnection.UserGuidReturns("", errors.New("some-error"))
		})

		It("returns the error", func() {
			_, err := userReporter.IsSpaceDeveloper()
			Expect(err).To(MatchError("some-error"))
		})
	})

	Context("getting the space users returns an error", func() {
		BeforeEach(func() {
			fakeCliConnection.GetSpaceUsersReturns(nil, errors.New("some-error"))
		})

		It("returns the error", func() {
			_, err := userReporter.IsSpaceDeveloper()
			Expect(err).To(MatchError("some-error"))
		})
	})
})
