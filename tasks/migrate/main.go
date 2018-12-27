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
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
)

var (
	VcapCredentials = os.Getenv("VCAP_SERVICES")
)

func RegisterTLSConfig(key, ca string) {
	var rootCAs *x509.CertPool
	if ca != "" {
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM([]byte(ca))
	}
	mysql.RegisterTLSConfig(key, &tls.Config{
		RootCAs: rootCAs,
	})
}

func main() {
	var (
		sourceInstance string
		destInstance   string
	)

	if len(os.Args) < 3 {
		log.Fatal("Usage: migrate <source service> <target service>")
	}

	sourceInstance = os.Args[1]
	destInstance = os.Args[2]

	sourceCredentials, err := InstanceCredentials(sourceInstance, VcapCredentials)
	if err != nil {
		log.Fatalf("Failed to lookup source credentials: %v", err)
	}

	destCredentials, err := InstanceCredentials(destInstance, VcapCredentials)
	if err != nil {
		log.Fatalf("Failed to lookup destination credentials: %v", err)
	}

	RegisterTLSConfig("default", sourceCredentials.TLS.Cert.CA)

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
