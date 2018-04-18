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
	"log"
	"os"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/migrate"
	"github.com/pivotal-cf/mysql-cli-plugin/unpack"
	"github.com/pkg/errors"
)

var (
	version = "built from source"
	gitSHA  = "unknown"
)

//go:generate counterfeiter . CFClient
type CFClient interface {
	CreateServiceInstance(planType, instanceName string) error
	GetHostnames(instanceName string) ([]string, error)
	UpdateServiceConfig(instanceName string, jsonParams string) error
	DeleteServiceInstance(instanceName string) error
	BindService(appName, serviceName string) error
	DeleteApp(appName string) error
	DumpLogs(appName string)
	PushApp(path, appName string) error
	RenameService(oldName, newName string) error
	RunTask(appName, command string) error
	ServiceExists(serviceName string) bool
	StartApp(appName string) error
}

//go:generate counterfeiter . migrator
type migrator interface {
	CheckServiceExists(donorInstanceName string) error
	CreateAndConfigureServiceInstance(planType, serviceName string) error
	MigrateData(donorInstanceName, recipientInstanceName string) error
	RenameServiceInstances(donorInstanceName, recipientInstanceName string) error
	CleanupOnError(recipientInstanceName string) error
}

type MySQLPlugin struct {
	err error
}

func (c *MySQLPlugin) Err() error {
	return c.err
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	if len(args) < 2 {
		log.Println("Please pass in a command [migrate|replace|version] to mysql-tools")
		os.Exit(1)
		return
	}

	command := args[1]
	migrator := migrate.NewMigrator(cf.NewClient(cliConnection), unpack.NewUnpacker())

	switch command {
	default:
		c.err = errors.Errorf("Unknown command '%s'", command)
	case "version":
		fmt.Printf("%s (%s)", version, gitSHA)
		os.Exit(0)
	case "replace":
		c.err = Replace(migrator, args)
	case "migrate":
		c.err = Migrate(migrator, args)
	}
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
					Usage: `mysql-tools
    cf mysql-tools migrate <v1-service-instance> <v2-service-instance>
    cf mysql-tools replace <v1-service-instance> <v2-service-instance>
`,
				},
			},
		},
	}
}

func Replace(migrator migrator, args []string) error {
	if len(args) != 4 {
		return errors.New("Usage: cf mysql-tools replace <v1-service-instance> <v2-service-instance>")
	}

	donorInstanceName := args[2]
	recipientInstanceName := args[3]

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	if err := migrator.CheckServiceExists(recipientInstanceName); err != nil {
		return err
	}

	if err := migrator.MigrateData(donorInstanceName, recipientInstanceName); err != nil {
		return err
	}

	return migrator.RenameServiceInstances(donorInstanceName, recipientInstanceName)
}

func Migrate(migrator migrator, args []string) error {
	if len(args) != 5 || args[3] != "--create" {
		return errors.New("Usage: cf mysql-tools migrate <v1-service-instance> --create <v2-plan>")
	}

	donorInstanceName := args[2]
	recipientInstanceName := donorInstanceName + "-new"
	destPlan := args[4]

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	log.Printf("Creating new service instance %q for service p.mysql using plan %s", recipientInstanceName, destPlan)
	if err := migrator.CreateAndConfigureServiceInstance(destPlan, recipientInstanceName); err != nil {
		return err
	}

	if err := migrator.MigrateData(donorInstanceName, recipientInstanceName); err != nil {

		migrator.CleanupOnError(recipientInstanceName)

		return fmt.Errorf("Error migrating data: %v. Attempting to clean up service %s",
			err,
			recipientInstanceName,
		)
	}

	return nil
}

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
