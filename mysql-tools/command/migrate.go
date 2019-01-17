// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under the terms of the under the Apache License,
// Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy
// of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
// TODO: delete this package
package command

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

//go:generate counterfeiter . Migrator
type Migrator interface {
	CheckServiceExists(donorInstanceName string) error
	CreateAndConfigureServiceInstance(planType, serviceName string) error
	MigrateData(donorInstanceName, recipientInstanceName string, cleanup bool) error
	RenameServiceInstances(donorInstanceName, recipientInstanceName string) error
	CleanupOnError(recipientInstanceName string) error
}

func Migrate(migrator Migrator, args []string) error {
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
		// show migrate usage?
		return errors.Errorf("%s", msg)
	}
	///  ^^ split this out


	donorInstanceName := opts.Args.Source
	tempRecipientInstanceName := donorInstanceName + "-new"
	destPlan := opts.Args.PlanName
	cleanup := !opts.NoCleanup

	if err := migrator.CheckServiceExists(donorInstanceName); err != nil {
		return err
	}

	log.Printf("Warning: The mysql-tools migrate command will not migrate any triggers, routines or events.")
	productName := os.Getenv("RECIPIENT_PRODUCT_NAME")
	if productName == "" {
		productName = "p.mysql"
	}

	log.Printf("Creating new service instance %q for service %s using plan %s", tempRecipientInstanceName, productName, destPlan)
	if err := migrator.CreateAndConfigureServiceInstance(destPlan, tempRecipientInstanceName); err != nil {
		if cleanup {
			migrator.CleanupOnError(tempRecipientInstanceName)
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

	if err := migrator.MigrateData(donorInstanceName, tempRecipientInstanceName, cleanup); err != nil {
		if cleanup {
			migrator.CleanupOnError(tempRecipientInstanceName)

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

	return migrator.RenameServiceInstances(donorInstanceName, tempRecipientInstanceName)
}
