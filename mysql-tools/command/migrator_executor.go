package command

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

func NewMigratorExecutor(client migrate.Client, unpacker migrate.Unpacker) *MigratorExecutor {
	migrator := migrate.NewMigrator(client, unpacker)
	return &MigratorExecutor{
		Migrator: migrator,
	}

}



type MigratorExecutor struct {
	Migrator Migrator
}



func (me *MigratorExecutor) Migrate(args []string) error {

	var opts struct {
		Args struct {
			Source   string `positional-arg-name:"<source-service-instance>"`
			PlanName string `positional-arg-name:"<p.mysql-plan-type>"`
		} `positional-args:"yes" required:"yes"`
		NoCleanup bool `long:"no-cleanup" description:"don't clean up migration app and new service instance after a failed migration'"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools migrate"
	parser.Args()
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return errors.Errorf("%s", msg)
	}
	donorInstanceName := opts.Args.Source
	tempRecipientInstanceName := donorInstanceName + "-new"
	destPlan := opts.Args.PlanName
	cleanup := !opts.NoCleanup

	if err := me.Migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	log.Printf("Warning: The mysql-tools migrate command will not migrate any triggers, routines or events.")
	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	log.Printf("Creating new service instance %q for service %s using plan %s", tempRecipientInstanceName, productName, destPlan)
	if err := me.Migrator.CreateAndConfigureServiceInstance(destPlan, tempRecipientInstanceName); err != nil {
		if cleanup {
			me.Migrator.CleanupOnError(tempRecipientInstanceName)
			return fmt.Errorf("error creating service instance: %v. Attempting to clean up service %s",
				err,
				tempRecipientInstanceName,
			)
		}

		return fmt.Errorf("error creating service instance: %v. Not cleaning up service %s",
			err,
			tempRecipientInstanceName,
		)
	}

	if err := me.Migrator.MigrateData(donorInstanceName, tempRecipientInstanceName, cleanup); err != nil {
		if cleanup {
			me.Migrator.CleanupOnError(tempRecipientInstanceName)

			return fmt.Errorf("error migrating data: %v. Attempting to clean up service %s",
				err,
				tempRecipientInstanceName,
			)
		}

		return fmt.Errorf("error migrating data: %v. Not cleaning up service %s",
			err,
			tempRecipientInstanceName,
		)
	}

	return me.Migrator.RenameServiceInstances(donorInstanceName, tempRecipientInstanceName)
}
