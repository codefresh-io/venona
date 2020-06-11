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

package agent

import (
	"errors"
	"time"

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/runtime"
)

var (
	errAlreadyStarted = errors.New("Already started")
)

type (
	// Agent holds all the references from Codefresh
	// in order to run the process
	Agent struct {
		ID                 string
		Codefresh          codefresh.Codefresh
		Runtimes           map[string]runtime.Runtime
		Logger             logger.Logger
		TaskPullerTicker   *time.Ticker
		ReportStatusTicker *time.Ticker
		running            bool
		lastStatus         Status
	}

	// Status of the agent
	Status struct {
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}

	workflowCandidate struct {
		tasks   []task.Task
		runtime string
	}
)

// Start starting the agent process
func (a Agent) Start() error {
	if a.running {
		return errAlreadyStarted
	}
	a.Logger.Debug("Starting agent")
	a.running = true
	go a.startTaskPullerRoutine()
	go a.startStatusReporterRoutine()
	return nil
}

// Stop stops the agent routines
func (a Agent) Stop() {
	a.running = false
}

// Status returns the last knows status of the agent and related runtimes
func (a Agent) Status() Status {
	return a.lastStatus
}

func (a Agent) startTaskPullerRoutine() {
	for {
		select {
		case <-a.TaskPullerTicker.C:
			if !a.running {
				continue
			}
			go func(client codefresh.Codefresh, runtimes map[string]runtime.Runtime, logger logger.Logger) {
				tasks := pullTasks(client, logger)
				startTasks(tasks, runtimes, logger)
			}(a.Codefresh, a.Runtimes, a.Logger)
		}
	}
}

func (a Agent) startStatusReporterRoutine() {
	for {
		select {
		case <-a.ReportStatusTicker.C:
			if !a.running {
				continue
			}
			go reportStatus(a.Codefresh, codefresh.AgentStatus{
				Message: "All good",
			}, a.Logger)
		}
	}
}

func reportStatus(client codefresh.Codefresh, status codefresh.AgentStatus, logger logger.Logger) {
	err := client.ReportStatus(status)
	if err != nil {
		logger.Error(err.Error())
	}
}

func pullTasks(client codefresh.Codefresh, logger logger.Logger) []task.Task {
	logger.Debug("Requesting tasks from API server")
	tasks, err := client.Tasks()
	if err != nil {
		logger.Error(err.Error())
		return []task.Task{}
	}
	if len(tasks) == 0 {
		logger.Debug("No new tasks received")
		return []task.Task{}
	}
	logger.Debug("Received new tasks", "len", len(tasks))
	return tasks
}

func startTasks(tasks []task.Task, runtimes map[string]runtime.Runtime, logger logger.Logger) {
	creationTasks := []task.Task{}
	deletionTasks := []task.Task{}
	for _, t := range tasks {
		logger.Debug("Starting tasks", "runtime", t.Metadata.ReName)
		if t.Type == task.TypeCreatePod || t.Type == task.TypeCreatePVC {
			creationTasks = append(creationTasks, t)
		}

		if t.Type == task.TypeDeletePod || t.Type == task.TypeDeletePVC {
			deletionTasks = append(deletionTasks, t)
		}
	}

	for _, tasks := range groupTasks(creationTasks) {
		reName := tasks[0].Metadata.ReName
		if err := runtimes[reName].StartWorkflow(tasks); err != nil {
			logger.Error(err.Error())
		}
	}
	for _, tasks := range groupTasks(deletionTasks) {
		reName := tasks[0].Metadata.ReName
		if err := runtimes[reName].TerminateWorkflow(tasks); err != nil {
			logger.Error(err.Error())
		}
	}
}

func groupTasks(tasks []task.Task) map[string][]task.Task {
	candidates := map[string][]task.Task{}
	for _, task := range tasks {
		name := task.Metadata.Workflow
		if name == "" {
			// If for some reason the task is not related to any workflow
			// Might heppen in older versions on Codefresh
			name = "_"
		}
		candidates[name] = append(candidates[name], task)
	}
	return candidates
}
