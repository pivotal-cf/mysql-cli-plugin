package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils"

	"code.cloudfoundry.org/cli/plugin"
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

		srcInstanceName := args[1]
		dstInstanceName := args[2]

		// TODO clean tempdir later?
		tmpDir, err := ioutil.TempDir(os.TempDir(), "mysql-migrate")
		if cli_utils.PushApp(cliConnection, tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error pushing app: %v", err)
			c.exitStatus = 1
			return
		}

		defer func() {
			if err = cli_utils.DeleteApp(cliConnection); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting app: %v", err)
				c.exitStatus = 1
			}
		}()

		if err = cli_utils.CreateServiceKey(cliConnection, srcInstanceName); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating service key for instance %s: %v", srcInstanceName, err)
			c.exitStatus = 1
			return
		}

		defer func() {
			if err = cli_utils.DeleteServiceKey(cliConnection, srcInstanceName); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting service key for instance %s: %v", srcInstanceName, err)
				c.exitStatus = 1
			}
		}()

		if err = cli_utils.CreateServiceKey(cliConnection, dstInstanceName); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating service key for instance %s: %v", dstInstanceName, err)
			c.exitStatus = 1
			return
		}

		defer func() {
			if err = cli_utils.DeleteServiceKey(cliConnection, dstInstanceName); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting service key for instance %s: %v", dstInstanceName, err)
				c.exitStatus = 1
			}
		}()

		srcInstanceKey, err := cli_utils.GetServiceKey(cliConnection, srcInstanceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching service key for %s: %v", srcInstanceName, err)
			c.exitStatus = 1
			return
		}

		dstInstanceKey, err := cli_utils.GetServiceKey(cliConnection, dstInstanceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching service key for %s: %v", dstInstanceName, err)
			c.exitStatus = 1
			return
		}

		tunnels := []cli_utils.Tunnel{
			{
				ServiceKey: *srcInstanceKey,
			},
			{
				ServiceKey: *dstInstanceKey,
			},
		}

		tunnel, err := cli_utils.NewTunnelManager(cliConnection, tunnels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tunnel: %v", err)
			c.exitStatus = 1
			return
		}

		tunnel.AppName = "static-app"

		go func() {
			err = tunnel.CreateSSHTunnel()
			// NOTE this will exit with an error when we delete the app
			// TODO replace the tunnel with something that uses golang/x/crypto/ssh
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				c.exitStatus = 1
			}
		}()

		log.Println("Waiting for tunnel to come online")
		err = tunnel.WaitForTunnel(60 * time.Second)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error waiting for tunnel to app %s: %v", tunnel.AppName, err)
			c.exitStatus = 1
			return
		}

		// TODO come up with better name
		path := "mysql-v2-migrate.sql"
		tmpFile, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tempfile %s: %v", path, err)
			c.exitStatus = 1
			return
		}

		defer tmpFile.Close()

		dumpArgs := []string{
			"--routines",
			"--set-gtid-purged=off",
			"-u", tunnel.Tunnels[0].ServiceKey.Username,
			"-h", "127.0.0.1",
			"-P", strconv.Itoa(tunnel.Tunnels[0].Port),
			tunnel.Tunnels[0].ServiceKey.DBName,
		}

		log.Printf("Executing 'mysqldump' with args %v", dumpArgs)
		dumpCmd := exec.Command("mysqldump", dumpArgs...)
		dumpCmd.Env = []string{"MYSQL_PWD=" + tunnel.Tunnels[0].ServiceKey.Password}
		dumpCmd.Stderr = os.Stderr
		dumpCmd.Stdout = tmpFile

		err = dumpCmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error dumping database: %v", err)
			c.exitStatus = 1
			return
		}

		tmpFile.Seek(0, 0)

		restoreArgs := []string{
			"-u", tunnel.Tunnels[1].ServiceKey.Username,
			"-h", "127.0.0.1",
			"-P", strconv.Itoa(tunnel.Tunnels[1].Port),
			"-D", tunnel.Tunnels[1].ServiceKey.DBName,
		}

		log.Printf("Executing 'mysql' with args %v", restoreArgs)
		restoreCmd := exec.Command("mysql", restoreArgs...)
		restoreCmd.Env = []string{"MYSQL_PWD=" + tunnel.Tunnels[1].ServiceKey.Password}
		restoreCmd.Stderr = os.Stderr
		restoreCmd.Stdout = os.Stdout
		restoreCmd.Stdin = tmpFile

		err = restoreCmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error restoring database: %v", err)
			c.exitStatus = 1
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

				// UsageDetails is optional
				// It is used to show help of usage of each command
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
