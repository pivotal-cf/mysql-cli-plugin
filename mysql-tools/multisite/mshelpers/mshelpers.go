package mshelpers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Note we pass in an extra argument targetsDir, which the existing SetupReplication()
// provides to its validation via ms.ReplicationConfigHome; this will get called as
// ValidateInstance(ms.ReplicationConfigHome, primaryFoundation, primaryInstance)
func ValidateInstance(targetsDir, cfTarget, instanceName string) error {

	_, err := os.Stat(targetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("error: target store %s is not a directory", targetsDir)
		}
	}
	//targetCFHome := filepath.Join(targetsDir, cfTarget)
	//_, err = executeCommand(targetCFHome, "cf", "service", instanceName)

	return err
}

func executeCommand(cfHome, command string, args ...string) (string, error) {
	var cmd *exec.Cmd

	if len(args) == 0 {
		return "", fmt.Errorf("insufficient arguments to cf command")
	}

	cmd = exec.Command(command, args...)

	currentEnv := os.Environ()
	appendedEnv := append(currentEnv,
		"CF_HOME="+cfHome, // Set CF_HOME for this command execution
	)
	cmd.Env = appendedEnv

	comOutput, err := cmd.CombinedOutput()

	// TODO: Consider bundling error diagnostic info within error itself:
	// return "", fmt.Errorf("%w\nFailed command: %s %s\nFailed command output: %s\n", command, strings.Join(args, " "), comOutput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed command: %s %s\n", command, strings.Join(args, " "))
		_, _ = fmt.Fprintf(os.Stderr, "Failed command output: %s\n", comOutput)
		return "", err
	}

	return string(comOutput), nil
}
