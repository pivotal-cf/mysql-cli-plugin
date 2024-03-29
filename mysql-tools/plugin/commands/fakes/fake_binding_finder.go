// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	find_bindings "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
)

type FakeBindingFinder struct {
	FindBindingsStub        func(string) ([]find_bindings.Binding, error)
	findBindingsMutex       sync.RWMutex
	findBindingsArgsForCall []struct {
		arg1 string
	}
	findBindingsReturns struct {
		result1 []find_bindings.Binding
		result2 error
	}
	findBindingsReturnsOnCall map[int]struct {
		result1 []find_bindings.Binding
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBindingFinder) FindBindings(arg1 string) ([]find_bindings.Binding, error) {
	fake.findBindingsMutex.Lock()
	ret, specificReturn := fake.findBindingsReturnsOnCall[len(fake.findBindingsArgsForCall)]
	fake.findBindingsArgsForCall = append(fake.findBindingsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.FindBindingsStub
	fakeReturns := fake.findBindingsReturns
	fake.recordInvocation("FindBindings", []interface{}{arg1})
	fake.findBindingsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBindingFinder) FindBindingsCallCount() int {
	fake.findBindingsMutex.RLock()
	defer fake.findBindingsMutex.RUnlock()
	return len(fake.findBindingsArgsForCall)
}

func (fake *FakeBindingFinder) FindBindingsCalls(stub func(string) ([]find_bindings.Binding, error)) {
	fake.findBindingsMutex.Lock()
	defer fake.findBindingsMutex.Unlock()
	fake.FindBindingsStub = stub
}

func (fake *FakeBindingFinder) FindBindingsArgsForCall(i int) string {
	fake.findBindingsMutex.RLock()
	defer fake.findBindingsMutex.RUnlock()
	argsForCall := fake.findBindingsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeBindingFinder) FindBindingsReturns(result1 []find_bindings.Binding, result2 error) {
	fake.findBindingsMutex.Lock()
	defer fake.findBindingsMutex.Unlock()
	fake.FindBindingsStub = nil
	fake.findBindingsReturns = struct {
		result1 []find_bindings.Binding
		result2 error
	}{result1, result2}
}

func (fake *FakeBindingFinder) FindBindingsReturnsOnCall(i int, result1 []find_bindings.Binding, result2 error) {
	fake.findBindingsMutex.Lock()
	defer fake.findBindingsMutex.Unlock()
	fake.FindBindingsStub = nil
	if fake.findBindingsReturnsOnCall == nil {
		fake.findBindingsReturnsOnCall = make(map[int]struct {
			result1 []find_bindings.Binding
			result2 error
		})
	}
	fake.findBindingsReturnsOnCall[i] = struct {
		result1 []find_bindings.Binding
		result2 error
	}{result1, result2}
}

func (fake *FakeBindingFinder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.findBindingsMutex.RLock()
	defer fake.findBindingsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBindingFinder) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ commands.BindingFinder = new(FakeBindingFinder)
