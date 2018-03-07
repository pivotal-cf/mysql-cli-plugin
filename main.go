package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/cfapi"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/user"
)

type MySQLPlugin struct {
	exitStatus int
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	if len(args) >= 2 && args[1] != "migrate" {
		log.Printf("Unknown command '%s'", args[1])
		c.exitStatus = 1
		return
	}

	if len(args) != 4 {
		log.Println("Usage: cf mysql-tools migrate <v1-service-instance> <v2-service-instance>")
		c.exitStatus = 1
		return
	}

	var (
		user = user.NewReporter(cliConnection)
	)

	ok, err := user.IsSpaceDeveloper()
	if err != nil {
		log.Printf("Error getting user information: %v", err)
		c.exitStatus = 1
		return
	}

	if !ok {
		log.Println("You must have the 'Space Developer' privilege to use the 'cf mysql migrate' command")
		c.exitStatus = 1
		return
	}

	_, err = cliConnection.CliCommand("push",
		"migrate-app",
		"-b", "binary_buildpack",
		"-u", "none",
		"-c", "sleep infinity",
		"-p", "./app",
		"--no-start",
	)
	if err != nil {
		log.Printf("failed to push application: %s", err)
		c.exitStatus = 1
		return
	}

	sourceServiceName := args[2]
	destServiceName := args[3]

	if _, err := cliConnection.CliCommand("bind-service", "migrate-app", sourceServiceName); err != nil {
		log.Printf("failed to bind-service %q to application %q: %s", "migrate-app", sourceServiceName, err)
		c.exitStatus = 1
		return
	}

	if _, err := cliConnection.CliCommand("bind-service", "migrate-app", destServiceName); err != nil {
		log.Printf("failed to bind-service %q to application %q: %s", "migrate-app", destServiceName, err)
		c.exitStatus = 1
		return
	}

	if _, err := cliConnection.CliCommand("start", "migrate-app"); err != nil {
		log.Printf("failed to start application %q: %s", "migrate-app", err)
		c.exitStatus = 1
		return
	}

	app := cfapi.GetAppByName("migrate-app", cliConnection)

	cmd := fmt.Sprintf(`bin/migrate %s %s`, sourceServiceName, destServiceName)
	task := cfapi.CreateTask(app, cmd, cliConnection)

	for task.State != "SUCCEEDED" && task.State != "FAILED" {
		time.Sleep(1 * time.Second)
		task = cfapi.GetTaskByGUID(task.Guid, cliConnection)
	}

	log.Printf("Done: %s", task.State)
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
	os.Exit(mysqlPlugin.exitStatus)
}
