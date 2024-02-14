package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
)

//counterfeiter:generate -o fakes/fake_migrator.go . Migrator
type Migrator interface {
	CheckServiceExists(instanceName string) error
	CreateServiceInstance(planName, instanceName string) error
	CleanupOnError(instanceName string) error
	MigrateData(opts migrate.MigrateOptions) error
	RenameServiceInstances(donorInstanceName, recipientInstanceName string) error
}

func Migrate(args []string, migrator Migrator) error {
	const (
		migrateUsage = `cf mysql-tools migrate [-h] [--no-cleanup] [--skip-tls-validation] <source-service-instance> <p.mysql-plan-type>`
	)

	var opts struct {
		Args struct {
			Source   string `positional-arg-name:"<source-service-instance>"`
			PlanName string `positional-arg-name:"<p.mysql-plan-type>"`
		} `positional-args:"yes" required:"yes"`
		NoCleanup         bool `long:"no-cleanup" description:"don't clean up migration app and new service instance after a failed migration"`
		SkipTLSValidation bool `long:"skip-tls-validation" short:"k" description:"Skip certificate validation of the MySQL server certificate. Not recommended!"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools migrate"
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", migrateUsage, msg)
	}
	donorInstanceName := opts.Args.Source
	tempRecipientInstanceName := donorInstanceName + "-new"
	destPlan := opts.Args.PlanName
	cleanup := !opts.NoCleanup
	skipTLSValidation := opts.SkipTLSValidation

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	log.Printf("Warning: The mysql-tools migrate command will not migrate any triggers, routines or events.")
	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	log.Printf("Creating new service instance %q for service %s using plan %s", tempRecipientInstanceName, productName, destPlan)
	if err := migrator.CreateServiceInstance(destPlan, tempRecipientInstanceName); err != nil {
		if cleanup {
			_ = migrator.CleanupOnError(tempRecipientInstanceName)
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

	migrationOptions := migrate.MigrateOptions{
		DonorInstanceName:     donorInstanceName,
		RecipientInstanceName: tempRecipientInstanceName,
		Cleanup:               cleanup,
		SkipTLSValidation:     skipTLSValidation,
	}

	if err := migrator.MigrateData(migrationOptions); err != nil {
		if cleanup {
			_ = migrator.CleanupOnError(tempRecipientInstanceName)

			return fmt.Errorf(
				"error migrating data: %w. Attempting to clean up service %s",
				err,
				tempRecipientInstanceName,
			)
		}

		return fmt.Errorf("error migrating data: %v. Not cleaning up service %s",
			err,
			tempRecipientInstanceName,
		)
	}

	return migrator.RenameServiceInstances(donorInstanceName, tempRecipientInstanceName)
}
