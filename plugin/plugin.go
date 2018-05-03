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
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver"
	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/migrate"
	"github.com/pivotal-cf/mysql-cli-plugin/unpack"
	"github.com/pkg/errors"
)

var (
	version = "built from source"
	gitSHA  = "unknown"
)

//go:generate counterfeiter . migrator
type migrator interface {
	CheckServiceExists(donorInstanceName string) error
	CreateAndConfigureServiceInstance(planType, serviceName string) error
	MigrateData(donorInstanceName, recipientInstanceName string, cleanup bool) error
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
		fmt.Fprintln(os.Stderr,"Please pass in a command [migrate|replace|version] to mysql-tools")
		os.Exit(1)
		return
	}

	command := args[1]
	migrator := migrate.NewMigrator(cf.NewClient(cliConnection), unpack.NewUnpacker())

	switch command {
	default:
		c.err = errors.Errorf("unknown command '%s'", command)
	case "version":
		fmt.Printf("%s (%s)", version, gitSHA)
		os.Exit(0)
	case "replace":
		c.err = Replace(migrator, args[2:])
	case "migrate":
		c.err = Migrate(migrator, args[2:])
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
	cf mysql-tools migrate [--no-cleanup] <v1-service-instance> --create <plan-type>
	cf mysql-tools replace [--no-cleanup] <v1-service-instance> <v2-service-instance>
`,
				},
			},
		},
	}
}

func Replace(migrator migrator, args []string) error {
	var opts struct {
		Args struct {
			Source string `positional-arg-name:"<v1-service-instance>"`
			Dest   string `positional-arg-name:"<v2-service-instance>"`
		} `positional-args:"yes" required:"yes"`
		NoCleanup bool `long:"no-cleanup" description:"don't clean up migration app after a failed migration'"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools replace"

	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		fmt.Fprintln(os.Stderr, `Usage: cf mysql-tools replace [--no-cleanup] <v1-service-instance> <v2-service-instance>`)
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return errors.Errorf("Usage: cf mysql-tools replace [--no-cleanup] <v1-service-instance> <v2-service-instance>\n%s", msg)
	}

	donorInstanceName := opts.Args.Source
	recipientInstanceName := opts.Args.Dest
	cleanup := ! opts.NoCleanup

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	if err := migrator.CheckServiceExists(recipientInstanceName); err != nil {
		return err
	}

	if err := migrator.MigrateData(donorInstanceName, recipientInstanceName, cleanup); err != nil {
		return err
	}

	return migrator.RenameServiceInstances(donorInstanceName, recipientInstanceName)
}

func Migrate(migrator migrator, args []string) error {
	var opts struct {
		Args struct {
			Source string `positional-arg-name:"<v1-service-instance>"`
		} `positional-args:"yes"`
		PlanName  string `long:"create" description:"create a new service instance with the given plan" required:"true"`
		NoCleanup bool   `long:"no-cleanup" description:"don't clean up migration app and new service instance after a failed migration'"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools migrate"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		fmt.Fprintln(os.Stderr, `Usage: cf mysql-tools migrate [--no-cleanup] <v1-service-instance> --create <plan-type>`)
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return errors.Errorf("Usage: cf mysql-tools migrate [--no-cleanup] <v1-service-instance> --create <plan-type>\n%s", msg)
	}
	donorInstanceName := opts.Args.Source
	recipientInstanceName := donorInstanceName + "-new"
	destPlan := opts.PlanName
	cleanup := ! opts.NoCleanup

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	log.Printf("Creating new service instance %q for service %s using plan %s", recipientInstanceName, productName, destPlan)
	if err := migrator.CreateAndConfigureServiceInstance(destPlan, recipientInstanceName); err != nil {
		if cleanup {
			migrator.CleanupOnError(recipientInstanceName)
			return fmt.Errorf("error creating service instance: %v. Attempting to clean up service %s",
				err,
				recipientInstanceName,
			)
		}

		return fmt.Errorf("error creating service instance: %v. Not cleaning up service %s",
			err,
			recipientInstanceName,
		)
	}

	if err := migrator.MigrateData(donorInstanceName, recipientInstanceName, cleanup); err != nil {
		if cleanup {
			migrator.CleanupOnError(recipientInstanceName)

			return fmt.Errorf("error migrating data: %v. Attempting to clean up service %s",
				err,
				recipientInstanceName,
			)
		}

		return fmt.Errorf("error migrating data: %v. Not cleaning up service %s",
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
