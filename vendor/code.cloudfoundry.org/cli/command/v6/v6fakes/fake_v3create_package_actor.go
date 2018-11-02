// Code generated by counterfeiter. DO NOT EDIT.
package v6fakes

import (
	sync "sync"

	v3action "code.cloudfoundry.org/cli/actor/v3action"
	v6 "code.cloudfoundry.org/cli/command/v6"
)

type FakeV3CreatePackageActor struct {
	CloudControllerAPIVersionStub        func() string
	cloudControllerAPIVersionMutex       sync.RWMutex
	cloudControllerAPIVersionArgsForCall []struct {
	}
	cloudControllerAPIVersionReturns struct {
		result1 string
	}
	cloudControllerAPIVersionReturnsOnCall map[int]struct {
		result1 string
	}
	CreateAndUploadBitsPackageByApplicationNameAndSpaceStub        func(string, string, string) (v3action.Package, v3action.Warnings, error)
	createAndUploadBitsPackageByApplicationNameAndSpaceMutex       sync.RWMutex
	createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
	}
	createAndUploadBitsPackageByApplicationNameAndSpaceReturns struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}
	createAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall map[int]struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}
	CreateDockerPackageByApplicationNameAndSpaceStub        func(string, string, v3action.DockerImageCredentials) (v3action.Package, v3action.Warnings, error)
	createDockerPackageByApplicationNameAndSpaceMutex       sync.RWMutex
	createDockerPackageByApplicationNameAndSpaceArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 v3action.DockerImageCredentials
	}
	createDockerPackageByApplicationNameAndSpaceReturns struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}
	createDockerPackageByApplicationNameAndSpaceReturnsOnCall map[int]struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeV3CreatePackageActor) CloudControllerAPIVersion() string {
	fake.cloudControllerAPIVersionMutex.Lock()
	ret, specificReturn := fake.cloudControllerAPIVersionReturnsOnCall[len(fake.cloudControllerAPIVersionArgsForCall)]
	fake.cloudControllerAPIVersionArgsForCall = append(fake.cloudControllerAPIVersionArgsForCall, struct {
	}{})
	fake.recordInvocation("CloudControllerAPIVersion", []interface{}{})
	fake.cloudControllerAPIVersionMutex.Unlock()
	if fake.CloudControllerAPIVersionStub != nil {
		return fake.CloudControllerAPIVersionStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.cloudControllerAPIVersionReturns
	return fakeReturns.result1
}

func (fake *FakeV3CreatePackageActor) CloudControllerAPIVersionCallCount() int {
	fake.cloudControllerAPIVersionMutex.RLock()
	defer fake.cloudControllerAPIVersionMutex.RUnlock()
	return len(fake.cloudControllerAPIVersionArgsForCall)
}

func (fake *FakeV3CreatePackageActor) CloudControllerAPIVersionCalls(stub func() string) {
	fake.cloudControllerAPIVersionMutex.Lock()
	defer fake.cloudControllerAPIVersionMutex.Unlock()
	fake.CloudControllerAPIVersionStub = stub
}

func (fake *FakeV3CreatePackageActor) CloudControllerAPIVersionReturns(result1 string) {
	fake.cloudControllerAPIVersionMutex.Lock()
	defer fake.cloudControllerAPIVersionMutex.Unlock()
	fake.CloudControllerAPIVersionStub = nil
	fake.cloudControllerAPIVersionReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeV3CreatePackageActor) CloudControllerAPIVersionReturnsOnCall(i int, result1 string) {
	fake.cloudControllerAPIVersionMutex.Lock()
	defer fake.cloudControllerAPIVersionMutex.Unlock()
	fake.CloudControllerAPIVersionStub = nil
	if fake.cloudControllerAPIVersionReturnsOnCall == nil {
		fake.cloudControllerAPIVersionReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.cloudControllerAPIVersionReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpace(arg1 string, arg2 string, arg3 string) (v3action.Package, v3action.Warnings, error) {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Lock()
	ret, specificReturn := fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall[len(fake.createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall)]
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall = append(fake.createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	fake.recordInvocation("CreateAndUploadBitsPackageByApplicationNameAndSpace", []interface{}{arg1, arg2, arg3})
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Unlock()
	if fake.CreateAndUploadBitsPackageByApplicationNameAndSpaceStub != nil {
		return fake.CreateAndUploadBitsPackageByApplicationNameAndSpaceStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	fakeReturns := fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturns
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpaceCallCount() int {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RUnlock()
	return len(fake.createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall)
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpaceCalls(stub func(string, string, string) (v3action.Package, v3action.Warnings, error)) {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateAndUploadBitsPackageByApplicationNameAndSpaceStub = stub
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall(i int) (string, string, string) {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RUnlock()
	argsForCall := fake.createAndUploadBitsPackageByApplicationNameAndSpaceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpaceReturns(result1 v3action.Package, result2 v3action.Warnings, result3 error) {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateAndUploadBitsPackageByApplicationNameAndSpaceStub = nil
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturns = struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeV3CreatePackageActor) CreateAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall(i int, result1 v3action.Package, result2 v3action.Warnings, result3 error) {
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateAndUploadBitsPackageByApplicationNameAndSpaceStub = nil
	if fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall == nil {
		fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall = make(map[int]struct {
			result1 v3action.Package
			result2 v3action.Warnings
			result3 error
		})
	}
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceReturnsOnCall[i] = struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpace(arg1 string, arg2 string, arg3 v3action.DockerImageCredentials) (v3action.Package, v3action.Warnings, error) {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.Lock()
	ret, specificReturn := fake.createDockerPackageByApplicationNameAndSpaceReturnsOnCall[len(fake.createDockerPackageByApplicationNameAndSpaceArgsForCall)]
	fake.createDockerPackageByApplicationNameAndSpaceArgsForCall = append(fake.createDockerPackageByApplicationNameAndSpaceArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 v3action.DockerImageCredentials
	}{arg1, arg2, arg3})
	fake.recordInvocation("CreateDockerPackageByApplicationNameAndSpace", []interface{}{arg1, arg2, arg3})
	fake.createDockerPackageByApplicationNameAndSpaceMutex.Unlock()
	if fake.CreateDockerPackageByApplicationNameAndSpaceStub != nil {
		return fake.CreateDockerPackageByApplicationNameAndSpaceStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	fakeReturns := fake.createDockerPackageByApplicationNameAndSpaceReturns
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpaceCallCount() int {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.RUnlock()
	return len(fake.createDockerPackageByApplicationNameAndSpaceArgsForCall)
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpaceCalls(stub func(string, string, v3action.DockerImageCredentials) (v3action.Package, v3action.Warnings, error)) {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateDockerPackageByApplicationNameAndSpaceStub = stub
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpaceArgsForCall(i int) (string, string, v3action.DockerImageCredentials) {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.RUnlock()
	argsForCall := fake.createDockerPackageByApplicationNameAndSpaceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpaceReturns(result1 v3action.Package, result2 v3action.Warnings, result3 error) {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateDockerPackageByApplicationNameAndSpaceStub = nil
	fake.createDockerPackageByApplicationNameAndSpaceReturns = struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeV3CreatePackageActor) CreateDockerPackageByApplicationNameAndSpaceReturnsOnCall(i int, result1 v3action.Package, result2 v3action.Warnings, result3 error) {
	fake.createDockerPackageByApplicationNameAndSpaceMutex.Lock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.Unlock()
	fake.CreateDockerPackageByApplicationNameAndSpaceStub = nil
	if fake.createDockerPackageByApplicationNameAndSpaceReturnsOnCall == nil {
		fake.createDockerPackageByApplicationNameAndSpaceReturnsOnCall = make(map[int]struct {
			result1 v3action.Package
			result2 v3action.Warnings
			result3 error
		})
	}
	fake.createDockerPackageByApplicationNameAndSpaceReturnsOnCall[i] = struct {
		result1 v3action.Package
		result2 v3action.Warnings
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeV3CreatePackageActor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cloudControllerAPIVersionMutex.RLock()
	defer fake.cloudControllerAPIVersionMutex.RUnlock()
	fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createAndUploadBitsPackageByApplicationNameAndSpaceMutex.RUnlock()
	fake.createDockerPackageByApplicationNameAndSpaceMutex.RLock()
	defer fake.createDockerPackageByApplicationNameAndSpaceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeV3CreatePackageActor) recordInvocation(key string, args []interface{}) {
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

var _ v6.V3CreatePackageActor = new(FakeV3CreatePackageActor)
