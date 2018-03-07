package cfapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

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

func GetAppByName(name string, cliConnection plugin.CliConnection) App {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/apps?names="+name)
	if err != nil {
		panic(err)
	}

	var appInfo struct {
		Resources []App
	}

	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &appInfo); err != nil {
		panic(err)
	}

	if len(appInfo.Resources) != 1 {
		panic("expected at most one app guid")
	}

	return appInfo.Resources[0]
}

func GetTaskByGUID(guid string, cliConnection plugin.CliConnection) Task {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/tasks/"+guid)
	if err != nil {
		panic(err)
	}

	taskInfo := Task{}
	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &taskInfo); err != nil {
		panic(err)
	}

	return taskInfo
}

func CreateTask(app App, command string, cliConnection plugin.CliConnection) Task {
	cfArgs := []string{
		"curl",
		"-X", "POST",
		"-d", fmt.Sprintf(`{"command":"%s"}`, command),
		"/v3/apps/" + app.Guid + "/tasks",
	}
	output, err := cliConnection.CliCommandWithoutTerminalOutput(cfArgs...)
	if err != nil {
		panic(err)
	}

	taskInfo := Task{}
	jsonRaw := strings.Join(output, "\n")
	if err := json.Unmarshal([]byte(jsonRaw), &taskInfo); err != nil {
		panic(err)
	}

	if taskInfo.Errors != nil {
		panic(taskInfo.Errors)
	}

	return taskInfo
}
