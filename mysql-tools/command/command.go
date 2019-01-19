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

package command

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin_errors"
	"github.com/pkg/errors"
)

var (
	version = "built from source"
	gitSHA  = "unknown"
)

const (
	usage = `cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools version`
)

func NewMySQLPlugin(routes map[string]interface{ CommandRunner }) *MySQLPlugin {

	cmdRouter := &MySQCmdLRouter{
		Routes: routes,
	}
	return &MySQLPlugin{
		CommandRouter: cmdRouter,
	}
}

type MySQLPlugin struct {
	err           error
	CommandRouter Router
}

func (c *MySQLPlugin) Err() error {
	return c.err
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 1 { // should we keep this it seems like it may never happen?
		c.err = errors.New("Error: plugin did not receive the expected input from the CLI")
		return
	}

	pluginName := args[0]

	if pluginName == "mysql-tools" {

		if len(args) == 1 {
			// Unfortunately there is no good way currently to show the usage on a plugin
			// without having `-h` added to the command line, so we hardcode it.
			c.err = plugin_errors.NewUsageError("")
			return
		}

		if searchForHelpFlag(args) {
			c.err = plugin_errors.NewUsageError("")
			return
		}

		command := args[1]

		err := c.CommandRouter.Match(command, cliConnection, args[2:])
		if err != nil {
			c.err = err
		}
	}

}

func searchForHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" {
			return true
		}
	}
	return false
}

func (c *MySQLPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "MysqlTools",
		Version: versionFromSemver(version),
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "mysql-tools",
				HelpText: "Plugin to migrate mysql instances",
				UsageDetails: plugin.Usage{
					Usage: usage,
				},
			},
		},
	}
}

//func (c *MySQLPlugin) PrintUsage(prefix string) {
//	var errMsg string
//	defaultErrMsg := `NAME:
//   mysql-tools - Plugin to migrate mysql instances
//
//USAGE:
//   cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>
//   cf mysql-tools version`
//	if prefix != "" {
//		errMsg = fmt.Sprintf("%s\n%s", prefix, defaultErrMsg)
//	} else {
//		errMsg = defaultErrMsg
//	}
//	c.err = errors.New(errMsg)
//}

func versionFromSemver(in string) plugin.VersionType {
	var unknownVersion = plugin.VersionType{
		Major: 0,
		Minor: 0,
		Build: 1,
	}

	if in == "built from source" {
		return unknownVersion
	}

	v, err := semver.Parse(in)
	if err != nil {
		return unknownVersion
	}

	return plugin.VersionType{
		Major: int(v.Major),
		Minor: int(v.Minor),
		Build: int(v.Patch),
	}
}
