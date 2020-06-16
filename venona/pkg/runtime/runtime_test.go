// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"testing"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createKubernetesMock() *mocks.Kubernetes {
	m := &mocks.Kubernetes{}
	m.On("CreateResource", mock.Anything).Return(nil)
	m.On("DeleteResource", mock.Anything).Return(nil)
	return m
}

func Test_runtime_StartWorkflow(t *testing.T) {
	type args struct {
		tasks []task.Task
	}
	tests := []struct {
		name    string
		runtime runtime
		args    args
		wantErr bool
	}{
		{
			name: "should call kube create resouce",
			runtime: runtime{
				client: createKubernetesMock(),
			},
			args: args{
				tasks: []task.Task{
					{
						Type: "runtime",
						Spec: "spec",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.runtime
			mo := r.client.(*mocks.Kubernetes)

			err := r.StartWorkflow(tt.args.tasks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			mo.AssertCalled(t, "CreateResource", tt.args.tasks[0].Spec)
		})
	}
}

func Test_runtime_TerminateWorkflow(t *testing.T) {
	type fields struct {
		client kubernetes.Kubernetes
	}
	type args struct {
		tasks []task.Task
	}
	tests := []struct {
		name        string
		runtime     runtime
		args        args
		wantErr     bool
		expectedOpt kubernetes.DeleteOptions
	}{
		{
			name: "should call kube delete resouce",
			runtime: runtime{
				client: createKubernetesMock(),
			},
			args: args{
				tasks: []task.Task{
					{
						Type: "runtime",
						Spec: `{"Name":"name", "Namespace":"ns"}`,
					},
				},
			},
			expectedOpt: kubernetes.DeleteOptions{
				Kind:      "runtime",
				Name:      "name",
				Namespace: "ns",
			},
		},
		{
			name: "should fail if spec is not string",
			runtime: runtime{
				client: createKubernetesMock(),
			},
			args: args{
				tasks: []task.Task{
					{
						Type: "runtime",
						Spec: 123,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.runtime
			mo := r.client.(*mocks.Kubernetes)

			err := r.TerminateWorkflow(tt.args.tasks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			mo.AssertCalled(t, "DeleteResource", tt.expectedOpt)
		})
	}
}
