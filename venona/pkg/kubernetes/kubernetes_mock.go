// Copyright 2023 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by mockery v2.32.0. DO NOT EDIT.

package kubernetes

import (
	context "context"

	task "github.com/codefresh-io/go/venona/pkg/task"
	mock "github.com/stretchr/testify/mock"
)

// MockKubernetes is an autogenerated mock type for the Kubernetes type
type MockKubernetes struct {
	mock.Mock
}

type MockKubernetes_Expecter struct {
	mock *mock.Mock
}

func (_m *MockKubernetes) EXPECT() *MockKubernetes_Expecter {
	return &MockKubernetes_Expecter{mock: &_m.Mock}
}

// CreateResource provides a mock function with given fields: ctx, taskType, spec
func (_m *MockKubernetes) CreateResource(ctx context.Context, taskType task.Type, spec interface{}) error {
	ret := _m.Called(ctx, taskType, spec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, task.Type, interface{}) error); ok {
		r0 = rf(ctx, taskType, spec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockKubernetes_CreateResource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateResource'
type MockKubernetes_CreateResource_Call struct {
	*mock.Call
}

// CreateResource is a helper method to define mock.On call
//   - ctx context.Context
//   - taskType task.Type
//   - spec interface{}
func (_e *MockKubernetes_Expecter) CreateResource(ctx interface{}, taskType interface{}, spec interface{}) *MockKubernetes_CreateResource_Call {
	return &MockKubernetes_CreateResource_Call{Call: _e.mock.On("CreateResource", ctx, taskType, spec)}
}

func (_c *MockKubernetes_CreateResource_Call) Run(run func(ctx context.Context, taskType task.Type, spec interface{})) *MockKubernetes_CreateResource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(task.Type), args[2].(interface{}))
	})
	return _c
}

func (_c *MockKubernetes_CreateResource_Call) Return(_a0 error) *MockKubernetes_CreateResource_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKubernetes_CreateResource_Call) RunAndReturn(run func(context.Context, task.Type, interface{}) error) *MockKubernetes_CreateResource_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteResource provides a mock function with given fields: ctx, opts
func (_m *MockKubernetes) DeleteResource(ctx context.Context, opts DeleteOptions) error {
	ret := _m.Called(ctx, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, DeleteOptions) error); ok {
		r0 = rf(ctx, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockKubernetes_DeleteResource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteResource'
type MockKubernetes_DeleteResource_Call struct {
	*mock.Call
}

// DeleteResource is a helper method to define mock.On call
//   - ctx context.Context
//   - opts DeleteOptions
func (_e *MockKubernetes_Expecter) DeleteResource(ctx interface{}, opts interface{}) *MockKubernetes_DeleteResource_Call {
	return &MockKubernetes_DeleteResource_Call{Call: _e.mock.On("DeleteResource", ctx, opts)}
}

func (_c *MockKubernetes_DeleteResource_Call) Run(run func(ctx context.Context, opts DeleteOptions)) *MockKubernetes_DeleteResource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(DeleteOptions))
	})
	return _c
}

func (_c *MockKubernetes_DeleteResource_Call) Return(_a0 error) *MockKubernetes_DeleteResource_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKubernetes_DeleteResource_Call) RunAndReturn(run func(context.Context, DeleteOptions) error) *MockKubernetes_DeleteResource_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockKubernetes creates a new instance of MockKubernetes. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockKubernetes(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockKubernetes {
	mock := &MockKubernetes{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
