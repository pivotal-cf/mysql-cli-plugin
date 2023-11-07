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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver/v4"
	"github.com/jessevdk/go-flags"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/cf"
	find_bindings "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/presentation"
)

var (
	version = "built from source"
	gitSHA  = "unknown"
)

const (
	usage = `cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools find-bindings [-h] <mysql-v1-service-name>
   cf mysql-tools version`
	migrateUsage          = `cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>`
	findUsage             = `cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`
	saveTargetUsage       = `cf mysql-tools save-target <target-name>`
	removeTargetUsage     = `cf mysql-tools remove-target <target-name>`
	setupReplicationUsage = `cf mysql-tools setup-replication <primary-foundation> <primary-instance> <secondary-foundation> <secondary-instance>`
)

//counterfeiter:generate . BindingFinder
type BindingFinder interface {
	FindBindings(serviceLabel string) ([]find_bindings.Binding, error)
}

//counterfeiter:generate . Migrator
type Migrator interface {
	CheckServiceExists(donorInstanceName string) error
	CreateServiceInstance(planType, serviceName string) error
	MigrateData(options migrate.MigrateOptions) error
	RenameServiceInstances(donorInstanceName, recipientInstanceName string) error
	CleanupOnError(recipientInstanceName string) error
}

//counterfeiter:generate . MultiSite
type MultiSite interface {
	ListConfigs() ([]string, error)
	SaveConfig(cfConfig, targetName string) error
	RemoveConfig(targetName string) error
	SetupReplication(primaryFoundation, primaryInstance, secondaryFoundation, secondaryInstance string) error
}

type MigrationAppExtractor interface {
	Unpack(directoryPath string) error
}

type MySQLPlugin struct {
	MigrationAppExtractor MigrationAppExtractor
	err                   error
}

func (c *MySQLPlugin) Err() error {
	return c.err
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	// Setting of variables each time the plugin runs
	cfConfig, _ := confighelpers.DefaultFilePath()
	replicationHome := filepath.Join(confighelpers.PluginRepoDir(), ".set-replication")
	if _, err := os.Stat(replicationHome); os.IsNotExist(err) {
		err = os.Mkdir(replicationHome, 0700)
		if err != nil {
			fmt.Printf("error trying to create %s to store the saved configurations: %v\n", replicationHome, err)
		}
	}

	if len(args) < 2 {
		// Unfortunately there is no good way currently to show the usage on a plugin
		// without having `-h` added to the command line, so we hardcode it.
		fmt.Fprintln(os.Stderr, `NAME:
   mysql-tools - Plugin to migrate mysql instances

USAGE:
   cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>
   cf mysql-tools find-bindings [-h] <mysql-v1-service-name>
   cf mysql-tools version`)
		os.Exit(1)
		return
	}

	command := args[1]

	switch command {
	default:
		c.err = fmt.Errorf("unknown command '%s'", command)
	case "version":
		fmt.Printf("%s (%s)\n", version, gitSHA)
		os.Exit(0)
	case "find-bindings":
		finder := find_bindings.NewBindingFinder(cf.NewFindBindingsClient(cliConnection))
		c.err = FindBindings(finder, args[2:])
	case "migrate":
		migrator := migrate.NewMigrator(cf.NewMigratorClient(cliConnection), c.MigrationAppExtractor)
		c.err = Migrate(migrator, args[2:])
	case "save-target":
		msSetup := multisite.NewMultiSite(replicationHome)
		c.err = SaveTarget(msSetup, cfConfig, args[2:])
	case "list-targets":
		msSetup := multisite.NewMultiSite(replicationHome)
		c.err = ListTargets(msSetup)
	case "remove-target":
		msSetup := multisite.NewMultiSite(replicationHome)
		c.err = RemoveTarget(msSetup, args[2:])
	case "setup-replication":
		msSetup := multisite.NewMultiSite(replicationHome)
		c.err = SetupReplication(msSetup, args[2:])
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
					Usage: usage,
				},
			},
		},
	}
}

func FindBindings(bf BindingFinder, args []string) error {
	var opts struct {
		Args struct {
			ServiceName string `positional-arg-name:"<mysql-v1-service-name>"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools find-bindings"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", findUsage, msg)
	}

	serviceName := opts.Args.ServiceName

	binding, err := bf.FindBindings(serviceName)
	if err != nil {
		return err
	}

	presentation.Report(os.Stdout, binding)

	return nil
}

func Migrate(migrator Migrator, args []string) error {
	var opts struct {
		Args struct {
			Source   string `positional-arg-name:"<source-service-instance>"`
			PlanName string `positional-arg-name:"<p.mysql-plan-type>"`
		} `positional-args:"yes" required:"yes"`
		NoCleanup         bool `long:"no-cleanup" description:"don't clean up migration app and new service instance after a failed migration"`
		SkipTLSValidation bool `long:"skip-tls-validation" short:"k" description:"Skip certificate validation of the MySQL server certificate. Not recommended!"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools migrate"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", migrateUsage, msg)
	}
	donorInstanceName := opts.Args.Source
	tempRecipientInstanceName := donorInstanceName + "-new"
	destPlan := opts.Args.PlanName
	cleanup := !opts.NoCleanup
	skipTLSValidation := opts.SkipTLSValidation

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	log.Printf("Warning: The mysql-tools migrate command will not migrate any triggers, routines or events.")
	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	log.Printf("Creating new service instance %q for service %s using plan %s", tempRecipientInstanceName, productName, destPlan)
	if err := migrator.CreateServiceInstance(destPlan, tempRecipientInstanceName); err != nil {
		if cleanup {
			migrator.CleanupOnError(tempRecipientInstanceName)
			return fmt.Errorf("error creating service instance: %v. Attempting to clean up service %s",
				err,
				tempRecipientInstanceName,
			)
		}

		return fmt.Errorf("error creating service instance: %v. Not cleaning up service %s",
			err,
			tempRecipientInstanceName,
		)
	}

	migrationOptions := migrate.MigrateOptions{
		DonorInstanceName:     donorInstanceName,
		RecipientInstanceName: tempRecipientInstanceName,
		Cleanup:               cleanup,
		SkipTLSValidation:     skipTLSValidation,
	}

	if err := migrator.MigrateData(migrationOptions); err != nil {
		if cleanup {
			migrator.CleanupOnError(tempRecipientInstanceName)

			return fmt.Errorf(
				"error migrating data: %w. Attempting to clean up service %s",
				err,
				tempRecipientInstanceName,
			)
		}

		return fmt.Errorf("error migrating data: %v. Not cleaning up service %s",
			err,
			tempRecipientInstanceName,
		)
	}

	return migrator.RenameServiceInstances(donorInstanceName, tempRecipientInstanceName)
}

func ListTargets(ms MultiSite) error {
	configs, err := ms.ListConfigs()
	if err != nil {
		return fmt.Errorf("error listing multisite targets: %v", err)
	}
	fmt.Printf("Configured Targets:\n%v\n", configs)
	return nil
}

func SaveTarget(ms MultiSite, cfConfig string, args []string) error {
	var opts struct {
		Args struct {
			TargetName string `positional-arg-name:"<target-name>"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools save-target"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", saveTargetUsage, msg)
	}

	targetConfigName := opts.Args.TargetName

	err = ms.SaveConfig(cfConfig, targetConfigName)
	if err != nil {
		return fmt.Errorf("error trying to save the target config: %v", err)
	}

	return nil
}

func RemoveTarget(ms MultiSite, args []string) error {
	var opts struct {
		Args struct {
			TargetName string `positional-arg-name:"<target-name>"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools remove-target"
	parser.Args()
	args, err := parser.ParseArgs(args)

	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", removeTargetUsage, msg)
	}

	removeConfigName := opts.Args.TargetName
	err = ms.RemoveConfig(removeConfigName)

	if err != nil {
		return fmt.Errorf("error trying to remove the target config: %v", err)
	}

	return nil
}

func SetupReplication(ms MultiSite, args []string) error {
	var opts struct {
		Args struct {
			PrimaryFoundation   string `positional-arg-name:"<primary-foundation>"`
			PrimaryInstance     string `positional-arg-name:"<primary-instance>"`
			SecondaryFoundation string `positional-arg-name:"<secondary-foundation>"`
			SecondaryInstance   string `positional-arg-name:"<secondary-instance>"`
		} `positional-args:"yes" required:"yes"`
	}
	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools setup-replication"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", setupReplicationUsage, msg)
	}

	primaryFoundation := opts.Args.PrimaryFoundation
	primaryInstance := opts.Args.PrimaryInstance
	secondaryFoundation := opts.Args.SecondaryFoundation
	secondaryInstance := opts.Args.SecondaryInstance

	err = ms.SetupReplication(primaryFoundation, primaryInstance, secondaryFoundation, secondaryInstance)
	if err != nil {
		return fmt.Errorf("replication setup error: %w", err)
	}

	return nil
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
