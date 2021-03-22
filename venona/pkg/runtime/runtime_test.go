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
	"context"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createKubernetesMock() *kubernetes.MockKubernetes {
	m := &kubernetes.MockKubernetes{}
	m.On("CreateResource", mock.Anything, mock.Anything).Return(nil)
	m.On("DeleteResource", mock.Anything, mock.Anything).Return(nil)
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
			mo := r.client.(*kubernetes.MockKubernetes)

			ctx := context.Background()
			err := r.StartWorkflow(ctx, tt.args.tasks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			mo.AssertCalled(t, "CreateResource", ctx, tt.args.tasks[0].Spec)
		})
	}
}

func Test_runtime_TerminateWorkflow(t *testing.T) {
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
						Spec: map[string]interface{}{
							"name":      "name",
							"namespace": "ns",
						},
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
			mo := r.client.(*kubernetes.MockKubernetes)

			ctx := context.Background()
			errs := r.TerminateWorkflow(ctx, tt.args.tasks)
			if tt.wantErr {
				assert.Equal(t, len(errs), 1)
				return
			}

			mo.AssertCalled(t, "DeleteResource", ctx, tt.expectedOpt)
		})
	}
}
