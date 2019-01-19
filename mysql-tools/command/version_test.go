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

package command_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/command"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate/migratefakes"
	"log"
)

var _ = Describe("Version Executor Commands", func() {
	var versionExecutor command.VersionExecutor
	var args []string
	var client *migratefakes.FakeClient

	BeforeEach(func() {
		versionExecutor = command.VersionExecutor{}
		client = &migratefakes.FakeClient{}
	})

	Context("Execute", func() {
		var (
			logOutput *bytes.Buffer
		)
		BeforeEach(func() {
			logOutput = &bytes.Buffer{}
			log.SetOutput(logOutput)
		})

		It("outputs the version to the user", func() {
			err := versionExecutor.Execute(client, args)
			Expect(err).NotTo(HaveOccurred())

			Expect(logOutput.String()).To(ContainSubstring(`built from source (unknown)`))
		})
	})
})
