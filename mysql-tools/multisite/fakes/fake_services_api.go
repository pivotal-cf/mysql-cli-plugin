package fakes

import (
	"fmt"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
)

type FakeFoundation struct {
	FoundationName string
	Operations     *[]string

	InstanceExistsResult struct {
		Err error
	}

	UpdateServiceResult struct {
		ErrFunc func(instanceName, arbitraryParams string) error
		Err     error
	}

	CreateHostInfoKeyResult struct {
		Key string
		Err error
	}

	CreateCredentialsKeyResult struct {
		Key string
		Err error
	}
}

func (f *FakeFoundation) ID() string {
	return f.FoundationName
}

func (f *FakeFoundation) UpdateServiceAndWait(instanceName, arbitraryParams string) error {
	*f.Operations = append(*f.Operations, fmt.Sprintf(f.FoundationName+".UpdateServiceAndWait(%q, %q)",
		instanceName, arbitraryParams))

	if f.UpdateServiceResult.ErrFunc != nil {
		return f.UpdateServiceResult.ErrFunc(instanceName, arbitraryParams)
	}

	return f.UpdateServiceResult.Err
}

func (f *FakeFoundation) CreateHostInfoKey(instanceName string) (key string, err error) {
	op := fmt.Sprintf("%s.CreateHostInfoKey(%q)",
		f.FoundationName, instanceName)
	*f.Operations = append(*f.Operations, op)

	return f.CreateHostInfoKeyResult.Key, f.CreateHostInfoKeyResult.Err
}

func (f *FakeFoundation) CreateCredentialsKey(instanceName string) (key string, err error) {
	op := fmt.Sprintf("%s.CreateCredentialsKey(%q)",
		f.FoundationName, instanceName)
	*f.Operations = append(*f.Operations, op)

	return f.CreateCredentialsKeyResult.Key, f.CreateCredentialsKeyResult.Err
}

func (f *FakeFoundation) InstanceExists(instanceName string) (err error) {
	op := fmt.Sprintf("%s.InstanceExists(%q)", f.FoundationName, instanceName)
	*f.Operations = append(*f.Operations, op)

	return f.InstanceExistsResult.Err
}

var _ multisite.ServiceAPI = (*FakeFoundation)(nil)
