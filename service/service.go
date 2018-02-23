package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ServiceKeyName = "service-key"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
}

// TODO distinguish between v1 and v2
type ServiceInstance struct {
	cliRunner    CfCommandRunner
	instanceName string
}

func NewServiceInstance(cliRunner CfCommandRunner, instanceName string) *ServiceInstance {
	return &ServiceInstance{
		cliRunner:    cliRunner,
		instanceName: instanceName,
	}
}

type ServiceInfo struct {
	Hostname     string `json:"hostname"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DBName       string `json:"name"`
	LocalSSHPort int
}

func (si *ServiceInstance) ServiceInfo() (ServiceInfo, error) {
	_, err := si.cliRunner.CliCommand("create-service-key", si.instanceName, ServiceKeyName)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("failed to create-service-key: %s", err)
	}

	outputElements, err := si.cliRunner.CliCommand("service-key", si.instanceName, ServiceKeyName)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("failed to create-service-key: %s", err)
	}

	output := strings.Join(outputElements[2:], "\n")

	var serviceInfo ServiceInfo
	if err := json.Unmarshal([]byte(output), &serviceInfo); err != nil {
		return ServiceInfo{}, fmt.Errorf("failed to get service-key: %s", err)
	}

	return serviceInfo, nil
}

func (si *ServiceInstance) Cleanup() {
	si.cliRunner.CliCommand("delete-service-key", "-f", si.instanceName, ServiceKeyName)
}
