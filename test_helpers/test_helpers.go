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

package test_helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	pollcf "github.com/pivotal-cf/mysql-cli-plugin/test_helpers/poll_cf/cf"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	cfCommandTimeout         = "4m"
	cfServiceWaitTimeout     = "15m"
	cfServicePollingInterval = "2s"
	curlTimeout              = "10s"
)

type Entity struct {
	AppGUID string `json:"app_guid"`
}

type Resource struct {
	Entity   Entity   `json:"entity"`
	Metadata Metadata `json:"metadata"`
}

type BindingResult struct {
	Resources []Resource `json:"resources"`
}

type Metadata struct {
	CreatedAt string `json:"created_at"`
	GUID      string `json:"guid"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
}

func CheckForRequiredEnvVars(envs []string) {
	var missingEnvs []string

	for _, v := range envs {
		if os.Getenv(v) == "" {
			missingEnvs = append(missingEnvs, v)
		}
	}

	Expect(missingEnvs).To(BeEmpty(), "Missing environment variables: %s", strings.Join(missingEnvs, ", "))
}

func CreateService(serviceName string, planName string, name string, args ...string) {
	createServiceArgs := []string{
		"create-service",
		serviceName,
		planName,
		name,
	}
	createServiceArgs = append(createServiceArgs, args...)
	output := ExecuteCfCmd(createServiceArgs...)
	Expect(output).To(SatisfyAny(
		ContainSubstring("Create in progress"),
		ContainSubstring("Creating service instance")))
}

func DeleteService(name string) {
	if ! ResourceDeleted("service", name) {
		output := ExecuteCfCmd("delete-service", name, "-f")
		Expect(output).To(SatisfyAny(
			ContainSubstring("Delete in progress"),
			ContainSubstring("Deleting service"),
		))
	}
}

func ResourceDeleted(resourceType string, resourceName string) bool {
	session := cf.Cf(resourceType, resourceName).Wait(cfCommandTimeout)
	output := string(session.Out.Contents())

	return strings.Contains(output, "not found") || strings.Contains(output, "delete in progress")
}

func WaitForService(name string, success string) {
	cf.Cf("service", name).Wait(cfCommandTimeout)
	commandReport := fmt.Sprintf("Polling `cf service %s` for '%s'", name, success)
	pollcf.ReportPoll(commandReport)
	Eventually(func() string {
		session := pollcf.PollCf("service", name).Wait(cfCommandTimeout)
		output := string(session.Out.Contents()) + string(session.Err.Contents())
		Expect(output).ToNot(ContainSubstring("failed"))
		return output
	}, cfServiceWaitTimeout, cfServicePollingInterval).Should(MatchRegexp(success))
	fmt.Fprintln(GinkgoWriter)
}

func InstanceUUID(name string) string {
	return resourceGUID("service", name)
}

func AppUUID(name string) string {
	return resourceGUID("app", name)
}

func resourceGUID(resourceType string, name string) string {
	output := ExecuteCfCmd(resourceType, name, "--guid")
	return strings.TrimSpace(output)
}

type ServiceKey struct {
	Hostname string `json:"hostname"`
	JbdcUrl  string `json:"jbdcUrl"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Uri      string `json:"uri"`
	Username string `json:"username"`
	TLS struct {
		Cert struct {
			CA string
		}
	}
}

func GetServiceKey(serviceInstanceName, serviceKeyName string) ServiceKey {
	CreateServiceKey(serviceInstanceName, serviceKeyName)

	session := cf.Cf("service-key", serviceInstanceName, serviceKeyName).Wait("10s")
	output := string(session.Out.Contents())
	Expect(output).To(ContainSubstring(fmt.Sprintf("Getting key %s for service instance %s", serviceKeyName, serviceInstanceName)))

	outputLines := strings.SplitN(output, "\n", 3)
	serviceKeyJson := outputLines[len(outputLines)-1]

	var serviceKey ServiceKey

	json.Unmarshal([]byte(serviceKeyJson), &serviceKey)

	return serviceKey
}

func CreateServiceKey(instanceName, keyName string) {
	output := ExecuteCfCmd("create-service-key", instanceName, keyName)

	Expect(output).To(ContainSubstring("Creating service key"))
	Expect(output).To(ContainSubstring("OK"))
}

func DeleteServiceKey(instanceName, keyName string) {
	output := ExecuteCfCmd("delete-service-key", instanceName, keyName, "-f")

	Expect(output).To(ContainSubstring("Deleting key"))
	Expect(output).To(ContainSubstring("OK"))
}

func PushApp(appName, appCodePath string) {
	ExecuteCfCmd("push", appName,
		"--no-start",
		"-f", fmt.Sprintf("%s/manifest.yml", appCodePath),
	)
}

func StartApp(appName string) {
	ExecuteCfCmd("start", appName)
}

func DeleteApp(appName string) {
	ExecuteCfCmd("delete", appName, "-f")
}

func AssertAppIsDeleted(appName string) {
	success := fmt.Sprintf("App %s not found", appName)
	commandReport := fmt.Sprintf("Polling `cf app %s` for '%s'", appName, success)
	pollcf.ReportPoll(commandReport)
	EventuallyWithOffset(1, func() string {
		return string(pollcf.PollCf("app", appName).Wait(cfCommandTimeout).Err.Contents())
	}, cfServiceWaitTimeout, curlTimeout).Should(ContainSubstring(success))
}

func BoundAppGUIDs(instanceGUID string) []string {
	bindingResponse := cf.Cf("curl", fmt.Sprintf("/v2/service_instances/%s/service_bindings", instanceGUID)).Wait(cfCommandTimeout).Out.Contents()
	var binding BindingResult
	err := json.Unmarshal(bindingResponse, &binding)
	Expect(err).NotTo(HaveOccurred())

	appGUIDs := make([]string, 0)
	for _, resource := range binding.Resources {
		appGUIDs = append(appGUIDs, resource.Entity.AppGUID)
	}

	return appGUIDs
}

func UnbindAllAppsFromService(instanceGUID string) {
	var binding BindingResult
	var session *gexec.Session
	bindingResponse := cf.Cf("curl", fmt.Sprintf("/v2/service_instances/%s/service_bindings", instanceGUID)).Wait(cfCommandTimeout).Out.Contents()
	err := json.Unmarshal(bindingResponse, &binding)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	for _, resource := range binding.Resources {
		session = cf.Cf("curl", "-X", "DELETE", resource.Metadata.URL).Wait(cfCommandTimeout)
		ExpectWithOffset(1, session.ExitCode()).To(Equal(0))
	}
}

func BindAppToService(appName string, instance string) {
	output := cf.Cf("bind-service", appName, instance).Wait(cfCommandTimeout).Out.Contents()
	Expect(string(output)).To(ContainSubstring("Binding service %s to app %s", instance, appName))
	Expect(string(output)).ToNot(SatisfyAny(ContainSubstring("FAILED"),
		ContainSubstring("Server error")))

	instanceGUID := InstanceUUID(instance)
	appGUID := AppUUID(appName)

	Eventually(BoundAppGUIDs(instanceGUID), cfServiceWaitTimeout, curlTimeout).Should(ContainElement(appGUID))
}

func UnbindAppFromService(appName string, instance string) {
	output := cf.Cf("unbind-service", appName, instance).Wait(cfCommandTimeout).Out
	ExpectWithOffset(1, output).To(Say("Unbinding app %s from service %s", appName, instance))
	ExpectWithOffset(1, output).ToNot(SatisfyAny(
		Say("FAILED"),
		Say("Server error")))

	instanceGUID := InstanceUUID(instance)
	appGUID := AppUUID(appName)

	Eventually(BoundAppGUIDs(instanceGUID), cfServiceWaitTimeout, curlTimeout).ShouldNot(ContainElement(appGUID))
}

func CheckAppInfo(skipSSLValidation bool, appURI string, instance string) {
	appInfoUri := fmt.Sprintf("https://%s/appinfo", appURI)
	resp, err := httpClient(skipSSLValidation).Get(appInfoUri)
	Expect(err).ToNot(HaveOccurred())
	appConfigurationInfo, _ := ioutil.ReadAll(resp.Body)

	Expect(string(appConfigurationInfo)).Should(SatisfyAll(ContainSubstring("mysql"), ContainSubstring(instance)))
}

func ReadData(skipSSLValidation bool, appURI string, id string) string {
	getUri := fmt.Sprintf("https://%s/albums/%s", appURI, id)

	resp, err := httpClient(skipSSLValidation).Get(getUri)
	Expect(err).ToNot(HaveOccurred())

	fetchedData, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK), string(fetchedData))

	var outputAlbum album
	json.Unmarshal([]byte(fetchedData), &outputAlbum)

	return outputAlbum.Title
}

func WriteData(skipSSLValidation bool, appURI string, value string) string {
	postUri := fmt.Sprintf("https://%s/albums", appURI)

	values := map[string]string{"title": value}
	jsonValue, err := json.Marshal(values)
	Expect(err).NotTo(HaveOccurred())

	resp, err := httpClient(skipSSLValidation).Post(postUri, "application/json", bytes.NewBuffer(jsonValue))
	Expect(err).ToNot(HaveOccurred())

	writtenData, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())

	Expect(resp.StatusCode).To(Equal(http.StatusOK), string(writtenData))

	var inputAlbum album
	Expect(json.Unmarshal([]byte(writtenData), &inputAlbum)).To(Succeed())
	Expect(inputAlbum.Title).Should(Equal(value))

	return inputAlbum.Id
}

func httpClient(skipSsl bool) *http.Client {
	if skipSsl {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
		return &http.Client{Transport: transCfg}
	} else {
		return &http.Client{}
	}

}

func ExecuteCfCmd(args ...string) string {
	session := cf.Cf(args...).Wait(cfCommandTimeout)
	EventuallyWithOffset(1, session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}

type album struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	ReleaseYear int    `json:"releaseYear"`
	Genre       string `json:"genre"`
	TrackCount  int    `json:"trackCount"`
	AlbumId     string `json:"albumId"`
}

func OpenDatabaseTunnelToApp(port int, appName string, serviceKey ServiceKey) context.CancelFunc {

	tunnelContext, tunnelCancel := context.WithCancel(context.Background())
	connectionString := fmt.Sprintf("%d:%s:3306", port, serviceKey.Hostname)
	tunnelCommand := exec.CommandContext(tunnelContext, "cf", "ssh", "--skip-remote-execution", "-L", connectionString, appName)
	err := tunnelCommand.Start()
	Expect(err).ToNot(HaveOccurred())

	waitForTunnel(port, serviceKey)

	return tunnelCancel
}

func waitForTunnel(port int, serviceKey ServiceKey) {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(127.0.0.1:%d)/%s?interpolateParams=true",
		serviceKey.Username,
		serviceKey.Password,
		port,
		serviceKey.Name,
	)

	Eventually(func() error {
		db, err := sql.Open("mysql", connectionString)
		if err != nil {
			return err
		}
		defer func() {
			e := db.Close()
			Expect(e).NotTo(HaveOccurred())
		}()

		_, err = db.Exec("SELECT 1")
		return err
	}, "1m", "5s").Should(Not(HaveOccurred()))
}
