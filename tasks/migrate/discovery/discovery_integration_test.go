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
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/mysql"

	"github.com/pivotal/mysql-test-utils/dockertest"
)

var _ = Describe("Discovery Integration Tests", func() {
	var (
		db             *sql.DB
		mysqlContainer *docker.Container
		credentials    mysql.Credentials
	)

	BeforeEach(func() {
		var err error
		mysqlContainer, err = createMySQLContainer("mysql")
		Expect(err).NotTo(HaveOccurred())

		db, err = dockertest.ContainerDBConnection(mysqlContainer, mySQLDockerPort)
		Expect(err).NotTo(HaveOccurred())

		Eventually(db.Ping, "1m", "1s").Should(Succeed(),
			`Expected MySQL instance to be reachable after 1m, but it was not`,
		)

		hostMySQLPort, err := strconv.Atoi(dockertest.HostPort("3306/tcp", mysqlContainer))
		Expect(err).NotTo(HaveOccurred())

		credentials = mysql.Credentials{
			Hostname: "127.0.0.1",
			Name:     "",
			Password: "",
			Port:     hostMySQLPort,
			Username: "root",
		}
	})

	AfterEach(func() {
		if mysqlContainer != nil {
			Expect(dockertest.RemoveContainer(dockerClient, mysqlContainer)).To(Succeed())
		}
	})

	Context("DiscoverDatabases", func() {
		When("querying the database is successful", func() {
			When("there are databases not in the list of filtered schemas", func() {
				BeforeEach(func() {
					Expect(createDatabases(db, "foo\tbar", "cf_metadata", "foo", "bar", "baz")).To(Succeed())
				})

				It("returns a slice of discovered database names", func() {
					Expect(DiscoverDatabases(credentials)).To(ConsistOf(
						"foo",
						"bar",
						"baz",
						"foo\tbar",
					), `Expected DiscoverDatabases to find 3 user databases, but it did not`)
				})
			})

			When("there are no databases not in the list of filtered schemas", func() {
				BeforeEach(func() {
					Expect(createDatabases(db, "cf_metadata")).To(Succeed())
				})

				It("returns an error", func() {
					_, err := DiscoverDatabases(credentials)
					Expect(err).To(MatchError("no databases found"))
				})
			})
		})
	})

	Context("DiscoverInvalidViews", func() {
		When("there are no invalid views", func() {
			It("returns an empty list without error", func() {
				invalidViews, err := DiscoverInvalidViews(credentials)
				Expect(err).NotTo(HaveOccurred())
				Expect(invalidViews).To(BeEmpty())
			})
		})

		When("there are invalid views in the source database", func() {
			var (
				schemasToMigrate []string
			)

			BeforeEach(func() {
				schemasToMigrate = []string{
					"service_instance_db",
					"custom_user_db",
				}

				Expect(createDatabases(db, schemasToMigrate...)).To(Succeed())

				Expect(createInvalidViews(db, []View{
					{Schema: "service_instance_db", TableName: "invalid_view"},
					{Schema: "custom_user_db", TableName: "invalid_view"},
				})).To(Succeed())

				Expect(createValidViews(db, []View{
					{Schema: "service_instance_db", TableName: "valid_view"},
				})).To(Succeed())
			})

			It("returns a list of qualified view names", func() {
				invalidViews, err := DiscoverInvalidViews(credentials)
				Expect(err).NotTo(HaveOccurred())
				Expect(invalidViews).To(ConsistOf([]string{
					"service_instance_db.invalid_view",
					"custom_user_db.invalid_view",
				}))
			})

			Context("when a schema exists with special characters in the name", func() {
				BeforeEach(func() {
					Expect(createDatabases(db, quoteIdentifier("bad_character_`_schema"))).To(Succeed())
					schemasToMigrate = append(schemasToMigrate, "bad_character_`_schema")
				})

				It("is able to migrate views", func() {
					invalidViews, err := DiscoverInvalidViews(credentials)
					Expect(err).NotTo(HaveOccurred())
					Expect(invalidViews).To(ConsistOf([]string{
						"service_instance_db.invalid_view",
						"custom_user_db.invalid_view",
					}))
				})
			})
		})
	})
})

func createDatabases(db *sql.DB, names ...string) error {
	var errs error
	for _, name := range names {
		if _, err := db.Exec(`CREATE DATABASE IF NOT EXISTS ` + quoteIdentifier(name)); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

type View struct {
	Schema    string
	TableName string
}

func quoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

func createInvalidViews(db *sql.DB, views []View) error {

	for _, view := range views {
		var err error
		schema := quoteIdentifier(view.Schema)

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
		schema := quoteIdentifier(view.Schema)
		_, err := db.Exec(`CREATE VIEW ` + schema + `.` + view.TableName + ` AS SELECT NOW()`)
		if err != nil {
			return err
		}
	}

	return nil
}
