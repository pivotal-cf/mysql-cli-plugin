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
	"github.com/pivotal-cf/mysql-cli-plugin/migrate/migratefakes"
	"github.com/pkg/errors"

	. "github.com/pivotal-cf/mysql-cli-plugin/migrate"
)

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
		migrator = NewMigrator(fakeClient, fakeUnpacker, donorName, recipientName)
	})

	Context("Given valid parameters", func() {
		BeforeEach(func() {
			fakeUnpacker.UnpackStub = func(path string) error {
				Expect(path).To(BeADirectory())
				return nil
			}
		})

		It("Migrates data from the donor instance to the recipient instance", func() {
			err := migrator.MigrateData()

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
			})

			By("Deleting the migration app afterwards", func() {
				Expect(fakeClient.DeleteAppCallCount()).To(Equal(1))
			})

			Expect(err).NotTo(HaveOccurred())
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
		migrator = NewMigrator(fakeClient, fakeUnpacker, donorName, recipientName)
	})

	Context("If a service instance with the donor name appended with '-old' already exists", func() {
		It("Fails", func() {
			fakeClient.RenameServiceReturns(errors.New("The service instance name is taken: some-donor-instance-old"))

			err := migrator.RenameServiceInstances()

			Expect(err).To(MatchError("Error renaming service instance some-donor-instance: The service instance name is taken: some-donor-instance-old"))
		})
	})

	It("Renames the recipient service instance to match the donor service instance's name, and appends '-old' to the donor service instance's name", func() {
		err := migrator.RenameServiceInstances()

		Expect(err).NotTo(HaveOccurred())
		Expect(fakeClient.RenameServiceCallCount()).To(Equal(2))

	})

})
