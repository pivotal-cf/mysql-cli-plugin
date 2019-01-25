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
	"os"
	"os/exec"

	"github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/mysql-test-utils/dockertest"
)

const dockerVcapServicesTemplate = `
{
  "p.mysql": [
    {
      "binding_name": null,
      "credentials": {
        "hostname": "127.0.0.1",
        "name": "mysql",
        "password": "",
        "port": %s,
        "username": "root"
      },
      "instance_name": "source",
      "label": "p.mysql",
      "name": "source"
    },
    {
      "credentials": {
        "hostname": "127.0.0.1",
        "name": "service_instance_db",
        "password": "",
        "port": %s,
        "tls": {
          "cert": {
            "ca": "-----BEGIN CERTIFICATE-----\nMIIDCzCCAfOgAwIBAgIUFmlOKyBuBXmtx5vHgAGLSCYts6UwDQYJKoZIhvcNAQEL\nBQAwFTETMBEGA1UEAwwKcHhjX3Rsc19jYTAeFw0xOTAxMjMxOTQwMjdaFw0yMDAx\nMjMxOTQwMjdaMBUxEzARBgNVBAMMCnB4Y190bHNfY2EwggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQDkXKOPl9Fn+igx99aNia4Gvsx2IXgcyYLFUH56iZEE\nn9V3gVS+5p5X4ByOphVxv+UemWcJCvlAyK7T9KKnwXDgUGunYbSqHQXv19eTXeKn\nA2/TkQ/7rXQWkakjo7mtLgJJ7BOtNHw/MDXgEbKM7ifkQLt0zFHmFEzfYrh7VNab\ncZs19IIRFb6BB9oYHQs6oslHQ79Xz6l3gjgyPFlpG/b3RLYqlsKYuK5mxQ/hkCaM\nh13VBITlsdowJiu/9eIn9eeDlxwQBg9VTN/PdQJ41t/7Hw2MpYxTHd04FK7AlGm0\nKLMRM7rRevuxiHXafKNx9tIJsScULOCJ0ssxfCl/Gc1JAgMBAAGjUzBRMB0GA1Ud\nDgQWBBSpWZwmhWpg3p0iqQODgA7EwlXENzAfBgNVHSMEGDAWgBSpWZwmhWpg3p0i\nqQODgA7EwlXENzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQA5\nx+ajEBKBwKylaCf1Lckug6YC5mkmHTfyzdOM95zH5Yp4Xh9oUxRbuGYq2mPv4SbN\nmN54Xf54g357AZuPF4xVX1LwbMNp0El+OXMUZoUncdDKewJaiINMqwYBeUwnE0Q5\nrP+jiGmBd9BZ2IJy9wi+a7JlRiRScSbDiQovCSv+T5AJd8+59x061Jc3mdxmWKtm\nyb1nit5DcDYFuXccvPkp0+L9EJ5mHH4TLXtq6Tf71WB8Tq1HP+/wA5XM5YJmpaAG\no2NdO8TonvPQbBjb3ntE7Kxlh7nT+PG9Mxa5CUch7fXmY7fFJL0ngQIDrBb7fzGY\nQl+cYjMSL4NsMjegQzQp\n-----END CERTIFICATE-----\n"
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
		db             *sql.DB
		mysqlContainer *docker.Container
		vcapServices   string
	)

	BeforeEach(func() {
		var err error
		mysqlContainer, err = createMySQLContainer("mysql")
		Expect(err).NotTo(HaveOccurred())

		mysqlPort := dockertest.HostPort(mySQLDockerPort, mysqlContainer)
		vcapServices = fmt.Sprintf(dockerVcapServicesTemplate, mysqlPort, mysqlPort)
		db, err = dockertest.ContainerDBConnection(mysqlContainer, mySQLDockerPort)
		Expect(err).NotTo(HaveOccurred())

		Eventually(db.Ping, "1m", "1s").Should(Succeed(),
			`Expected MySQL instance to be reachable after 1m, but it was not`,
		)

		_, err = db.Exec(`CREATE DATABASE foo`)
		Expect(err).NotTo(HaveOccurred())

		_, err = db.Exec(`CREATE DATABASE service_instance_db`)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if mysqlContainer != nil {
			dockertest.RemoveContainer(dockerClient, mysqlContainer)
		}
	})

	It("accepts a --skip-tls-validation option", func() {
		cmd := exec.Command(migrateTaskBinPath, "-skip-tls-validation", "source", "dest")
		cmd.Env = append(os.Environ(), "VCAP_SERVICES="+vcapServices)
		cmd.Stdout = GinkgoWriter
		cmd.Stderr = GinkgoWriter
		Expect(cmd.Run()).To(Succeed())
	})
})
