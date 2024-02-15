package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/jessevdk/go-flags"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite/foundation"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/must"
)

const (
	SwitchoverReplicationUsage = `cf mysql-tools switchover [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ] [ --force | -f ]`
)

func SwitchoverReplication(args []string, cfg MultisiteConfig, out io.Writer, in io.Reader) error {
	var opts struct {
		PrimaryTarget     string `short:"P" long:"primary-target" required:"true"`
		PrimaryInstance   string `short:"p" long:"primary-instance" required:"true"`
		SecondaryTarget   string `short:"S" long:"secondary-target" required:"true"`
		SecondaryInstance string `short:"s" long:"secondary-instance" required:"true"`
		Force             bool   `short:"f" long:"force"`
	}
	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools setup-replication"
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", SwitchoverReplicationUsage, msg)
	}

	if !opts.Force {
		u := must.SucceedWithValue(ui.NewUI(&configv3.Config{}))
		u.Out = out
		u.OutForInteration = out
		u.In = in

		shouldProceed, err := u.DisplayBoolPrompt(false, fmt.Sprintf("When successful, %s will become secondary and %s will become primary. Do you want to continue?", opts.PrimaryInstance, opts.SecondaryInstance), nil)

		if !shouldProceed || err != nil {
			u.DisplayText("Operation cancelled")
			return nil
		}
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	primary := foundation.New(opts.PrimaryTarget, cfg.ConfigDir(opts.PrimaryTarget))
	secondary := foundation.New(opts.SecondaryTarget, cfg.ConfigDir(opts.SecondaryTarget))
	workflow := multisite.NewWorkflow(primary, secondary, logger)

	if err = workflow.SwitchoverReplication(opts.PrimaryInstance, opts.SecondaryInstance); err != nil {
		return err
	}

	return nil
}
