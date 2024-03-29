// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
)

type FakeMultisiteConfig struct {
	ConfigDirStub        func(string) string
	configDirMutex       sync.RWMutex
	configDirArgsForCall []struct {
		arg1 string
	}
	configDirReturns struct {
		result1 string
	}
	configDirReturnsOnCall map[int]struct {
		result1 string
	}
	ListConfigsStub        func() ([]multisite.Target, error)
	listConfigsMutex       sync.RWMutex
	listConfigsArgsForCall []struct {
	}
	listConfigsReturns struct {
		result1 []multisite.Target
		result2 error
	}
	listConfigsReturnsOnCall map[int]struct {
		result1 []multisite.Target
		result2 error
	}
	RemoveConfigStub        func(string) error
	removeConfigMutex       sync.RWMutex
	removeConfigArgsForCall []struct {
		arg1 string
	}
	removeConfigReturns struct {
		result1 error
	}
	removeConfigReturnsOnCall map[int]struct {
		result1 error
	}
	SaveConfigStub        func(string, string) (multisite.Target, error)
	saveConfigMutex       sync.RWMutex
	saveConfigArgsForCall []struct {
		arg1 string
		arg2 string
	}
	saveConfigReturns struct {
		result1 multisite.Target
		result2 error
	}
	saveConfigReturnsOnCall map[int]struct {
		result1 multisite.Target
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMultisiteConfig) ConfigDir(arg1 string) string {
	fake.configDirMutex.Lock()
	ret, specificReturn := fake.configDirReturnsOnCall[len(fake.configDirArgsForCall)]
	fake.configDirArgsForCall = append(fake.configDirArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ConfigDirStub
	fakeReturns := fake.configDirReturns
	fake.recordInvocation("ConfigDir", []interface{}{arg1})
	fake.configDirMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeMultisiteConfig) ConfigDirCallCount() int {
	fake.configDirMutex.RLock()
	defer fake.configDirMutex.RUnlock()
	return len(fake.configDirArgsForCall)
}

func (fake *FakeMultisiteConfig) ConfigDirCalls(stub func(string) string) {
	fake.configDirMutex.Lock()
	defer fake.configDirMutex.Unlock()
	fake.ConfigDirStub = stub
}

func (fake *FakeMultisiteConfig) ConfigDirArgsForCall(i int) string {
	fake.configDirMutex.RLock()
	defer fake.configDirMutex.RUnlock()
	argsForCall := fake.configDirArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeMultisiteConfig) ConfigDirReturns(result1 string) {
	fake.configDirMutex.Lock()
	defer fake.configDirMutex.Unlock()
	fake.ConfigDirStub = nil
	fake.configDirReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeMultisiteConfig) ConfigDirReturnsOnCall(i int, result1 string) {
	fake.configDirMutex.Lock()
	defer fake.configDirMutex.Unlock()
	fake.ConfigDirStub = nil
	if fake.configDirReturnsOnCall == nil {
		fake.configDirReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.configDirReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeMultisiteConfig) ListConfigs() ([]multisite.Target, error) {
	fake.listConfigsMutex.Lock()
	ret, specificReturn := fake.listConfigsReturnsOnCall[len(fake.listConfigsArgsForCall)]
	fake.listConfigsArgsForCall = append(fake.listConfigsArgsForCall, struct {
	}{})
	stub := fake.ListConfigsStub
	fakeReturns := fake.listConfigsReturns
	fake.recordInvocation("ListConfigs", []interface{}{})
	fake.listConfigsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeMultisiteConfig) ListConfigsCallCount() int {
	fake.listConfigsMutex.RLock()
	defer fake.listConfigsMutex.RUnlock()
	return len(fake.listConfigsArgsForCall)
}

func (fake *FakeMultisiteConfig) ListConfigsCalls(stub func() ([]multisite.Target, error)) {
	fake.listConfigsMutex.Lock()
	defer fake.listConfigsMutex.Unlock()
	fake.ListConfigsStub = stub
}

func (fake *FakeMultisiteConfig) ListConfigsReturns(result1 []multisite.Target, result2 error) {
	fake.listConfigsMutex.Lock()
	defer fake.listConfigsMutex.Unlock()
	fake.ListConfigsStub = nil
	fake.listConfigsReturns = struct {
		result1 []multisite.Target
		result2 error
	}{result1, result2}
}

func (fake *FakeMultisiteConfig) ListConfigsReturnsOnCall(i int, result1 []multisite.Target, result2 error) {
	fake.listConfigsMutex.Lock()
	defer fake.listConfigsMutex.Unlock()
	fake.ListConfigsStub = nil
	if fake.listConfigsReturnsOnCall == nil {
		fake.listConfigsReturnsOnCall = make(map[int]struct {
			result1 []multisite.Target
			result2 error
		})
	}
	fake.listConfigsReturnsOnCall[i] = struct {
		result1 []multisite.Target
		result2 error
	}{result1, result2}
}

func (fake *FakeMultisiteConfig) RemoveConfig(arg1 string) error {
	fake.removeConfigMutex.Lock()
	ret, specificReturn := fake.removeConfigReturnsOnCall[len(fake.removeConfigArgsForCall)]
	fake.removeConfigArgsForCall = append(fake.removeConfigArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.RemoveConfigStub
	fakeReturns := fake.removeConfigReturns
	fake.recordInvocation("RemoveConfig", []interface{}{arg1})
	fake.removeConfigMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeMultisiteConfig) RemoveConfigCallCount() int {
	fake.removeConfigMutex.RLock()
	defer fake.removeConfigMutex.RUnlock()
	return len(fake.removeConfigArgsForCall)
}

func (fake *FakeMultisiteConfig) RemoveConfigCalls(stub func(string) error) {
	fake.removeConfigMutex.Lock()
	defer fake.removeConfigMutex.Unlock()
	fake.RemoveConfigStub = stub
}

func (fake *FakeMultisiteConfig) RemoveConfigArgsForCall(i int) string {
	fake.removeConfigMutex.RLock()
	defer fake.removeConfigMutex.RUnlock()
	argsForCall := fake.removeConfigArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeMultisiteConfig) RemoveConfigReturns(result1 error) {
	fake.removeConfigMutex.Lock()
	defer fake.removeConfigMutex.Unlock()
	fake.RemoveConfigStub = nil
	fake.removeConfigReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeMultisiteConfig) RemoveConfigReturnsOnCall(i int, result1 error) {
	fake.removeConfigMutex.Lock()
	defer fake.removeConfigMutex.Unlock()
	fake.RemoveConfigStub = nil
	if fake.removeConfigReturnsOnCall == nil {
		fake.removeConfigReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.removeConfigReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeMultisiteConfig) SaveConfig(arg1 string, arg2 string) (multisite.Target, error) {
	fake.saveConfigMutex.Lock()
	ret, specificReturn := fake.saveConfigReturnsOnCall[len(fake.saveConfigArgsForCall)]
	fake.saveConfigArgsForCall = append(fake.saveConfigArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.SaveConfigStub
	fakeReturns := fake.saveConfigReturns
	fake.recordInvocation("SaveConfig", []interface{}{arg1, arg2})
	fake.saveConfigMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeMultisiteConfig) SaveConfigCallCount() int {
	fake.saveConfigMutex.RLock()
	defer fake.saveConfigMutex.RUnlock()
	return len(fake.saveConfigArgsForCall)
}

func (fake *FakeMultisiteConfig) SaveConfigCalls(stub func(string, string) (multisite.Target, error)) {
	fake.saveConfigMutex.Lock()
	defer fake.saveConfigMutex.Unlock()
	fake.SaveConfigStub = stub
}

func (fake *FakeMultisiteConfig) SaveConfigArgsForCall(i int) (string, string) {
	fake.saveConfigMutex.RLock()
	defer fake.saveConfigMutex.RUnlock()
	argsForCall := fake.saveConfigArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeMultisiteConfig) SaveConfigReturns(result1 multisite.Target, result2 error) {
	fake.saveConfigMutex.Lock()
	defer fake.saveConfigMutex.Unlock()
	fake.SaveConfigStub = nil
	fake.saveConfigReturns = struct {
		result1 multisite.Target
		result2 error
	}{result1, result2}
}

func (fake *FakeMultisiteConfig) SaveConfigReturnsOnCall(i int, result1 multisite.Target, result2 error) {
	fake.saveConfigMutex.Lock()
	defer fake.saveConfigMutex.Unlock()
	fake.SaveConfigStub = nil
	if fake.saveConfigReturnsOnCall == nil {
		fake.saveConfigReturnsOnCall = make(map[int]struct {
			result1 multisite.Target
			result2 error
		})
	}
	fake.saveConfigReturnsOnCall[i] = struct {
		result1 multisite.Target
		result2 error
	}{result1, result2}
}

func (fake *FakeMultisiteConfig) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.configDirMutex.RLock()
	defer fake.configDirMutex.RUnlock()
	fake.listConfigsMutex.RLock()
	defer fake.listConfigsMutex.RUnlock()
	fake.removeConfigMutex.RLock()
	defer fake.removeConfigMutex.RUnlock()
	fake.saveConfigMutex.RLock()
	defer fake.saveConfigMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMultisiteConfig) recordInvocation(key string, args []interface{}) {
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

var _ commands.MultisiteConfig = new(FakeMultisiteConfig)
