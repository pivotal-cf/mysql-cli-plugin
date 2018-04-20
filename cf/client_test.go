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

package cf_test

import (
	"errors"
	"os"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/cf/cffakes"
)

type FakeClock struct {
	sleepCount    int
	sleepCallArgs []time.Duration
}

func (c *FakeClock) Sleep(d time.Duration) {
	c.sleepCallArgs = append(c.sleepCallArgs, d)
	c.sleepCount++
}

func (c *FakeClock) SleepCallCount() int {
	return c.sleepCount
}

func (c *FakeClock) SleepCallArgs(i int) time.Duration {
	return c.sleepCallArgs[i]
}

var _ = Describe("Client", func() {
	var (
		client          *cf.Client
		fakeCFPluginAPI *cffakes.FakeCFPluginAPI
		fakeClock       *FakeClock
		buffer          *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeCFPluginAPI = new(cffakes.FakeCFPluginAPI)
		client = cf.NewClient(fakeCFPluginAPI)
		fakeClock = &FakeClock{}
		client.Sleep = fakeClock.Sleep
		buffer = gbytes.NewBuffer()
		client.Log.SetOutput(buffer)
	})

	Context("BindService", func() {
		It("binds an app to a service", func() {
			Expect(client.BindService("some-app", "some-service")).
				To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal(
				[]string{
					"bind-service",
					"some-app",
					"some-service",
				},
			))
		})

		It("returns an error when the binding request fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"),
			)

			err := client.BindService("some-app", "some-service")
			Expect(err).To(MatchError(`failed to bind-service "some-service" to application "some-app": some-error`))
		})
	})

	Context("CreateServiceInstance", func() {
		Context("When we create a service instance", func() {
			var (
				inProgressService plugin_models.GetService_Model
				completedService  plugin_models.GetService_Model
			)

			BeforeEach(func() {
				inProgressService = plugin_models.GetService_Model{
					LastOperation: plugin_models.GetService_LastOperation{
						Type:  "create",
						State: "in progress",
					},
				}
				completedService = plugin_models.GetService_Model{
					LastOperation: plugin_models.GetService_LastOperation{
						Type:  "create",
						State: "succeeded",
					},
				}

				fakeCFPluginAPI.GetServiceReturnsOnCall(0, plugin_models.GetService_Model{}, errors.New("not-exist"))
				fakeCFPluginAPI.GetServiceReturnsOnCall(1, inProgressService, nil)
				fakeCFPluginAPI.GetServiceReturnsOnCall(2, inProgressService, nil)
				fakeCFPluginAPI.GetServiceReturnsOnCall(3, completedService, nil)
			})

			It("We wait until the service instance has been successfully created", func() {
				err := client.CreateServiceInstance("plan-type", "service-instance-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(4))
			})

			Context("when the service polling fails continuously", func() {
				BeforeEach(func() {
					fakeCFPluginAPI.GetServiceReturnsOnCall(3, plugin_models.GetService_Model{}, errors.New("boom!"))
					fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, errors.New("boom!"))
				})

				It("keeps trying until a timeout is reached", func() {
					err := client.CreateServiceInstance("plan-type", "service-instance-name")
					Expect(err).To(MatchError("failed to look up status of service instance 'service-instance-name'"))
					Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(6))
				})
			})

			Context("when the service polling fails intermittently", func() {
				BeforeEach(func() {
					fakeCFPluginAPI.GetServiceReturnsOnCall(3, plugin_models.GetService_Model{}, errors.New("boom!"))
					fakeCFPluginAPI.GetServiceReturnsOnCall(4, completedService, nil)
				})

				It("keeps trying until a definitive answer is reached", func() {
					err := client.CreateServiceInstance("plan-type", "service-instance-name")
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(5))
				})
			})

			Context("when the service fails to create successfully", func() {
				BeforeEach(func() {
					failedService := plugin_models.GetService_Model{
						LastOperation: plugin_models.GetService_LastOperation{
							Type:        "create",
							State:       "failed",
							Description: "description",
						},
					}
					fakeCFPluginAPI.GetServiceReturnsOnCall(3, failedService, nil)
				})

				It("returns an error", func() {
					err := client.CreateServiceInstance("plan-type", "service-instance-name")
					Expect(err).To(MatchError("failed to create service instance 'service-instance-name': description"))
					Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(4))
				})
			})

			Context("when RECIPIENT_PRODUCT_NAME is set", func() {
				var (
					originalProductName string
				)
				BeforeEach(func() {
					originalProductName = os.Getenv("RECIPIENT_PRODUCT_NAME")
					os.Setenv("RECIPIENT_PRODUCT_NAME", "some-fake-product")
				})

				It("Uses the product name from RECIPIENT_PRODUCT_NAME when creating the service instance", func() {
					err := client.CreateServiceInstance("plan-type", "service-instance-name")

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
						To(Equal([]string{
						"create-service",
						"some-fake-product",
						"plan-type",
						"service-instance-name",
					}))
					Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				})

				AfterEach(func() {
					os.Setenv("RECIPIENT_PRODUCT_NAME", originalProductName)
				})
			})
		})

		Context("When an invalid plan type is specified", func() {
			It("returns an error", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, errors.New("does not exist"))
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{}, errors.New("Invalid service plan"))

				err := client.CreateServiceInstance("invalid-plan-type", "service-instance-name")
				Expect(err).To(MatchError("Invalid service plan"))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
					To(Equal([]string{
					"create-service",
					"p.mysql",
					"invalid-plan-type",
					"service-instance-name",
				}))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
			})
		})

		Context("When a pre-existing service name is requested", func() {
			It("Fails", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{Guid: "some-guid"}, nil)

				err := client.CreateServiceInstance("plan-type", "preexisting-service-instance-name")
				Expect(err).To(MatchError("service instance 'preexisting-service-instance-name' already exists"))
				Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(1))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(0))
			})
		})

	})

	Context("GetHostnames", func() {
		BeforeEach(func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{
					`Getting key some-service-key for service instance test-tls as admin...`,
					``,
					`{`,
					` "hostname": "some-host-name",`,
					` "jdbcUrl": "jdbc:mysql://some-host-name:3306/service_instance_db?user=8efbb5299eae4b8698fdfaca0e07e0d1\u0026password=1jms9jmd7i0twuk7\u0026useSSL=true\u0026requireSSL=true",`,
					` "name": "service_instance_db",`,
					` "password": "1jms9jmd7i0twuk7",`,
					` "port": 3306,`,
					` "uri": "mysql://8efbb5299eae4b8698fdfaca0e07e0d1:1jms9jmd7i0twuk7@some-host-name:3306/service_instance_db?reconnect=true",`,
					` "username": "8efbb5299eae4b8698fdfaca0e07e0d1"`,
					`}`,
				}, nil,
			)
		})

		It("returns the hostname for a single service instance", func() {
			hostnames, err := client.GetHostnames("some-instance")
			Expect(err).NotTo(HaveOccurred())

			Expect(hostnames).To(ConsistOf(`some-host-name`))
			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(3))

			createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
			Expect(createServiceArgs).To(HaveLen(3))
			Expect(createServiceArgs[0]).To(Equal("create-service-key"))
			Expect(createServiceArgs[1]).To(Equal("some-instance"))
			Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))

			serviceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)
			Expect(serviceKeyArgs).To(HaveLen(3))
			Expect(serviceKeyArgs[0]).To(Equal("service-key"))
			Expect(serviceKeyArgs[1]).To(Equal("some-instance"))
			Expect(serviceKeyArgs[2]).To(HavePrefix("MIGRATE-"))

			deleteServiceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)
			Expect(deleteServiceKeyArgs).To(HaveLen(4))
			Expect(deleteServiceKeyArgs[0]).To(Equal("delete-service-key"))
			Expect(deleteServiceKeyArgs[1]).To(Equal("-f"))
			Expect(deleteServiceKeyArgs[2]).To(Equal("some-instance"))
			Expect(deleteServiceKeyArgs[3]).To(HavePrefix("MIGRATE-"))
		})

		It("returns all the hostnames for a leader/follower service instance", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{
					`Getting key some-service-key for service instance test-tls as admin...`,
					``,
					`{`,
					` "hostname": "some-host-name",`,
					` "hostnames": [`,
					`  "some-host-name",`,
					`  "some-other-host-name"`,
					` ],`,
					` "jdbcUrl": "jdbc:mysql://some-host-name:3306/service_instance_db?user=8efbb5299eae4b8698fdfaca0e07e0d1\u0026password=1jms9jmd7i0twuk7\u0026useSSL=true\u0026requireSSL=true",`,
					` "name": "service_instance_db",`,
					` "password": "1jms9jmd7i0twuk7",`,
					` "port": 3306,`,
					` "uri": "mysql://8efbb5299eae4b8698fdfaca0e07e0d1:1jms9jmd7i0twuk7@some-host-name:3306/service_instance_db?reconnect=true",`,
					` "username": "8efbb5299eae4b8698fdfaca0e07e0d1"`,
					`}`,
				}, nil,
			)

			hostnames, err := client.GetHostnames("some-instance")
			Expect(err).NotTo(HaveOccurred())

			Expect(hostnames).To(ConsistOf(`some-host-name`, `some-other-host-name`))
			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(3))

			createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
			Expect(createServiceArgs).To(HaveLen(3))
			Expect(createServiceArgs[0]).To(Equal("create-service-key"))
			Expect(createServiceArgs[1]).To(Equal("some-instance"))
			Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))

			serviceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)
			Expect(serviceKeyArgs).To(HaveLen(3))
			Expect(serviceKeyArgs[0]).To(Equal("service-key"))
			Expect(serviceKeyArgs[1]).To(Equal("some-instance"))
			Expect(serviceKeyArgs[2]).To(HavePrefix("MIGRATE-"))

			deleteServiceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)
			Expect(deleteServiceKeyArgs).To(HaveLen(4))
			Expect(deleteServiceKeyArgs[0]).To(Equal("delete-service-key"))
			Expect(deleteServiceKeyArgs[1]).To(Equal("-f"))
			Expect(deleteServiceKeyArgs[2]).To(Equal("some-instance"))
			Expect(deleteServiceKeyArgs[3]).To(HavePrefix("MIGRATE-"))
		})

		Context("When the service key fails to be created", func() {
			It("fails", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(nil, errors.New("cannot create service key"))

				_, err := client.GetHostnames("some-instance")

				Expect(err).To(MatchError("Cannot get the hostnames for some-instance: cannot create service key"))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))

				createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(createServiceArgs).To(HaveLen(3))
				Expect(createServiceArgs[0]).To(Equal("create-service-key"))
				Expect(createServiceArgs[1]).To(Equal("some-instance"))
				Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))
			})
		})

		Context("When the new service key cannot be read", func() {
			It("fails when the cli command returns invalid json", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1, []string{"not json"}, nil)
				_, err := client.GetHostnames("some-instance")

				Expect(err).To(MatchError("Cannot get the hostnames for some-instance: invalid response: not json"))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(3))

				createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(createServiceArgs).To(HaveLen(3))
				Expect(createServiceArgs[0]).To(Equal("create-service-key"))
				Expect(createServiceArgs[1]).To(Equal("some-instance"))
				Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))

				serviceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)
				Expect(serviceKeyArgs).To(HaveLen(3))
				Expect(serviceKeyArgs[0]).To(Equal("service-key"))
				Expect(serviceKeyArgs[1]).To(Equal("some-instance"))
				Expect(serviceKeyArgs[2]).To(HavePrefix("MIGRATE-"))

				deleteServiceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)
				Expect(deleteServiceKeyArgs).To(HaveLen(4))
				Expect(deleteServiceKeyArgs[0]).To(Equal("delete-service-key"))
				Expect(deleteServiceKeyArgs[1]).To(Equal("-f"))
				Expect(deleteServiceKeyArgs[2]).To(Equal("some-instance"))
				Expect(deleteServiceKeyArgs[3]).To(HavePrefix("MIGRATE-"))
			})

			It("fails when the cli command fails", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1, nil, errors.New("cannot read service key"))

				_, err := client.GetHostnames("some-instance")

				Expect(err).To(MatchError("Cannot get the hostnames for some-instance: cannot read service key"))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(3))

				createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(createServiceArgs).To(HaveLen(3))
				Expect(createServiceArgs[0]).To(Equal("create-service-key"))
				Expect(createServiceArgs[1]).To(Equal("some-instance"))
				Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))

				serviceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)
				Expect(serviceKeyArgs).To(HaveLen(3))
				Expect(serviceKeyArgs[0]).To(Equal("service-key"))
				Expect(serviceKeyArgs[1]).To(Equal("some-instance"))
				Expect(serviceKeyArgs[2]).To(HavePrefix("MIGRATE-"))

				deleteServiceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)
				Expect(deleteServiceKeyArgs).To(HaveLen(4))
				Expect(deleteServiceKeyArgs[0]).To(Equal("delete-service-key"))
				Expect(deleteServiceKeyArgs[1]).To(Equal("-f"))
				Expect(deleteServiceKeyArgs[2]).To(Equal("some-instance"))
				Expect(deleteServiceKeyArgs[3]).To(HavePrefix("MIGRATE-"))

			})
		})

		Context("When the service key fails to be deleted", func() {
			It("succeeds anyway", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(2, nil, errors.New("cannot delete service key"))

				hostnames, err := client.GetHostnames("some-instance")

				Expect(err).NotTo(HaveOccurred())
				Expect(hostnames).To(ConsistOf(`some-host-name`))
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(3))

				createServiceArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(createServiceArgs).To(HaveLen(3))
				Expect(createServiceArgs[0]).To(Equal("create-service-key"))
				Expect(createServiceArgs[1]).To(Equal("some-instance"))
				Expect(createServiceArgs[2]).To(HavePrefix("MIGRATE-"))

				serviceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)
				Expect(serviceKeyArgs).To(HaveLen(3))
				Expect(serviceKeyArgs[0]).To(Equal("service-key"))
				Expect(serviceKeyArgs[1]).To(Equal("some-instance"))
				Expect(serviceKeyArgs[2]).To(HavePrefix("MIGRATE-"))

				deleteServiceKeyArgs := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)
				Expect(deleteServiceKeyArgs).To(HaveLen(4))
				Expect(deleteServiceKeyArgs[0]).To(Equal("delete-service-key"))
				Expect(deleteServiceKeyArgs[1]).To(Equal("-f"))
				Expect(deleteServiceKeyArgs[2]).To(Equal("some-instance"))
				Expect(deleteServiceKeyArgs[3]).To(HavePrefix("MIGRATE-"))
			})
		})
	})

	Context("UpdateServiceConfig", func() {
		var (
			inProgressUpdate plugin_models.GetService_Model
			completedUpdate  plugin_models.GetService_Model
		)

		BeforeEach(func() {
			inProgressUpdate = plugin_models.GetService_Model{
				LastOperation: plugin_models.GetService_LastOperation{
					Type:  "update",
					State: "in progress",
				},
			}
			completedUpdate = plugin_models.GetService_Model{
				LastOperation: plugin_models.GetService_LastOperation{
					Type:  "update",
					State: "succeeded",
				},
			}

			fakeCFPluginAPI.GetServiceReturnsOnCall(0, plugin_models.GetService_Model{}, errors.New("not-exist"))
			fakeCFPluginAPI.GetServiceReturnsOnCall(1, inProgressUpdate, nil)
			fakeCFPluginAPI.GetServiceReturnsOnCall(2, inProgressUpdate, nil)
			fakeCFPluginAPI.GetServiceReturnsOnCall(3, completedUpdate, nil)
		})

		It("Waits until the service has been successfully updated", func() {
			err := client.UpdateServiceConfig("service-instance-name", `{"config-key":"config-value"}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(4))
		})

		Context("when the service polling fails continuously", func() {
			BeforeEach(func() {
				fakeCFPluginAPI.GetServiceReturnsOnCall(3, plugin_models.GetService_Model{}, errors.New("boom!"))
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, errors.New("boom!"))
			})

			It("keeps trying until a timeout is reached", func() {
				err := client.UpdateServiceConfig("service-instance-name", `{"config-key":"config-value"}`)
				Expect(err).To(MatchError("failed to look up status of service instance 'service-instance-name'"))
				Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(6))
			})
		})

		Context("when the service polling fails intermittently", func() {
			BeforeEach(func() {
				fakeCFPluginAPI.GetServiceReturnsOnCall(3, plugin_models.GetService_Model{}, errors.New("boom!"))
				fakeCFPluginAPI.GetServiceReturnsOnCall(4, completedUpdate, nil)
			})

			It("keeps trying until a definitive answer is reached", func() {
				err := client.UpdateServiceConfig("service-instance-name", `{"config-key":"config-value"}`)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(5))
			})
		})

		Context("when the service fails to update successfully", func() {
			BeforeEach(func() {
				failedService := plugin_models.GetService_Model{
					LastOperation: plugin_models.GetService_LastOperation{
						Type:        "update",
						State:       "failed",
						Description: "description",
					},
				}
				fakeCFPluginAPI.GetServiceReturnsOnCall(3, failedService, nil)
			})

			It("returns an error", func() {
				err := client.UpdateServiceConfig("service-instance-name", `{"config-key":"config-value"}`)
				Expect(err).To(MatchError("failed to update service config 'service-instance-name': description"))
				Expect(fakeCFPluginAPI.GetServiceCallCount()).To(Equal(4))
			})
		})
	})

	Context("DeleteServiceInstance", func() {
		Context("When the specified instance exists", func() {
			It("Runs the delete-service command", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, nil)

				err := client.DeleteServiceInstance("service-instance-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
			})
		})

		Context("When the specified instance doesn't exist", func() {
			It("Succeeds anyway", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, errors.New("invalid instance"))

				err := client.DeleteServiceInstance("invalid-service-instance-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
			})
		})

	})

	Context("CreateTask", func() {
		It("creates a task", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
				`{`,
				`"guid": "abc-123",`,
				`"state": "RUNNING"`,
				`}`,
			}, nil)

			task, err := client.CreateTask(cf.App{
				Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
			}, "some-command")
			Expect(err).NotTo(HaveOccurred())
			Expect(task.State).To(Equal("RUNNING"))
			Expect(task.Guid).To(Equal("abc-123"))

			args := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
			Expect(args).To(Equal([]string{
				"curl", "-X", "POST", "-d",
				`{"command":"some-command"}`,
				"/v3/apps/6aef0cf0-c5d5-4ec1-89ae-73971d24241c/tasks"}))
		})

		Context("when there is an error creating the task", func() {
			It("returns an error", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
					nil, errors.New("some-error"))

				_, err := client.CreateTask(cf.App{
					Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
				}, "some-command")
				Expect(err).To(MatchError("failed to create a task: some-error"))
			})
		})

		Context("when invalid json is returned", func() {
			It("returns an error with the contents", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
					`something bad happened`,
				}, nil)

				_, err := client.CreateTask(cf.App{
					Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
				}, "some-command")
				Expect(err).To(MatchError("failed to parse the following api response: something bad happened"))
			})
		})

		Context("when the returned task has errors", func() {
			It("returns the errors", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
					`{`,
					`"guid": "abc-123",`,
					`"state": "RUNNING",`,
					`"errors": [{`,
					`"detail": "some-detail",`,
					`"title": "some-title",`,
					`"code": 404`,
					`}]`,
					`}`,
				}, nil)

				_, err := client.CreateTask(cf.App{
					Guid: "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",
				}, "some-command")
				Expect(err).To(MatchError("failed to create a task: 404: some-title - some-detail"))
			})

		})
	})

	Context("DeleteApp", func() {
		It("deletes an application", func() {
			Expect(client.DeleteApp("some-app")).
				To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal([]string{"delete", "-f", "some-app"}))
		})

		It("returns an error when deleting the application fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"),
			)

			err := client.DeleteApp("some-app")
			Expect(err).To(MatchError(`failed to delete application "some-app": some-error`))
		})
	})

	Context("DumpLogs", func() {
		It("dumps logs for an app", func() {
			client.DumpLogs("some-app")
			Expect(fakeCFPluginAPI.CliCommandArgsForCall(0)).
				To(Equal([]string{"logs", "--recent", "some-app"}))
		})
	})

	Context("GetAppByName", func() {
		It("returns an application by its name", func() {
			fakeCFPluginAPI.GetCurrentSpaceReturns(plugin_models.Space{
				plugin_models.SpaceFields{
					Guid: "some-guid",
					Name: "some-name",
				},
			}, nil)

			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
				`{"resources": [{`,
				`"guid": "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",`,
				`"name": "mysql-migrate"`,
				`}]}`,
			}, nil)

			app, err := client.GetAppByName("some-app")
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Name).To(Equal("mysql-migrate"))
			Expect(app.Guid).To(Equal("6aef0cf0-c5d5-4ec1-89ae-73971d24241c"))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal([]string{
				"curl",
				"/v3/apps?names=some-app&space_guids=some-guid",
			}))
		})

		Context("when there is an error getting the current space", func() {
			BeforeEach(func() {
				fakeCFPluginAPI.GetCurrentSpaceReturns(
					plugin_models.Space{},
					errors.New("bad space"),
				)
			})

			It("returns an error", func() {
				_, err := client.GetAppByName("mysql-migrate")
				Expect(err).To(MatchError("failed to lookup current space: bad space"))
			})
		})

		Context("when there is an error getting an application by name", func() {
			It("returns an error", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
					nil, errors.New("some-error"))

				_, err := client.GetAppByName("mysql-migrate")
				Expect(err).To(MatchError("failed to retrieve an app by name: some-error"))
			})
		})

		Context("when there are no applications returned by name", func() {
			It("returns an error", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
					`{"resources": [`,
					`]}`,
				}, nil)

				_, err := client.GetAppByName("mysql-migrate")
				Expect(err).To(MatchError("failed to retrieve an app by name: none found"))
			})
		})

		Context("when invalid json is returned", func() {
			It("returns an error with the contents", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
					`something bad happened`,
				}, nil)

				_, err := client.GetAppByName("mysql-migrate")
				Expect(err).To(MatchError("failed to parse the following api response: something bad happened"))
			})
		})
	})

	Context("GetTaskByGUID", func() {
		Context("when /v3/tasks returns the task", func() {
			It("Returns the task", func() {
				fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
					`{`,
					`"guid": "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",`,
					`"state": "RUNNING"`,
					`}`,
				}, nil)

				task, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
				Expect(err).NotTo(HaveOccurred())
				Expect(task.State).To(Equal("RUNNING"))
				Expect(task.Guid).To(Equal("6aef0cf0-c5d5-4ec1-89ae-73971d24241c"))

				args := fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(args).To(Equal([]string{"curl", "/v3/tasks/6aef0cf0-c5d5-4ec1-89ae-73971d24241c"}))
				Expect(fakeClock.SleepCallCount()).To(BeZero())
			})

			Context("When invalid json is returned", func() {
				It("Returns an error with the contents", func() {
					fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
						`something bad happened`,
					}, nil)

					_, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
					Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: failed to parse the following api response: something bad happened`))
					Expect(buffer).To(gbytes.Say(`Attempt 2/3: failed to retrieve task by GUID: failed to parse the following api response: something bad happened`))
					Expect(buffer).To(gbytes.Say(`Attempt 3/3: failed to retrieve task by GUID: failed to parse the following api response: something bad happened`))
					Expect(err).To(MatchError("failed to retrieve task by GUID"))
				})
			})
		})

		Context("when /v3/tasks cannot return the task on the first try", func() {
			 Context("Because the oauth token expired", func() {
				 BeforeEach(func() {
					 fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns([]string{
						 `{
						   "errors": [
							 {
							  "detail": "Invalid Auth Token",
							  "title": "CF-InvalidAuthToken",
								"code": 1000
							 }
						   ]
						 }`,
					 }, nil)

				 })

				 It("Tries to auto-refresh the token", func() {
					 client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")

					 Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: \(error code 1000: CF-InvalidAuthToken - Invalid Auth Token\)`))
					 Expect(fakeCFPluginAPI.AccessTokenCallCount()).To(Equal(3)) //TODO: decouple this test
				 })

				 Context("When automatic token refresh succeeds", func() {
					 It("Returns the task after successfully auto-refreshing the token", func() {
						 fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1, []string{
							 `{`,
							 `"guid": "some-guid",`,
							 `"state": "some-state"`,
							 `}`,
						 }, nil)

						 task, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
						 Expect(err).NotTo(HaveOccurred())

						 Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: \(error code 1000: CF-InvalidAuthToken - Invalid Auth Token\)`)) //TODO: decouple this test
						 Expect(buffer).NotTo(gbytes.Say(`Attempt 2/3:`))

						 Expect(task.Guid).To(Equal(`some-guid`))
						 Expect(task.State).To(Equal(`some-state`))

						 Expect(fakeCFPluginAPI.AccessTokenCallCount()).To(Equal(1))
					 })
				 })

				 Context("When automatic token refresh fails", func(){
					 It("Logs a failure", func() {
						 fakeCFPluginAPI.AccessTokenReturns("", errors.New("something"))

						 _, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")
						 Expect(err).To(HaveOccurred())

						 Expect(buffer).To(gbytes.Say(`failed to refresh the access token: something`))
						 Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: \(error code 1000: CF-InvalidAuthToken - Invalid Auth Token\)`)) //TODO: decouple this test
						 Expect(buffer).To(gbytes.Say(`Attempt 2/3: failed to retrieve task by GUID: \(error code 1000: CF-InvalidAuthToken - Invalid Auth Token\)`))
						 Expect(buffer).To(gbytes.Say(`Attempt 3/3: failed to retrieve task by GUID: \(error code 1000: CF-InvalidAuthToken - Invalid Auth Token\)`))
					 })
				 })
			 })

			 Context("Because /v3/tasks fails", func() {
			     Context("And we retry the maximum allowed number of times without success", func() {
					 It("Returns an error", func() {
						 fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
							 nil, errors.New("some-error"))

						 _, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")

						 Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(client.MaxAttempts))
						 Expect(err).To(MatchError("failed to retrieve task by GUID"))
						 Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: some-error`))
						 Expect(buffer).To(gbytes.Say(`Attempt 2/3: failed to retrieve task by GUID: some-error`))
						 Expect(buffer).To(gbytes.Say(`Attempt 3/3: failed to retrieve task by GUID: some-error`))
						 Expect(fakeClock.SleepCallCount()).To(Equal(2))
						 Expect(fakeClock.SleepCallArgs(0)).Should(Equal(2 * time.Second))
						 Expect(fakeClock.SleepCallArgs(1)).Should(Equal(4 * time.Second))
					 })
			     })

			     Context("And we get a successful response after retrying", func() {
			         It("Returns the task", func() {
						 fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(
							 0, nil, errors.New("some-error"))
						 fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(
							 1,
							 []string{
								 `{`,
								 `"guid": "6aef0cf0-c5d5-4ec1-89ae-73971d24241c",`,
								 `"state": "RUNNING"`,
								 `}`,
							 },
							 nil)

						 task, err := client.GetTaskByGUID("6aef0cf0-c5d5-4ec1-89ae-73971d24241c")

						 Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).To(Equal(2))
						 Expect(err).NotTo(HaveOccurred())
						 Expect(task.State).To(Equal("RUNNING"))
						 Expect(task.Guid).To(Equal("6aef0cf0-c5d5-4ec1-89ae-73971d24241c"))
						 Expect(buffer).To(gbytes.Say(`Attempt 1/3: failed to retrieve task by GUID: some-error`))
						 Expect(fakeClock.SleepCallCount()).To(Equal(1))
						 Expect(fakeClock.SleepCallArgs(0)).To(Equal(2 * time.Second))
			         })
			     })
			 })
		})
	})

	Context("PushApp", func() {
		It("pushes an app with --no when give a path and application name", func() {
			Expect(client.PushApp("some-path", "some-app-name")).To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal(
				[]string{
					"push",
					"some-app-name",
					"-b", "binary_buildpack",
					"-u", "none",
					"-c", "sleep infinity",
					"-p", "some-path",
					"--no-route",
					"--no-start",
				},
			))
		})

		It("returns an error when pushing an app fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"),
			)

			err := client.PushApp("some-path", "some-app-name")
			Expect(err).To(MatchError(`failed to push application: some-error`))
		})
	})

	Context("RenameService", func() {
		It("renames a service instance", func() {
			Expect(client.RenameService("some-service-name", "some-new-service-name")).
				To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal([]string{
				"rename-service", "some-service-name", "some-new-service-name",
			}))
		})

		It("returns an error when renaming a service fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"),
			)

			err := client.RenameService("some-service-name", "some-new-service-name")
			Expect(err).To(MatchError(`failed to rename-service "some-service-name" to "some-new-service-name": some-error`))
		})
	})

	Context("RunTask", func() {
		BeforeEach(func() {
			fakeCFPluginAPI.GetCurrentSpaceReturns(plugin_models.Space{
				plugin_models.SpaceFields{
					Guid: "some-guid",
					Name: "some-name",
				},
			}, nil)

			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(0,
				[]string{
					`{"resources": [{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"name": "some-app"`,
					`}]}`,
				}, nil)
		})

		It("runs a task and waits for it to finish", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{
					`{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"state": "RUNNING"`,
					`}`,
				}, nil)

			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(2,
				[]string{
					`{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"state": "SUCCEEDED"`,
					`}`,
				}, nil)

			Expect(client.RunTask("some-app", "some-command")).
				To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(3))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(1)).
				To(
				Equal([]string{
					"curl", "-X", "POST", "-d",
					`{"command":"some-command"}`,
					"/v3/apps/be5077ed-abba-bea7-deb7-50f7ba110000/tasks",
				}))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(2)).
				To(Equal([]string{
				"curl",
				"/v3/tasks/be5077ed-abba-bea7-deb7-50f7ba110000",
			}))
		})

		It("returns an error when looking up an app guid fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(0,
				nil, errors.New("app guid not found error"),
			)

			err := client.RunTask("some-app", "some-command")
			Expect(err).To(MatchError(`Error: failed to retrieve an app by name: app guid not found error`))
		})

		It("returns an error when creating a task fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				nil, errors.New("create task failed"),
			)

			err := client.RunTask("some-app", "some-command")
			Expect(err).To(MatchError(`Error: failed to create a task: create task failed`))
		})

		It("returns an error when waiting for a task fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{
					`{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"state": "RUNNING"`,
					`}`,
				}, nil)

			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-api-error"))

			err := client.RunTask("some-app", "some-command")
			Expect(err).To(MatchError(`Error when waiting for task to complete: failed to retrieve task by GUID`))
		})

		It("returns an error when a tasks finishes with a failed state", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(1,
				[]string{
					`{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"state": "RUNNING"`,
					`}`,
				}, nil)

			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturnsOnCall(2,
				[]string{
					`{`,
					`"guid": "be5077ed-abba-bea7-deb7-50f7ba110000",`,
					`"state": "FAILED"`,
					`}`,
				}, nil)

			err := client.RunTask("some-app", "some-command")
			Expect(err).To(MatchError(`task completed with status "FAILED"`))
		})
	})

	Context("ServiceExists", func() {
		Context("When the service does not exist", func() {
			It("Returns false", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, errors.New("Service does not exist"))

				serviceExists := client.ServiceExists("service-name")
				Expect(serviceExists).To(Equal(false))
			})
		})

		Context("When the service exists", func() {
			It("Returns true", func() {
				fakeCFPluginAPI.GetServiceReturns(plugin_models.GetService_Model{}, nil)

				serviceExists := client.ServiceExists("service-name")
				Expect(serviceExists).To(Equal(true))
			})
		})
	})

	Context("StartApp", func() {
		It("starts an application", func() {
			Expect(client.StartApp("some-app")).
				To(Succeed())

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputCallCount()).
				To(Equal(1))

			Expect(fakeCFPluginAPI.CliCommandWithoutTerminalOutputArgsForCall(0)).
				To(Equal([]string{"start", "some-app"}))
		})

		It("returns an error when starting the application fails", func() {
			fakeCFPluginAPI.CliCommandWithoutTerminalOutputReturns(
				nil, errors.New("some-error"),
			)

			err := client.StartApp("some-app")
			Expect(err).To(MatchError(`failed to start application "some-app": some-error`))
		})
	})
})
