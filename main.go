package main

import (
	"github.com/pivotal-cf/mysql-v2-cli-plugin/cli_utils"

	"fmt"

	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/plugin"
)

// BasicPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type BasicPlugin struct{}

// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.

func (c *BasicPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	command := args[0]

	switch command {
	case "mysql-migrate":
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: cf mysql-migrate <v1-service-instance> <v2-service-instance>")
			os.Exit(1)
		}

		srcInstanceName := args[1]
		//dstInstanceName := args[2]

		tmpDir, err := ioutil.TempDir(os.TempDir(), "mysql-migrate")
		if cli_utils.PushApp(cliConnection, tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		defer func() {
			if err = cli_utils.DeleteApp(cliConnection); err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
		}()

		if err = cli_utils.CreateServiceKey(cliConnection, srcInstanceName); err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		defer func() {
			if err = cli_utils.DeleteServiceKey(cliConnection, srcInstanceName); err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
		}()

		serviceKey, err := cli_utils.GetServiceKey(cliConnection, srcInstanceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		closeTunnel, err := cli_utils.CreateSshTunnel(*serviceKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}

		defer closeTunnel()

		//create ssh tunnel instance1
		//mysqldump --single-transaction --add-drop-table --add-locks --create-options --disable-keys --extended-insert --quick --set-charset --routines --flush-privileges -u USERNAME -p -h 0 -P 63306 DB_NAME > backup.sql
		//create ssh tunnel instance2
		//mysql -u USERNAME -p -h 0 -P 63306 -D service_instance_db < backup.sql
		//validations
	}
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *BasicPlugin) GetMetadata() plugin.PluginMetadata {
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

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(BasicPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
