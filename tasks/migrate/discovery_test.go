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

package main

import (
	"database/sql"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var _ = Describe("Discovery", func() {
	var (
		db   *sql.DB
		mock sqlmock.Sqlmock
	)

	BeforeEach(func() {
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("DiscoverDatabases", func() {
		When("querying the database is successful", func() {
			When("there are databases not in the list of filtered schemas", func() {
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
							AddRow("baz"),
						)
				})

				It("returns a slice of discovered database names", func() {
					Expect(DiscoverDatabases(db)).To(ConsistOf(
						"foo",
						"bar",
						"baz",
					), `Expected DiscoverDatabases to find 3 user databases, but it did not`)
				})
			})

			When("there are no databases not in the list of filtered schemas", func() {
				BeforeEach(func() {
					mock.ExpectQuery(`SHOW DATABASES`).
						WillReturnRows(sqlmock.NewRows([]string{"Database"}).
							AddRow("information_schema").
							AddRow("mysql").
							AddRow("performance_schema").
							AddRow("cf_metadata").
							AddRow("sys"),
						)
				})

				It("returns an error", func() {
					_, err := DiscoverDatabases(db)
					Expect(err).To(MatchError("no databases found"))
				})
			})
		})

		When("querying the database fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SHOW DATABASES`).WillReturnError(fmt.Errorf("database error"))
			})

			It("returns an error", func() {
				_, err := DiscoverDatabases(db)
				Expect(err).To(MatchError("failed to query the database: database error"))
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
				_, err := DiscoverDatabases(db)
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
				_, err := DiscoverDatabases(db)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to scan the list of databases"))
			})
		})
	})

	Context("DiscoverInvalidViews", func() {
		var (
			schemasToMigrate     []string
			expectedInvalidViews []View
		)

		BeforeEach(func() {
			schemasToMigrate = []string{
				"service_instance_db",
				"custom_user_db",
			}
			expectedInvalidViews = []View{
				{Schema: "service_instance_db", TableName: "invalid_view"},
				{Schema: "custom_user_db", TableName: "invalid_view"},
			}
		})

		It("returns invalid views for each specified databases", func() {
			for _, schema := range schemasToMigrate {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs(schema).
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow("valid_view").
						AddRow("invalid_view"),
					)

				mock.ExpectExec(`SHOW FIELDS FROM \? FROM \?`).
					WithArgs("valid_view", schema).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(`SHOW FIELDS FROM \? FROM \?`).
					WithArgs("invalid_view", schema).
					WillReturnError(errors.New("invalid view"))
			}

			invalidViews, err := DiscoverInvalidViews(db, schemasToMigrate)
			Expect(err).NotTo(HaveOccurred())
			Expect(invalidViews).To(Equal(expectedInvalidViews))
			Expect(mock.ExpectationsWereMet()).To(Succeed())
		})

		Context("when querying views fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnError(errors.New("failed to query views"))
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(db, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to query views"))
			})
		})

		Context("when preparing the list of views fails", func() {
			BeforeEach(func() {
				schemasToMigrate = []string{
					"service_instance_db",
				}

				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow("valid_view").
						AddRow("invalid_view").
						RowError(0, errors.New("failed to prepare view")),
					)
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(db, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to prepare the list of views"))
			})
		})

		Context("when scanning the list of views fails", func() {
			BeforeEach(func() {
				schemasToMigrate = []string{
					"service_instance_db",
				}

				mock.ExpectQuery(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`).
					WithArgs("service_instance_db").
					WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
						AddRow("valid_view").
						AddRow(nil),
					)
			})

			It("returns the error", func() {
				_, err := DiscoverInvalidViews(db, schemasToMigrate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to scan the list of views"))
			})
		})
	})
})
