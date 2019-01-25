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
	"bytes"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-binmock"
	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
)

var _ = Describe("MySQLDumpCmd", func() {
	var (
		credentials Credentials
		schemas     []string
	)

	BeforeEach(func() {
		schemas = nil
		credentials = Credentials{
			Username: "some-user-name",
			Password: "some-password",
			Hostname: "some-hostname",
			Port:     3307,
		}
	})

	It("configures stdio", func() {
		mysqldump := MySQLDumpCmd(credentials, nil, schemas...)
		By("directing stderr to os.Stderr", func() {
			Expect(mysqldump.Stderr).To(Equal(os.Stderr))
		})

		By("not configuring stdout", func() {
			Expect(mysqldump.Stdout).To(BeNil())
		})
	})

	When("dumping multiple schemas", func() {
		BeforeEach(func() {
			schemas = []string{"foo", "bar", "baz"}
		})

		It("adds the mysqldump --databases option", func() {
			mysqldump := MySQLDumpCmd(credentials, nil, schemas...)
			Expect(mysqldump).ToNot(BeNil())
			Expect(mysqldump.Args).To(Equal([]string{
				"mysqldump",
				"--user=some-user-name",
				"--host=some-hostname",
				"--port=3307",
				"--max-allowed-packet=1G",
				"--single-transaction",
				"--skip-routines",
				"--skip-events",
				"--set-gtid-purged=off",
				"--skip-triggers",
				"--databases",
				"foo",
				"bar",
				"baz",
			}))
			Expect(mysqldump.Env).To(ContainElement("MYSQL_PWD=some-password"))
		})

		When("TLS is enabled", func() {
			BeforeEach(func() {
				credentials.CA = "-----BEGIN CERTIFICATE-----\nMIIDcTCCAlmgAwIBAgIUdOO5sOa14a6sjU8/3BdrH5wgY7wwDQYJKoZIhvcNAQEL\nBQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTE4\nMDEyNTE3NDI0M1oXDTE5MDEyNTE3NDI0M1owJjEkMCIGA1UEAxMbZG0tcm9vdC5k\nZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAw2f9mjCtEHnjNUrNPxk9K2GXrBEOd6FT5RQOzQ9hN64OQp2q9sqJ3sQDuxhv\nqj8H5neaKmpz9yYQERUol1j+lIcZz2XSySAIEl9gwj2Ifj7W8RZZ2zLgu2atqXjG\n0/Kx74gwT3DssktXctDmTA9qvRHggvkafUJDsqFixAVtd3vuX+73qfonn79ACnBR\n8w5/wCoh5JW449w7v7Ix1tlPEaN1PK82yUgJdW2jOSQ3FQfgwJGCt45qFQSpNYok\n1CmmZ9m0ZtMNCYThsfInU4kWPNigH6dekmrJQwO4Q84h0EmMMebUeaP6havS7gT3\nEsNpeIQvm+aUdDLXllFCx52npwIDAQABo4GWMIGTMB0GA1UdDgQWBBTSVUBoLjWu\n1axkj373gNzrq1QGuzBhBgNVHSMEWjBYgBTSVUBoLjWu1axkj373gNzrq1QGu6Eq\npCgwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tghR047mw\n5rXhrqyNTz/cF2sfnCBjvDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA\nA4IBAQBgoGK9SOECIEssWcd0bQrJrTJGH6ZXzDLMxalpXoGockpvX0awAFNDJ654\nGezOBAJ7TPmDLdRDZFtITwP6Bjaz0HeLz5bkaFsiDyJxkULRgI2kYI9pADu9Uo74\nk6CgIaupoBHrRXR7aVGrWYeN840IFSZB1TCnrCuPne4UVEzGsTnFfUjyOgs0Mqo6\nqAkD6ZTVUPu0SwBDoY2TWD1UuH4rOIDwWzVV7u3vY6HY7rFtOGhiNHdiav7RjpsY\nXmsHnVSaa+5iOgi04VsOF3JhpxuvbdMmEe+sOfBmjv+NwNR+ngXelyCzvR7w74NQ\ndjPvEZQgEt7w6DpovGp8cwtKIMAx\n-----END CERTIFICATE-----\n"
			})

			It("specifies the TLS options in the mysqldump command", func() {
				mysqldump := MySQLDumpCmd(credentials, nil, schemas...)
				Expect(mysqldump).ToNot(BeNil())
				Expect(mysqldump.Args).To(Equal([]string{
					"mysqldump",
					"--user=some-user-name",
					"--host=some-hostname",
					"--port=3307",
					"--ssl-mode=VERIFY_IDENTITY",
					"--ssl-capath=/etc/ssl/certs",
					"--max-allowed-packet=1G",
					"--single-transaction",
					"--skip-routines",
					"--skip-events",
					"--set-gtid-purged=off",
					"--skip-triggers",
					"--databases",
					"foo",
					"bar",
					"baz",
				}))
				Expect(mysqldump.Env).To(ContainElement("MYSQL_PWD=some-password"))
			})
		})

		When("views to ignore are specified", func() {
			var (
				invalidViews []discovery.View
			)

			BeforeEach(func() {
				invalidViews = []discovery.View{
					{Schema: "foo", TableName: "view1"},
					{Schema: "bar", TableName: "view1"},
					{Schema: "baz", TableName: "view1"},
				}
			})

			It("add the mysqldump --ignore-table option the correct number of times", func() {
				mysqldump := MySQLDumpCmd(credentials, invalidViews, schemas...)
				Expect(mysqldump).ToNot(BeNil())
				Expect(mysqldump.Args).To(Equal([]string{
					"mysqldump",
					"--user=some-user-name",
					"--host=some-hostname",
					"--port=3307",
					"--max-allowed-packet=1G",
					"--single-transaction",
					"--skip-routines",
					"--skip-events",
					"--set-gtid-purged=off",
					"--skip-triggers",
					"--ignore-table=foo.view1",
					"--ignore-table=bar.view1",
					"--ignore-table=baz.view1",
					"--databases",
					"foo",
					"bar",
					"baz",
				}))
				Expect(mysqldump.Env).To(ContainElement("MYSQL_PWD=some-password"))
			})
		})
	})

	When("dumping a single schema", func() {
		BeforeEach(func() {
			schemas = []string{"one-database"}
		})

		It("dumps only a single database without the --databases option", func() {
			mysqldump := MySQLDumpCmd(credentials, nil, schemas...)
			Expect(mysqldump).ToNot(BeNil())
			Expect(mysqldump.Args).To(Equal([]string{
				"mysqldump",
				"--user=some-user-name",
				"--host=some-hostname",
				"--port=3307",
				"--max-allowed-packet=1G",
				"--single-transaction",
				"--skip-routines",
				"--skip-events",
				"--set-gtid-purged=off",
				"--skip-triggers",
				"one-database",
			}))
			Expect(mysqldump.Env).To(ContainElement("MYSQL_PWD=some-password"))
		})

		When("TLS is enabled", func() {
			BeforeEach(func() {
				credentials.CA = "-----BEGIN CERTIFICATE-----\nMIIDcTCCAlmgAwIBAgIUdOO5sOa14a6sjU8/3BdrH5wgY7wwDQYJKoZIhvcNAQEL\nBQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTE4\nMDEyNTE3NDI0M1oXDTE5MDEyNTE3NDI0M1owJjEkMCIGA1UEAxMbZG0tcm9vdC5k\nZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAw2f9mjCtEHnjNUrNPxk9K2GXrBEOd6FT5RQOzQ9hN64OQp2q9sqJ3sQDuxhv\nqj8H5neaKmpz9yYQERUol1j+lIcZz2XSySAIEl9gwj2Ifj7W8RZZ2zLgu2atqXjG\n0/Kx74gwT3DssktXctDmTA9qvRHggvkafUJDsqFixAVtd3vuX+73qfonn79ACnBR\n8w5/wCoh5JW449w7v7Ix1tlPEaN1PK82yUgJdW2jOSQ3FQfgwJGCt45qFQSpNYok\n1CmmZ9m0ZtMNCYThsfInU4kWPNigH6dekmrJQwO4Q84h0EmMMebUeaP6havS7gT3\nEsNpeIQvm+aUdDLXllFCx52npwIDAQABo4GWMIGTMB0GA1UdDgQWBBTSVUBoLjWu\n1axkj373gNzrq1QGuzBhBgNVHSMEWjBYgBTSVUBoLjWu1axkj373gNzrq1QGu6Eq\npCgwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tghR047mw\n5rXhrqyNTz/cF2sfnCBjvDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA\nA4IBAQBgoGK9SOECIEssWcd0bQrJrTJGH6ZXzDLMxalpXoGockpvX0awAFNDJ654\nGezOBAJ7TPmDLdRDZFtITwP6Bjaz0HeLz5bkaFsiDyJxkULRgI2kYI9pADu9Uo74\nk6CgIaupoBHrRXR7aVGrWYeN840IFSZB1TCnrCuPne4UVEzGsTnFfUjyOgs0Mqo6\nqAkD6ZTVUPu0SwBDoY2TWD1UuH4rOIDwWzVV7u3vY6HY7rFtOGhiNHdiav7RjpsY\nXmsHnVSaa+5iOgi04VsOF3JhpxuvbdMmEe+sOfBmjv+NwNR+ngXelyCzvR7w74NQ\ndjPvEZQgEt7w6DpovGp8cwtKIMAx\n-----END CERTIFICATE-----\n"
			})

			It("specifies the TLS options in the mysqldump command", func() {
				mysqldump := MySQLDumpCmd(credentials, nil, schemas...)
				Expect(mysqldump).ToNot(BeNil())
				Expect(mysqldump.Args).To(Equal([]string{
					"mysqldump",
					"--user=some-user-name",
					"--host=some-hostname",
					"--port=3307",
					"--ssl-mode=VERIFY_IDENTITY",
					"--ssl-capath=/etc/ssl/certs",
					"--max-allowed-packet=1G",
					"--single-transaction",
					"--skip-routines",
					"--skip-events",
					"--set-gtid-purged=off",
					"--skip-triggers",
					"one-database",
				}))
				Expect(mysqldump.Env).To(ContainElement("MYSQL_PWD=some-password"))
			})
		})
	})
})

var _ = Describe("MySQLCmd", func() {
	var (
		credentials Credentials
	)

	BeforeEach(func() {
		credentials = Credentials{
			Username: "some-user-name",
			Name:     "some-db-name",
			Password: "some-password",
			Hostname: "some-hostname",
			Port:     3307,
		}
	})

	It("configures stdio", func() {
		mysql := MySQLCmd(credentials)
		By("directing stderr to os.Stderr", func() {
			Expect(mysql.Stderr).To(Equal(os.Stderr))
		})

		By("directing stdout to os.Stdout", func() {
			Expect(mysql.Stdout).To(Equal(os.Stdout))
		})
	})

	It("builds the mysql command", func() {
		mysql := MySQLCmd(credentials)
		Expect(mysql).ToNot(BeNil())
		Expect(mysql.Args).To(Equal([]string{
			"mysql",
			"--user=some-user-name",
			"--host=some-hostname",
			"--port=3307",
			"some-db-name",
		}))
		Expect(mysql.Env).To(ContainElement("MYSQL_PWD=some-password"))
	})

	When("TLS is enabled", func() {
		BeforeEach(func() {
			credentials.CA = "-----BEGIN CERTIFICATE-----\nMIIDcTCCAlmgAwIBAgIUdOO5sOa14a6sjU8/3BdrH5wgY7wwDQYJKoZIhvcNAQEL\nBQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTE4\nMDEyNTE3NDI0M1oXDTE5MDEyNTE3NDI0M1owJjEkMCIGA1UEAxMbZG0tcm9vdC5k\nZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAw2f9mjCtEHnjNUrNPxk9K2GXrBEOd6FT5RQOzQ9hN64OQp2q9sqJ3sQDuxhv\nqj8H5neaKmpz9yYQERUol1j+lIcZz2XSySAIEl9gwj2Ifj7W8RZZ2zLgu2atqXjG\n0/Kx74gwT3DssktXctDmTA9qvRHggvkafUJDsqFixAVtd3vuX+73qfonn79ACnBR\n8w5/wCoh5JW449w7v7Ix1tlPEaN1PK82yUgJdW2jOSQ3FQfgwJGCt45qFQSpNYok\n1CmmZ9m0ZtMNCYThsfInU4kWPNigH6dekmrJQwO4Q84h0EmMMebUeaP6havS7gT3\nEsNpeIQvm+aUdDLXllFCx52npwIDAQABo4GWMIGTMB0GA1UdDgQWBBTSVUBoLjWu\n1axkj373gNzrq1QGuzBhBgNVHSMEWjBYgBTSVUBoLjWu1axkj373gNzrq1QGu6Eq\npCgwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tghR047mw\n5rXhrqyNTz/cF2sfnCBjvDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA\nA4IBAQBgoGK9SOECIEssWcd0bQrJrTJGH6ZXzDLMxalpXoGockpvX0awAFNDJ654\nGezOBAJ7TPmDLdRDZFtITwP6Bjaz0HeLz5bkaFsiDyJxkULRgI2kYI9pADu9Uo74\nk6CgIaupoBHrRXR7aVGrWYeN840IFSZB1TCnrCuPne4UVEzGsTnFfUjyOgs0Mqo6\nqAkD6ZTVUPu0SwBDoY2TWD1UuH4rOIDwWzVV7u3vY6HY7rFtOGhiNHdiav7RjpsY\nXmsHnVSaa+5iOgi04VsOF3JhpxuvbdMmEe+sOfBmjv+NwNR+ngXelyCzvR7w74NQ\ndjPvEZQgEt7w6DpovGp8cwtKIMAx\n-----END CERTIFICATE-----\n"
		})

		It("specifies the TLS options in the mysql command", func() {
			mysql := MySQLCmd(credentials)
			Expect(mysql).ToNot(BeNil())
			Expect(mysql.Args).To(Equal([]string{
				"mysql",
				"--user=some-user-name",
				"--host=some-hostname",
				"--port=3307",
				"--ssl-mode=VERIFY_IDENTITY",
				"--ssl-capath=/etc/ssl/certs",
				"some-db-name",
			}))
			Expect(mysql.Env).To(ContainElement("MYSQL_PWD=some-password"))
		})
	})
})

var _ = Describe("ReplaceDefinerCmd", func() {
	It("builds the sed command", func() {
		replace := ReplaceDefinerCmd()
		Expect(replace).ToNot(BeNil())
		Expect(replace.Args).To(Equal([]string{
			"sed",
			"-e",
			"s/DEFINER=.* SQL SECURITY .*/SQL SECURITY INVOKER/",
		}))
	})

})

var _ = Describe("CopyData", func() {
	var (
		mySQLDumpMock      *binmock.Mock
		mySQLDumpCmd       *exec.Cmd
		mySQLMock          *binmock.Mock
		mySQLCmd           *exec.Cmd
		replaceDefinerMock *binmock.Mock
		replaceDefinerCmd  *exec.Cmd
	)

	BeforeEach(func() {
		mySQLDumpMock = binmock.NewBinMock(Fail)
		mySQLDumpMock.
			WhenCalled().
			WillPrintToStdOut(`something`).
			WillExitWith(0)
		mySQLDumpCmd = exec.Command(mySQLDumpMock.Path)

		replaceDefinerMock = binmock.NewBinMock(Fail)
		replaceDefinerMock.
			WhenCalled().
			WillPrintToStdOut(`replaced`).
			WillExitWith(0)
		replaceDefinerCmd = exec.Command(replaceDefinerMock.Path)

		mySQLMock = binmock.NewBinMock(Fail)
		mySQLMock.WhenCalled().WillExitWith(0)
		mySQLCmd = exec.Command(mySQLMock.Path)
	})

	It("Pipes the output of the mySQLDumpCmd command into the mySQLCmd command", func() {
		Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).To(Succeed())

		Expect(mySQLDumpMock.Invocations()).To(HaveLen(1))
		Expect(replaceDefinerMock.Invocations()).To(HaveLen(1))
		Expect(mySQLMock.Invocations()).To(HaveLen(1))

		Expect(replaceDefinerMock.Invocations()[0].Stdin()).
			To(ConsistOf(`something`))
		Expect(mySQLMock.Invocations()[0].Stdin()).
			To(ConsistOf(`replaced`))
	})

	When("piping the output of mysqldump fails", func() {
		BeforeEach(func() {
			mySQLDumpMock.Reset()
			mySQLDumpMock.
				WhenCalled().
				WillPrintToStdOut(`something`).
				WillExitWith(1)
			mySQLDumpCmd = exec.Command(mySQLDumpMock.Path)

			mySQLMock.WhenCalled().WillExitWith(0)
			mySQLCmd = exec.Command(mySQLMock.Path)
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError(`mysqldump command failed: exit status 1`))
		})
	})

	When("starting the mysqldump command fails", func() {
		BeforeEach(func() {
			mySQLDumpCmd.Path = "/invalid/path/to/mysqldump"
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError(`couldn't start mysqldump: fork/exec /invalid/path/to/mysqldump: no such file or directory`))
		})
	})

	When("starting the mysql command fails", func() {
		BeforeEach(func() {
			mySQLCmd.Path = "/invalid/path/to/mysql"
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError(`couldn't start mysql: fork/exec /invalid/path/to/mysql: no such file or directory`))
		})
	})

	When("the mysqldump command fails", func() {
		BeforeEach(func() {
			mySQLDumpMock.Reset()
			mySQLDumpMock.WhenCalled().WillExitWith(1)
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError("mysqldump command failed: exit status 1"))
		})
	})

	When("the mysql command fails", func() {
		BeforeEach(func() {
			mySQLMock.Reset()
			mySQLMock.WhenCalled().WillExitWith(1)
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError("mysql command failed: exit status 1"))
		})
	})

	When("creating a pipe for mysqldump stdout fails", func() {
		BeforeEach(func() {
			mySQLDumpCmd.Stdout = &bytes.Buffer{}
		})

		It("returns an error", func() {
			Expect(CopyData(mySQLDumpCmd, replaceDefinerCmd, mySQLCmd)).
				To(MatchError("couldn't pipe the output of mysqldump: exec: Stdout already set"))
		})
	})
})
