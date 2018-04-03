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

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver"
	"github.com/gobuffalo/packr"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pkg/errors"
)

type MySQLPlugin struct{
	Replace bool
}

var (
	version = "built from source"
	gitSHA  = "unknown"
)

//go:generate go install github.com/pivotal-cf/mysql-cli-plugin/vendor/github.com/gobuffalo/packr/...
//go:generate $GOPATH/bin/packr --compress
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

	switch command {
	default:
		log.Printf("Unknown command '%s'", command)
		os.Exit(1)
		return
	case "version":
		fmt.Printf("%s (%s)", version, gitSHA)
		os.Exit(0)
	case "replace":
		c.Replace = true
		fallthrough
	case "migrate":
		if len(args) != 4 {
			log.Printf("Usage: cf mysql-tools %s <v1-service-instance> <v2-service-instance>", command)
			os.Exit(1)
			return
		}

		sourceServiceName := args[2]
		destServiceName := args[3]

		err := c.run(cliConnection, sourceServiceName, destServiceName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}

func (c *MySQLPlugin) run(cliConnection plugin.CliConnection, sourceServiceName, destServiceName string) error {
	var (
		api     = cf.NewApi(cliConnection)
		appName = "migrate-app-" + uuid.New()
	)

	box := packr.NewBox("./app")
	tmpDir, err := ioutil.TempDir(os.TempDir(), "migrate_app_")
	if err != nil {
		return errors.Errorf("Error creating temp directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	err = box.Walk(func(name string, file packr.File) error {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(tmpDir, name)), 0700); err != nil {
			return err
		}

		dest, err := os.Create(filepath.Join(tmpDir, name))
		if err != nil {
			return err
		}

		if _, err := io.Copy(dest, file); err != nil {
			return err
		}

		if runtime.GOOS != "windows" {
			return dest.Chmod(0700)
		}

		return nil
	})

	if err != nil {
		return errors.Errorf("Error extracting migrate assets: %s", err)
	}

	log.Print("Started to push app")

	_, err = cliConnection.CliCommandWithoutTerminalOutput("push",
		appName,
		"-b", "binary_buildpack",
		"-u", "none",
		"-c", "sleep infinity",
		"-p", tmpDir,
		"--no-route",
		"--no-start",
	)
	if err != nil {
		return errors.Errorf("failed to push application: %s", err)
	}
	defer func() {
		cliConnection.CliCommandWithoutTerminalOutput("delete", appName, "-f")
		log.Print("Cleaning up...")
	}()
	log.Print("Sucessfully pushed app")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("bind-service", appName, sourceServiceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", appName, sourceServiceName, err)
	}
	log.Print("Sucessfully bound app to v1 instance")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("bind-service", appName, destServiceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", appName, destServiceName, err)
	}
	log.Print("Sucessfully bound app to v2 instance")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("start", appName); err != nil {
		return errors.Errorf("failed to start application %q: %s", appName, err)
	}

	app, err := api.GetAppByName(appName)
	if err != nil {
		return errors.Errorf("Error: %s", err)
	}

	log.Print("Started to run migration task")
	cmd := fmt.Sprintf(`./migrate %s %s`, sourceServiceName, destServiceName)
	task, err := api.CreateTask(app, cmd)
	if err != nil {
		return errors.Errorf("Error: %s", err)
	}

	finalState, err := api.WaitForTask(task)
	if err != nil {
		return errors.Errorf("Error when waiting for task to complete: %s", err)
	}

	log.Printf("Final migration task state: %s", finalState)

	if finalState == "SUCCEEDED" {
		log.Print("Migration completed successfully")
	} else {
		log.Print("Migration failed. Fetching log output...")
		time.Sleep(5 * time.Second)
		cliConnection.CliCommand("logs", "--recent", appName)
		return errors.New("FAILED")
	}

	if c.Replace {
		renamedSourceServiceName := fmt.Sprintf("%s-old", sourceServiceName)

		log.Printf("Renaming source service from %s to %s", sourceServiceName, renamedSourceServiceName)
		if _, err := cliConnection.CliCommandWithoutTerminalOutput("rename-service", sourceServiceName, renamedSourceServiceName); err != nil {
			return errors.Errorf("Error renaming source service from %s to %s: %s", sourceServiceName, renamedSourceServiceName, err)
		}

		log.Printf("Renaming destination service from %s to %s", destServiceName, sourceServiceName)
		if _, err := cliConnection.CliCommandWithoutTerminalOutput("rename-service", destServiceName, sourceServiceName); err != nil {
			return errors.Errorf("Error renaming destination service from %s to %s: %s", destServiceName, sourceServiceName, err)
		}
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

func main() {
	mysqlPlugin := &MySQLPlugin{}
	plugin.Start(mysqlPlugin)
}
