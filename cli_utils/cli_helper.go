package cli_utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	// TODO don't hardcode app name
	AppName    = "static-app"
	StaticFile = "Staticfile"
	IndexFile  = "index.html"
	// TODO don't hardcode service-key name
	ServiceKeyName = "service-key"
)

type ServiceKey struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"name"`
}

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}

func PushApp(cfCommandRunner CfCommandRunner, tmpDir string) error {
	appDir := filepath.Join(tmpDir, AppName)
	err := os.Mkdir(appDir, 0700)
	if err != nil {
		return fmt.Errorf("Failed to create app directory: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(appDir, StaticFile), nil, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(appDir, IndexFile), nil, 0600)
	if err != nil {
		return err
	}

	_, err = cfCommandRunner.CliCommand("push", AppName, "--random-route", "-b", "staticfile_buildpack", "-p", appDir)

	if err != nil {
		return fmt.Errorf("Failed to push app: %s", err)
	}
	return nil
}

func DeleteApp(cfCommandRunner CfCommandRunner) error {
	_, err := cfCommandRunner.CliCommand("delete", AppName, "-f")

	if err != nil {
		return fmt.Errorf("Failed to delete app: %s", err)
	}

	return nil
}

func CreateServiceKey(cfCommandRunner CfCommandRunner, instanceName string) error {

	_, err := cfCommandRunner.CliCommand("create-service-key", instanceName, ServiceKeyName)

	if err != nil {
		return fmt.Errorf("Failed to create-service-key: %s", err)
	}

	return nil
}

func DeleteServiceKey(cfCommandRunner CfCommandRunner, instanceName string) error {
	_, err := cfCommandRunner.CliCommand("delete-service-key", instanceName, ServiceKeyName, "-f")

	if err != nil {
		return fmt.Errorf("Failed to delete-service-key: %s", err)
	}

	return nil
}


func GetServiceKey(cfCommandRunner CfCommandRunner, instanceName string) (*ServiceKey, error) {
	outputElements, err := cfCommandRunner.CliCommand("service-key", instanceName, ServiceKeyName)
	if err != nil {
		return nil, fmt.Errorf("Failed to get service-key: %s", err)
	}

	output := strings.Join(outputElements[2:], "\n")

	var serviceKey ServiceKey
	err = json.Unmarshal([]byte(output), &serviceKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to get service-key: %s", err)
	}

	return &serviceKey, nil
}
