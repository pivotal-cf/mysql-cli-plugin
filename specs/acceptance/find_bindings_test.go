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

package acceptance

import (
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers"
	"os"
	"os/exec"
)

var _ = Describe("Find bindings Integration tests", func() {
	Context("When there are app bindings and service keys", func() {
		var (
			bindingInstance     string
			bindingInstanceGUID string
			bindingApp          string
			bindingKey          string
		)
		BeforeEach(func() {
			bindingInstance = generator.PrefixedRandomName("MYSQL", "FIND_BINDING")
			bindingApp = generator.PrefixedRandomName("MYSQL", "FIND_BINDING_APP")
			bindingKey = generator.PrefixedRandomName("MYSQL", "FIND_BINDING_KEY")

			test_helpers.CreateService(os.Getenv("FIND_BINDING_SERVICE_NAME"), os.Getenv("FIND_BINDING_PLAN_NAME"), bindingInstance)
			test_helpers.WaitForService(bindingInstance, `[Ss]tatus:\s+create succeeded`)
			test_helpers.PushApp(bindingApp, "../assets/spring-music")
			test_helpers.BindAppToService(bindingApp, bindingInstance)
			test_helpers.CreateServiceKey(bindingInstance, bindingKey)
		})

		AfterEach(func() {
			test_helpers.UnbindAppFromService(bindingApp, bindingInstance)
			test_helpers.DeleteApp(bindingApp)
			test_helpers.DeleteServiceKey(bindingInstance, bindingKey)
			test_helpers.DeleteService(bindingInstance)
			test_helpers.WaitForService(bindingInstance, fmt.Sprintf("Service instance %s not found", bindingInstance))
		})

		It("Prints out the app bindings and the service keys", func() {
			cmd := exec.Command("cf", "mysql-tools", "find-bindings", os.Getenv("FIND_BINDING_SERVICE_NAME"))
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "1m", "1s").Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring(bindingInstanceGUID))
			Expect(session.Out.Contents()).To(ContainSubstring(bindingInstance))
			Expect(session.Out.Contents()).To(ContainSubstring(TestSetup.GetOrganizationName()))
			Expect(session.Out.Contents()).To(ContainSubstring(TestSetup.TestSpace.SpaceName()))
			Expect(session.Out.Contents()).To(ContainSubstring(bindingApp))
			Expect(session.Out.Contents()).To(ContainSubstring(bindingKey))
		})
	})

	Context("When there are no bindings found", func() {
		It("Shows no bindings", func() {
			cmd := exec.Command("cf", "mysql-tools", "find-bindings", os.Getenv("FIND_BINDING_SERVICE_NAME"))
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "1m", "1s").Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("No bindings found."))
		})
	})

	Context("When an invalid service label is given", func() {
		It("Returns an error", func() {
			fakeServiceLabel := "this isn't a real service"

			cmd := exec.Command("cf", "mysql-tools", "find-bindings", fakeServiceLabel)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "1m", "1s").Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring(
				fmt.Sprintf(`failed to lookup service matching label "%s"`, fakeServiceLabel),
			))
		})
	})

})
