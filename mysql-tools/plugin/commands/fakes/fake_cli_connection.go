package fakes

import (
	"code.cloudfoundry.org/cli/plugin"
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
)

type FakeCliConnection struct{}

func (f FakeCliConnection) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) CliCommand(args ...string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetCurrentOrg() (plugin_models.Organization, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetCurrentSpace() (plugin_models.Space, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) Username() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) UserGuid() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) UserEmail() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) IsLoggedIn() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) IsSSLDisabled() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) HasOrganization() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) HasSpace() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) ApiEndpoint() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) ApiVersion() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) HasAPIEndpoint() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) LoggregatorEndpoint() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) DopplerEndpoint() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) AccessToken() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetApp(s string) (plugin_models.GetAppModel, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetApps() ([]plugin_models.GetAppsModel, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetOrgs() ([]plugin_models.GetOrgs_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetSpaces() ([]plugin_models.GetSpaces_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetOrgUsers(s string, s2 ...string) ([]plugin_models.GetOrgUsers_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetSpaceUsers(s string, s2 string) ([]plugin_models.GetSpaceUsers_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetServices() ([]plugin_models.GetServices_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetService(s string) (plugin_models.GetService_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetOrg(s string) (plugin_models.GetOrg_Model, error) {
	//TODO implement me
	panic("implement me")
}

func (f FakeCliConnection) GetSpace(s string) (plugin_models.GetSpace_Model, error) {
	//TODO implement me
	panic("implement me")
}

var _ plugin.CliConnection = (*FakeCliConnection)(nil)
