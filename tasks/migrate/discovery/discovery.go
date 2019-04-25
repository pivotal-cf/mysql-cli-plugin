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

package discovery

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"strings"

	"github.com/pkg/errors"
)

func DiscoverDatabases(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SHOW DATABASES`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query the database")
	}

	var dbs []string

	filterSchemas := map[string]struct{}{
		"cf_metadata":        {},
		"information_schema": {},
		"mysql":              {},
		"performance_schema": {},
		"sys":                {},
	}

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, errors.Wrap(err, "failed to scan the list of databases")
		}

		if _, ok := filterSchemas[dbName]; ok {
			continue
		}

		dbs = append(dbs, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to parse the list of databases")
	}

	if len(dbs) == 0 {
		return nil, fmt.Errorf("no databases found")
	}

	return dbs, nil
}

type View struct {
	Schema    string
	TableName string
}

func (v View) String() string {
	return fmt.Sprintf("%s.%s", v.Schema, v.TableName)
}

func QuoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

func discoverViews(db *sql.DB, schema string) (views []View, err error) {
	findViewsQuery := `SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`
	rows, err := db.Query(findViewsQuery, schema)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve views for %s schema", schema)
	}

	for rows.Next() {
		var (
			view View
		)
		if err := rows.Scan(&view.TableName); err != nil {
			return nil, errors.Wrap(err, "failed to scan the list of views")
		}
		view.Schema = schema

		views = append(views, view)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to prepare the list of views")
	}

	return views, nil
}

func DiscoverInvalidViews(db *sql.DB, schemas []string) ([]View, error) {
	var invalidViews []View
	for _, schema := range schemas {
		views, err := discoverViews(db, schema)
		if err != nil {
			return nil, err
		}

		for _, view := range views {
			checkInvalidViewQuery := fmt.Sprintf(`SHOW FIELDS FROM %s IN %s`, QuoteIdentifier(view.TableName), QuoteIdentifier(view.Schema))
			if _, err := db.Exec(checkInvalidViewQuery); err != nil {
				if _, ok := err.(*mysql.MySQLError); ok {
					invalidViews = append(invalidViews, view)
				} else {
					return nil, errors.Wrapf(err, "Unexpected error when validating view %q.%q", view.Schema, view.TableName)
				}
			}
		}
	}

	return invalidViews, nil
}

func DiscoverExistingData(destinationDb *sql.DB, sourceSchemas []string) ([]string, error) {
	var badSchemas []string
	var returnErr error

	destinationSchemas, err := DiscoverDatabases(destinationDb)
	if err != nil {
		return nil, err
	}

	// We only care about testing schemas in the target server that exist in the source server
	var candidateSchemas []string
	for _, sourceSchema := range sourceSchemas {
		for _, destinationSchema := range destinationSchemas {
			if sourceSchema == destinationSchema {
				candidateSchemas = append(candidateSchemas, sourceSchema)
			}
		}
	}

	for _, schema := range candidateSchemas {
		checkExistingTables := fmt.Sprintf(`SHOW TABLES FROM %s`, schema)
		rows, err := destinationDb.Query(checkExistingTables)
		if err != nil {
			return nil, err
		}
		if rows.Next() {
			badSchemas = append(badSchemas, schema)
			returnErr = errors.Errorf("Migration target database already contains tables!  Giving up...")
		}
	}

	return badSchemas, returnErr
}
