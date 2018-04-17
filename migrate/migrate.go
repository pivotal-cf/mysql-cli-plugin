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

package migrate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . client
type client interface {
	BindService(appName, serviceName string) error
	DeleteApp(appName string) error
	DumpLogs(appName string)
	PushApp(path, appName string) error
	RenameService(oldName, newName string) error
	RunTask(appName, command string) error
	StartApp(appName string) error
}

type unpacker interface {
	Unpack(destDir string) error
}

func NewMigrator(client client, unpacker unpacker, donorInstanceName, recipientInstanceName string) *Migrator {
	return &Migrator{
		client:                client,
		donorInstanceName:     donorInstanceName,
		recipientInstanceName: recipientInstanceName,
		unpacker:              unpacker,
	}
}

type Migrator struct {
	AppName               string
	client                client
	donorInstanceName     string
	recipientInstanceName string
	unpacker              unpacker
}

func (m *Migrator) CreateServiceInstance(planType string) error {
	return errors.New("not implemented") // todo
}

func (m *Migrator) MigrateData() error {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "migrate_app_")
	if err != nil {
		return errors.Errorf("Error creating temp directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	log.Printf("Unpacking assets for migration to %s", tmpDir)
	if err = m.unpacker.Unpack(tmpDir); err != nil {
		return errors.Errorf("Error extracting migrate assets: %s", err)
	}

	log.Print("Started to push app")
	m.AppName = "migrate-app-" + uuid.New()
	if err = m.client.PushApp(tmpDir, m.AppName); err != nil {
		return errors.Errorf("failed to push application: %s", err)
	}
	defer func() {
		m.client.DeleteApp(m.AppName)
		log.Print("Cleaning up...")
	}()
	log.Print("Successfully pushed app")

	if err = m.client.BindService(m.AppName, m.donorInstanceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", m.AppName, m.donorInstanceName, err)
	}
	log.Print("Successfully bound app to v1 instance")

	if err = m.client.BindService(m.AppName, m.recipientInstanceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", m.AppName, m.recipientInstanceName, err)
	}
	log.Print("Successfully bound app to v2 instance")

	log.Print("Starting migration app")
	if err = m.client.StartApp(m.AppName); err != nil {
		return errors.Errorf("failed to start application %q: %s", m.AppName, err)
	}

	log.Print("Started to run migration task")
	command := fmt.Sprintf("./migrate %s %s", m.donorInstanceName, m.recipientInstanceName)
	if err = m.client.RunTask(m.AppName, command); err != nil {
		log.Printf("Migration failed: %s", err)
		log.Print("Fetching log output...")
		time.Sleep(5 * time.Second)
		m.client.DumpLogs(m.AppName)
		return err
	}

	log.Print("Migration completed successfully")

	return nil
}

func (m *Migrator) RenameServiceInstances() error {
	newDonorInstanceName := m.donorInstanceName + "-old"
	if err := m.client.RenameService(m.donorInstanceName, newDonorInstanceName); err != nil {
		return fmt.Errorf("Error renaming service instance %s: %s", m.donorInstanceName, err)
	}

	if err := m.client.RenameService(m.recipientInstanceName, m.donorInstanceName); err != nil {
		return fmt.Errorf("Error renaming service instance %s: %s", m.recipientInstanceName, err)
	}

	return nil
}
