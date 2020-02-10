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
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/mysql"
)

func DiscoverDatabases(creds mysql.Credentials) ([]string, error) {
	var (
		sql    = "SHOW DATABASES WHERE `Database` NOT IN ('cf_metadata', 'information_schema', 'mysql', 'performance_schema', 'sys')"
		input  = bytes.NewBufferString(sql)
		output = bytes.Buffer{}
	)

	cmd := mysql.MySQLCmd(creds)
	cmd.Args = append(cmd.Args, "--batch", "--skip-column-names")
	cmd.Stdin = input
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return nil, err // wraap. aalso log stderr?
	}

	var result []string

	r := strings.NewReplacer(`\t`, "\t", `\\n`, "\n", `\\`, `\`)
	for _, name := range strings.Split(output.String(), "\n") {
		if name == "" {
			continue
		}
		result = append(result, r.Replace(name))
	}

	if len(result) == 0 {
		return nil, errors.New("no databases found")
	}

	return result, nil
}

func DiscoverInvalidViews(creds mysql.Credentials) ([]string, error) {
	var (
		sql = "SELECT `vws`.`TABLE_SCHEMA`, `vws`.`TABLE_NAME` " +
			"FROM (" +
			"SELECT `TABLE_SCHEMA`, `TABLE_NAME` " +
			"FROM `information_schema`.`TABLES` " +
			"WHERE `TABLE_SCHEMA` NOT IN ('cf_metadata', 'information_schema', 'mysql', 'performance_schema', 'sys') " +
			"AND `TABLE_TYPE`='VIEW' " +
			"AND `TABLE_ROWS` IS NULL " +
			"AND `TABLE_COMMENT` LIKE '%invalid%') vws"
		input  = bytes.NewBufferString(sql)
		output = bytes.Buffer{}
	)

	cmd := mysql.MySQLCmd(creds)
	cmd.Args = append(cmd.Args, "--batch", "--skip-column-names")
	cmd.Stdin = input
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var result []string

	r := strings.NewReplacer(`\t`, "\t", `\\n`, "\n", `\\`, `\`)
	for _, row := range strings.Split(output.String(), "\n") {
		if row == "" {
			continue
		}

		fields := strings.SplitN(row, "\t", 2)
		if len(fields) != 2 {
			log.Printf("Error while scanning for invalid views: expected exactly one schema name and view name but found %q", row)
			continue
		}

		result = append(result, fmt.Sprintf("%s.%s",
			r.Replace(fields[0]),
			r.Replace(fields[1]),
		))
	}

	return result, nil
}
