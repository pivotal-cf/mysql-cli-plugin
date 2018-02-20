package cli_utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	AppName        = "static-app"
	StaticFile     = "Staticfile"
	IndexFile      = "index.html"
	ServiceKeyName = "service-key"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
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

	_, err = cfCommandRunner.CliCommand("push", AppName, "-p", appDir)

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
	_, err := cfCommandRunner.CliCommand("delete-service-key", instanceName, ServiceKeyName, "f")

	if err != nil {
		return fmt.Errorf("Failed to delete-service-key: %s", err)
	}

	return nil
}

type ServiceKey struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"name"`
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

//func CreateSshTunnel(cfCommandRunner CfCommandRunner, hostName string) error {
//	_, err := cfCommandRunner.CliCommand("ssh", "static-app", "-N", "-L", fmt.Sprintf("63306:%s:3306", hostName), "&")
//	if err != nil {
//		return fmt.Errorf("Failed to open ssh tunnel for service host %s: %s", hostName, err)
//	}
//
//	return nil
//}

type TunnelManager struct {
	db      sql.DB
	timeout time.Duration
}

func (t *TunnelManager) CreateSshTunnel(serviceKey ServiceKey) (context.CancelFunc, error) {
	tunnelContext, tunnelCancel := context.WithCancel(context.Background())
	connectionString := fmt.Sprintf("63306:%s:3306", serviceKey.Hostname)
	tunnelCommand := exec.CommandContext(tunnelContext, "cf", "ssh", "--skip-remote-execution", "-L", connectionString, AppName)
	if err := tunnelCommand.Start(); err != nil {
		return nil, err
	}

	//waitForTunnel(serviceKey)

	return tunnelCancel, nil
}

func WaitForTunnel(db *sql.DB, timeout time.Duration) error {
	//connectionString := fmt.Sprintf(
	//	"%s:%s@tcp(127.0.0.1:63306)/%s?interpolateParams=true&tls=skip-verify",
	//	serviceKey.Username,
	//	serviceKey.Password,
	//	serviceKey.DBName,
	//)

	//selectErr := errors.New("temp-error")
	//for selectErr != nil {
	//	db, err := sql.Open("mysql", connectionString)
	//	if err != nil {
	//		return err
	//	}
	//
	//	_, selectErr = db.Exec("SELECT 1")
	//	if err = db.Close(); err != nil {
	//		return err
	//	}
	//}

	timerCh := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-timerCh:
			return errors.New("Timeout")
		case <-ticker.C:
			var unused int
			if err := db.QueryRow("SELECT 1").Scan(&unused); err == nil {
				return nil
			}
		}

	}

	return nil
}

func DeleteSshTunnel(cfCommandRunner CfCommandRunner, hostName string) error {
	_, err := cfCommandRunner.CliCommand("ssh", "static-app", "-N", "-L", fmt.Sprintf("63306:%s:3306", hostName), "&")
	if err != nil {
		return fmt.Errorf("Failed to open ssh tunnel for service host %s: %s", hostName, err)
	}

	return nil
}
