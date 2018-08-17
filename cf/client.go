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

//go:generate counterfeiter . CFPluginAPI
type CFPluginAPI interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	GetCurrentSpace() (plugin_models.Space, error)
	GetService(string) (plugin_models.GetService_Model, error)
	AccessToken() (string, error)
}

type SleepFunc func(time.Duration)

type Error struct {
	Detail string
	Title  string
	Code   int
}

type Task struct {
	Errors []Error
	State  string
	Guid   string
}

type App struct {
	Name string
	Guid string
}

type Client struct {
	pluginAPI   CFPluginAPI
	MaxAttempts int
	Log         *log.Logger
	Sleep       SleepFunc
}

func NewClient(pluginAPI CFPluginAPI) *Client {
	return &Client{
		pluginAPI:   pluginAPI,
		MaxAttempts: 3,
		Log:         log.New(os.Stderr, "", log.LstdFlags),
		Sleep:       time.Sleep,
	}
}

func (c *Client) CreateServiceInstance(planType, instanceName string) error {
	if _, err := c.pluginAPI.GetService(instanceName); err == nil {
		return fmt.Errorf("service instance '%s' already exists", instanceName)
	}

	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	if _, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
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
		_ = c.deleteServiceKey(instanceName, serviceKeyName)
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


func (c *Client) GetSingleHostname(instanceName string) (string, error) {
	serviceKeyName := "MIGRATE-" + uuid.New()
	if err := c.createServiceKey(instanceName, serviceKeyName); err != nil {
		return "", errors.Wrapf(err, "Cannot get the hostname for %s", instanceName)
	}
	defer func() {
		_ = c.deleteServiceKey(instanceName, serviceKeyName)
	}()

	jsonRaw, err := c.serviceKey(instanceName, serviceKeyName)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot get the hostname for %s", instanceName)
	}

	var serviceKey struct {
		Hostname  string   `json:"hostname"`
		Hostnames []string `json:"hostnames"`
	}

	if err = json.Unmarshal([]byte(jsonRaw), &serviceKey); err != nil {
		return "", fmt.Errorf("Cannot get the hostname for %s: invalid response: %s", instanceName, jsonRaw)
	}

	return serviceKey.Hostname, nil
}

func (c *Client) createServiceKey(instanceName, serviceKeyName string) error {
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput("create-service-key", instanceName, serviceKeyName)
	return err
}

func (c *Client) serviceKey(instanceName, serviceKeyName string) (string, error) {
	output, err := c.pluginAPI.CliCommandWithoutTerminalOutput("service-key", instanceName, serviceKeyName)

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
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput("delete-service-key", "-f", instanceName, serviceKeyName)
	return err
}

func (c *Client) UpdateServiceConfig(instanceName string, jsonParams string) error {
	if _, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
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
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
		"delete-service",
		instanceName,
		"-f",
	)

	return err
}

func (c *Client) ServiceExists(serviceName string) bool {
	_, err := c.pluginAPI.GetService(serviceName)
	return err == nil
}

func (c *Client) PushApp(path, appName string) error {
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
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
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
		"bind-service", appName, serviceName,
	)

	return errors.Wrapf(
		err,
		"failed to bind-service %q to application %q",
		serviceName, appName,
	)
}

func (c *Client) GetAppByName(name string) (App, error) {
	space, err := c.pluginAPI.GetCurrentSpace()
	if err != nil {
		return App{}, fmt.Errorf("failed to lookup current space: %s", err)
	}

	output, err := c.pluginAPI.CliCommandWithoutTerminalOutput("curl", "/v3/apps?names="+name+"&space_guids="+space.Guid)
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
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
		"start", appName,
	)

	return errors.Wrapf(
		err,
		"failed to start application %q",
		appName,
	)
}

// cf run-task migration-app "./migrate source dest"

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

type cliTask func(string) (*Task, error) // TODO: make the first arg generic

func (c *Client) GetTaskByGUID(guid string) (*Task, error) {
	response, err := c.retryCfCLIRequestWithExponentialBackoff(c.requestTask, guid, "failed to retrieve task by GUID")
	task, _ := response.(*Task)

	return task, err
}

func (c *Client) retryCfCLIRequestWithExponentialBackoff(requestFunc cliTask, requestArg string, failureMessage string) (interface{}, error) { // TODO: in an ideal world, the arguments passed in wouldn't know about Tasks
	for attempt := 0; attempt < c.MaxAttempts; attempt ++ {
		if attempt > 0 {
			c.Sleep(time.Second << uint(attempt))
		}

		response, err := requestFunc(requestArg)
		if err != nil {
			if strings.Contains(err.Error(), "CF-InvalidAuthToken") {

				if _, e := c.pluginAPI.AccessToken(); e != nil {
					c.Log.Printf("failed to refresh the access token: %s", e.Error())
				}
			}
			c.Log.Printf("Attempt %d/%d: %s: %s", attempt+1, c.MaxAttempts, failureMessage, err)

			continue
		}

		return response, nil
	}

	return nil, errors.New(failureMessage)
}

func (c *Client) requestTask(guid string) (*Task, error) {
	output, err := c.pluginAPI.CliCommandWithoutTerminalOutput("curl", "/v3/tasks/"+guid)

	if err != nil {
		return nil, err
	}

	jsonRaw := strings.Join(output, "\n")

	task := Task{}
	err = json.Unmarshal([]byte(jsonRaw), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the following api response: %s", jsonRaw)
	}

	if len(task.Errors) != 0 {
		if task.Errors[0].Title == "CF-InvalidAuthToken" {
			return nil, errors.New("(error code 1000: CF-InvalidAuthToken - Invalid Auth Token)")
		}
		return nil, fmt.Errorf("cc error code %d: %s - %s", task.Errors[0].Code, task.Errors[0].Title, task.Errors[0].Detail)
	}
	return &task, err
}

func (c *Client) CreateTask(app App, command string) (*Task, error) {
	cfArgs := []string{
		"curl",
		"-X", "POST",
		"-d", fmt.Sprintf(`{"command":"%s"}`, command),
		"/v3/apps/" + app.Guid + "/tasks",
	}

	output, err := c.pluginAPI.CliCommandWithoutTerminalOutput(cfArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to create a task: %s", err)
	}

	taskInfo := Task{}
	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &taskInfo); err != nil {
		return nil, fmt.Errorf("failed to parse the following api response: %s", jsonRaw)
	}

	if len(taskInfo.Errors) != 0 {
		err := taskInfo.Errors[0]
		return nil, fmt.Errorf("failed to create a task: %d: %s - %s", err.Code, err.Title, err.Detail)
	}

	return &taskInfo, nil
}

func (c *Client) waitForTask(task *Task) (string, error) {
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
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
		"delete", "-f", appName,
	)

	return errors.Wrapf(
		err,
		"failed to delete application %q",
		appName,
	)
}

func (c *Client) RenameService(oldName, newName string) error {
	_, err := c.pluginAPI.CliCommandWithoutTerminalOutput(
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
		service, err := c.pluginAPI.GetService(instanceName)

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
	c.pluginAPI.CliCommand("logs", "--recent", appName)
}
