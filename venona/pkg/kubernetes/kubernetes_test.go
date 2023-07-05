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

package kubernetes

import (
	"context"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/codefresh-io/go/venona/pkg/task"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		opts    Options
		want    kube
		wantErr string
	}{
		"should succeed with a valid type": {
			opts: Options{
				Type: "runtime",
			},
			want: kube{},
		},
		"should fail with an invalid type": {
			opts: Options{
				Type: "secret",
			},
			wantErr: "not a valid type",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := New(tt.opts)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func createMockLogger() *mocks.Logger {
	l := &mocks.Logger{}
	l.On("Info", mock.Anything, "namespace", mock.AnythingOfType("string"), "name", mock.AnythingOfType("string")).Return(nil)
	return l
}

func Test_kube_CreateResource(t *testing.T) {
	tests := map[string]struct {
		client  kubernetes.Interface
		spec    interface{}
		wantErr string
	}{
		"Should succesfully create a pod": {
			client: fake.NewSimpleClientset(),
			spec: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "some-pod",
					"namespace": "some-namespace",
				},
			},
		},
		"Should succesfully create a PCV": {
			client: fake.NewSimpleClientset(),
			spec: map[string]interface{}{
				"kind":       "PersistentVolumeClaim",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      "some-pvc",
					"namespace": "some-namespace",
				},
			},
		},
		"Should fail creating a Deployment": {
			client:  fake.NewSimpleClientset(),
			wantErr: "failed creating resource of type apps/v1, Kind=Deployment",
			spec: map[string]interface{}{
				"kind":       "Deployment",
				"apiVersion": "apps/v1",
				"metadata": map[string]interface{}{
					"name":      "some-deployment",
					"namespace": "some-namespace",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			k := kube{
				client: tt.client,
				logger: createMockLogger(),
			}
			err := k.CreateResource(context.Background(), tt.spec)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func Test_kube_DeleteResource(t *testing.T) {
	tests := map[string]struct {
		client  kubernetes.Interface
		opts    DeleteOptions
		wantErr string
	}{
		"Should successfully delete an existing Pod": {
			client: fake.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-pod",
				},
			}),
			opts: DeleteOptions{
				Kind:      task.TypeDeletePod,
				Namespace: "some-namespace",
				Name:      "some-pod",
			},
		},
		"Should successfully delete an existing PVC": {
			client: fake.NewSimpleClientset(&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-namespace",
					Name:      "some-pvc",
				},
			}),
			opts: DeleteOptions{
				Kind:      task.TypeDeletePVC,
				Namespace: "some-namespace",
				Name:      "some-pvc",
			},
		},
		"Should fail deleting an unexisting Pod": {
			client: fake.NewSimpleClientset(),
			opts: DeleteOptions{
				Kind:      task.TypeDeletePod,
				Namespace: "some-namespace",
				Name:      "some-pod",
			},
			wantErr: "failed deleting pod \"some-namespace\\some-pod\": pods \"some-pod\" not found",
		},
		"Should fail deleting an unexisting PVC": {
			client: fake.NewSimpleClientset(),
			opts: DeleteOptions{
				Kind:      task.TypeDeletePod,
				Namespace: "some-namespace",
				Name:      "some-pvc",
			},
			wantErr: "failed deleting pod \"some-namespace\\some-pvc\": pods \"some-pvc\" not found",
		},
		"Should fail deleting an unknown type": {
			client: fake.NewSimpleClientset(),
			opts: DeleteOptions{
				Kind:      "unknown-type",
				Namespace: "some-namespace",
				Name:      "some-pvc",
			},
			wantErr: "failed deleting resource of type unknown-type",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			k := kube{
				client: tt.client,
				logger: createMockLogger(),
			}
			err := k.DeleteResource(context.Background(), tt.opts)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
