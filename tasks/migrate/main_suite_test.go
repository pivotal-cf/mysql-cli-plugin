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
	"io"
	"log"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal/mysql-test-utils/dockertest"
)

func TestMigrate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migrate Suite")
}

const (
	mysqlDockerImage             = "percona:5.7"
	mySQLDockerPort  docker.Port = "3306/tcp"
)

var (
	migrateTaskBinPath string
	dockerClient       *docker.Client
	dockerNetwork      *docker.Network
	sessionID          string
)

var _ = BeforeSuite(func() {
	format.TruncatedDiff = false

	log.SetOutput(GinkgoWriter)
	_ = mysql.SetLogger(log.New(GinkgoWriter, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile))

	var err error
	dockerClient, err = docker.NewClientFromEnv()
	Expect(err).NotTo(HaveOccurred())
	Expect(PullImage(dockerClient, mysqlDockerImage)).To(Succeed())

	Expect(os.Setenv("TMPDIR", "/tmp")).To(Succeed())

	migrateTaskBinPath, err = gexec.BuildWithEnvironment(
		"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate",
		[]string{
			"GOOS=linux",
			"GOARCH=amd64",
			"CGO_ENABLED=0",
		},
	)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	sessionID = uuid.New().String()

	var err error
	dockerNetwork, err = CreateNetwork(dockerClient, "mysql-net."+sessionID)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(dockerClient.RemoveNetwork(dockerNetwork.ID)).To(Succeed())
})

func createMySQLContainer(name string, extraOptions ...ContainerOption) (*docker.Container, error) {
	options := []ContainerOption{
		AddExposedPorts(mySQLDockerPort),
		WithImage(mysqlDockerImage),
		AddEnvVars(
			"MYSQL_ALLOW_EMPTY_PASSWORD=1",
		),
		WithNetwork(dockerNetwork),
	}

	options = append(options, extraOptions...)

	return RunContainer(
		dockerClient,
		name+"."+sessionID,
		options...,
	)
}

func runCommand(containerName string, extraOptions ...ContainerOption) (*gbytes.Buffer, int, error) {
	options := []ContainerOption{
		AddExposedPorts(mySQLDockerPort),
		WithImage(mysqlDockerImage),
		WithNetwork(dockerNetwork),
	}

	options = append(options, extraOptions...)

	container, err := RunContainer(
		dockerClient,
		containerName+"."+sessionID,
		options...,
	)
	if err != nil {
		return nil, -1, err
	}
	defer RemoveContainer(dockerClient, container)

	output := gbytes.NewBuffer()

	go func() {
		err := dockerClient.AttachToContainer(docker.AttachToContainerOptions{
			Container:    container.ID,
			OutputStream: io.MultiWriter(output, GinkgoWriter),
			ErrorStream:  io.MultiWriter(output, GinkgoWriter),
			Stream:       true,
			Stdout:       true,
			Stderr:       true,
		})
		if err != nil {
			log.Printf("error when streaming logs: %v", err)
		}
	}()

	exitStatus, err := dockerClient.WaitContainer(container.ID)

	return output, exitStatus, err
}
