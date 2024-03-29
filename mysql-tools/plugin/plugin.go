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

package plugin

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver/v4"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/cf"
	findbindings "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/version"
)

const (
	usage = `NAME:
mysql-tools - Plugin to manage mysql instances

USAGE:
cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>
cf mysql-tools find-bindings [-h] <mysql-v1-service-name>
cf mysql-tools save-target <target-name>
cf mysql-tools remove-target <target-name>
cf mysql-tools list-targets
cf mysql-tools setup-replication [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ]
cf mysql-tools switchover [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ] [ --force | -f ]
cf mysql-tools version`
)

type MigrationAppExtractor interface {
	Unpack(directoryPath string) error
}

type MySQLPlugin struct {
	MigrationAppExtractor MigrationAppExtractor
	MultisiteConfig       commands.MultisiteConfig
	err                   error
}

func (c *MySQLPlugin) Err() error {
	return c.err
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	if len(args) < 2 {
		// Unfortunately there is no good way currently to show the usage on a plugin
		// without having `-h` added to the command line, so we hardcode it.
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
		return
	}

	command := args[1]
	options := args[2:]
	switch command {
	default:
		c.err = fmt.Errorf("unknown command '%s'", command)
	case "version":
		c.err = commands.Version()
	case "find-bindings":
		bf := findbindings.NewBindingFinder(cf.NewFindBindingsClient(cliConnection))
		c.err = commands.FindBindings(options, bf)
	case "migrate":
		c.err = commands.Migrate(
			options,
			migrate.NewMigrator(cf.NewMigratorClient(cliConnection), c.MigrationAppExtractor),
		)
	case "save-target":
		c.err = commands.SaveTarget(options, c.MultisiteConfig)
	case "list-targets":
		c.err = commands.ListTargets(c.MultisiteConfig)
	case "remove-target":
		c.err = commands.RemoveTarget(options, c.MultisiteConfig)
	case "setup-replication":
		c.err = commands.SetupReplication(options, c.MultisiteConfig)
	case "switchover":
		c.err = commands.SwitchoverReplication(options, c.MultisiteConfig, os.Stdout, os.Stdin)
	}
}

func (c *MySQLPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "MysqlTools",
		Version: versionFromSemver(version.Version),
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

func versionFromSemver(in string) plugin.VersionType {
	unknownVersion := plugin.VersionType{
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
