package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/gobuffalo/packr"
	"github.com/pivotal-cf/mysql-cli-plugin/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/user"
)

type MySQLPlugin struct {
	exitStatus int
}

//go:generate packr --compress
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
		api = cf.NewApi(cliConnection)
		sourceServiceName = args[2]
		destServiceName = args[3]
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

	box := packr.NewBox("./app")
	tmpDir, err := ioutil.TempDir(os.TempDir(), "migrate_app_")
	if err != nil {
		log.Printf("Error creating temp directory: %s", err)
		c.exitStatus = 1
		return
	}

	err = box.Walk(func(name string, file packr.File) error {
		info, err := file.FileInfo()
		if err != nil {
			log.Printf("Failed to state fileinfo: %s", err)
			return err
		}
		log.Printf("box.path: %s [%d]", name, info.Size())

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
		log.Printf("Error extracting migrate assets: %s", err)
		c.exitStatus = 1
		return
	}

	log.Printf("Pushing app from %s", tmpDir)

	_, err = cliConnection.CliCommand("push",
		"migrate-app",
		"-b", "binary_buildpack",
		"-u", "none",
		"-c", "sleep infinity",
		"-p", tmpDir,
		"--no-route",
		"--no-start",
	)
	if err != nil {
		log.Printf("failed to push application: %s", err)
		c.exitStatus = 1
		return
	}
	defer cliConnection.CliCommand("delete", "migrate-app", "-f")

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

	app, err := api.GetAppByName("migrate-app")
	if err != nil {
		log.Printf("Error: %s", err)
		c.exitStatus = 1
		return
	}

	cmd := fmt.Sprintf(`./migrate %s %s`, sourceServiceName, destServiceName)
	task, err := api.CreateTask(app, cmd)
	if err != nil {
		log.Printf("Error: %s", err)
		c.exitStatus = 1
		return
	}

	for task.State != "SUCCEEDED" && task.State != "FAILED" {
		time.Sleep(1 * time.Second)
		task, err = api.GetTaskByGUID(task.Guid)
		if err != nil {
			log.Printf("Error: %s", err)
			c.exitStatus = 1
			return
		}
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
