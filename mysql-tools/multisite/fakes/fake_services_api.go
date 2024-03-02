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

	InstancePlanNameResult struct {
		PlanName string
		Err      error
	}

	PlanExistsResult struct {
		PlanExists bool
		Err        error
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

func (f *FakeFoundation) UpdateServiceAndWait(instanceName string, arbitraryParams string, planName *string) error {
	planString := "<nil>"
	if planName != nil {
		planString = *planName
	}
	*f.Operations = append(*f.Operations, fmt.Sprintf(f.FoundationName+".UpdateServiceAndWait(%q, %q, %s)",
		instanceName, arbitraryParams, planString))

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

func (f *FakeFoundation) InstancePlanName(instanceName string) (planName string, err error) {
	op := fmt.Sprintf("%s.InstancePlanName(%q)", f.FoundationName, instanceName)
	*f.Operations = append(*f.Operations, op)

	return f.InstancePlanNameResult.PlanName, f.InstancePlanNameResult.Err
}

func (f *FakeFoundation) PlanExists(planName string) (exists bool, err error) {
	op := fmt.Sprintf("%s.PlanExists(%q)", f.FoundationName, planName)
	*f.Operations = append(*f.Operations, op)

	return f.PlanExistsResult.PlanExists, f.PlanExistsResult.Err
}

var _ multisite.ServiceAPI = (*FakeFoundation)(nil)
