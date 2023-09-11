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

package discovery_test

import (
	"database/sql"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
)

var _ = Describe("Discovery Unit Tests", func() {
	var (
		mockDB *sql.DB
		mock   sqlmock.Sqlmock
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Context("DiscoverDatabases", func() {
		When("listing databases fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SHOW DATABASES`).
					WillReturnError(errors.New(`some database error`))
			})

			It("returns an error", func() {
				_, err := DiscoverDatabases(mockDB)
				Expect(err).To(MatchError("failed to query the database: some database error"))
			})
		})

		When("we are unable to parse the list of databases", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SHOW DATABASES`).
					WillReturnRows(sqlmock.NewRows([]string{"Database"}).
						AddRow("information_schema").
						AddRow("mysql").
						AddRow("performance_schema").
						AddRow("cf_metadata").
						AddRow("sys").
						AddRow("foo").
						AddRow("bar").
						AddRow("baz").
						RowError(2, fmt.Errorf("some error")),
					)
			})

			It("returns an error", func() {
				_, err := DiscoverDatabases(mockDB)
				Expect(err).To(MatchError("failed to parse the list of databases: some error"))
			})
		})

		When("the list of databases has an invalid data type", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SHOW DATABASES`).
					WillReturnRows(sqlmock.NewRows([]string{"Database"}).
						AddRow("information_schema").
						AddRow("mysql").
						AddRow("performance_schema").
						AddRow("cf_metadata").
						AddRow("sys").
						AddRow(nil).
						AddRow("bar").
						AddRow("baz"),
					)
			})

			It("returns an error", func() {
				_, err := DiscoverDatabases(mockDB)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to scan the list of databases"))
			})
		})
	})

	Context("DiscoverInvalidViews", func() {
		var schemasToMigrate []string

		BeforeEach(func() {
			schemasToMigrate = []string{
				"service_instance_db",
			}
		})

		Context("when querying views fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnError(errors.New("failed to query views"))
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(mockDB, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to query views"))
			})
		})

		Context("when preparing the list of views fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow("valid_view").
						AddRow("invalid_view").
						RowError(0, errors.New("failed to prepare view")),
					)
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(mockDB, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to prepare the list of views"))
			})
		})

		Context("when scanning the list of views fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow("valid_view").
						AddRow(nil),
					)
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(mockDB, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to scan the list of views"))
			})
		})

		Context("when DiscoverInvalidViews returns non MySQLError", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow(`some_view_name`))

				mock.ExpectExec("SHOW FIELDS FROM `some_view_name` IN `service_instance_db`").
					WillReturnError(errors.New(`some network error`))
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(mockDB, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(`Unexpected error when validating view "service_instance_db"."some_view_name": some network error`))
			})
		})
	})
})
