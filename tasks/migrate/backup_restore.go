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
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/pivotal-cf/mysql-cli-plugin/tasks/migrate/discovery"
	"github.com/pkg/errors"
)

func baseCmd(cmdName string, credentials Credentials) *exec.Cmd {
	args := []string{
		"--user=" + credentials.Username,
		"--host=" + credentials.Hostname,
		fmt.Sprintf("--port=%d", credentials.Port),
	}

	if credentials.HasTLS() && !credentials.SkipTLSValidation {
		tlsArgs := []string{
			"--ssl-mode=VERIFY_IDENTITY",
			"--ssl-capath=/etc/ssl/certs",
		}
		args = append(args, tlsArgs...)
	}

	cmd := exec.Command(cmdName, args...)
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+credentials.Password)

	return cmd
}

func MySQLDumpCmd(credentials Credentials, invalidViews []discovery.View, schemas ...string) *exec.Cmd {
	cmd := baseCmd("mysqldump", credentials)

	cmd.Args = append(cmd.Args,
		"--max-allowed-packet=1G",
		"--single-transaction",
		"--skip-routines",
		"--skip-events",
		"--set-gtid-purged=off",
		"--skip-triggers",
		"--databases",
	)

	for _, view := range invalidViews {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--ignore-table=%s", view))
	}

	cmd.Args = append(cmd.Args, schemas...)

	return cmd
}

func MySQLCmd(credentials Credentials) *exec.Cmd {
	cmd := baseCmd("mysql", credentials)

	cmd.Args = append(cmd.Args, credentials.Name)
	cmd.Stdout = os.Stdout

	return cmd
}

func ValidateHost(credentials Credentials, timeout time.Duration) ([]string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return nil, errors.New("timeout")
		case <-ticker.C:
			addrs, err := net.LookupHost(credentials.Hostname)
			if err == nil {
				return addrs, nil
			}

			if e, ok := err.(*net.DNSError); !ok || !e.Temporary() {
				return nil, err
			}
		}
	}
}

func ReplaceDefinerCmd() *exec.Cmd {
	args := []string{
		"-e",
		"s/DEFINER=.* SQL SECURITY .*/SQL SECURITY INVOKER/",
	}

	cmd := exec.Command("sed", args...)
	cmd.Stderr = os.Stderr
	return cmd
}

func CopyData(mysqldump, replaceDefinerCmd, mysql *exec.Cmd) error {
	dumpOut, err := mysqldump.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "couldn't pipe the output of mysqldump")
	}

	replaceDefinerCmd.Stdin = dumpOut

	replaceOut, err := replaceDefinerCmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "couldn't pipe the output of sed")
	}

	mysql.Stdin = replaceOut

	if err := mysqldump.Start(); err != nil {
		return errors.Wrap(err, "couldn't start mysqldump")
	}

	if err := replaceDefinerCmd.Start(); err != nil {
		return errors.Wrap(err, "couldn't start sed")
	}

	if err := mysql.Start(); err != nil {
		return errors.Wrap(err, "couldn't start mysql")
	}

	if err := mysql.Wait(); err != nil {
		return errors.Wrap(err, "mysql command failed")
	}

	if err := replaceDefinerCmd.Wait(); err != nil {
		return errors.Wrap(err, "sed command failed")
	}

	if err := mysqldump.Wait(); err != nil {
		return errors.Wrap(err, "mysqldump command failed")
	}

	return nil
}
