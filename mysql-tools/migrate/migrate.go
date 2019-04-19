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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . Client
type Client interface {
	ServiceExists(serviceName string) bool
	CreateServiceInstance(planType, instanceName string) error
	GetHostnames(instanceName string) ([]string, error)
	UpdateServiceConfig(instanceName string, jsonParams string) error
	BindService(appName, serviceName string) error
	DeleteApp(appName string) error
	DeleteServiceInstance(instanceName string) error
	GetLogs(appName, filter string) ([]string, error)
	PushApp(path, appName string) error
	RenameService(oldName, newName string) error
	RunTask(appName, command string) error
	StartApp(appName string) error
}

//go:generate counterfeiter . Unpacker
type Unpacker interface {
	Unpack(destDir string) error
}

func NewMigrator(client Client, unpacker Unpacker) *Migrator {
	return &Migrator{
		client:   client,
		unpacker: unpacker,
	}
}

type Migrator struct {
	appName  string
	client   Client
	unpacker Unpacker
}

type MigrateOptions struct {
	DonorInstanceName     string
	RecipientInstanceName string
	Cleanup               bool
	SkipTLSValidation     bool
}

func (m *Migrator) CheckServiceExists(donorInstanceName string) error {
	if !m.client.ServiceExists(donorInstanceName) {
		return fmt.Errorf("Service instance %s not found", donorInstanceName)
	}

	return nil
}

func (m *Migrator) ConfigureServiceInstance(serviceName string) error {
	return errors.New("unimplemented'")
}

func (m *Migrator) CreateAndConfigureServiceInstance(planType, serviceName string) error {
	if err := m.client.CreateServiceInstance(planType, serviceName); err != nil {
		return errors.Wrap(err, "Error creating service instance")
	}

	hostnames, err := m.client.GetHostnames(serviceName)
	if err != nil {
		return errors.Wrap(err, "Error obtaining hostname for new service instance")
	}

	jsonEncodedHostnames, err := json.Marshal(hostnames)
	if err != nil {
		return errors.Wrapf(err, "Error JSON encoding hostnames: %s", hostnames)
	}

	_ = m.client.UpdateServiceConfig(serviceName,
		fmt.Sprintf(`{"enable_tls": %s}`, jsonEncodedHostnames))

	return nil
}

func (m *Migrator) MigrateData(opts MigrateOptions) error {
	cleanup := opts.Cleanup
	donorInstanceName := opts.DonorInstanceName
	recipientInstanceName := opts.RecipientInstanceName

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
	m.appName = "migrate-app-" + uuid.New()
	if err = m.client.PushApp(tmpDir, m.appName); err != nil {
		return errors.Errorf("failed to push application: %s", err)
	}
	if cleanup {
		defer func() {
			m.client.DeleteApp(m.appName)
			log.Print("Cleaning up...")
		}()
	}
	log.Print("Successfully pushed app")

	if err = m.client.BindService(m.appName, donorInstanceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", m.appName, donorInstanceName, err)
	}
	log.Print("Successfully bound app to v1 instance")

	if err = m.client.BindService(m.appName, recipientInstanceName); err != nil {
		return errors.Errorf("failed to bind-service %q to application %q: %s", m.appName, recipientInstanceName, err)
	}
	log.Print("Successfully bound app to v2 instance")

	log.Print("Starting migration app")
	if err = m.client.StartApp(m.appName); err != nil {
		return errors.Errorf("failed to start application %q: %s", m.appName, err)
	}

	log.Print("Started to run migration task")
	command := fmt.Sprintf("migrate %s %s", donorInstanceName, recipientInstanceName)

	if opts.SkipTLSValidation {
		command = fmt.Sprintf("migrate -skip-tls-validation %s %s", donorInstanceName, recipientInstanceName)
	}

	if err = m.client.RunTask(m.appName, command); err != nil {
		log.Printf("Migration failed: %s", err)
		// Make best effort to retrieve logs in case of failure, but migration
		// error has priority over logging errors.
		_ = m.outputMigrationLogs("")
	} else {
		log.Print("Migration completed successfully")
		err = m.outputMigrationLogs("APP/TASK/")
	}

	return err
}

func (m *Migrator) outputMigrationLogs(filter string) error {
	log.Print("Fetching log output...")
	time.Sleep(5 * time.Second)
	output, err := m.client.GetLogs(m.appName, filter)
	if err != nil {
		return err
	}

	for _, line := range output {
		fmt.Println(line)
	}
	return nil
}

func (m *Migrator) RenameServiceInstances(donorInstanceName, recipientInstanceName string) error {
	newDonorInstanceName := donorInstanceName + "-old"
	if err := m.client.RenameService(donorInstanceName, newDonorInstanceName); err != nil {
		renameError := `Error renaming service instance %[1]s: %[2]s.
The migration of data from %[1]s to a newly created service instance with name: %[1]s-new has successfully completed.

In order to complete the data migration, please run 'cf rename-service %[1]s %[1]s-old' and
'cf rename-service %[1]s-new %[1]s' to complete the migration process.`

		return fmt.Errorf(renameError, donorInstanceName, err)
	}

	if err := m.client.RenameService(recipientInstanceName, donorInstanceName); err != nil {
		renameError := `Error renaming service instance %[1]s: %[2]s.
The migration of data from %[1]s to a newly created service instance with name: %[1]s-new has successfully completed.

In order to complete the data migration, please run 'cf rename-service %[1]s-new %[1]s' to complete the migration process.`

		return fmt.Errorf(renameError, donorInstanceName, err)
	}

	return nil
}

func (m *Migrator) DeleteServiceInstanceOnError(recipientServiceInstance string) error {
	return m.client.DeleteServiceInstance(recipientServiceInstance)
}
