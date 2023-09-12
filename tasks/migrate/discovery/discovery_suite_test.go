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

	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/internal/testing/docker"
)

func TestDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Discovery Test Suite")
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
	_ = mysql.SetLogger(log.New(GinkgoWriter, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile))
})

func createMySQLContainer(name string) error {
	_, err := docker.CreateContainer(docker.ContainerSpec{
		Name:  name,
		Image: "mariadb:10.1",
		Env:   []string{"MYSQL_ALLOW_EMPTY_PASSWORD=1"},
		Args:  nil,
	})

	return err
}
