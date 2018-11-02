// This file was generated by counterfeiter
package appfilesfakes

import (
	"sync"

	"code.cloudfoundry.org/cli/cf/appfiles"
)

type FakeCfIgnore struct {
	FileShouldBeIgnoredStub        func(path string) bool
	fileShouldBeIgnoredMutex       sync.RWMutex
	fileShouldBeIgnoredArgsForCall []struct {
		path string
	}
	fileShouldBeIgnoredReturns struct {
		result1 bool
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCfIgnore) FileShouldBeIgnored(path string) bool {
	fake.fileShouldBeIgnoredMutex.Lock()
	fake.fileShouldBeIgnoredArgsForCall = append(fake.fileShouldBeIgnoredArgsForCall, struct {
		path string
	}{path})
	fake.recordInvocation("FileShouldBeIgnored", []interface{}{path})
	fake.fileShouldBeIgnoredMutex.Unlock()
	if fake.FileShouldBeIgnoredStub != nil {
		return fake.FileShouldBeIgnoredStub(path)
	} else {
		return fake.fileShouldBeIgnoredReturns.result1
	}
}

func (fake *FakeCfIgnore) FileShouldBeIgnoredCallCount() int {
	fake.fileShouldBeIgnoredMutex.RLock()
	defer fake.fileShouldBeIgnoredMutex.RUnlock()
	return len(fake.fileShouldBeIgnoredArgsForCall)
}

func (fake *FakeCfIgnore) FileShouldBeIgnoredArgsForCall(i int) string {
	fake.fileShouldBeIgnoredMutex.RLock()
	defer fake.fileShouldBeIgnoredMutex.RUnlock()
	return fake.fileShouldBeIgnoredArgsForCall[i].path
}

func (fake *FakeCfIgnore) FileShouldBeIgnoredReturns(result1 bool) {
	fake.FileShouldBeIgnoredStub = nil
	fake.fileShouldBeIgnoredReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeCfIgnore) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.fileShouldBeIgnoredMutex.RLock()
	defer fake.fileShouldBeIgnoredMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCfIgnore) recordInvocation(key string, args []interface{}) {
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

var _ appfiles.CfIgnore = new(FakeCfIgnore)
