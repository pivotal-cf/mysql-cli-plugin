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
	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	BoshPath                 = "/usr/local/bin/bosh"
	cfCommandTimeout         = "4m"
	cfServiceWaitTimeout     = "15m"
	cfServicePollingInterval = "2s"
	curlTimeout              = "10s"
)

func GetBrokerDeploymentName() string {
	brokerDeploymentName := os.Getenv("BROKER_DEPLOYMENT_NAME")

	if brokerDeploymentName != "" {
		return brokerDeploymentName
	}

	return "dedicated-mysql-broker"
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
	Expect(output).To(ContainSubstring("Create in progress"))
}

func CreateInstanceAndWait(args ...string) string {
	instanceName := generator.PrefixedRandomName("MYSQL", "DED")
	CreateService(os.Getenv("SERVICE_NAME"), os.Getenv("PLAN_NAME"), instanceName, args...)
	WaitForService(instanceName, "Status: create succeeded")
	return instanceName
}

func DeleteService(name string) {
	if ResourceExists("service", name) {
		output := ExecuteCfCmd("delete-service", name, "-f")
		Expect(output).To(ContainSubstring("Delete in progress"))
	}
}

func DeleteInstanceAndWait(instanceName string) {
	ExpectWithOffset(1, instanceName).NotTo(BeEmpty())
	DeleteService(instanceName)
	WaitForService(instanceName, fmt.Sprintf("Service instance %s not found", instanceName))
}

func ResourceExists(resourceType string, resourceName string) bool {
	session := cf.Cf(resourceType, resourceName).Wait(cfCommandTimeout)
	output := string(session.Out.Contents())

	return !strings.Contains(output, "not found")
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
	}, cfServiceWaitTimeout, cfServicePollingInterval).Should(ContainSubstring(success))
	fmt.Fprintln(GinkgoWriter)
}

func InstanceUUID(name string) string {
	output := ExecuteCfCmd("service", name, "--guid")
	return strings.TrimSpace(output)
}

func InstanceDeploymentName(instanceUUID string) string {
	return fmt.Sprintf("service-instance_%s", instanceUUID)
}

func GetDeploymentName(instanceName string) string {
	serviceInstanceUUID := InstanceUUID(instanceName)
	return InstanceDeploymentName(serviceInstanceUUID)
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
	createServiceKey(serviceInstanceName, serviceKeyName)

	session := cf.Cf("service-key", serviceInstanceName, serviceKeyName).Wait("10s")
	output := string(session.Out.Contents())
	Expect(output).To(ContainSubstring(fmt.Sprintf("Getting key %s for service instance %s", serviceKeyName, serviceInstanceName)))

	outputLines := strings.SplitN(output, "\n", 3)
	serviceKeyJson := outputLines[len(outputLines)-1]

	var serviceKey ServiceKey

	json.Unmarshal([]byte(serviceKeyJson), &serviceKey)

	return serviceKey
}

func createServiceKey(instanceName, keyName string) {
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

func DeployApp(appName string) {
	PushApp(appName, "../assets/spring-music/")
	StartApp(appName)
	cf.Cf("enable-ssh", appName).Wait(cfCommandTimeout)
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

func BindAppToService(appName string, instance string) {
	output := cf.Cf("bind-service", appName, instance).Wait(cfCommandTimeout).Out.Contents()
	Expect(string(output)).To(ContainSubstring("Binding service %s to app %s", instance, appName))
	Expect(string(output)).ToNot(SatisfyAny(ContainSubstring("FAILED"),
		ContainSubstring("Server error")))

	Eventually(func() string {
		return string(cf.Cf("service", instance).Wait(cfCommandTimeout).Out.Contents())
	}, cfServiceWaitTimeout, curlTimeout).Should(ContainSubstring("Bound apps: %s", appName))
}

func BindAppToServiceWithUsername(appName, instance, username string) {
	usernameArgs := fmt.Sprintf(`{"username":"%s"}`, username)
	output := cf.Cf("bind-service", appName, instance, "-c", usernameArgs).Wait(cfCommandTimeout).Out.Contents()
	Expect(string(output)).To(ContainSubstring("Binding service %s to app %s", instance, appName))
	Expect(string(output)).ToNot(SatisfyAny(ContainSubstring("FAILED"),
		ContainSubstring("Server error")))

	Eventually(func() string {
		return string(cf.Cf("service", instance).Wait(cfCommandTimeout).Out.Contents())
	}, cfServiceWaitTimeout, curlTimeout).Should(ContainSubstring("Bound apps: %s", appName))
}

func UnbindAppFromService(appName string, instance string) {
	output := cf.Cf("unbind-service", appName, instance).Wait(cfCommandTimeout).Out
	ExpectWithOffset(1, output).To(Say("Unbinding app %s from service %s", appName, instance))
	ExpectWithOffset(1, output).ToNot(SatisfyAny(
		Say("FAILED"),
		Say("Server error")))

	EventuallyWithOffset(1, func() *Buffer {
		return cf.Cf("service", instance).Wait(cfCommandTimeout).Out
	}, cfServiceWaitTimeout, curlTimeout).ShouldNot(Say("Bound apps: %s", appName))
}

func CreateAndBindServiceToApp(serviceName, planName, appName, appPath, instanceName string) {
	CreateService(serviceName, planName, instanceName)
	WaitForService(instanceName, "Status: create succeeded")

	PushApp(appName, appPath)
	BindAppToService(appName, instanceName)
	StartApp(appName)
}

func ManageInstanceProcesses(deploymentName, task, instance string) {
	command := exec.Command(
		BoshPath, "-d", deploymentName, task, instance, "-n")
	command.Stdout = GinkgoWriter
	command.Stderr = GinkgoWriter
	err := command.Run()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

func DeleteAppAndService(appName, instance string) {
	DeleteApp(appName)
	AssertAppIsDeleted(appName)
	DeleteInstanceAndWait(instance)
}

func CheckAppInfo(cfg *config.Config, appURI string, instance string) {
	appInfoUri := fmt.Sprintf("https://%s/appinfo", appURI)
	resp, err := httpClient(cfg.GetSkipSSLValidation()).Get(appInfoUri)
	Expect(err).ToNot(HaveOccurred())
	appConfigurationInfo, _ := ioutil.ReadAll(resp.Body)

	Expect(string(appConfigurationInfo)).Should(SatisfyAll(ContainSubstring("mysql"), ContainSubstring(instance)))
}

func ReadData(cfg *config.Config, appURI string, id string) string {
	getUri := fmt.Sprintf("https://%s/albums/%s", appURI, id)

	resp, err := httpClient(cfg.GetSkipSSLValidation()).Get(getUri)
	Expect(err).ToNot(HaveOccurred())
	fetchedData, _ := ioutil.ReadAll(resp.Body)
	var outputAlbum album
	json.Unmarshal([]byte(fetchedData), &outputAlbum)
	return outputAlbum.Title
}

func WriteData(cfg *config.Config, appURI string, value string) string {
	postUri := fmt.Sprintf("https://%s/albums", appURI)
	values := map[string]string{"title": value}
	jsonValue, _ := json.Marshal(values)
	resp, err := httpClient(cfg.GetSkipSSLValidation()).Post(postUri, "application/json", bytes.NewBuffer(jsonValue))

	Expect(err).ToNot(HaveOccurred())
	writtenData, _ := ioutil.ReadAll(resp.Body)
	var inputAlbum album
	json.Unmarshal([]byte(writtenData), &inputAlbum)
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
	TrackCount  string `json:"trackCount"`
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
		"%s:%s@tcp(127.0.0.1:%d)/%s?interpolateParams=true&tls=skip-verify",
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
