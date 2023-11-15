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

const cfInitialConfig = `{
	"OrganizationFields": {
		"GUID": "ignored-Org-GUID",
		"Name": "InitialOrgName"
	},
	"SpaceFields": {
		"GUID": "ignored-Space-GUID",
		"Name": "InitialSpaceName"
	},
	"Target": "https://api.sys.initial-domain.com"
}`
const secondOrgName = "SecondOrgName"
const secondSpaceName = "SecondSpaceName"
const secondAPI = "https://api.sys.second-domain.com"
const cfSecondConfig = `{
	"OrganizationFields": {
		"GUID": "ignored-Org-GUID",
		"Name": "` + secondOrgName + `"
	},
	"SpaceFields": {
		"GUID": "ignored-Space-GUID",
		"Name": "` + secondSpaceName + `"
	},
	"Target": "` + secondAPI + `"
}`

var (
	MultiSiteSetup multisite.MultiSite
)

var _ = Describe("Setup Replication Tests", func() {
	returnedConfigSummary := multisite.ConfigCoreSubset{
		OrganizationFields: struct {
			Name string
		}{
			Name: "InitialOrgName",
		},
		SpaceFields: struct {
			Name string
		}{
			Name: "InitialSpaceName",
		},
		Target: "https://api.sys.initial-domain.com",
		Name:   "foundation1",
	}
	BeforeEach(func() {
		clearAllMSConfigs()
		MultiSiteSetup = multisite.MultiSite{ReplicationConfigHome: cfReplicationHome}
		initializeCFHomeConfig(cfInitialConfig)
	})

	It("saves, lists and deletes multiple configurations", func() {
		By("saving a config without an error")
		var target = "foundation1"

		_, err := MultiSiteSetup.SaveConfig(cfConfig, target)
		Expect(err).NotTo(HaveOccurred())

		By("listing results of that save")
		configs, err := MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(1))

		Expect(*configs[0]).To(Equal(returnedConfigSummary))

		By("saving a second config without an error")
		target = "foundation2"
		_, err = MultiSiteSetup.SaveConfig(cfConfig, target)
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(2))
		Expect(configs).To(ContainElement(&returnedConfigSummary))
		returnedConfigSummary.Name = "foundation2"
		Expect(configs).To(ContainElement(&returnedConfigSummary))

		By("removing the initial config")
		err = MultiSiteSetup.RemoveConfig("foundation1")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(len(configs)).To(Equal(1))
		Expect(*configs[0]).To(Equal(returnedConfigSummary))

		By("updates an existing config")
		_, err = MultiSiteSetup.SaveConfig(cfConfig, "foundation2")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(configs)).To(Equal(1))
		Expect(*configs[0]).To(Equal(returnedConfigSummary))

		By("removing the last remaining config")
		err = MultiSiteSetup.RemoveConfig("foundation2")
		Expect(err).NotTo(HaveOccurred())

		configs, err = MultiSiteSetup.ListConfigs()
		Expect(len(configs)).To(Equal(0))
	})

	Context("Listing configs", func() {
		When("config file path does not exist", func() {
			BeforeEach(func() {
				_, err := MultiSiteSetup.SaveConfig(cfConfig, "testFoundation")
				Expect(err).NotTo(HaveOccurred())
				// delete the config to force a failure
				Expect(os.Remove(filepath.Join(MultiSiteSetup.ReplicationConfigHome, "testFoundation", ".cf", "config.json"))).To(Succeed())
			})
			It("returns an error", func() {
				_, err := MultiSiteSetup.ListConfigs()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
		When("one of the config file paths does not exist", func() {
			var tmp string
			BeforeEach(func() {
				_, err := MultiSiteSetup.SaveConfig(cfConfig, "testFoundation1")
				Expect(err).NotTo(HaveOccurred())
				// delete the config to force a failure
				Expect(os.Remove(filepath.Join(MultiSiteSetup.ReplicationConfigHome, "testFoundation1", ".cf", "config.json"))).To(Succeed())

				_, err = MultiSiteSetup.SaveConfig(cfConfig, "testFoundation2")
				tmp = returnedConfigSummary.Name
				returnedConfigSummary.Name = "testFoundation2"
				Expect(err).NotTo(HaveOccurred())
			})
			It("returns an error and the existing configs", func() {
				configs, err := MultiSiteSetup.ListConfigs()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				Expect(configs).To(ContainElement(
					&returnedConfigSummary,
				))
			})
			AfterEach(func() {
				returnedConfigSummary.Name = tmp
			})
		})
	})

	Context("Saving configs", func() {
		When("saving a valid config", func() {
			It("returns a JSON subset of relevant config info", func() {
				savedSummary, err := MultiSiteSetup.SaveConfig(cfConfig, "testFoundation")
				Expect(err).NotTo(HaveOccurred())
				Expect(savedSummary).NotTo(BeNil())
				Expect(savedSummary.OrganizationFields.Name).To(Equal("InitialOrgName"))
				Expect(savedSummary.SpaceFields.Name).To(Equal("InitialSpaceName"))
			})
		})
		When("saving two configs to the same name", func() {
			It("the second save successfully overwrites the first", func() {
				_, err := MultiSiteSetup.SaveConfig(cfConfig, "reusedName")
				Expect(err).NotTo(HaveOccurred())
				initializeCFHomeConfig(cfSecondConfig)

				savedSummary, err := MultiSiteSetup.SaveConfig(cfConfig, "reusedName")
				Expect(err).NotTo(HaveOccurred())
				Expect(savedSummary).NotTo(BeNil())

				Expect(savedSummary.OrganizationFields.Name).To(Equal(secondOrgName))
				Expect(savedSummary.SpaceFields.Name).To(Equal(secondSpaceName))
				Expect(savedSummary.Target).To(Equal(secondAPI))
				savedConfigFile, err := savedMSConfigContents("reusedName")
				Expect(err).NotTo(HaveOccurred())
				Expect(savedConfigFile).To(Equal(cfSecondConfig))
			})
		})

		When("saving a non-existent config file", func() {
			It("returns a descriptive error", func() {
				result, err := MultiSiteSetup.SaveConfig(cfConfig+"wrong_name", "reusedFoundation")
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})
		})
		When("saving a config with corrupted JSON", func() {
			It("returns a descriptive error", func() {
				initializeCFHomeConfig(`{"invalidJSON": {"non-JSON-assignment" = true}}`)
				result, err := MultiSiteSetup.SaveConfig(cfConfig, "reusedFoundation")
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
		When("saving a config missing required values", func() {
			It("returns a descriptive error", func() {
				initializeCFHomeConfig(`{"validJSON": {"UnexpectedStructure": true}}`)
				result, err := MultiSiteSetup.SaveConfig(cfConfig, "reusedFoundation")
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("saved configuration must target Cloudfoundry"))
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
