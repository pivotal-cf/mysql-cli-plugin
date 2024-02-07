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

package multisite_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var cfHome, cfConfig, cfPluginsHome, cfReplicationHome string

func TestMultiSite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multisite Suite")
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
	var err error

	cfHome, err = os.MkdirTemp("", "tempCFHome")
	if err != nil {
		panic("Failed to create temp directory:" + err.Error())
	}

	cfDir := filepath.Join(cfHome, ".cf")
	err = os.Mkdir(cfDir, 0755)
	if err != nil {
		fmt.Println("Failed to create .cf directory:", err)
		return
	}

	cfConfig = filepath.Join(cfDir, "config.json")

	cfPluginsHome, err = os.MkdirTemp("", "tempCFPluginsHome")
	if err != nil {
		panic("Failed to create temp directory:" + err.Error())
	}

	cfReplicationHome = filepath.Join(cfPluginsHome, ".set-replication")

	err = os.Setenv("CF_PLUGIN_HOME", cfPluginsHome)
	if err != nil {
		panic("Failed to set CF_PLUGIN_HOME environment variable: " + err.Error())
	}

	err = os.Setenv("CF_HOME", cfHome)
	if err != nil {
		panic("Failed to set CF_HOME environment variable: " + err.Error())
	}

})

var _ = AfterSuite(func() {
	if strings.HasPrefix(cfHome, os.TempDir()) {
		os.RemoveAll(cfHome)
	}
	if strings.HasPrefix(cfPluginsHome, os.TempDir()) {
		os.RemoveAll(cfPluginsHome)
	}
})
