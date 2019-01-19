package command

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/cf"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin_errors"
)

const migrateUsage = `cf mysql-tools migrate [-h] [--no-cleanup] <source-service-instance> <p.mysql-plan-type>`

//go:generate counterfeiter . Router
type Router interface {
	Match(command string, cliConnection plugin.CliConnection, args []string) error
}

//go:generate counterfeiter . CommandRunner
type CommandRunner interface {
	Execute(client migrate.Client, args []string) error
}

type MySQCmdLRouter struct {
	Routes map[string]interface{ CommandRunner }
}

func (mcmdr *MySQCmdLRouter) Match(command string, cliConnection plugin.CliConnection, args []string) error {
	client := cf.NewClient(cliConnection)

	if selectedCmd, ok := mcmdr.Routes[command]; ok {
		err := selectedCmd.Execute(client, args)

		if err != nil {
			return plugin_errors.NewCustomUsageError(err.Error(), migrateUsage)
		}
	} else {
		return plugin_errors.NewUsageError(fmt.Sprintf("unknown command: %q", command))
	}

	return nil
}
