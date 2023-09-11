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

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/internal/testing/docker"
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
		sourceDB         *sql.DB
		destDB           *sql.DB
		containerNetwork string
		sourceContainer  string
		destContainer    string
		vcapServices     string
		sourceChecksums  string
	)

	BeforeEach(func() {
		var err error
		fixturesPath, err := filepath.Abs("fixtures")
		Expect(err).NotTo(HaveOccurred())

		containerNetwork = "mysql-net." + uuid.NewString()
		Expect(docker.CreateNetwork(containerNetwork)).To(Succeed())

		sourceContainer = "mysql.source." + uuid.NewString()
		destContainer = "mysql.source." + uuid.NewString()

		Expect(docker.CreateContainer(docker.ContainerSpec{
			Name:    sourceContainer,
			Image:   "percona:5.7",
			Network: containerNetwork,
			Env:     []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1", "MYSQL_DATABASE=service_instance_db"},
			Volumes: []string{
				filepath.Join(fixturesPath, "sakila-schema.sql:/docker-entrypoint-initdb.d/sakila-schema.sql"),
			},
		})).Error().NotTo(HaveOccurred())

		Expect(docker.CreateContainer(docker.ContainerSpec{
			Name:    destContainer,
			Image:   "percona:5.7",
			Network: containerNetwork,
			Env:     []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1", "MYSQL_DATABASE=service_instance_db"},
		})).Error().NotTo(HaveOccurred())

		vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, sourceContainer, destContainer, "")

		sourcePort, err := docker.ContainerPort(sourceContainer, "3306/tcp")
		Expect(err).NotTo(HaveOccurred())
		sourceDSN := `root@tcp(localhost:` + sourcePort + `)/`
		sourceDB, err = sql.Open("mysql", sourceDSN)
		Expect(err).NotTo(HaveOccurred())

		destPort, err := docker.ContainerPort(destContainer, "3306/tcp")
		Expect(err).NotTo(HaveOccurred())
		destDSN := `root@tcp(localhost:` + destPort + `)/`
		destDB, err = sql.Open("mysql", destDSN)
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
		Expect(docker.RemoveContainer(sourceContainer)).To(Succeed())
		Expect(docker.RemoveContainer(destContainer)).To(Succeed())
		Expect(docker.RemoveNetwork(containerNetwork)).To(Succeed())
	})

	It("migrates data between the source and destination", func() {
		_, err := docker.Run(
			"--env=VCAP_SERVICES="+vcapServices,
			"--name=migrate.command."+uuid.NewString(),
			"--network="+containerNetwork,
			"--rm",
			"--volume="+migrateTaskBinPath+":/usr/local/bin/migrate",
			"percona:5.7",
			"migrate", "source", "dest",
		)
		Expect(err).NotTo(HaveOccurred())

		destChecksums, err := schemaChecksum(destDB, "sakila")
		Expect(err).NotTo(HaveOccurred())

		Expect(destChecksums).To(Equal(sourceChecksums))
	})

	Context("when resolving mysql host keep failing", func() {
		BeforeEach(func() {
			vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, "non-existing-source", "non-existing-destination", "")
		})

		It("Validate the host and print out failure message", func() {
			output, err := docker.Run(
				"--env=VCAP_SERVICES="+vcapServices,
				"--name=migrate.command."+uuid.NewString(),
				"--network="+containerNetwork,
				"--rm",
				"--volume="+migrateTaskBinPath+":/usr/local/bin/migrate",
				"--tty",
				"percona:5.7",
				"migrate", "source", "dest",
			)
			Expect(err).To(MatchError(`exit status 1`))

			Expect(output).To(SatisfyAll(
				MatchRegexp(`Failed to resolve source host "non-existing-source": Timed out - failing with error: lookup non-existing-source.*: no such host`),
				MatchRegexp(`Failed to resolve destination host "non-existing-destination": Timed out - failing with error: lookup non-existing-destination.*: no such host`),
			))
		})
	})

	Context("when a TLS CA certificate is provided", func() {
		BeforeEach(func() {
			vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, sourceContainer, destContainer, "some-ca-cert")
		})

		It("accepts a --skip-tls-validation option", func() {
			_, err := docker.Run(
				"--env=VCAP_SERVICES="+vcapServices,
				"--name=migrate.command."+uuid.NewString(),
				"--network="+containerNetwork,
				"--rm",
				"--volume="+migrateTaskBinPath+":/usr/local/bin/migrate",
				"--tty",
				"percona:5.7",
				"migrate", "--skip-tls-validation", "source", "dest",
			)
			Expect(err).NotTo(HaveOccurred())

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
