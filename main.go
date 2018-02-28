package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/service"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/user"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/ssh"
	"time"
)

type MySQLPlugin struct {
	exitStatus int
}

func (c *MySQLPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	command := args[0]

	switch command {
	case "mysql-migrate":
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: cf mysql-migrate <v1-service-instance> <v2-service-instance>")
			c.exitStatus = 1
			return
		}

		tmpDir, err := ioutil.TempDir(os.TempDir(), "mysql-migrate")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary directory: %v", err)
			c.exitStatus = 1
			return
		}

		var (
			appName         = "static-app"
			srcInstanceName = args[1]
			dstInstanceName = args[2]
			user            = user.NewReporter(cliConnection)
			srcInstance     = service.NewServiceInstance(cliConnection, srcInstanceName)
			dstInstance     = service.NewServiceInstance(cliConnection, dstInstanceName)
			tunnerManager   = ssh.NewTunnerManager(cliConnection, ssh.NewDB(), tmpDir, time.Minute)
		)

		ok, err := user.IsSpaceDeveloper()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting user information: %v", err)
			c.exitStatus = 1
			return
		}

		if !ok {
			fmt.Fprintln(os.Stderr, "You must have the 'Space Developer' privilege to use the 'cf mysql migrate' command")
			c.exitStatus = 1
			return
		}

		defer func() {
			os.RemoveAll(tmpDir)
			srcInstance.Cleanup()
			dstInstance.Cleanup()
		}()

		srcServiceKey, err := srcInstance.ServiceInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			c.exitStatus = 1
			return
		}

		dstServiceKey, err := dstInstance.ServiceInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			c.exitStatus = 1
			return
		}

		err = tunnerManager.PushApp(tmpDir)
		if err != nil {
			c.exitStatus = 1
			return
		}

		agentUser := "lf-agent"
		agentPassword := "REPLACE-ME"

		curlCmd := fmt.Sprintf(`curl -X POST http://%s:%s@%s:8443/migrate -d '{"source":"%s", "dest": "%s"}'`,
			agentUser, agentPassword, dstServiceKey.Hostname, srcServiceKey, dstServiceKey)

		_, err = cliConnection.CliCommand("ssh", appName, "-c", curlCmd)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error running ssh: %v", err)
			return
		}
	}
}

func (c *MySQLPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "MysqlMigrate",
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
				Name:     "mysql-migrate",
				HelpText: "Plugin to migrate mysql instances",
				UsageDetails: plugin.Usage{
					Usage: "mysql-migrate\n   cf mysql-migrate <v1-service-instance> <v2-service-instance>",
				},
			},
		},
	}
}

func main() {

	mysqlPlugin := new(MySQLPlugin)
	plugin.Start(mysqlPlugin)
	os.Exit(mysqlPlugin.exitStatus)

}
