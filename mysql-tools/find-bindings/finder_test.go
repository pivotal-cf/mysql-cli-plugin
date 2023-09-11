// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
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

package find_bindings_test

import (
	"errors"
	"net/url"

	"github.com/cloudfoundry-community/go-cfclient/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings/find-bindingsfakes"
)

var _ = Describe("BindingFinder", func() {
	Context("FindBindings", func() {
		var (
			serviceName            string
			expectedBindings       []find_bindings.Binding
			fakeClient             *findbindingsfakes.FakeClient
			service                cfclient.Service
			servicePlans           []cfclient.ServicePlan
			smallServiceInstances  []cfclient.ServiceInstance
			mediumServiceInstances []cfclient.ServiceInstance
			smallServiceBindings   []cfclient.ServiceBinding
			mediumServiceBindings  []cfclient.ServiceBinding
			smallServiceKey        []cfclient.ServiceKey
			mediumServiceKey       []cfclient.ServiceKey
			smallApp               cfclient.App
			mediumApp              cfclient.App
		)

		BeforeEach(func() {
			fakeClient = &findbindingsfakes.FakeClient{}
			serviceName = "p.mysql"

			expectedBindings = []find_bindings.Binding{
				{
					Name:                "app1",
					ServiceInstanceName: "instance1",
					ServiceInstanceGuid: "instance1-guid",
					OrgName:             "app1-org",
					SpaceName:           "app1-space",
					Type:                "AppBinding",
				},
				{
					Name:                "key1",
					ServiceInstanceName: "instance1",
					ServiceInstanceGuid: "instance1-guid",
					OrgName:             "app1-org",
					SpaceName:           "app1-space",
					Type:                "ServiceKeyBinding",
				},
				{
					Name:                "app3",
					ServiceInstanceName: "instance3",
					ServiceInstanceGuid: "instance3-guid",
					OrgName:             "app3-org",
					SpaceName:           "app3-space",
					Type:                "AppBinding",
				},
				{
					Name:                "key3",
					ServiceInstanceName: "instance3",
					ServiceInstanceGuid: "instance3-guid",
					OrgName:             "app3-org",
					SpaceName:           "app3-space",
					Type:                "ServiceKeyBinding",
				},
			}

			service = cfclient.Service{
				Label: "p.mysql",
				Guid:  "service-guid",
			}

			fakeClient.ListServicesByQueryReturns([]cfclient.Service{service}, nil)

			servicePlans = []cfclient.ServicePlan{
				{Name: "small", Guid: "small-guid", ServiceGuid: "service-guid"},
				{Name: "medium", Guid: "medium-guid", ServiceGuid: "service-guid"},
				{Name: "large", Guid: "large-guid", ServiceGuid: "service-guid"},
			}

			fakeClient.ListServicePlansByQueryReturns(servicePlans, nil)

			smallServiceInstances = []cfclient.ServiceInstance{
				{Name: "instance1", Guid: "instance1-guid", ServicePlanGuid: "small-guid", SpaceGuid: "space1-guid"},
				{Name: "instance2", Guid: "instance2-guid", ServicePlanGuid: "small-guid", SpaceGuid: "space2-guid"},
			}

			fakeClient.ListServiceInstancesByQueryReturnsOnCall(0, smallServiceInstances, nil)

			mediumServiceInstances = []cfclient.ServiceInstance{
				{Name: "instance3", Guid: "instance3-guid", ServicePlanGuid: "medium-guid", SpaceGuid: "space3-guid"},
			}

			fakeClient.ListServiceInstancesByQueryReturnsOnCall(1, mediumServiceInstances, nil)
			fakeClient.ListServiceInstancesByQueryReturnsOnCall(2, []cfclient.ServiceInstance{}, nil)

			smallServiceBindings = []cfclient.ServiceBinding{
				{Guid: "binding1-guid", AppGuid: "app1-guid", ServiceInstanceGuid: "instance1-guid"},
			}
			mediumServiceBindings = []cfclient.ServiceBinding{
				{Guid: "binding3-guid", AppGuid: "app3-guid", ServiceInstanceGuid: "instance3-guid"},
			}

			fakeClient.ListServiceBindingsByQueryReturnsOnCall(0, smallServiceBindings, nil)
			fakeClient.ListServiceBindingsByQueryReturnsOnCall(1, []cfclient.ServiceBinding{}, nil)
			fakeClient.ListServiceBindingsByQueryReturnsOnCall(2, mediumServiceBindings, nil)

			smallServiceKey = []cfclient.ServiceKey{
				{Name: "key1"},
			}

			mediumServiceKey = []cfclient.ServiceKey{
				{Name: "key3"},
			}
			fakeClient.ListServiceKeysByQueryReturnsOnCall(0, smallServiceKey, nil)
			fakeClient.ListServiceKeysByQueryReturnsOnCall(1, []cfclient.ServiceKey{}, nil)
			fakeClient.ListServiceKeysByQueryReturnsOnCall(2, mediumServiceKey, nil)

			smallApp = cfclient.App{
				Guid: "app1-guid",
				Name: "app1",
				SpaceData: cfclient.SpaceResource{
					Entity: cfclient.Space{
						Name:             "app1-space",
						OrganizationGuid: "app1-org-guid",
						OrgData: cfclient.OrgResource{
							Entity: cfclient.Org{
								Name: "app1-org",
							},
						},
					},
				},
			}

			mediumApp = cfclient.App{
				Guid: "app3-guid",
				Name: "app3",
				SpaceData: cfclient.SpaceResource{
					Entity: cfclient.Space{
						Name:             "app3-space",
						OrganizationGuid: "app3-org-guid",
						OrgData: cfclient.OrgResource{
							Entity: cfclient.Org{
								Name: "app3-org",
							},
						},
					},
				},
			}

			fakeClient.GetAppByGuidReturnsOnCall(0, smallApp, nil)
			fakeClient.GetAppByGuidReturnsOnCall(1, mediumApp, nil)

			fakeClient.GetSpaceByGuidReturnsOnCall(0, smallApp.SpaceData.Entity, nil)
			fakeClient.GetSpaceByGuidReturnsOnCall(1, mediumApp.SpaceData.Entity, nil)

			fakeClient.GetOrgByGuidReturnsOnCall(0, smallApp.SpaceData.Entity.OrgData.Entity, nil)
			fakeClient.GetOrgByGuidReturnsOnCall(1, mediumApp.SpaceData.Entity.OrgData.Entity, nil)
		})

		It("returns a list of applications and service keys associated with the service", func() {
			finder := find_bindings.NewBindingFinder(fakeClient)
			listOfBindings, err := finder.FindBindings(serviceName)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeClient.ListServicesByQueryCallCount()).To(Equal(1))
			query := url.Values{}
			query.Set("q", "label:p.mysql")
			Expect(fakeClient.ListServicesByQueryArgsForCall(0)).To(Equal(query))

			Expect(fakeClient.ListServicePlansByQueryCallCount()).To(Equal(1))
			query = url.Values{}
			query.Set("q", "service_guid:service-guid")
			Expect(fakeClient.ListServicePlansByQueryArgsForCall(0)).To(Equal(query))

			Expect(fakeClient.ListServiceInstancesByQueryCallCount()).To(Equal(3))
			query = url.Values{}
			query.Set("q", "service_plan_guid:small-guid")
			Expect(fakeClient.ListServiceInstancesByQueryArgsForCall(0)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_plan_guid:medium-guid")
			Expect(fakeClient.ListServiceInstancesByQueryArgsForCall(1)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_plan_guid:large-guid")
			Expect(fakeClient.ListServiceInstancesByQueryArgsForCall(2)).To(Equal(query))

			Expect(fakeClient.ListServiceBindingsByQueryCallCount()).To(Equal(3))
			query = url.Values{}
			query.Set("q", "service_instance_guid:instance1-guid")
			Expect(fakeClient.ListServiceBindingsByQueryArgsForCall(0)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_instance_guid:instance2-guid")
			Expect(fakeClient.ListServiceBindingsByQueryArgsForCall(1)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_instance_guid:instance3-guid")
			Expect(fakeClient.ListServiceBindingsByQueryArgsForCall(2)).To(Equal(query))

			Expect(fakeClient.GetAppByGuidCallCount()).To(Equal(2))
			Expect(fakeClient.GetAppByGuidArgsForCall(0)).To(Equal("app1-guid"))
			Expect(fakeClient.GetAppByGuidArgsForCall(1)).To(Equal("app3-guid"))

			Expect(fakeClient.ListServiceKeysByQueryCallCount()).To(Equal(3))
			query = url.Values{}
			query.Set("q", "service_instance_guid:instance1-guid")
			Expect(fakeClient.ListServiceKeysByQueryArgsForCall(0)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_instance_guid:instance2-guid")
			Expect(fakeClient.ListServiceKeysByQueryArgsForCall(1)).To(Equal(query))

			query = url.Values{}
			query.Set("q", "service_instance_guid:instance3-guid")
			Expect(fakeClient.ListServiceKeysByQueryArgsForCall(2)).To(Equal(query))

			Expect(fakeClient.GetSpaceByGuidCallCount()).To(Equal(2))
			Expect(fakeClient.GetSpaceByGuidArgsForCall(0)).To(Equal("space1-guid"))
			Expect(fakeClient.GetSpaceByGuidArgsForCall(1)).To(Equal("space3-guid"))

			Expect(fakeClient.GetOrgByGuidCallCount()).To(Equal(2))
			Expect(fakeClient.GetOrgByGuidArgsForCall(0)).To(Equal("app1-org-guid"))
			Expect(fakeClient.GetOrgByGuidArgsForCall(1)).To(Equal("app3-org-guid"))

			Expect(listOfBindings).To(Equal(expectedBindings))
		})

		// TODO: add failure cases here
		Context("when ListService fails", func() {
			BeforeEach(func() {
				fakeClient.ListServicesByQueryReturns([]cfclient.Service{}, errors.New("listServicesByQueryError"))
			})

			It("returns an error", func() {
				finder := find_bindings.NewBindingFinder(fakeClient)
				_, err := finder.FindBindings(serviceName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listServicesByQueryError"))
			})
		})

		Context("when ListServicePlan fails", func() {
			BeforeEach(func() {
				fakeClient.ListServicePlansByQueryReturns([]cfclient.ServicePlan{}, errors.New("listServicePlansByQueryError"))
			})

			It("returns an error", func() {
				finder := find_bindings.NewBindingFinder(fakeClient)
				_, err := finder.FindBindings(serviceName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listServicePlansByQueryError"))
			})
		})

		Context("when ListServiceInstances fails", func() {
			BeforeEach(func() {
				fakeClient.ListServiceInstancesByQueryReturnsOnCall(0, []cfclient.ServiceInstance{}, errors.New("listServiceInstancesByQueryError"))
			})

			It("returns an error", func() {
				finder := find_bindings.NewBindingFinder(fakeClient)
				_, err := finder.FindBindings(serviceName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listServiceInstancesByQueryError"))
			})
		})

		Context("when ListServiceBindings fails", func() {
			BeforeEach(func() {
				fakeClient.ListServiceBindingsByQueryReturnsOnCall(0, []cfclient.ServiceBinding{}, errors.New("listServiceBindingsByQueryError"))
			})

			It("returns an error", func() {
				finder := find_bindings.NewBindingFinder(fakeClient)
				_, err := finder.FindBindings(serviceName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listServiceBindingsByQueryError"))
			})
		})

		Context("when ListServiceKeys fails", func() {
			BeforeEach(func() {
				fakeClient.ListServiceKeysByQueryReturnsOnCall(0, []cfclient.ServiceKey{}, errors.New("listServiceKeysByQueryError"))
			})

			It("returns an error", func() {
				finder := find_bindings.NewBindingFinder(fakeClient)
				_, err := finder.FindBindings(serviceName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("listServiceKeysByQueryError"))
			})
		})
	})
})
