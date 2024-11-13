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

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/metrics"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
)

type (
	// WorkflowQueue handles incoming workflow batch tasks
	WorkflowQueue interface {
		Start(ctx context.Context)
		Stop()
		Size() int
		Enqueue(wf *workflow.Workflow)
	}

	// Options to create a new WorkflowQueue
	Options struct {
		Runtimes    map[string]runtime.Runtime
		Log         logger.Logger
		WG          *sync.WaitGroup
		Monitor     monitoring.Monitor
		Concurrency int
		BufferSize  int
		Codefresh   codefresh.Codefresh
	}

	wfQueueImpl struct {
		runtimes        map[string]runtime.Runtime
		log             logger.Logger
		wg              *sync.WaitGroup
		monitor         monitoring.Monitor
		queue           chan *workflow.Workflow
		concurrency     int
		stop            []chan bool
		activeWorkflows map[string]struct{}
		mutex           sync.Mutex
		cf              codefresh.Codefresh
	}
)

var errRuntimeNotFound = errors.New("Runtime environment not found")

// New creates a new TaskQueue instance
func New(opts *Options) WorkflowQueue {
	return &wfQueueImpl{
		runtimes:        opts.Runtimes,
		log:             opts.Log,
		wg:              opts.WG,
		monitor:         opts.Monitor,
		queue:           make(chan *workflow.Workflow, opts.BufferSize),
		concurrency:     opts.Concurrency,
		stop:            make([]chan bool, opts.Concurrency),
		activeWorkflows: make(map[string]struct{}),
		cf:              opts.Codefresh,
	}
}

// Start creates the workflow handlers that will handle the incoming Workflows
func (wfq *wfQueueImpl) Start(ctx context.Context) {
	wfq.log.Info("starting workflow queue", "concurrency", wfq.concurrency)
	for i := 0; i < wfq.concurrency; i++ {
		stopChan := make(chan bool, 1)
		wfq.stop[i] = stopChan
		handlerID := i
		wfq.wg.Add(1)
		go wfq.handleChannel(ctx, stopChan, handlerID)
	}
}

// Stop sends a signal to each of the handler to notify it to stop once the queue is empty
func (wfq *wfQueueImpl) Stop() {
	for i := 0; i < wfq.concurrency; i++ {
		wfq.stop[i] <- true
	}
}

// Size returns the current size of the queue (used for logs)
func (wfq *wfQueueImpl) Size() int {
	return len(wfq.queue)
}

// Enqueue adds another task to be handled, internally using or creating a channel for the task's workflow
func (wfq *wfQueueImpl) Enqueue(wf *workflow.Workflow) {
	wfq.queue <- wf
}

func (wfq *wfQueueImpl) handleChannel(ctx context.Context, stopChan chan bool, id int) {
	ctxCancelled := false

	defer wfq.wg.Done()
	for {
		select {
		case <-stopChan:
			wfq.log.Info("stopping workflow handler", "handlerId", id)
			ctxCancelled = true
		case wf := <-wfq.queue:
			wfq.mutex.Lock()
			if _, ok := wfq.activeWorkflows[wf.Metadata.WorkflowId]; ok {
				// Workflow is already being handled, enqueue it again and skip processing
				wfq.mutex.Unlock()
				wfq.log.Info("Workflow", wf.Metadata.WorkflowId, " is already being handled, enqueue it again and skip processing")
				time.Sleep(100 * time.Millisecond)
				wfq.Enqueue(wf)
				continue
			}
			// Mark the workflow as active
			wfq.activeWorkflows[wf.Metadata.WorkflowId] = struct{}{}
			wfq.mutex.Unlock()

			wfq.log.Info("handling workflow", "handlerId", id, "workflow", wf.Metadata.WorkflowId)
			wfq.handleWorkflow(ctx, wf)
			wfq.mutex.Lock()
			delete(wfq.activeWorkflows, wf.Metadata.WorkflowId)
			wfq.mutex.Unlock()
		default:
			if ctxCancelled {
				wfq.log.Info("stopped workflow handler", "handlerId", id)
				return
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (wfq *wfQueueImpl) handleWorkflow(ctx context.Context, wf *workflow.Workflow) {
	wf.Timeline.Started = time.Now()
	txn := task.NewTaskTransaction(wfq.monitor, wf.Metadata)
	defer txn.End()

	workflow := wf.Metadata.WorkflowId
	reName := wf.Metadata.ReName
	runtime, ok := wfq.runtimes[reName]
	if !ok {
		wfq.log.Error("failed handling task", "error", errRuntimeNotFound, "workflow", workflow)
		txn.NoticeError(errRuntimeNotFound)
		return
	}

	for i := range wf.Tasks {
		taskDef := wf.Tasks[i]
		err := runtime.HandleTask(ctx, taskDef)
		status := task.TaskStatus{
			OccurredAt:     time.Now(),
			StatusRevision: taskDef.Metadata.CurrentStatusRevision + 1,
		}
		if err != nil {
			wfq.log.Error("failed handling task", "error", err, "workflow", workflow)
			txn.NoticeError(errRuntimeNotFound)
			status.Status = task.StatusError
			status.Reason = err.Error()
			status.IsRetriable = true // TODO: make this configurable depending on the error
		} else {
			status.Status = task.StatusSuccess
		}
		statusErr := wfq.cf.ReportTaskStatus(ctx, taskDef.Id, status)
		if statusErr != nil {
			wfq.log.Error("failed reporting task status", "error", statusErr, "task", taskDef.Id, "workflow", workflow)
			txn.NoticeError(statusErr)
		}
	}

	sinceCreation, inRunner, processed := wf.GetLatency()
	wfq.log.Info("Done handling workflow",
		"workflow", wf.Metadata.WorkflowId,
		"runtime", wf.Metadata.ReName,
		"time since creation", sinceCreation,
		"time in runner", inRunner,
		"processing time", processed,
	)
	metrics.ObserveWorkflowMetrics(wf.Type, sinceCreation, inRunner, processed)
}
