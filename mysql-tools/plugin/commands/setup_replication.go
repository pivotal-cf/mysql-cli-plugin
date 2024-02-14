package commands

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

const (
	SetupReplicationUsage = `cf mysql-tools setup-replication [ --primary-target | -P ] [ --primary-instance | -p ] [ --secondary-target | -S ] [ --secondary-instance | -s ]`
)

func SetupReplication(args []string, ms MultiSite) error {
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

	primaryFoundation := opts.PrimaryTarget
	primaryInstance := opts.PrimaryInstance
	secondaryFoundation := opts.SecondaryTarget
	secondaryInstance := opts.SecondaryInstance

	err = ms.SetupReplication(primaryFoundation, primaryInstance, secondaryFoundation, secondaryInstance)
	if err != nil {
		return fmt.Errorf("replication setup error: %w", err)
	}

	return nil
}
