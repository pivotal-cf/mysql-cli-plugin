package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/service"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/ssh"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/user"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
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
			tunnerManager.Close()
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

		err = tunnerManager.Start(&srcServiceKey, &dstServiceKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
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
			"-u", srcServiceKey.Username,
			"-h", "127.0.0.1",
			"-P", strconv.Itoa(srcServiceKey.LocalSSHPort),
			srcServiceKey.DBName,
		}

		log.Printf("Executing 'mysqldump' with args %v", dumpArgs)
		dumpCmd := exec.Command("mysqldump", dumpArgs...)
		dumpCmd.Env = []string{"MYSQL_PWD=" + srcServiceKey.Password}
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
			"-u", dstServiceKey.Username,
			"-h", "127.0.0.1",
			"-P", strconv.Itoa(dstServiceKey.LocalSSHPort),
			"-D", dstServiceKey.DBName,
		}

		log.Printf("Executing 'mysql' with args %v", restoreArgs)
		restoreCmd := exec.Command("mysql", restoreArgs...)
		restoreCmd.Env = []string{"MYSQL_PWD=" + dstServiceKey.Password}
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
