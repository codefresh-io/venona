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

package queue

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func checkPodInFakeClientset(client *fake.Clientset, name string) bool {
	r, err := client.Tracker().Get(v1.SchemeGroupVersion.WithResource("pods"), "some-namespace", name)
	if err != nil || r == nil {
		return false
	}

	return true
}

func TestTaskQueue_Enqueue(t *testing.T) {
	type taskOrSleep struct {
		t     *task.Task
		sleep time.Duration
	}
	tests := map[string]struct {
		tasks   []taskOrSleep
		afterFn func(t *testing.T, client *fake.Clientset)
	}{
		"should create a single pod for a single task": {
			tasks: []taskOrSleep{
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod-wf1",
							},
						},
					},
				},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.True(t, checkPodInFakeClientset(client, "pod-wf1"))
			},
		},
		"should create and delete the pod": {
			tasks: []taskOrSleep{
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod-wf1",
							},
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeDeletePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: map[string]string{
							"Namespace": "some-namespace",
							"Name":      "pod-wf1",
						},
					},
				},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.False(t, checkPodInFakeClientset(client, "pod-wf1"))
			},
		},
		"should create and delete the pod after a sleep": {
			tasks: []taskOrSleep{
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod-wf1",
							},
						},
					},
				},
				{
					sleep: time.Millisecond * 100,
				},
				{
					t: &task.Task{
						Type: task.TypeDeletePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: map[string]string{
							"Namespace": "some-namespace",
							"Name":      "pod-wf1",
						},
					},
				},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.False(t, checkPodInFakeClientset(client, "pod-wf1"))
			},
		},
		"should create multiple pods, sleep, delete some of them, create new ones": {
			tasks: []taskOrSleep{
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod1-wf1",
							},
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod2-wf1",
							},
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-2",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod-wf2",
							},
						},
					},
				},
				{
					sleep: time.Millisecond * 100,
				},
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod1-wf1-retry-1",
							},
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeCreatePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: &v1.Pod{
							TypeMeta: metav1.TypeMeta{
								APIVersion: v1.SchemeGroupVersion.String(),
								Kind:       "Pod",
							},
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "some-namespace",
								Name:      "pod2-wf1-retry-1",
							},
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeDeletePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-1",
						},
						Spec: map[string]string{
							"Namespace": "some-namespace",
							"Name":      "pod1-wf1",
						},
					},
				},
				{
					t: &task.Task{
						Type: task.TypeDeletePod,
						Metadata: task.Metadata{
							ReName:   "some-rt",
							Workflow: "wf-2",
						},
						Spec: map[string]string{
							"Namespace": "some-namespace",
							"Name":      "pod2-wf1",
						},
					},
				},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.True(t, checkPodInFakeClientset(client, "pod1-wf1-retry-1"))
				assert.True(t, checkPodInFakeClientset(client, "pod2-wf1-retry-1"))
				assert.False(t, checkPodInFakeClientset(client, "pod1-wf1"))
				assert.False(t, checkPodInFakeClientset(client, "pod2-wf1"))
				assert.False(t, checkPodInFakeClientset(client, "pod1-wf2"))
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			runtimes := map[string]runtime.Runtime{
				"some-rt": runtime.New(runtime.Options{
					Kubernetes: kubernetes.NewWithClient(client),
				}),
			}
			log := logger.New(logger.Options{})
			wg := &sync.WaitGroup{}
			tq := New(runtimes, log, wg, monitoring.NewEmpty())
			wg.Add(1)
			for _, tOrS := range tt.tasks {
				if tOrS.t != nil {
					tq.Enqueue(context.Background(), tOrS.t)
				} else {
					time.Sleep(tOrS.sleep)
				}
			}

			wg.Done()
			wg.Wait()
			tt.afterFn(t, client)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		runtimes map[string]runtime.Runtime
		log      logger.Logger
		wg       *sync.WaitGroup
		monitor  monitoring.Monitor
	}
	tests := []struct {
		name string
		args args
		want *TaskQueue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.runtimes, tt.args.log, tt.args.wg, tt.args.monitor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
