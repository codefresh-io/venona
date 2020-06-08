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
		started            bool
	}

	workflowCandidate struct {
		tasks   []codefresh.Task
		runtime string
	}
)

// Start starting the agent process
func (a Agent) Start() error {
	if a.started {
		return errAlreadyStarted
	}
	a.Logger.Debug("Starting agent")
	a.started = true
	go a.fetchTasks()
	go a.reportStatus()
	time.Sleep(30 * time.Second)
	return nil
}

func (a Agent) fetchTasks() {
	for {
		select {
		case <-a.TaskPullerTicker.C:
			a.Logger.Debug("Requesting tasks from API server")
			tasks, err := a.Codefresh.Tasks()
			if err != nil {
				a.Logger.Error(err.Error())
				continue
			}
			a.Logger.Debug("Received new tasks", "len", len(tasks))
			creationTasks := []codefresh.Task{}
			deletionTasks := []codefresh.Task{}
			for _, t := range tasks {
				a.Logger.Debug("Starting tasks", "runtime", t.Metadata.ReName)
				if t.Type == codefresh.TypeCreatePod || t.Type == codefresh.TypeCreatePVC {
					creationTasks = append(creationTasks, t)
				}

				if t.Type == codefresh.TypeDeletePod || t.Type == codefresh.TypeDeletePVC {
					deletionTasks = append(deletionTasks, t)
				}
			}

			for _, tasks := range groupTasks(creationTasks) {
				reName := tasks[0].Metadata.ReName
				if err := a.Runtimes[reName].StartWorkflow(tasks); err != nil {
					a.Logger.Error(err.Error())
				}
			}
			for _, tasks := range groupTasks(deletionTasks) {
				reName := tasks[0].Metadata.ReName
				if err := a.Runtimes[reName].TerminateWorkflow(tasks); err != nil {
					a.Logger.Error(err.Error())
				}
			}

		}
	}
}

func (a Agent) reportStatus() {
	for {
		select {
		case <-a.ReportStatusTicker.C:
			err := a.Codefresh.ReportStatus(codefresh.AgentStatus{
				Message: "All good",
			})
			if err != nil {
				a.Logger.Error(err.Error())
				continue
			}
		}
	}
}

func groupTasks(tasks []codefresh.Task) map[string][]codefresh.Task {
	candidates := map[string][]codefresh.Task{}
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
