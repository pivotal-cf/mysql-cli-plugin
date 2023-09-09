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
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func DiscoverDatabases(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SHOW DATABASES`)
	if err != nil {
		return nil, fmt.Errorf("failed to query the database: %w", err)
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
			return nil, fmt.Errorf("failed to scan the list of databases: %w", err)
		}

		if _, ok := filterSchemas[dbName]; ok {
			continue
		}

		dbs = append(dbs, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse the list of databases: %w", err)
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
		return nil, fmt.Errorf("failed to retrieve views for %s schema: %w", schema, err)
	}

	for rows.Next() {
		var view View
		if err := rows.Scan(&view.TableName); err != nil {
			return nil, fmt.Errorf("failed to scan the list of views: %w", err)
		}
		view.Schema = schema

		views = append(views, view)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to prepare the list of views: %w", err)
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
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &mysqlErr) {
					invalidViews = append(invalidViews, view)
				} else {
					return nil, fmt.Errorf("Unexpected error when validating view %q.%q: %w", view.Schema, view.TableName, err)
				}
			}
		}
	}

	return invalidViews, nil
}
