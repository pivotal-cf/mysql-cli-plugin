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

package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
)

var (
	VcapCredentials = os.Getenv("VCAP_SERVICES")
)

func main() {
	var (
		sourceInstance    string
		destInstance      string
		skipTLSValidation bool
	)

	flag.BoolVar(&skipTLSValidation, "skip-tls-validation", false, "Skip certificate validation of the MySQL server certificate.  Not recommended!")
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		log.Fatal("Usage: migrate <source service> <target service>")
	}

	sourceInstance = args[0]
	destInstance = args[1]

	sourceCredentials, err := InstanceCredentials(sourceInstance, VcapCredentials)
	if err != nil {
		log.Fatalf("Failed to lookup source credentials: %v", err)
	}

	destCredentials, err := InstanceCredentials(destInstance, VcapCredentials)
	if err != nil {
		log.Fatalf("Failed to lookup destination credentials: %v", err)
	}

	db, err := sql.Open("mysql", sourceCredentials.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize source connection: %v", err)
	}

	sourceSchemas, err := discovery.DiscoverDatabases(db)
	if err != nil {
		log.Fatalf("Failed to discover schemas: %v", err)
	}

	invalidViews, err := discovery.DiscoverInvalidViews(db, sourceSchemas)
	if err != nil {
		log.Fatalf("Failed to retrieve invalid views: %v", err)
	}

	if len(invalidViews) > 0 {
		log.Printf("The following views are invalid, and will not be migrated: %s\n", invalidViews)
	}

	mySQLDumpCmd := MySQLDumpCmd(sourceCredentials, invalidViews, sourceSchemas...)
	mySQLCmd := MySQLCmd(destCredentials)
	replaceCmd := ReplaceDefinerCmd()

	if err := CopyData(mySQLDumpCmd, replaceCmd, mySQLCmd); err != nil {
		log.Fatalf("Failed to copy data: %v", err)
	}
}
