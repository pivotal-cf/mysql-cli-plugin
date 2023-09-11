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

package commandreporter

import (
	"fmt"
	"io"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
)

const timeFormat = "2006-01-02 15:04:05.00 (MST)"

type CommandReporter struct {
	Writer io.Writer
}

func NewCommandReporter(writers ...io.Writer) *CommandReporter {
	var writer io.Writer
	switch len(writers) {
	case 0:
		writer = ginkgo.GinkgoWriter
	case 1:
		writer = writers[0]
	default:
		panic("NewCommandReporter should only take one writer")
	}

	return &CommandReporter{
		Writer: writer,
	}
}

func (r *CommandReporter) Report(startTime time.Time, command string) {
	startColor := ""
	endColor := ""
	if !config.DefaultReporterConfig.NoColor {
		startColor = "\x1b[32m"
		endColor = "\x1b[0m"
	}

	fmt.Fprintf(
		r.Writer,
		"\n%s[%s]> %s %s",
		startColor,
		startTime.UTC().Format(timeFormat),
		command,
		endColor,
	)
}

func (r *CommandReporter) Polling() {
	startColor := ""
	endColor := ""
	if !config.DefaultReporterConfig.NoColor {
		startColor = "\x1b[32m"
		endColor = "\x1b[0m"
	}

	fmt.Fprintf(
		r.Writer,
		"%s.%s",
		startColor,
		endColor,
	)
}
