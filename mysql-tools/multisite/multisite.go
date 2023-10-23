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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	targetF1CFHome := filepath.Join(ms.ReplicationConfigHome, primaryFoundation)
	targetF2CFHome := filepath.Join(ms.ReplicationConfigHome, secondaryFoundation)
	executeCommand(targetF1CFHome, "cf", "target")
	executeCommand(targetF1CFHome, "cf", "service", primaryInstance)
	executeCommand(targetF2CFHome, "cf", "target")
	executeCommand(targetF2CFHome, "cf", "service", secondaryInstance)
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

func executeCommand(cfHome, command string, args ...string) error {
	var cmd *exec.Cmd
	currentEnv := os.Environ()
	appendedEnv := append(currentEnv,
		"CF_HOME="+cfHome, // Set the global variable here
	)

	if len(args) > 0 {
		cmd = exec.Command(command, args...)
	} else {
		cmd = exec.Command(command)
	}

	cmd.Env = appendedEnv

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return fmt.Errorf("error executing command %s %v: %w", command, args, err)
	}

	fmt.Println("Result: " + out.String())
	return nil
}
