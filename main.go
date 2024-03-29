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
	"fmt"
	"os"

	cliplugin "code.cloudfoundry.org/cli/plugin"

	"github.com/pivotal-cf/mysql-cli-plugin/app"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin"
)

func main() {
	mysqlPlugin := &plugin.MySQLPlugin{
		MigrationAppExtractor: app.NewExtractor(),
		MultisiteConfig:       multisite.NewConfig(),
	}

	cliplugin.Start(mysqlPlugin)
	if err := mysqlPlugin.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
