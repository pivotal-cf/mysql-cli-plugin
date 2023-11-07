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

package multisite

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MultiSite struct {
	ReplicationConfigHome string
}

func NewMultiSite(replicationHome string) *MultiSite {
	return &MultiSite{ReplicationConfigHome: replicationHome}
}

func (ms *MultiSite) SaveConfig(cfConfig, targetName string) error {
	foundationHome := filepath.Join(ms.ReplicationConfigHome, targetName, ".cf")
	err := os.MkdirAll(foundationHome, 0700)
	if err != nil {
		return err
	}
	configFile := filepath.Join(foundationHome, "config.json")
	err = copyContents(cfConfig, configFile)
	return err
}

func (ms *MultiSite) ListConfigs() ([]string, error) {
	var configs []string
	files, err := os.ReadDir(ms.ReplicationConfigHome)
	if err != nil {
		return nil, err
	}
	for _, content := range files {
		if content.IsDir() {
			configs = append(configs, content.Name())
		}
	}
	return configs, nil
}

func (ms *MultiSite) RemoveConfig(targetName string) error {
	var err error
	files, err := os.ReadDir(ms.ReplicationConfigHome)
	if err != nil {
		return err
	}

	for _, content := range files {
		if content.IsDir() && content.Name() == targetName {
			err = os.RemoveAll(filepath.Join(ms.ReplicationConfigHome, content.Name()))
		}
	}
	return err
}

func (ms *MultiSite) SetupReplication(primaryFoundation, primaryInstance,
	secondaryFoundation, secondaryInstance string) error {

	var err error

	targetF1CFHome := filepath.Join(ms.ReplicationConfigHome, primaryFoundation)
	targetF2CFHome := filepath.Join(ms.ReplicationConfigHome, secondaryFoundation)

	// Validate primaryInstance
	_, err = executeCommand(targetF1CFHome, "cf", "service", primaryInstance)
	if err != nil {
		return fmt.Errorf("instance '%s' validation error: %w",
			primaryInstance, err)
	}

	// Validate secondaryInstance
	_, err = executeCommand(targetF2CFHome, "cf", "service", secondaryInstance)
	if err != nil {
		return fmt.Errorf("instance '%s' validation error: %w",
			secondaryInstance, err)
	}

	// Create secondary's host-info service key
	hostKeyName := "MSHostInfo-" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	_, err = executeCommand(targetF2CFHome, "cf", "create-service-key",
		secondaryInstance, hostKeyName, "-c", `{"replication-request": "host-info"}`)
	if err != nil {
		return fmt.Errorf("instance '%s' host-info key '%s' creation error: %w",
			secondaryInstance, hostKeyName, err)
	}

	// Recover host-info key contents
	keyOutput, err := executeCommand(targetF2CFHome, "cf", "service-key",
		secondaryInstance, hostKeyName)
	if err != nil {
		return fmt.Errorf("instance '%s' host-info key '%s' retrieval error: %w",
			secondaryInstance, hostKeyName, err)
	}
	keyContents, err := extractKeyContents(keyOutput)
	if err != nil {
		return fmt.Errorf("error extracting host-key info from output: %s\n", keyOutput)
	}

	// Update primary with that host-info service key
	_, err = executeCommand(targetF1CFHome, "cf", "update-service", primaryInstance,
		"-c", keyContents, "--wait")
	if err != nil {
		return fmt.Errorf("error updating primary instance %s with host-info key %s: %w\n", primaryInstance, keyContents, err)
	}

	// Create primary's credentials service key
	credKeyName := "MSCredInfo-" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	_, err = executeCommand(targetF1CFHome, "cf", "create-service-key",
		primaryInstance, credKeyName, "-c", `{"replication-request": "credentials"}`)
	if err != nil {
		return fmt.Errorf("instance '%s' credentials key '%s' creation error: %w",
			primaryInstance, hostKeyName, err)
	}

	// Recover credentials key contents
	credOutput, err := executeCommand(targetF1CFHome, "cf", "service-key",
		primaryInstance, credKeyName)
	if err != nil {
		return fmt.Errorf("instance '%s' credential key '%s' retrieval error: %w",
			primaryInstance, credKeyName, err)
	}
	keyContents, err = extractKeyContents(credOutput)
	if err != nil {
		return fmt.Errorf("error extracting credentials key info from output: %s\n", keyOutput)
	}

	// Update secondary with that credentials service key
	_, err = executeCommand(targetF2CFHome, "cf", "update-service", secondaryInstance,
		"-c", keyContents, "--wait")
	if err != nil {
		return fmt.Errorf("error updating secondary instance %s with credentials key %s: %w\n", secondaryInstance, keyContents, err)
	}

	return nil
}

func copyContents(source, destination string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	err = os.WriteFile(destination, content, 0600)
	return nil
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
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed command: %s %s\n", command, strings.Join(args, " "))
		_, _ = fmt.Fprintf(os.Stderr, "Failed command output: %s\n", comOutput)
		return "", err
	}

	return string(comOutput), nil
}

func executeWithPoll(pollMinutes, pollIntervalSecs int, targetOutput string,
	cfHome, command string, args ...string) error {
	var (
		commandOutput string
		targetFound   bool
		err           error
	)

	pollEndTime := time.Now().Add(time.Duration(pollMinutes) * time.Minute)

	for startTime := time.Now(); startTime.Compare(pollEndTime) > 0 && !targetFound && err == nil; startTime = time.Now() {
		// Run command & look for targeted output
		commandOutput, err = executeCommand(cfHome, command, args...)
		if err != nil {
			continue
		}
		for targetFound = strings.Contains(commandOutput, targetOutput); targetFound; {
			continue
		}

		// Sleep for remaining polling interval
		sleepTime := time.Duration(pollIntervalSecs)*time.Second - time.Now().Sub(startTime)*time.Second
		time.Sleep(sleepTime)
	}

	return err
}

func extractKeyContents(keyOutput string) (string, error) {
	outputLines := strings.SplitN(keyOutput, "\n", 3)
	serviceKeyJson := outputLines[len(outputLines)-1]
	var serviceKey struct {
		Credentials map[string]any `json:"credentials"`
	}

	err := json.Unmarshal([]byte(serviceKeyJson), &serviceKey)
	if err != nil {
		return "", err
	}
	keyInterior, err := json.Marshal(serviceKey.Credentials)
	if err != nil {
		return "", err
	}

	return string(keyInterior), nil
}
