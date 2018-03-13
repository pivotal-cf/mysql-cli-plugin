package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/gobuffalo/packr"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/user"
	"github.com/pkg/errors"
)

type MySQLPlugin struct{}

//go:generate go install github.com/pivotal-cf/mysql-cli-plugin/vendor/github.com/gobuffalo/packr/...
//go:generate $GOPATH/bin/packr --compress
func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	if len(args) >= 2 && args[1] != "migrate" {
		log.Printf("Unknown command '%s'", args[1])
		os.Exit(1)
		return
	}

	if len(args) != 4 {
		log.Println("Usage: cf mysql-tools migrate <v1-service-instance> <v2-service-instance>")
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

func (c *MySQLPlugin) run(cliConnection plugin.CliConnection, sourceServiceName, destServiceName string) error {
	var (
		user = user.NewReporter(cliConnection)
		api  = cf.NewApi(cliConnection)
	)

	ok, err := user.IsSpaceDeveloper()
	if err != nil {
		return errors.Errorf("Error getting user information: %v", err)
	}

	if !ok {
		return errors.New("You must have the 'Space Developer' privilege to use the 'cf mysql migrate' command")
	}

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

		return dest.Chmod(0700)
	})

	if err != nil {
		return errors.Errorf("Error extracting migrate assets: %s", err)
	}

	log.Print("Started to push app")

	_, err = cliConnection.CliCommandWithoutTerminalOutput("push",
		"migrate-app",
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
		cliConnection.CliCommandWithoutTerminalOutput("delete", "migrate-app", "-f")
		log.Print("Cleaning up...")
	}()
	log.Print("Sucessfully pushed app")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("bind-service", "migrate-app", sourceServiceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", "migrate-app", sourceServiceName, err)
	}
	log.Print("Sucessfully bound app to v1 instance")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("bind-service", "migrate-app", destServiceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", "migrate-app", destServiceName, err)
	}
	log.Print("Sucessfully bound app to v2 instance")

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("start", "migrate-app"); err != nil {
		return errors.Errorf("failed to start application %q: %s", "migrate-app", err)
	}

	app, err := api.GetAppByName("migrate-app")
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
		return nil
	} else {
		log.Print("Migration failed")
		cliConnection.CliCommand("logs", "--recent", "migrate-app")
		return errors.New("FAILED")
	}
}

func (c *MySQLPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "MysqlTools",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
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
					Usage: "mysql-tools\n   cf mysql-tools migrate <v1-service-instance> <v2-service-instance>",
				},
			},
		},
	}
}

func main() {
	mysqlPlugin := &MySQLPlugin{}
	plugin.Start(mysqlPlugin)
}
