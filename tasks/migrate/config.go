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
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Credentials struct {
	Hostname          string
	Name              string
	Password          string
	Port              int
	Username          string
	CA                string
	SkipTLSValidation bool
}

func (d Credentials) HasTLS() bool {
	return d.CA != ""
}

func (d Credentials) DSN() string {
	var tlsConfig = "false"

	if d.HasTLS() && d.SkipTLSValidation {
		tlsConfig = "skip-verify"
	} else if d.HasTLS() && !d.SkipTLSValidation {
		tlsConfig = "true"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?tls=%s",
		d.Username,
		d.Password,
		d.Hostname,
		d.Port,
		d.Name,
		tlsConfig,
	)
}

func InstanceCredentials(instanceName, vcapCredentials string) (Credentials, error) {
	var vcapServices map[string][]struct {
		InstanceName string `json:"instance_name"`
		Credentials  struct {
			Hostname string `json:"hostname"`
			Name     string `json:"name"`
			Password string `json:"password"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			TLS      struct {
				Cert struct {
					CA string `json:"ca"`
				} `json:"cert"`
			} `json:"tls"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal([]byte(vcapCredentials), &vcapServices); err != nil {
		return Credentials{}, errors.Wrapf(
			err,
			`failed to parse VCAP_SERVICES json when looking up credentials for instance_name=%s`,
			instanceName,
		)
	}

	for _, svc := range vcapServices {
		for _, binding := range svc {
			if binding.InstanceName == instanceName {
				return Credentials{
					Hostname: binding.Credentials.Hostname,
					Name:     binding.Credentials.Name,
					Password: binding.Credentials.Password,
					Port:     binding.Credentials.Port,
					Username: binding.Credentials.Username,
					CA:       binding.Credentials.TLS.Cert.CA,
				}, nil
			}
		}
	}

	return Credentials{}, errors.Errorf("instance_name '%s' not found in VCAP_SERVICES", instanceName)
}
