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
	"sync"
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func makePodCreationTask(workflow, pod string) *task.Task {
	return &task.Task{
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
				Name:      workflow + "-" + pod,
			},
		},
	}
}

func makePodDeletionTask(workflow, pod string) *task.Task {
	return &task.Task{
		Type: task.TypeDeletePod,
		Metadata: task.Metadata{
			ReName:   "some-rt",
			Workflow: workflow,
		},
		Spec: map[string]string{
			"Namespace": "some-namespace",
			"Name":      workflow + "-" + pod,
		},
	}
}

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
				{t: makePodCreationTask("wf1", "pod1")},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.True(t, checkPodInFakeClientset(client, "wf1-pod1"))
			},
		},
		"should create and delete the pod": {
			tasks: []taskOrSleep{
				{t: makePodCreationTask("wf1", "pod1")},
				{t: makePodDeletionTask("wf1", "pod1")},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.False(t, checkPodInFakeClientset(client, "pod-wf1"))
			},
		},
		"should create and delete the pod after a sleep": {
			tasks: []taskOrSleep{
				{t: makePodCreationTask("wf1", "pod1")},
				{sleep: time.Millisecond * 100},
				{t: makePodDeletionTask("wf1", "pod1")},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.False(t, checkPodInFakeClientset(client, "pod-wf1"))
			},
		},
		"should create multiple pods, sleep, delete some of them, create new ones": {
			tasks: []taskOrSleep{
				{t: makePodCreationTask("wf1", "pod1")},
				{t: makePodCreationTask("wf1", "pod2")},
				{t: makePodCreationTask("wf2", "pod1")},
				{sleep: time.Millisecond * 100},
				{t: makePodDeletionTask("wf1", "pod1")},
				{t: makePodCreationTask("wf1", "pod1-retry-1")},
				{t: makePodDeletionTask("wf1", "pod2")},
				{t: makePodCreationTask("wf1", "pod2-retry-1")},
				{t: makePodDeletionTask("wf2", "pod1")},
			},
			afterFn: func(t *testing.T, client *fake.Clientset) {
				assert.True(t, checkPodInFakeClientset(client, "wf1-pod1-retry-1"))
				assert.True(t, checkPodInFakeClientset(client, "wf1-pod2-retry-1"))
				assert.False(t, checkPodInFakeClientset(client, "wf1-pod"))
				assert.False(t, checkPodInFakeClientset(client, "wf1-pod2"))
				assert.False(t, checkPodInFakeClientset(client, "wf2-pod1"))
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
			for _, tOrS := range tt.tasks {
				if tOrS.t != nil {
					tq.Enqueue(context.Background(), tOrS.t)
				} else {
					time.Sleep(tOrS.sleep)
				}
			}

			wg.Wait()
			tt.afterFn(t, client)
		})
	}
}

func TestTaskQueue_NoDeadlock(t *testing.T) {
	type dummySpec struct {
		workflow string
		order    int
	}

	k := &kubernetes.MockKubernetes{}
	runtimes := map[string]runtime.Runtime{
		"some-rt": runtime.New(runtime.Options{
			Kubernetes: k,
		}),
	}
	log := logger.New(logger.Options{})
	wg := &sync.WaitGroup{}
	tq := New(runtimes, log, wg, monitoring.NewEmpty())

	doneTasks := make(map[string]int)
	testLock := sync.Mutex{}
	wf1Chan := make(chan bool)
	wf2Chan := make(chan bool)
	k.On("CreateResource", mock.Anything, mock.AnythingOfType("queue.dummySpec")).Return(func(_ context.Context, spec interface{}) error {
		s, _ := spec.(dummySpec)
		testLock.Lock()
		doneTask := doneTasks[s.workflow]
		testLock.Unlock()
		assert.Equal(t, doneTask+1, s.order, "workflow %s", s.workflow)
		if s.workflow == "wf1" && s.order == 3 {
			// wait for a signal before continuing with the 3rd wf1 task
			<-wf1Chan
		}

		if s.workflow == "wf2" && s.order == 3 {
			// notify the last wf2 task is done
			wf2Chan <- true
		}

		testLock.Lock()
		doneTasks[s.workflow] = s.order
		testLock.Unlock()
		return nil
	})

	ctx := context.Background()
	// queue 20 tasks for wf1, to fill up wf1 channel (but not overflow it)
	// the 3rd will hang until later
	for i := 1; i < 21; i++ {
		tq.Enqueue(ctx, &task.Task{
			Type: task.TypeCreatePod,
			Metadata: task.Metadata{
				ReName:   "some-rt",
				Workflow: "wf1",
			},
			Spec: dummySpec{
				workflow: "wf1",
				order:    i,
			},
		})
	}

	// and also run 4 more tasks on a different wf
	for i := 1; i < 4; i++ {
		tq.Enqueue(ctx, &task.Task{
			Type: task.TypeCreatePod,
			Metadata: task.Metadata{
				ReName:   "some-rt",
				Workflow: "wf2",
			},
			Spec: dummySpec{
				workflow: "wf2",
				order:    i,
			},
		})
	}

	// wait for wf2 to finish all 3 tasks
	<-wf2Chan
	// signal 3rd task of wf1 to stop hanging
	wf1Chan <- true
	// wait until all task queues are empty
	tq.wg.Wait()
	assert.Equal(t, doneTasks["wf1"], 20, "wf1 was not completed")
	assert.Equal(t, doneTasks["wf2"], 3, "wf2 was not completed")
}
