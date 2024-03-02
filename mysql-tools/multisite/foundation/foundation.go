package foundation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	Name      string
	CfHomeDir string
	CF        func(cfHomeDir string, args ...string) (string, error)
}

func New(name, cfHomeDir string) Handler {
	return Handler{
		Name:      name,
		CfHomeDir: cfHomeDir,
		CF:        cf,
	}
}

func (h Handler) ID() string {
	return h.Name
}

func (h Handler) UpdateServiceAndWait(instanceName string, arbitraryParams string, planName *string) error {
	var cfArgs = []string{"update-service", instanceName, "-c", arbitraryParams, "--wait"}

	if planName != nil {
		cfArgs = append(cfArgs, "-p", *planName)
	}
	if _, err := h.CF(h.CfHomeDir, cfArgs...); err != nil {
		return err
	}

	return nil
}

func (h Handler) CreateHostInfoKey(instanceName string) (key string, err error) {
	keyName := "host-info-" + strconv.FormatInt(time.Now().UTC().Unix(), 10)

	if _, err := h.CF(h.CfHomeDir, "create-service-key", instanceName, keyName, "-c", `{"replication-request": "host-info" }`); err != nil {
		return "", fmt.Errorf("failed to create service key: %s", err)
	}

	key, err = h.CF(h.CfHomeDir, "service-key", instanceName, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve service-key '%s' on instance '%s': %w", keyName, instanceName, err)
	}

	return extractNestedKey(key)
}

func (h Handler) CreateCredentialsKey(instanceName string) (key string, err error) {
	keyName := "credentials-" + strconv.FormatInt(time.Now().UTC().Unix(), 10)

	if _, err := h.CF(h.CfHomeDir, "create-service-key", instanceName, keyName, "-c", `{"replication-request": "credentials" }`); err != nil {
		return "", fmt.Errorf("failed to create service key: %w", err)
	}

	key, err = h.CF(h.CfHomeDir, "service-key", instanceName, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve service-key '%s' on instance '%s': %w", keyName, instanceName, err)
	}

	return extractNestedKey(key)
}

func (h Handler) InstanceExists(instanceName string) error {
	out, err := h.CF(h.CfHomeDir, "service", instanceName)

	if strings.Contains(out, `Service instance '`+instanceName+`' not found`) {
		return fmt.Errorf("instance '%s' does not exist", instanceName)
	}

	if err != nil {
		return fmt.Errorf("error when checking whether instance exists: %w", err)
	}

	return nil
}

func (h Handler) InstancePlanName(instanceName string) (string, error) {
	out, err := h.CF(h.CfHomeDir, "service", instanceName)
	if err != nil {
		return "", fmt.Errorf("error when checking plan name of instance '%s': %w", instanceName, err)
	}

	for _, line := range strings.Split(out, "\n") {
		// Check if the line contains the plan
		if strings.Contains(line, "plan:") {
			// Split the line to get the plan value
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				// Trim any leading/trailing whitespace from the plan value
				plan := strings.TrimSpace(parts[1])
				return plan, nil
			}
		}
	}

	return "", fmt.Errorf("plan not found for service instance '%s'", instanceName)
}

func (h Handler) PlanExists(planName string) (bool, error) {
	out, err := h.CF(h.CfHomeDir, "marketplace", "-e", "p.mysql")
	if err != nil {
		return false, err
	}
	if !strings.Contains(out, planName) {
		return false, fmt.Errorf(`[%s] Plan '%s' does not exist`, h.ID(), planName)
	}
	return true, nil
}

func extractNestedKey(rawServiceKey string) (usefulContents string, err error) {
	parts := strings.SplitN(rawServiceKey, "\n", 3)
	var serviceKey struct {
		Credentials map[string]any `json:"credentials"`
	}

	if err = json.Unmarshal([]byte(parts[len(parts)-1]), &serviceKey); err != nil {
		return "", fmt.Errorf("failed to parse host-info service key: %w", err)
	}

	// This operation should never fail with an error as we just unmarshalled a superset of the value being remarshalled here
	// If the value is garbage, the workflow will fail later on with a helpful error message
	extractedKey, _ := json.Marshal(serviceKey.Credentials)

	return string(extractedKey), nil
}

// cf is a helper function tested via specs/contract_tests/
func cf(cfHome string, args ...string) (string, error) {
	cmd := exec.Command("cf", args...)
	cmd.Env = append(os.Environ(), "CF_HOME="+cfHome)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("cf %s failed: %s\noutput:\n%s", args[0], err, output)
	}

	return strings.TrimSpace(string(output)), err
}
