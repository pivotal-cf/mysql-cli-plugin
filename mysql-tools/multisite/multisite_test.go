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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
)

var (
	MultiSiteSetup multisite.MultiSite
)

var _ = Describe("Setup Replication Tests", func() {
	BeforeEach(func() {
		clearAllMSConfigs()
		MultiSiteSetup = multisite.MultiSite{ReplicationConfigHome: cfReplicationHome}
		initializeCFHomeConfig("initial CF config")
	})

	It("saves, lists and deletes multiple configurations", func() {
		By("saving a config without an error")
		var target = "foundation1"
		err := MultiSiteSetup.SaveConfig(cfConfig, target)
		Expect(err).NotTo(HaveOccurred())

		By("listing results of that save")
		configs, err := MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(1))
		Expect(configs[0]).To(Equal("foundation1"))

		By("saving a second config without an error")
		target = "foundation2"
		err = MultiSiteSetup.SaveConfig(cfConfig, target)
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(2))
		Expect(configs).To(ContainElement("foundation1"))
		Expect(configs).To(ContainElement("foundation2"))

		By("removing the initial config")
		err = MultiSiteSetup.RemoveConfig("foundation1")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(len(configs)).To(Equal(1))
		Expect(configs[0]).To(Equal("foundation2"))

		By("updates an existing config")
		err = MultiSiteSetup.SaveConfig(cfConfig, "foundation2")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(1))
		Expect(configs[0]).To(Equal("foundation2"))

		By("removing the last remaining config")
		err = MultiSiteSetup.RemoveConfig("foundation2")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(len(configs)).To(Equal(0))
	})

	Context("Saving configs", func() {
		When("two configs are saved with the same name", func() {
			It("succeeds and overwrites the first config with the second", func() {
				const firstConfigContents = "initial config"
				initializeCFHomeConfig(firstConfigContents)
				err := MultiSiteSetup.SaveConfig(cfConfig, "reusedFoundation")
				Expect(err).NotTo(HaveOccurred())

				const secondConfigContents = "overwritten config"
				initializeCFHomeConfig(secondConfigContents)
				err = MultiSiteSetup.SaveConfig(cfConfig, "reusedFoundation")
				Expect(err).NotTo(HaveOccurred())

				config, err := savedMSConfigContents("reusedFoundation")
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(secondConfigContents))
			})
		})
	})
})

func initializeCFHomeConfig(configContents string) {
	err := os.WriteFile(cfConfig, []byte(configContents), 0775)
	Expect(err).NotTo(HaveOccurred())
	return
}

func savedMSConfigContents(configName string) (string, error) {
	configFilename := filepath.Join(MultiSiteSetup.ReplicationConfigHome, configName,
		".cf", "config.json")
	contents, err := os.ReadFile(configFilename)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func clearAllMSConfigs() {
	err := os.RemoveAll(MultiSiteSetup.ReplicationConfigHome)
	Expect(err).NotTo(HaveOccurred())
}
