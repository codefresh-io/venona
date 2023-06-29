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
)

type (
	// TaskQueue manages a map of workflow (id) -> tasks
	TaskQueue struct {
		runtimes map[string]runtime.Runtime
		log      logger.Logger
		wg       *sync.WaitGroup
		mutex    sync.Mutex
		monitor  monitoring.Monitor
		tasks    map[string]chan *task.Task
	}
)

const defaultWfTaskBufferSize = 20

var (
	errRuntimeNotFound = errors.New("Runtime environment not found")
)

// New creates a new TaskQueue instance
func New(runtimes map[string]runtime.Runtime, log logger.Logger, wg *sync.WaitGroup, monitor monitoring.Monitor) *TaskQueue {
	return &TaskQueue{
		runtimes: runtimes,
		log:      log,
		wg:       wg,
		monitor:  monitor,
		tasks:    make(map[string]chan *task.Task),
	}
}

// Enqueue adds another task to be handled, internally using or creating a channel for the task's workflow
func (tq *TaskQueue) Enqueue(ctx context.Context, t *task.Task) {
	workflow := t.Metadata.Workflow
	tq.mutex.Lock()
	c, ok := tq.tasks[workflow]
	tq.mutex.Unlock()
	if !ok {
		tq.log.Info("Creating new queue", "workflow", workflow)
		c = make(chan *task.Task, defaultWfTaskBufferSize)
		tq.tasks[workflow] = c
		tq.wg.Add(1)
		go tq.handleChannel(ctx, c, workflow)
	}

	select {
	case c <- t:
		// sent task to queue
	case <-time.After(5 * time.Second):
		tq.log.Error("Send operation timed out", "workflow", workflow)
	}
}

func (tq *TaskQueue) handleChannel(ctx context.Context, c chan *task.Task, workflow string) {
	var txn monitoring.Transaction

	defer tq.wg.Done()
	for {
		select {
		case <-ctx.Done():
			tq.log.Info("stopping wf task handler routine", "workflow", workflow)
			return
		case t := <-c:
			if txn == nil {
				txn = task.NewTaskTransaction(tq.monitor, t)
				defer txn.End()
			}

			err := tq.handleTask(ctx, t)
			if err != nil {
				tq.log.Error("failed handling task", "error", err, "workflow", workflow)
				txn.NoticeError(err)
			}
		default:
			if txn == nil {
				// if there is no transaction yet, it means we haven't handled any tasks yet
				continue
			}

			tq.mutex.Lock()
			{
				// making sure the channel is still empty after the lock
				if len(c) == 0 {
					tq.log.Info("workflow tasks channel empty, stopping task handler", "workflow", workflow)
					delete(tq.tasks, workflow)
					close(c)
					tq.mutex.Unlock()
					return
				}
			}
			tq.mutex.Unlock()
		}
	}
}

func (tq *TaskQueue) handleTask(ctx context.Context, t *task.Task) error {
	reName := t.Metadata.ReName
	runtime, ok := tq.runtimes[reName]
	if !ok {
		return errRuntimeNotFound
	}

	return runtime.HandleTask(ctx, t)
}
