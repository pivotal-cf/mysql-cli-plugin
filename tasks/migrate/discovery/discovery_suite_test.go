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
	"log"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/mysql-cli-plugin/test_helpers/dockertest"
)

func TestDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Integration Test Suite")
}

const (
	mysqlDockerImage             = "mariadb:10.1"
	mySQLDockerPort  docker.Port = "3306/tcp"
)

var (
	dockerClient *docker.Client
	sessionID    string
)

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
	_ = mysql.SetLogger(log.New(GinkgoWriter, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile))

	var err error
	dockerClient, err = docker.NewClientFromEnv()
	Expect(err).NotTo(HaveOccurred())
	Expect(PullImage(dockerClient, mysqlDockerImage)).To(Succeed())

})

var _ = BeforeEach(func() {
	sessionID = uuid.New().String()
})

func createMySQLContainer(name string) (*docker.Container, error) {
	return RunContainer(
		dockerClient,
		name+"."+sessionID,
		AddExposedPorts(mySQLDockerPort),
		WithImage(mysqlDockerImage),
		AddEnvVars(
			"MYSQL_ALLOW_EMPTY_PASSWORD=1",
		),
	)
}
