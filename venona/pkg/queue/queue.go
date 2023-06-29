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
	"errors"
	"sync"
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
)

type (
	// WorkflowQueue manages a map of workflow (id) -> workflow
	WorkflowQueue struct {
		runtimes    map[string]runtime.Runtime
		log         logger.Logger
		wg          *sync.WaitGroup
		monitor     monitoring.Monitor
		queue       chan *workflow.Workflow
		concurrency int
		stop        []chan bool
	}
)

const defaultWfTaskBufferSize = 20

var (
	errRuntimeNotFound = errors.New("Runtime environment not found")
)

// New creates a new TaskQueue instance
func New(runtimes map[string]runtime.Runtime, log logger.Logger, wg *sync.WaitGroup, monitor monitoring.Monitor, concurrency int) *WorkflowQueue {
	return &WorkflowQueue{
		runtimes:    runtimes,
		log:         log,
		wg:          wg,
		monitor:     monitor,
		queue:       make(chan *workflow.Workflow, 1000),
		concurrency: concurrency,
		stop:        make([]chan bool, concurrency),
	}
}

// Start creates the workflow handlers that will handle the incoming Workflows
func (wfq *WorkflowQueue) Start(ctx context.Context) {
	for i := 0; i < wfq.concurrency; i++ {
		stopChan := make(chan bool, 1)
		wfq.stop[i] = stopChan
		handlerID := i
		wfq.wg.Add(1)
		go wfq.handleChannel(ctx, stopChan, handlerID)
	}
}

// Stop sends a signal to each of the handler to notify it to stop once the queue is empty
func (wfq *WorkflowQueue) Stop() {
	for i := 0; i < wfq.concurrency; i++ {
		wfq.stop[i] <- true
	}
}

// Enqueue adds another task to be handled, internally using or creating a channel for the task's workflow
func (wfq *WorkflowQueue) Enqueue(wf *workflow.Workflow) {
	wfq.queue <- wf
}

func (wfq *WorkflowQueue) handleChannel(ctx context.Context, stopChan chan bool, id int) {
	wfq.log.Info("starting workflow handler", "handlerId", id)
	ctxCancelled := false

	defer wfq.wg.Done()
	for {
		select {
		case <-stopChan:
			wfq.log.Info("stopping workflow handler", "handlerId", id)
			ctxCancelled = true
		case wf := <-wfq.queue:
			wfq.log.Info("handling workflow", "handlerId", id, "workflow", wf.Metadata.Workflow)
			start := time.Now()
			wfq.handleWorkflow(ctx, wf)
			end := time.Now()
			created, err := time.Parse(time.RFC3339, wf.Metadata.CreatedAt)
			if err != nil {
				wfq.log.Error("failed parsing CreatedAt", "handlerId", id, "workflow", wf.Metadata.Workflow, "createdAt", wf.Metadata.CreatedAt)
			}

			wfq.log.Info("Done handling workflow",
				"handlerId", id,
				"workflow", wf.Metadata.Workflow,
				"runtime", wf.Metadata.ReName,
				"time since creation", end.Sub(created),
				"time in runner", end.Sub(wf.Metadata.Pulled),
				"processing time", end.Sub(start),
			)
		default:
			if ctxCancelled {
				wfq.log.Info("stopped workflow handler", "handlerId", id)
				return
			}
		}
	}
}

func (wfq *WorkflowQueue) handleWorkflow(ctx context.Context, wf *workflow.Workflow) {
	txn := task.NewTaskTransaction(wfq.monitor, wf.Metadata)
	defer txn.End()

	workflow := wf.Metadata.Workflow
	reName := wf.Metadata.ReName
	runtime, ok := wfq.runtimes[reName]
	if !ok {
		wfq.log.Error("failed handling task", "error", errRuntimeNotFound, "workflow", workflow)
		txn.NoticeError(errRuntimeNotFound)
		return
	}

	for i := range wf.Tasks {
		err := runtime.HandleTask(ctx, wf.Tasks[i])
		if err != nil {
			wfq.log.Error("failed handling task", "error", err, "workflow", workflow)
			txn.NoticeError(errRuntimeNotFound)
		}
	}
}
