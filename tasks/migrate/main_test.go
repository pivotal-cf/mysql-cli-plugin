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

package main_test

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers/dockertest"
)

const dockerVcapServicesTemplate = `
{
  "p.mysql": [
    {
      "binding_name": null,
      "credentials": {
        "hostname": %q,
        "name": "service_instance_db",
        "password": "",
        "port": 3306,
        "username": "root"
      },
      "instance_name": "source",
      "label": "p.mysql",
      "name": "source"
    },
    {
      "credentials": {
        "hostname": %q,
        "name": "service_instance_db",
        "password": "",
        "port": 3306,
        "tls": {
          "cert": {
            "ca": %q
          }
        },
        "username": "root"
      },
      "instance_name": "dest",
      "label": "p.mysql",
      "name": "dest"
    }
  ]
}
`

var _ = Describe("Migrate Task", func() {
	var (
		sourceDB        *sql.DB
		destDB          *sql.DB
		sourceContainer *docker.Container
		destContainer   *docker.Container
		vcapServices    string
		sourceChecksums string
	)

	BeforeEach(func() {
		var err error
		fixturesPath, err := filepath.Abs("fixtures")
		Expect(err).NotTo(HaveOccurred())

		sourceContainer, err = createMySQLContainer(
			"mysql.source",
			dockertest.AddEnvVars(`MYSQL_DATABASE=service_instance_db`),
			dockertest.AddBinds(
				filepath.Join(fixturesPath, "sakila-schema.sql:/docker-entrypoint-initdb.d/sakila-schema.sql"),
			),
		)
		Expect(err).NotTo(HaveOccurred())

		destContainer, err = createMySQLContainer(
			"mysql.dest",
			dockertest.AddEnvVars(`MYSQL_DATABASE=service_instance_db`),
		)
		Expect(err).NotTo(HaveOccurred())

		vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, "mysql.source."+sessionID, "mysql.dest."+sessionID, "")

		sourceDB, err = dockertest.ContainerDBConnection(sourceContainer, mySQLDockerPort)
		Expect(err).NotTo(HaveOccurred())
		destDB, err = dockertest.ContainerDBConnection(destContainer, mySQLDockerPort)
		Expect(err).NotTo(HaveOccurred())

		Eventually(sourceDB.Ping, "1m", "1s").Should(Succeed(),
			`Expected MySQL instance to be reachable after 1m, but it was not`,
		)
		Eventually(destDB.Ping, "1m", "1s").Should(Succeed(),
			`Expected MySQL instance to be reachable after 1m, but it was not`,
		)

		sourceChecksums, err = schemaChecksum(sourceDB, "sakila")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if sourceContainer != nil {
			dockertest.RemoveContainer(dockerClient, sourceContainer)
		}

		if destContainer != nil {
			dockertest.RemoveContainer(dockerClient, destContainer)
		}
	})

	It("migrates data between the source and destination", func() {
		_, exitStatus, err := runCommand(
			"migrate.command",
			dockertest.AddBinds(
				migrateTaskBinPath+":/usr/local/bin/migrate",
			),
			dockertest.WithCmd("migrate", "source", "dest"),
			dockertest.AddEnvVars("VCAP_SERVICES="+vcapServices),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(exitStatus).To(Equal(0))

		destChecksums, err := schemaChecksum(destDB, "sakila")
		Expect(err).NotTo(HaveOccurred())

		Expect(destChecksums).To(Equal(sourceChecksums))
	})

	Context("when resolving mysql host keep failing", func() {
		BeforeEach(func() {
			vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, "nonexist-source", "nonexist-destination", "")
		})

		It("Validate the host and print out failure message", func() {
			output, exitStatus, err := runCommand(
				"migrate.command",
				dockertest.AddBinds(
					migrateTaskBinPath+":/usr/local/bin/migrate",
				),
				dockertest.WithCmd("migrate", "source", "dest"),
				dockertest.AddEnvVars("VCAP_SERVICES="+vcapServices),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(exitStatus).To(Equal(1))
			Eventually(output).Should(gbytes.Say(`Failed to resolve source host "nonexist-source": Timed out - failing with error: lookup nonexist-source.*: no such host`))
			Eventually(output).Should(gbytes.Say(`Failed to resolve destination host "nonexist-destination": Timed out - failing with error: lookup nonexist-destination.*: no such host`))
		})
	})

	Context("when a TLS CA certificate is provided", func() {
		BeforeEach(func() {
			vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, "mysql.source."+sessionID, "mysql.dest."+sessionID, "some-ca-cert")
		})

		It("accepts a --skip-tls-validation option", func() {
			_, exitStatus, err := runCommand(
				"migrate.command",
				dockertest.AddBinds(
					migrateTaskBinPath+":/usr/local/bin/migrate",
				),
				dockertest.WithCmd("migrate", "-skip-tls-validation", "source", "dest"),
				dockertest.AddEnvVars("VCAP_SERVICES="+vcapServices),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(exitStatus).To(Equal(0))

			destChecksums, err := schemaChecksum(destDB, "sakila")
			Expect(err).NotTo(HaveOccurred())

			Expect(destChecksums).To(Equal(sourceChecksums))
		})
	})
})

func schemaChecksum(db *sql.DB, schemaName string) (string, error) {
	rows, err := db.Query(`SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?`, schemaName)
	if err != nil {
		return "", err
	}

	var result []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return "", err
		}
		checksum, err := tableChecksum(db, schemaName, tableName)
		if err != nil {
			return "", err
		}
		result = append(result, checksum)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return strings.Join(result, "\n"), nil
}

func tableChecksum(db *sql.DB, schemaName, tableName string) (string, error) {
	var (
		unused   string
		checksum sql.NullString
	)

	if err := db.QueryRow(`CHECKSUM TABLE `+schemaName+"."+tableName).Scan(&unused, &checksum); err != nil {
		return "", err
	}

	result := schemaName + "." + tableName + ":"
	if checksum.Valid {
		result += checksum.String
	} else {
		// MySQL VIEWs will always generate a NULL checksum
		result += "N/A"
	}

	return result, nil
}
