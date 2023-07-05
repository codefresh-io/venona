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
	"errors"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/task"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_runtime_HandleTask(t *testing.T) {
	tests := map[string]struct {
		task     task.Task
		wantErr  string
		beforeFn func(k *kubernetes.MockKubernetes)
	}{
		"should successfully create a resource on TypeCreatePVC task": {
			task: task.Task{
				Type: task.TypeCreatePVC,
				Spec: "some spec",
			},
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("CreateResource", mock.Anything, "some spec").Return(nil)
			},
		},
		"should successfully create a resource on TypeCreatePod task": {
			task: task.Task{
				Type: task.TypeCreatePod,
				Spec: "some spec",
			},
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("CreateResource", mock.Anything, "some spec").Return(nil)
			},
		},
		"should successfully delete a resource on TypeDeletePVC task": {
			task: task.Task{
				Type: task.TypeDeletePVC,
				Spec: map[string]string{
					"Name":      "some-name",
					"Namespace": "some-namespace",
				},
			},
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("DeleteResource", mock.Anything, kubernetes.DeleteOptions{
					Kind:      task.TypeDeletePVC,
					Name:      "some-name",
					Namespace: "some-namespace",
				}).Return(nil)
			},
		},
		"should successfully delete a resource on TypeDeletePod task": {
			task: task.Task{
				Type: task.TypeDeletePod,
				Spec: map[string]string{
					"Name":      "some-name",
					"Namespace": "some-namespace",
				},
			},
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("DeleteResource", mock.Anything, kubernetes.DeleteOptions{
					Kind:      task.TypeDeletePod,
					Name:      "some-name",
					Namespace: "some-namespace",
				}).Return(nil)
			},
		},
		"should fail for unknown type": {
			task: task.Task{
				Type: "some-type",
			},
			wantErr: "unknown task type \"some-type\"",
		},
		"should fail creating if k8s client fails":{
			task: task.Task{
				Type: task.TypeCreatePod,
				Spec: "some spec",
			},
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("CreateResource", mock.Anything, "some spec").Return(errors.New("some error"))
			},
			wantErr: "failed creating resource: some error",
		},
		"should fail deleting if json.unmarshal fails": {
			task: task.Task{
				Type: task.TypeDeletePod,
				Spec: "bad spec",
			},
			wantErr: "failed to unmarshal task spec: json: cannot unmarshal string into Go value of type kubernetes.DeleteOptions",
		},
		"should fail deleting if client fails": {
			task: task.Task{
				Type: task.TypeDeletePod,
				Spec: map[string]string{
					"Name":      "some-name",
					"Namespace": "some-namespace",
				},
			},
			wantErr: "failed deleting resource: some error",
			beforeFn: func(k *kubernetes.MockKubernetes) {
				k.On("DeleteResource", mock.Anything, kubernetes.DeleteOptions{
					Kind:      task.TypeDeletePod,
					Name:      "some-name",
					Namespace: "some-namespace",
				}).Return(errors.New("some error"))
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			k := &kubernetes.MockKubernetes{}
			if tt.beforeFn != nil {
				tt.beforeFn(k)
			}

			r := runtime{
				client: k,
			}
			err := r.HandleTask(context.Background(), tt.task)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
