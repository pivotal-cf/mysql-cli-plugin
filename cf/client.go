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

package cf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	GetCurrentSpace() (plugin_models.Space, error)
	GetService(string) (plugin_models.GetService_Model, error)
}

//go:generate counterfeiter . Unpacker
type Unpacker interface {
	Unpack(destDir string) error
}

type SleepFunc func(time.Duration)

type Task struct {
	Errors []struct {
		Detail string
		Title  string
		Code   int
	}
	State string
	Guid  string
}

type App struct {
	Name string
	Guid string
}

type Client struct {
	cfCommandRunner CfCommandRunner
	unpacker        Unpacker
	MaxAttempts     int
	Log             *log.Logger
	Sleep           SleepFunc
}

func NewClient(cfCommandRunner CfCommandRunner) *Client {
	return &Client{
		cfCommandRunner: cfCommandRunner,
		MaxAttempts:     3,
		Log:             log.New(os.Stderr, "", log.LstdFlags),
		Sleep:           time.Sleep,
	}
}

func (c *Client) CreateServiceInstance(planType, instanceName string) error {
	if _, err := c.cfCommandRunner.GetService(instanceName); err == nil {
		return fmt.Errorf("service instance '%s' already exists", instanceName)
	}

	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	if _, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"create-service",
		productName,
		planType,
		instanceName,
	); err != nil {
		return err
	}

	return c.waitForOperationCompletion("create service instance", instanceName)
}

func (c *Client) GetHostnames(instanceName string) ([]string, error) {
	serviceKeyName := "MIGRATE-" + uuid.New()
	if err := c.createServiceKey(instanceName, serviceKeyName); err != nil {
		return nil, errors.Wrapf(err, "Cannot get the hostnames for %s", instanceName)
	}
	defer func() {
		c.deleteServiceKey(instanceName, serviceKeyName)
	}()

	jsonRaw, err := c.serviceKey(instanceName, serviceKeyName)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get the hostnames for %s", instanceName)
	}

	var serviceKey struct {
		Hostname  string   `json:"hostname"`
		Hostnames []string `json:"hostnames"`
	}

	if err = json.Unmarshal([]byte(jsonRaw), &serviceKey); err != nil {
		return nil, fmt.Errorf("Cannot get the hostnames for %s: invalid response: %s", instanceName, jsonRaw)
	}

	if len(serviceKey.Hostnames) != 0 {
		return serviceKey.Hostnames, nil
	}

	return []string{serviceKey.Hostname}, nil
}

func (c *Client) createServiceKey(instanceName, serviceKeyName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput("create-service-key", instanceName, serviceKeyName)
	return err
}

func (c *Client) serviceKey(instanceName, serviceKeyName string) (string, error) {
	output, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput("service-key", instanceName, serviceKeyName)

	if err != nil {
		return "", err
	}

	// skip non-json message in service-key output
	if len(output) > 2 {
		output = output[2:]
	}

	return strings.Join(output, "\n"), nil
}

func (c *Client) deleteServiceKey(instanceName, serviceKeyName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput("delete-service-key", "-f", instanceName, serviceKeyName)
	return err
}

func (c *Client) UpdateServiceConfig(instanceName string, jsonParams string) error {
	if _, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"update-service",
		instanceName,
		"-c",
		jsonParams,
	); err != nil {
		return err
	}

	return c.waitForOperationCompletion("update service config", instanceName)
}

func (c *Client) DeleteServiceInstance(instanceName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"delete-service",
		instanceName,
		"-f",
	)

	return err
}

func (c *Client) ServiceExists(serviceName string) bool {
	_, err := c.cfCommandRunner.GetService(serviceName)
	return err == nil
}

func (c *Client) PushApp(path, appName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"push",
		appName,
		"-b", "binary_buildpack",
		"-u", "none",
		"-c", "sleep infinity",
		"-p", path,
		"--no-route",
		"--no-start",
	)

	return errors.Wrap(err, "failed to push application")
}

func (c *Client) BindService(appName, serviceName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"bind-service", appName, serviceName,
	)

	return errors.Wrapf(
		err,
		"failed to bind-service %q to application %q",
		serviceName, appName,
	)
}

func (c *Client) GetAppByName(name string) (App, error) {
	space, err := c.cfCommandRunner.GetCurrentSpace()
	if err != nil {
		return App{}, fmt.Errorf("failed to lookup current space: %s", err)
	}

	output, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput("curl", "/v3/apps?names="+name+"&space_guids="+space.Guid)
	if err != nil {
		return App{}, fmt.Errorf("failed to retrieve an app by name: %s", err)
	}

	var appInfo struct {
		Resources []App
	}

	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &appInfo); err != nil {
		return App{}, fmt.Errorf("failed to parse the following api response: %s", jsonRaw)
	}

	if len(appInfo.Resources) != 1 {
		return App{}, errors.New("failed to retrieve an app by name: none found")
	}

	return appInfo.Resources[0], nil
}

func (c *Client) StartApp(appName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"start", appName,
	)

	return errors.Wrapf(
		err,
		"failed to start application %q",
		appName,
	)
}

func (c *Client) RunTask(appName, command string) error {
	app, err := c.GetAppByName(appName)
	if err != nil {
		return errors.Wrap(err, "Error")
	}

	task, err := c.CreateTask(app, command)
	if err != nil {
		return errors.Wrap(err, "Error")
	}

	finalState, err := c.waitForTask(task)
	if err != nil {
		return errors.Wrap(err, "Error when waiting for task to complete")
	}

	if finalState != "SUCCEEDED" {
		return errors.Errorf("task completed with status %q", finalState)
	}

	return nil
}

func (c *Client) GetTaskByGUID(guid string) (Task, error) {
	maxAttempts := c.MaxAttempts

	for attempt := 0; attempt < maxAttempts; attempt++ {

		if attempt > 0 {
			c.Sleep(time.Second << uint(attempt))
		}

		output, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput("curl", "/v3/tasks/"+guid)
		if err != nil {
			c.Log.Printf("Attempt %d/%d: failed to retrieve task by guid: %s",
				attempt+1, maxAttempts, err)
			continue
		}

		task := Task{}
		jsonRaw := strings.Join(output, "\n")
		if err := json.Unmarshal([]byte(jsonRaw), &task); err != nil {
			c.Log.Printf("Attempt %d/%d: failed to parse the following api response: %s",
				attempt+1, maxAttempts, jsonRaw)
			continue
		}

		if len(task.Errors) != 0 {
			err := task.Errors[0]
			c.Log.Printf("Attempt %d/%d: failed to look up task (error code %d: %s - %s)",
				attempt+1, maxAttempts, err.Code, err.Title, err.Detail)
			continue
		}

		return task, nil
	}

	return Task{}, errors.New("failed to get task by GUID")
}

func (c *Client) CreateTask(app App, command string) (Task, error) {
	cfArgs := []string{
		"curl",
		"-X", "POST",
		"-d", fmt.Sprintf(`{"command":"%s"}`, command),
		"/v3/apps/" + app.Guid + "/tasks",
	}

	output, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(cfArgs...)
	if err != nil {
		return Task{}, fmt.Errorf("failed to create a task: %s", err)
	}

	taskInfo := Task{}
	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &taskInfo); err != nil {
		return Task{}, fmt.Errorf("failed to parse the following api response: %s", jsonRaw)
	}

	if len(taskInfo.Errors) != 0 {
		err := taskInfo.Errors[0]
		return Task{}, fmt.Errorf("failed to create a task: %d: %s - %s", err.Code, err.Title, err.Detail)
	}

	return taskInfo, nil
}

func (c *Client) waitForTask(task Task) (string, error) {
	var (
		taskGUID = task.Guid
		err      error
	)

	for task.State != "SUCCEEDED" && task.State != "FAILED" {
		c.Sleep(time.Second)
		task, err = c.GetTaskByGUID(taskGUID)
		if err != nil {
			return "", err
		}
	}
	return task.State, nil
}

func (c *Client) DeleteApp(appName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"delete", "-f", appName,
	)

	return errors.Wrapf(
		err,
		"failed to delete application %q",
		appName,
	)
}

func (c *Client) RenameService(oldName, newName string) error {
	_, err := c.cfCommandRunner.CliCommandWithoutTerminalOutput(
		"rename-service", oldName, newName,
	)

	return errors.Wrapf(
		err,
		"failed to rename-service %q to %q",
		oldName, newName,
	)
}

func (c *Client) waitForOperationCompletion(operationName, instanceName string) error {
	attempt := 0

	for {
		service, err := c.cfCommandRunner.GetService(instanceName)

		if err != nil {
			attempt++

			if attempt == c.MaxAttempts {
				return fmt.Errorf("failed to look up status of service instance '%s'", instanceName)
			}

			c.Sleep(time.Second << uint(attempt))
			continue
		}

		attempt = 0

		switch service.LastOperation.State {
		default:
			return nil
		case "failed":
			return fmt.Errorf("failed to %s '%s': %s",
				operationName, instanceName, service.LastOperation.Description)
		case "in progress":
			continue
		}
	}
}

func (c *Client) DumpLogs(appName string) {
	c.cfCommandRunner.CliCommand("logs", "--recent", appName)
}
