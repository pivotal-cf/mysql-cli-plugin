// Copyright (C) 2019-Present Pivotal Software, Inc. All rights reserved.
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
	"code.cloudfoundry.org/cli/plugin/models"
	"time"
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

type cliTask func(string) (*Task, error)
