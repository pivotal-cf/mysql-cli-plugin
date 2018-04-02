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
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	GetCurrentSpace() (plugin_models.Space, error)
}

type Api struct {
	cfCommandRunner CfCommandRunner
}

func NewApi(cfCommandRunner CfCommandRunner) *Api {
	return &Api{
		cfCommandRunner: cfCommandRunner,
	}
}

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

func (a *Api) GetAppByName(name string) (App, error) {
	space, err := a.cfCommandRunner.GetCurrentSpace()
	if err != nil {
		return App{}, fmt.Errorf("failed to lookup current space: %s", err)
	}

	output, err := a.cfCommandRunner.CliCommandWithoutTerminalOutput("curl", "/v3/apps?names="+name+"&space_guids="+space.Guid)
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

func (a *Api) GetTaskByGUID(guid string) (Task, error) {
	output, err := a.cfCommandRunner.CliCommandWithoutTerminalOutput("curl", "/v3/tasks/"+guid)
	if err != nil {
		return Task{}, fmt.Errorf("failed to retrieve a task by guid: %s", err)
	}

	taskInfo := Task{}
	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &taskInfo); err != nil {
		return Task{}, fmt.Errorf("failed to parse the following api response: %s", jsonRaw)
	}

	return taskInfo, nil
}

func (a *Api) CreateTask(app App, command string) (Task, error) {
	cfArgs := []string{
		"curl",
		"-X", "POST",
		"-d", fmt.Sprintf(`{"command":"%s"}`, command),
		"/v3/apps/" + app.Guid + "/tasks",
	}

	output, err := a.cfCommandRunner.CliCommandWithoutTerminalOutput(cfArgs...)
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

func (a *Api) WaitForTask(task Task) (string, error) {
	var t Task
	var err error
	for t.State != "SUCCEEDED" && t.State != "FAILED" {
		t, err = a.GetTaskByGUID(task.Guid)
		if err != nil {
			return "", err
		}
	}
	return t.State, nil
}
