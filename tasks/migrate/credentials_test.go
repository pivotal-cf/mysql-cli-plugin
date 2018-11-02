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
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func loadFixture(name string) (string, error) {
	content, err := ioutil.ReadFile(filepath.Join("fixtures", name))
	if err != nil {
		return "", err
	}

	return string(content), nil
}

var _ = Describe("InstanceCredentials", func() {
	var (
		vcapServices string
	)
	BeforeEach(func() {
		var err error
		vcapServices, err = loadFixture("vcap_services.json")
		Expect(err).NotTo(HaveOccurred())
	})

	When("given an instance name and VCAP_SERVICES", func() {
		It("unmarshals credentials", func() {
			sourceCreds, err := InstanceCredentials("source", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(sourceCreds).To(Equal(Credentials{
				Hostname: "q-n3s3y1.q-g662.bosh",
				Name:     "service_instance_db",
				Password: "um9bgh1ni4uajz6t",
				Port:     3306,
				Username: "6750244ed5954287990d7a255dc84d9d",
				TLS: TLSInfo{
					Cert: struct {
						CA string `json:"ca"`
					}{
						CA: "-----BEGIN CERTIFICATE-----\nMIIDcTCCAlmgAwIBAgIUdOO5sOa14a6sjU8/3BdrH5wgY7wwDQYJKoZIhvcNAQEL\nBQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTE4\nMDEyNTE3NDI0M1oXDTE5MDEyNTE3NDI0M1owJjEkMCIGA1UEAxMbZG0tcm9vdC5k\nZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAw2f9mjCtEHnjNUrNPxk9K2GXrBEOd6FT5RQOzQ9hN64OQp2q9sqJ3sQDuxhv\nqj8H5neaKmpz9yYQERUol1j+lIcZz2XSySAIEl9gwj2Ifj7W8RZZ2zLgu2atqXjG\n0/Kx74gwT3DssktXctDmTA9qvRHggvkafUJDsqFixAVtd3vuX+73qfonn79ACnBR\n8w5/wCoh5JW449w7v7Ix1tlPEaN1PK82yUgJdW2jOSQ3FQfgwJGCt45qFQSpNYok\n1CmmZ9m0ZtMNCYThsfInU4kWPNigH6dekmrJQwO4Q84h0EmMMebUeaP6havS7gT3\nEsNpeIQvm+aUdDLXllFCx52npwIDAQABo4GWMIGTMB0GA1UdDgQWBBTSVUBoLjWu\n1axkj373gNzrq1QGuzBhBgNVHSMEWjBYgBTSVUBoLjWu1axkj373gNzrq1QGu6Eq\npCgwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tghR047mw\n5rXhrqyNTz/cF2sfnCBjvDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA\nA4IBAQBgoGK9SOECIEssWcd0bQrJrTJGH6ZXzDLMxalpXoGockpvX0awAFNDJ654\nGezOBAJ7TPmDLdRDZFtITwP6Bjaz0HeLz5bkaFsiDyJxkULRgI2kYI9pADu9Uo74\nk6CgIaupoBHrRXR7aVGrWYeN840IFSZB1TCnrCuPne4UVEzGsTnFfUjyOgs0Mqo6\nqAkD6ZTVUPu0SwBDoY2TWD1UuH4rOIDwWzVV7u3vY6HY7rFtOGhiNHdiav7RjpsY\nXmsHnVSaa+5iOgi04VsOF3JhpxuvbdMmEe+sOfBmjv+NwNR+ngXelyCzvR7w74NQ\ndjPvEZQgEt7w6DpovGp8cwtKIMAx\n-----END CERTIFICATE-----\n",
					},
				},
			}))

			destCreds, err := InstanceCredentials("dest", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(destCreds).To(Equal(Credentials{
				Hostname: "q-n3s3y1.q-g663.bosh",
				Name:     "service_instance_db",
				Password: "egs9n8bi5rzj2ego",
				Port:     3306,
				Username: "aab9b3d47be84c4fa7aa7003ac3907c9",
				TLS: TLSInfo{
					Cert: struct {
						CA string `json:"ca"`
					}{
						CA: "-----BEGIN CERTIFICATE-----\nMIIDcTCCAlmgAwIBAgIUdOO5sOa14a6sjU8/3BdrH5wgY7wwDQYJKoZIhvcNAQEL\nBQAwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tMB4XDTE4\nMDEyNTE3NDI0M1oXDTE5MDEyNTE3NDI0M1owJjEkMCIGA1UEAxMbZG0tcm9vdC5k\nZWRpY2F0ZWQtbXlzcWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAw2f9mjCtEHnjNUrNPxk9K2GXrBEOd6FT5RQOzQ9hN64OQp2q9sqJ3sQDuxhv\nqj8H5neaKmpz9yYQERUol1j+lIcZz2XSySAIEl9gwj2Ifj7W8RZZ2zLgu2atqXjG\n0/Kx74gwT3DssktXctDmTA9qvRHggvkafUJDsqFixAVtd3vuX+73qfonn79ACnBR\n8w5/wCoh5JW449w7v7Ix1tlPEaN1PK82yUgJdW2jOSQ3FQfgwJGCt45qFQSpNYok\n1CmmZ9m0ZtMNCYThsfInU4kWPNigH6dekmrJQwO4Q84h0EmMMebUeaP6havS7gT3\nEsNpeIQvm+aUdDLXllFCx52npwIDAQABo4GWMIGTMB0GA1UdDgQWBBTSVUBoLjWu\n1axkj373gNzrq1QGuzBhBgNVHSMEWjBYgBTSVUBoLjWu1axkj373gNzrq1QGu6Eq\npCgwJjEkMCIGA1UEAxMbZG0tcm9vdC5kZWRpY2F0ZWQtbXlzcWwuY29tghR047mw\n5rXhrqyNTz/cF2sfnCBjvDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA\nA4IBAQBgoGK9SOECIEssWcd0bQrJrTJGH6ZXzDLMxalpXoGockpvX0awAFNDJ654\nGezOBAJ7TPmDLdRDZFtITwP6Bjaz0HeLz5bkaFsiDyJxkULRgI2kYI9pADu9Uo74\nk6CgIaupoBHrRXR7aVGrWYeN840IFSZB1TCnrCuPne4UVEzGsTnFfUjyOgs0Mqo6\nqAkD6ZTVUPu0SwBDoY2TWD1UuH4rOIDwWzVV7u3vY6HY7rFtOGhiNHdiav7RjpsY\nXmsHnVSaa+5iOgi04VsOF3JhpxuvbdMmEe+sOfBmjv+NwNR+ngXelyCzvR7w74NQ\ndjPvEZQgEt7w6DpovGp8cwtKIMAx\n-----END CERTIFICATE-----\n",
					},
				},
			}))
		})
	})

	When("TLS is enabled in the binding", func() {
		Specify("The connection string enables the default tls profile in the mysql connector", func() {
			creds, err := InstanceCredentials("dest", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.DSN()).To(HaveSuffix(`?tls=default`))
		})

		Specify("HasTLS() returns true", func() {
			creds, err := InstanceCredentials("dest", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.HasTLS()).To(BeTrue())
		})
	})

	When("TLS is not enabled in the binding", func() {
		Specify("Credentials.HasTLS() returns false", func() {
			creds, err := InstanceCredentials("nonTls", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.HasTLS()).To(BeFalse())
		})

		Specify("The connection string disables tls for the mysql connector", func() {
			creds, err := InstanceCredentials("nonTls", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.DSN()).To(HaveSuffix(`?tls=false`))
		})

		Specify("HasTLS() returns false", func() {
			creds, err := InstanceCredentials("nonTls", vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.HasTLS()).To(BeFalse())
		})
	})

	When("invalid VCAP_SERVICES json is provided", func() {
		BeforeEach(func() {
			vcapServices = `{ invalid_json `
		})
		It("returns an error", func() {
			_, err := InstanceCredentials("some-instance-name", vcapServices)
			Expect(err).To(MatchError(`failed to parse VCAP_SERVICES json when looking up credentials for instance_name=some-instance-name: invalid character 'i' looking for beginning of object key string`))

		})
	})

	When("given an unknown instance name", func() {
		It("returns an error", func() {
			_, err := InstanceCredentials("instance-does-not-exist", vcapServices)
			Expect(err).To(MatchError(`instance_name 'instance-does-not-exist' not found in VCAP_SERVICES`))
		})
	})
})
