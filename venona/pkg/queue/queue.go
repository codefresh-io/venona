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
		queue:       make(chan *workflow.Workflow),
		concurrency: concurrency,
		stop:        make([]chan bool, concurrency),
	}
}

func (wfq *WorkflowQueue) Start(ctx context.Context) {
	for i := 0; i < wfq.concurrency; i++ {
		stopChan := make(chan bool, 1)
		wfq.stop[i] = stopChan
		handlerId := i
		wfq.wg.Add(1)
		go wfq.handleChannel(ctx, stopChan, handlerId)
	}
}

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
			wfq.handleWorkflow(ctx, wf)
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

	for _, t := range wf.Tasks {
		err := runtime.HandleTask(ctx, &t)
		if err != nil {
			wfq.log.Error("failed handling task", "error", err, "workflow", workflow)
			txn.NoticeError(errRuntimeNotFound)
		}
	}
}
