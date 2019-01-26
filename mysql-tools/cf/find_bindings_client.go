// Copyright (C) 2019-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package cf

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"net/url"
	"strings"
)

type FindBindingsClient struct {
	cliConnection plugin.CliConnection
	cfClient      find_bindings.Client
}

func NewFindBindingsClient(cliConnection plugin.CliConnection) *FindBindingsClient {
	return &FindBindingsClient{
		cliConnection: cliConnection,
		cfClient:      nil, // lazily initialized
	}
}

func (c *FindBindingsClient) GetAppByGuid(guid string) (cfclient.App, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return cfclient.App{}, err
	}

	return c.cfClient.GetAppByGuid(guid)
}

func (c *FindBindingsClient) GetOrgByGuid(spaceGUID string) (cfclient.Org, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return cfclient.Org{}, err
	}

	return c.cfClient.GetOrgByGuid(spaceGUID)
}

func (c *FindBindingsClient) GetSpaceByGuid(spaceGUID string) (cfclient.Space, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return cfclient.Space{}, err
	}

	return c.cfClient.GetSpaceByGuid(spaceGUID)
}

func (c *FindBindingsClient) ListServicesByQuery(query url.Values) ([]cfclient.Service, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return nil, err
	}

	return c.cfClient.ListServicesByQuery(query)
}

func (c *FindBindingsClient) ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return nil, err
	}

	return c.cfClient.ListServiceBindingsByQuery(query)
}

func (c *FindBindingsClient) ListServicePlansByQuery(query url.Values) ([]cfclient.ServicePlan, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return nil, err
	}

	return c.cfClient.ListServicePlansByQuery(query)
}

func (c *FindBindingsClient) ListServiceKeysByQuery(query url.Values) ([]cfclient.ServiceKey, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return nil, err
	}

	return c.cfClient.ListServiceKeysByQuery(query)
}

func (c *FindBindingsClient) ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error) {
	err := c.lazyInitializeCFClient()
	if err != nil {
		return nil, err
	}

	return c.cfClient.ListServiceInstancesByQuery(query)
}

func (c *FindBindingsClient) lazyInitializeCFClient() error {
	if c.cfClient != nil {
		return nil
	}

	apiEndpoint, err := c.cliConnection.ApiEndpoint()
	if err != nil {
		return err
	}

	bearToken, err := c.cliConnection.AccessToken()
	if err != nil {
		return err
	}

	tokens := strings.Fields(bearToken)

	sslDisabled, err := c.cliConnection.IsSSLDisabled()
	if err != nil {
		return err
	}

	newClient, err := cfclient.NewClient(&cfclient.Config{
		ApiAddress:        apiEndpoint,
		SkipSslValidation: sslDisabled,
		Token:             tokens[1],
	})

	if err != nil {
		return err
	}

	c.cfClient = newClient

	return nil
}
