package cf

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

//go:generate counterfeiter . CfCommandRunner
type CfCommandRunner interface {
	CliCommand(...string) ([]string, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
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
	output, err := a.cfCommandRunner.CliCommandWithoutTerminalOutput("curl", "/v3/apps?names="+name)
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
