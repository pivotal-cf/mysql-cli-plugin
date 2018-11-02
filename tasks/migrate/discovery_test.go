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
