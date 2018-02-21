package test_helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

type UpgradeTestConfiguration struct {
	AlbumId     string `json:"albumId"`
	RandomValue string `json:"randomValue"`
}

func SaveUpgradeTestConfig(randomValue, albumId, appName, appDomain string) {
	outputFile := filepath.Join(os.Getenv("UPGRADE_TEST_CONFIG_DIR"), "upgrade-test-config.json")

	upgradeTestConfig := UpgradeTestConfiguration{
		AlbumId:     albumId,
		RandomValue: randomValue,
	}

	jsonStruct, err := json.Marshal(&upgradeTestConfig)
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(outputFile, jsonStruct, 0644)
	Expect(err).NotTo(HaveOccurred())
}

func LoadUpgradeTestConfig() *UpgradeTestConfiguration {
	configFile := filepath.Join(os.Getenv("UPGRADE_TEST_CONFIG_DIR"), "upgrade-test-config.json")

	fileContents, err := ioutil.ReadFile(configFile)
	Expect(err).NotTo(HaveOccurred())

	upgradeTestConfig := UpgradeTestConfiguration{}
	err = json.Unmarshal(fileContents, &upgradeTestConfig)
	Expect(err).NotTo(HaveOccurred())

	return &upgradeTestConfig
}
