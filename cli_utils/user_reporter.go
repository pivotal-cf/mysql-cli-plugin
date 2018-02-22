package cli_utils

import "code.cloudfoundry.org/cli/plugin/models"

//go:generate counterfeiter . CliConnection
type CliConnection interface {
	GetCurrentOrg() (plugin_models.Organization, error)
	GetCurrentSpace() (plugin_models.Space, error)
	UserGuid() (string, error)
	GetSpaceUsers(string, string) ([]plugin_models.GetSpaceUsers_Model, error)
}

const RoleSpaceDeveloper = "RoleSpaceDeveloper"

type UserReporter struct {
	cliConnection CliConnection
}

func NewUserReporter(cliConnection CliConnection) *UserReporter {
	return &UserReporter{
		cliConnection: cliConnection,
	}
}

func (u *UserReporter) IsSpaceDeveloper() (bool, error) {
	org, err := u.cliConnection.GetCurrentOrg()
	if err != nil {
		return false, err
	}

	space, err := u.cliConnection.GetCurrentSpace()
	if err != nil {
		return false, err
	}

	userGUID, err := u.cliConnection.UserGuid()
	if err != nil {
		return false, err
	}

	spaceUsers, err := u.cliConnection.GetSpaceUsers(org.Name, space.Name)
	if err != nil {
		return false, err
	}

	return spaceUsersAny(spaceUsers, func(spaceUser plugin_models.GetSpaceUsers_Model) bool {
		return spaceUser.Guid == userGUID && any(spaceUser.Roles, func(role string) bool {
			return role == RoleSpaceDeveloper
		})
	}), nil
}

func spaceUsersAny(vs []plugin_models.GetSpaceUsers_Model, f func(plugin_models.GetSpaceUsers_Model) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}


func any(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}