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

package presentation_test

import (
	"bytes"
	"io/ioutil"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/presentation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("Report", func() {
	var fixtureTable string

	BeforeEach(func() {
		content, err := ioutil.ReadFile("fixtures/table.txt")
		Expect(err).NotTo(HaveOccurred())

		fixtureTable = string(content)

		format.TruncatedDiff = false
	})

	It("Prints a table to a write", func() {
		writer := bytes.Buffer{}
		bindings := []find_bindings.Binding{
			{
				Name:                "binding1",
				ServiceInstanceName: "instance1",
				ServiceInstanceGuid: "instance1-guid",
				OrgName:             "instance1-org",
				SpaceName:           "instance1-space",
				Type:                "ServiceKeyBinding",
			},
			{
				Name:                "app1",
				ServiceInstanceName: "instance2",
				ServiceInstanceGuid: "instance2-guid",
				OrgName:             "app1-org",
				SpaceName:           "app1-space",
				Type:                "AppBinding",
			},
		}

		presentation.Report(&writer, bindings)

		Expect(writer.String()).To(Equal(fixtureTable))
	})

	When("there are no bindings provided", func() {
		It("prints a helpful message", func() {
			writer := bytes.Buffer{}
			presentation.Report(&writer, nil)

			Expect(writer.String()).To(Equal("No bindings found.\n"))
		})
	})
})
