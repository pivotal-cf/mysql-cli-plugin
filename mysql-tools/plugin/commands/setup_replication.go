package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/foundation"
)

const (
	SetupReplicationUsage = `cf mysql-tools setup-replication [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ]`
)

func SetupReplication(args []string, cfg MultisiteConfig) error {
	var opts struct {
		PrimaryTarget     string `short:"P" long:"primary-target" required:"true"`
		PrimaryInstance   string `short:"p" long:"primary-instance" required:"true"`
		SecondaryTarget   string `short:"S" long:"secondary-target" required:"true"`
		SecondaryInstance string `short:"s" long:"secondary-instance" required:"true"`
	}
	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools setup-replication"
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", SetupReplicationUsage, msg)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	primary := foundation.New(opts.PrimaryTarget, cfg.ConfigDir(opts.PrimaryTarget))
	secondary := foundation.New(opts.SecondaryTarget, cfg.ConfigDir(opts.SecondaryTarget))
	workflow := multisite.NewWorkflow(primary, secondary, logger)

	if err = workflow.SetupReplication(opts.PrimaryInstance, opts.SecondaryInstance); err != nil {
		return err
	}

	return nil
}
