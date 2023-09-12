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

package commandstarter

import (
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers/poll_cf/internal"
)

type CommandStarter struct{}

func NewCommandStarter() *CommandStarter {
	return &CommandStarter{}
}

func (r *CommandStarter) Start(reporter internal.Reporter, executable string, args ...string) (*gexec.Session, error) {
	cmd := exec.Command(executable, args...)
	reporter.Polling()
	return gexec.Start(cmd, nil, ginkgo.GinkgoWriter)
}
