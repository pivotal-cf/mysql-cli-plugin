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

package migrate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate/migratefakes"
	"github.com/pkg/errors"

	. "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
)

var _ = Describe("CheckServiceExists", func() {
	var (
		donorInstanceName string
		fakeClient        *migratefakes.FakeClient
		migrator          *Migrator
	)

	BeforeEach(func() {
		donorInstanceName = "some-donor-instance"
		fakeClient = new(migratefakes.FakeClient)
		migrator = NewMigrator(fakeClient, nil)
	})

	It("Confirms we have an existing donor service instance", func() {
		fakeClient.ServiceExistsReturns(true)
		err := migrator.CheckServiceExists(donorInstanceName)

		Expect(err).NotTo(HaveOccurred())
		Expect(fakeClient.ServiceExistsCallCount()).To(Equal(1))
	})

	Context("When the donor service instance doesn't exist", func() {
		It("fails", func() {
			fakeClient.ServiceExistsReturns(false)
			err := migrator.CheckServiceExists(donorInstanceName)

			Expect(err).To(MatchError("Service instance some-donor-instance not found"))
			Expect(fakeClient.ServiceExistsCallCount()).To(Equal(1))
		})
	})
})

var _ = Describe("CreateAndConfigureServiceInstance", func() {
	var (
		planType      string
		recipientName string
		fakeClient    *migratefakes.FakeClient
		fakeUnpacker  *migratefakes.FakeUnpacker
		migrator      *Migrator
	)

	BeforeEach(func() {
		planType = "plan-type"
		recipientName = "some-recipient-instance"
		fakeClient = new(migratefakes.FakeClient)
		fakeUnpacker = new(migratefakes.FakeUnpacker)
		migrator = NewMigrator(fakeClient, fakeUnpacker)
	})

	It("Creates a new service instance and updates it to enable TLS", func() {
		hostnames := []string{
			"some-leader-hostname",
			"some-follower-hostname",
		}
		fakeClient.GetHostnamesReturns(hostnames, nil)
		err := migrator.CreateAndConfigureServiceInstance(planType, recipientName)

		Expect(err).NotTo(HaveOccurred())

		By("Creating a service instance", func() {
			Expect(fakeClient.CreateServiceInstanceCallCount()).To(Equal(1))
		})

		By("Obtaining its hostname", func() {
			Expect(fakeClient.GetHostnamesCallCount()).To(Equal(1))
			Expect(fakeClient.GetHostnamesArgsForCall(0)).To(Equal(recipientName))
		})

		By("Updating the service to enable TLS on its hostname", func() {
			Expect(fakeClient.UpdateServiceConfigCallCount()).To(Equal(1))
			updatedServiceName, updatedJSONParams := fakeClient.UpdateServiceConfigArgsForCall(0)
			Expect(updatedServiceName).To(Equal(recipientName))
			Expect(updatedJSONParams).To(Equal(
				`{"enable_tls": ["some-leader-hostname","some-follower-hostname"]}`,
			))
		})
	})

	Context("When we cannot create a new service instance", func() {
		BeforeEach(func() {
			fakeClient.CreateServiceInstanceReturns(errors.New("create service failed"))
		})

		It("Fails", func() {
			err := migrator.CreateAndConfigureServiceInstance(planType, recipientName)

			Expect(err).To(MatchError("Error creating service instance: create service failed"))
			Expect(fakeClient.CreateServiceInstanceCallCount()).To(Equal(1))
		})
	})

	Context("When we cannot obtain the new service instance's hostname", func() {
		BeforeEach(func() {
			fakeClient.GetHostnamesReturns(nil, errors.New("get hostname failed"))
		})

		It("Fails", func() {
			err := migrator.CreateAndConfigureServiceInstance(planType, recipientName)

			Expect(err).To(MatchError("Error obtaining hostname for new service instance: get hostname failed"))
			Expect(fakeClient.CreateServiceInstanceCallCount()).To(Equal(1))
			Expect(fakeClient.GetHostnamesCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteServiceInstanceCallCount()).To(Equal(0))
			Expect(fakeClient.UpdateServiceConfigCallCount()).To(Equal(0))

		})
	})
	Context("When we cannot update a new service instance", func() {
		BeforeEach(func() {
			fakeClient.UpdateServiceConfigReturns(errors.New("update service failed"))
		})

		It("does not fail", func() {
			err := migrator.CreateAndConfigureServiceInstance(planType, recipientName)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeClient.CreateServiceInstanceCallCount()).To(Equal(1))
			Expect(fakeClient.GetHostnamesCallCount()).To(Equal(1))
			Expect(fakeClient.DeleteServiceInstanceCallCount()).To(Equal(0))
			Expect(fakeClient.UpdateServiceConfigCallCount()).To(Equal(1))
		})
	})
})

var _ = Describe("MigrateData", func() {
	var (
		donorName     string
		recipientName string
		fakeClient    *migratefakes.FakeClient
		fakeUnpacker  *migratefakes.FakeUnpacker
		migrator      *Migrator
	)

	BeforeEach(func() {
		donorName = "some-donor-instance"
		recipientName = "some-recipient-instance"
		fakeClient = new(migratefakes.FakeClient)
		fakeUnpacker = new(migratefakes.FakeUnpacker)
		migrator = NewMigrator(fakeClient, fakeUnpacker)
	})

	Context("Given valid parameters", func() {
		BeforeEach(func() {
			fakeUnpacker.UnpackStub = func(path string) error {
				Expect(path).To(BeADirectory())
				return nil
			}
		})

		It("Migrates data from the donor instance to the recipient instance", func() {
			err := migrator.MigrateData(donorName, recipientName, true)

			By("Unpacking the migration app", func() {
				Expect(fakeUnpacker.UnpackCallCount()).To(Equal(1))
			})

			By("Pushing the migration app to cf", func() {
				Expect(fakeClient.PushAppCallCount()).To(Equal(1))
			})

			By("Binding the migration app to the donor and recipient instances", func() {
				Expect(fakeClient.BindServiceCallCount()).To(Equal(2))
			})

			By("Starting the migration app", func() {
				Expect(fakeClient.StartAppCallCount()).To(Equal(1))
			})

			By("Running the migration app", func() {
				Expect(fakeClient.RunTaskCallCount()).To(Equal(1))
				migrateAppName, migrateTaskCmd := fakeClient.RunTaskArgsForCall(0)
				Expect(migrateAppName).To(HavePrefix(`migrate-app-`))
				Expect(migrateTaskCmd).To(MatchRegexp(`^migrate %s %s`, donorName, recipientName))
			})

			By("Deleting the migration app afterwards", func() {
				Expect(fakeClient.DeleteAppCallCount()).To(Equal(1))
			})

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when told to not cleanup", func() {
			It("keeps the application around for inspection", func() {
				err := migrator.MigrateData(donorName, recipientName, false)

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeClient.DeleteAppCallCount()).To(BeZero())
			})
		})
	})
})

var _ = Describe("RenameServiceInstances", func() {
	var (
		donorName     string
		recipientName string
		fakeClient    *migratefakes.FakeClient
		fakeUnpacker  *migratefakes.FakeUnpacker
		migrator      *Migrator
	)

	BeforeEach(func() {
		donorName = "some-donor-instance"
		recipientName = "some-recipient-instance"
		fakeClient = new(migratefakes.FakeClient)
		fakeUnpacker = new(migratefakes.FakeUnpacker)
		migrator = NewMigrator(fakeClient, fakeUnpacker)
	})

	Context("When renaming the donor instance fails", func() {
		It("tells the operator what command to run to complete the migration", func() {
			fakeClient.RenameServiceReturnsOnCall(0,
				errors.New("The service instance name is taken: some-donor-instance-old"))

			err := migrator.RenameServiceInstances(donorName, recipientName)

			renameError := `Error renaming service instance some-donor-instance: The service instance name is taken: some-donor-instance-old.
The migration of data from some-donor-instance to a newly created service instance with name: some-donor-instance-new has successfully completed.

In order to complete the data migration, please run 'cf rename-service some-donor-instance some-donor-instance-old' and
'cf rename-service some-donor-instance-new some-donor-instance' to complete the migration process.`
			Expect(err).To(MatchError(renameError))
		})
	})

	Context("When renaming the recipient instance fails", func() {
		It("tells the operator what command to run to complete the migration", func() {
			fakeClient.RenameServiceReturnsOnCall(1,
				errors.New("The service instance name is taken: some-donor-instance"))

			err := migrator.RenameServiceInstances(donorName, recipientName)

			renameError := `Error renaming service instance some-donor-instance: The service instance name is taken: some-donor-instance.
The migration of data from some-donor-instance to a newly created service instance with name: some-donor-instance-new has successfully completed.

In order to complete the data migration, please run 'cf rename-service some-donor-instance-new some-donor-instance' to complete the migration process.`
			Expect(err).To(MatchError(renameError))
		})
	})

	It("Renames the recipient service instance to match the donor service instance's name, and appends '-old' to the donor service instance's name", func() {
		err := migrator.RenameServiceInstances(donorName, recipientName)

		Expect(err).NotTo(HaveOccurred())
		Expect(fakeClient.RenameServiceCallCount()).To(Equal(2))

		previousDonorName, newDonorName := fakeClient.RenameServiceArgsForCall(0)
		Expect(previousDonorName).To(Equal(donorName))
		Expect(newDonorName).To(Equal("some-donor-instance-old"))

		previousRecipientName, newRecipientName := fakeClient.RenameServiceArgsForCall(1)
		Expect(previousRecipientName).To(Equal(recipientName))
		Expect(newRecipientName).To(Equal(donorName))
	})
})

var _ = Describe("CleanupOnError", func() {
	var (
		recipientServiceInstance string
		fakeClient               *migratefakes.FakeClient
		migrator                 *Migrator
	)

	BeforeEach(func() {
		recipientServiceInstance = "some-recipient-instance"
		fakeClient = new(migratefakes.FakeClient)
		migrator = NewMigrator(fakeClient, nil)
	})

	It("deletes the service instance", func() {
		err := migrator.CleanupOnError(recipientServiceInstance)

		Expect(err).NotTo(HaveOccurred())
		Expect(fakeClient.DeleteServiceInstanceCallCount()).To(Equal(1))
	})
})
