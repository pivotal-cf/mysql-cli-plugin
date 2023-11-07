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

package acceptance

import (
	"os"
	"testing"

	"github.com/cloudfoundry/cf-test-helpers/v2/config"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite", Label("acceptance"))
}

var (
	TestSetup        *workflowhelpers.ReproducibleTestSuiteSetup
	Config           *config.Config
	migrationTimeout string
)

var _ = BeforeSuite(func() {
	test_helpers.CheckForRequiredEnvVars([]string{
		"CONFIG",
		"DC1_CONFIG",
		"DC2_CONFIG",
		"DONOR_PLAN_NAME",
		"DONOR_SERVICE_NAME",
		"FIND_BINDING_PLAN_NAME",
		"FIND_BINDING_SERVICE_NAME",
		"RECIPIENT_PLAN_NAME",
		"RECIPIENT_SERVICE_NAME",
		"SERVICE_NAME",
		"SINGLE_NODE_PLAN_NAME",
		"V2_DONOR_SERVICE_NAME",
		"V2_DONOR_PLAN_NAME",
	})

	Config = config.LoadConfig()

	TestSetup = workflowhelpers.NewTestSuiteSetup(Config)
	TestSetup.Setup()

	migrationTimeout = "25m"
	if os.Getenv("MIGRATION_TIMEOUT") != "" {
		migrationTimeout = os.Getenv("MIGRATION_TIMEOUT")
	}

	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())
})

var _ = AfterSuite(func() {
	if TestSetup != nil {
		TestSetup.Teardown()
	}
})
