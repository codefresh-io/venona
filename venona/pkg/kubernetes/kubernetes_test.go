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

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNew(t *testing.T) {
	type args struct {
		opt Options
	}
	tests := []struct {
		name        string
		args        args
		want        kube
		wantErr     bool
		errorString string
	}{
		{
			name: "on valid input retun kube",
			args: args{
				opt: Options{
					Type: "runtime",
				},
			},
			want:    kube{},
			wantErr: false,
		},
		{
			name: "on non valid type return errNotValidType",
			args: args{
				opt: Options{
					Type: "secret",
				},
			},
			wantErr:     true,
			errorString: "not a valid type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.opt)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errorString)
			}
		})
	}
}

func createFakeClientSetForPodOperation(t *testing.T, ns string) kubernetes.Interface {
	client := fake.NewSimpleClientset()
	client.Fake.PrependReactor("create", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		assert.Equal(t, ns, action.GetNamespace())
		return true, nil, nil
	})
	client.Fake.PrependReactor("delete", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		assert.Equal(t, ns, action.GetNamespace())
		return true, nil, nil
	})
	return client
}

func createFakeClientSetForPvcOperation(t *testing.T, ns string) kubernetes.Interface {
	client := fake.NewSimpleClientset()
	client.Fake.PrependReactor("create", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		assert.Equal(t, ns, action.GetNamespace())
		return true, nil, nil
	})
	client.Fake.PrependReactor("delete", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		assert.Equal(t, ns, action.GetNamespace())
		return true, nil, nil
	})
	return client
}

func createMockLogger() *mocks.Logger {
	l := &mocks.Logger{}
	l.On("Info", mock.Anything).Return(nil)
	return l
}

func Test_kube_CreateResource(t *testing.T) {
	type fields struct {
		client kubernetes.Interface
		logger logger.Logger
	}
	type args struct {
		spec interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantMsg string
	}{
		{
			name: "shoul call create pod if spec type is create pod	",
			fields: fields{
				client: createFakeClientSetForPodOperation(t, "ns"),
				logger: createMockLogger(),
			},
			wantMsg: "Pod has been created",
			args: args{
				spec: map[string]interface{}{
					"kind":       "Pod",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "dind",
						"namespace": "ns",
					},
				},
			},
		},
		{
			name: "shoul call create PersistentVolumeClaim if spec type is create PersistentVolumeClaim	",
			fields: fields{
				client: createFakeClientSetForPvcOperation(t, "ns"),
				logger: createMockLogger(),
			},
			wantMsg: "PersistentVolumeClaim has been created",
			args: args{
				spec: map[string]interface{}{
					"kind":       "PersistentVolumeClaim",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name":      "dind",
						"namespace": "ns",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := kube{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			mo := k.logger.(*mocks.Logger)
			err := k.CreateResource(context.Background(), tt.args.spec)
			mo.AssertCalled(t, "Info", tt.wantMsg)
			if tt.wantErr {
				assert.Error(t, err)
				//	assert.EqualError(t, err, tt.errorString)
			}
		})
	}
}

func Test_kube_DeleteResource(t *testing.T) {
	type fields struct {
		client kubernetes.Interface
		logger logger.Logger
	}
	type args struct {
		opt DeleteOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		wantMsg string
	}{
		{
			name: "shoul call create pod if spec type is create pod	",
			fields: fields{
				client: createFakeClientSetForPodOperation(t, "ns"),
				logger: createMockLogger(),
			},
			wantMsg: "Pod has been deleted",
			args: args{
				DeleteOptions{
					Kind:      task.TypeDeletePod,
					Namespace: "ns",
					Name:      "name",
				},
			},
		},
		{
			name: "shoul call create PersistentVolumeClaim if spec type is create PersistentVolumeClaim	",
			fields: fields{
				client: createFakeClientSetForPvcOperation(t, "ns"),
				logger: createMockLogger(),
			},
			wantMsg: "PersistentVolumeClaim has been deleted",
			args: args{
				DeleteOptions{
					Kind:      task.TypeDeletePVC,
					Namespace: "ns",
					Name:      "name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := kube{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := k.DeleteResource(context.Background(), tt.args.opt); (err != nil) != tt.wantErr {
				t.Errorf("kube.DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
			}
			mo := k.logger.(*mocks.Logger)
			mo.AssertCalled(t, "Info", tt.wantMsg)
		})
	}
}
