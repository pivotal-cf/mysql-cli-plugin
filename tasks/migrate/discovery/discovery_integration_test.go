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

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/internal/testing/docker"
	. "github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
)

var _ = Describe("Discovery Integration Tests", func() {
	var (
		db             *sql.DB
		mysqlContainer string
	)

	BeforeEach(func() {
		mysqlContainer = "mysql." + uuid.NewString()
		Expect(createMySQLContainer(mysqlContainer)).To(Succeed())

		mysqlPort, err := docker.ContainerPort(mysqlContainer, "3306/tcp")
		Expect(err).NotTo(HaveOccurred())

		dsn := `root@tcp(localhost:` + mysqlPort + `)/`

		db, err = sql.Open("mysql", dsn)
		Expect(err).NotTo(HaveOccurred())

		Eventually(db.Ping, "1m", "1s").Should(Succeed(),
			`Expected MySQL instance to be reachable after 1m, but it was not`,
		)
	})

	AfterEach(func() {
		Expect(docker.RemoveContainer(mysqlContainer)).To(Succeed())
	})

	Context("DiscoverDatabases", func() {
		When("querying the database is successful", func() {
			When("there are databases not in the list of filtered schemas", func() {
				BeforeEach(func() {
					Expect(createDatabases(db, "cf_metadata", "foo", "bar", "baz")).To(Succeed())
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
					Expect(createDatabases(db, "cf_metadata")).To(Succeed())
				})

				It("returns an error", func() {
					_, err := DiscoverDatabases(db)
					Expect(err).To(MatchError("no databases found"))
				})
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

			Expect(createDatabases(db, schemasToMigrate...)).To(Succeed())
			Expect(createInvalidViews(db, expectedInvalidViews)).To(Succeed())
			Expect(createValidViews(db, []View{
				{Schema: "service_instance_db", TableName: "valid_view"},
			})).To(Succeed())
		})

		It("returns invalid views for each specified databases", func() {
			invalidViews, err := DiscoverInvalidViews(db, schemasToMigrate)
			Expect(err).NotTo(HaveOccurred())
			Expect(invalidViews).To(Equal(expectedInvalidViews))
		})

		Context("when a schema exists with special characters in the name", func() {
			BeforeEach(func() {
				Expect(createDatabases(db, QuoteIdentifier("bad_character_`_schema"))).To(Succeed())
				schemasToMigrate = append(schemasToMigrate, "bad_character_`_schema")
			})

			It("is able to migrate views", func() {
				invalidViews, err := DiscoverInvalidViews(db, schemasToMigrate)
				Expect(err).NotTo(HaveOccurred())
				Expect(invalidViews).To(Equal(expectedInvalidViews))
			})
		})
	})
})

func createDatabases(db *sql.DB, names ...string) error {
	var errs error
	for _, name := range names {
		if _, err := db.Exec(`CREATE DATABASE IF NOT EXISTS ` + name); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func createInvalidViews(db *sql.DB, views []View) error {
	for _, view := range views {
		var err error
		schema := QuoteIdentifier(view.Schema)

		_, err = db.Exec(`CREATE TABLE ` + schema + `.t1 (id int, data text)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`CREATE VIEW ` + schema + `.` + view.TableName + ` AS SELECT * FROM ` + schema + `.t1`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`ALTER TABLE ` + schema + `.t1 DROP COLUMN data`)
		if err != nil {
			return err
		}
	}

	return nil
}

func createValidViews(db *sql.DB, views []View) error {
	for _, view := range views {
		schema := QuoteIdentifier(view.Schema)
		_, err := db.Exec(`CREATE VIEW ` + schema + `.` + view.TableName + ` AS SELECT NOW()`)
		if err != nil {
			return err
		}
	}

	return nil
}
