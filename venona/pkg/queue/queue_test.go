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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeWorkflow(wfID string, numOfTasks int) *workflow.Workflow {
	metadata := task.Metadata{
		Workflow: wfID,
		ReName:   "some-rt",
	}
	wf := workflow.New(metadata)
	for i := 0; i < numOfTasks; i++ {
		_ = wf.AddTask(&task.Task{
			Type:     task.TypeCreatePod,
			Metadata: metadata,
			Spec:     fmt.Sprintf("%s-%d", wfID, i),
		})
	}

	return wf
}

func TestWorkflowQueue_Enqueue(t *testing.T) {
	type wfOrSleep struct {
		wf    *workflow.Workflow
		sleep time.Duration
	}
	tests := map[string]struct {
		workflows   []wfOrSleep
		concurrency int
		want        []string
		afterFn     func(t *testing.T, createdPods []string)
	}{
		"should create a single workflow with a single task": {
			workflows: []wfOrSleep{
				{wf: makeWorkflow("wf1", 1)},
			},
			concurrency: 1,
			want:        []string{"wf1-0"},
		},
		"should create a single workflow with several tasks": {
			workflows: []wfOrSleep{
				{wf: makeWorkflow("wf1", 3)},
			},
			concurrency: 1,
			want:        []string{"wf1-0", "wf1-1", "wf1-2"},
		},
		"should create multiple workflows with concurrency 1": {
			workflows: []wfOrSleep{
				{wf: makeWorkflow("wf1", 3)},
				{wf: makeWorkflow("wf2", 3)},
				{wf: makeWorkflow("wf3", 3)},
			},
			concurrency: 1,
			want:        []string{"wf1-0", "wf1-1", "wf1-2", "wf2-0", "wf2-1", "wf2-2", "wf3-0", "wf3-1", "wf3-2"},
		},
		"should create multiple workflows with higher concurrency": {
			workflows: []wfOrSleep{
				{wf: makeWorkflow("wf1", 2)},
				{wf: makeWorkflow("wf2", 2)},
				{wf: makeWorkflow("wf3", 2)},
				{wf: makeWorkflow("wf4", 2)},
				{wf: makeWorkflow("wf5", 2)},
				{wf: makeWorkflow("wf6", 2)},
				{sleep: 100},
				{wf: makeWorkflow("wf7", 2)},
				{wf: makeWorkflow("wf8", 2)},
				{wf: makeWorkflow("wf9", 2)},
			},
			concurrency: 3,
			want: []string{
				"wf1-0", "wf1-1", "wf2-0", "wf2-1", "wf3-0", "wf3-1",
				"wf4-0", "wf4-1", "wf5-0", "wf5-1", "wf6-0", "wf6-1",
				"wf7-0", "wf7-1", "wf8-0", "wf8-1", "wf9-0", "wf9-1",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			createdPods := []string{}
			testLock := sync.Mutex{}
			mockKubernetes := kubernetes.NewMockKubernetes(t)
			mockKubernetes.EXPECT().CreateResource(mock.Anything, task.TypeCreatePod, mock.AnythingOfType("string")).RunAndReturn(func(_ context.Context, _ task.Type, spec interface{}) error {
				s, _ := spec.(string)
				testLock.Lock()
				createdPods = append(createdPods, s)
				testLock.Unlock()
				return nil
			})
			runtimes := map[string]runtime.Runtime{
				"some-rt": runtime.New(runtime.Options{
					Kubernetes: mockKubernetes,
				}),
			}
			log := logger.New(logger.Options{})
			wg := &sync.WaitGroup{}
			opts := &Options{
				Runtimes:    runtimes,
				Log:         log,
				WG:          wg,
				Monitor:     monitoring.NewEmpty(),
				Concurrency: tt.concurrency,
				BufferSize:  100,
			}
			tq := New(opts)
			tq.Start(context.Background())
			for _, tOrS := range tt.workflows {
				if tOrS.wf != nil {
					tq.Enqueue(tOrS.wf)
				} else {
					time.Sleep(tOrS.sleep)
				}
			}

			tq.Stop()
			wg.Wait()
			assert.ElementsMatch(t, createdPods, tt.want)
		})
	}
}
