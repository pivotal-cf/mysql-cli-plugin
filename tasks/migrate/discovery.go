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
	"fmt"

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

func DiscoverInvalidViews(db *sql.DB, schemas []string) ([]View, error) {
	var invalidViews []View
	for _, schema := range schemas {
		rows, err := db.Query(`SELECT table_name from INFORMATION_SCHEMA.VIEWS WHERE table_schema = ?`, schema)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to retrieve views for %s schema", schema)
		}

		for rows.Next() {
			var view string
			if err := rows.Scan(&view); err != nil {
				return nil, errors.Wrap(err, "failed to scan the list of views")
			}

			_, err := db.Exec(`SHOW FIELDS FROM ? FROM ?`, view, schema)
			if err != nil {
				invalidViews = append(invalidViews, View{Schema: schema, TableName: view})
			}
		}

		if err := rows.Err(); err != nil {
			return nil, errors.Wrap(err, "failed to prepare the list of views")
		}
	}
	return invalidViews, nil
}
