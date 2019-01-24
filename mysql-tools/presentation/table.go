// Copyright (C) 2019-Present Pivotal Software, Inc. All rights reserved.
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

package presentation

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
)

type BindingSet []find_bindings.Binding

func (bs BindingSet) ToRows() [][]string {
	var result [][]string
	for _, b := range bs {
			result = append(result, []string{
				b.ServiceInstanceName,
				b.ServiceInstanceGuid,
				b.OrgName,
				b.SpaceName,
				b.Name,
				b.Type,
			})
	}

	return result
}

func (bs BindingSet) Header() []string {
	return []string{
		"Service",
		"Service GUID",
		"Org",
		"Space",
		"App or Service Key",
		"Type",
	}
}

func Report(w io.Writer, bindings BindingSet) {
	if len(bindings) == 0 {
		fmt.Fprintln(w, "No bindings found.")
		return
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(bindings.Header())
	table.SetRowLine(true)
	table.AppendBulk(bindings.ToRows())
	table.Render()
}
