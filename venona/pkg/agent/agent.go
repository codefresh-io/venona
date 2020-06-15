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
	"sync"
	"time"

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
)

var (
	errAlreadyRunning                         = errors.New("Already running")
	errAlreadyStopped                         = errors.New("Already stopped")
	errOptionsRequired                        = errors.New("Options are required")
	errIDRequired                             = errors.New("ID options is required")
	errRuntimesRequired                       = errors.New("Runtimes options is required")
	errLoggerRequired                         = errors.New("Logger options is required")
	errTaskPullingSecondsIntervalRequired     = errors.New("TaskPullingSecondsInterval options is required")
	errStatusReportingSecondsIntervalRequired = errors.New("StatusReportingSecondsInterval options is required")
)

type (
	// Options for creating a new Agent instance
	Options struct {
		ID                             string
		Codefresh                      codefresh.Codefresh
		Runtimes                       map[string]runtime.Runtime
		Logger                         logger.Logger
		TaskPullingSecondsInterval     int64
		StatusReportingSecondsInterval int64
	}

	// Agent holds all the references from Codefresh
	// in order to run the process
	Agent struct {
		id                 string
		cf                 codefresh.Codefresh
		runtimes           map[string]runtime.Runtime
		log                logger.Logger
		taskPullerTicker   *time.Ticker
		reportStatusTicker *time.Ticker
		running            bool
		lastStatus         Status
		stoppedChan        chan struct{}
		wg                 *sync.WaitGroup
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

// New creates a new Agent instance
func New(opt *Options) (Agent, error) {
	a := Agent{}
	if err := checkOptions(opt); err != nil {
		return a, err
	}

	a.id = opt.ID
	a.cf = opt.Codefresh
	a.runtimes = opt.Runtimes
	a.log = opt.Logger
	a.taskPullerTicker = time.NewTicker(time.Duration(opt.TaskPullingSecondsInterval) * time.Second)
	a.reportStatusTicker = time.NewTicker(time.Duration(opt.StatusReportingSecondsInterval) * time.Second)
	a.stoppedChan = make(chan struct{})
	a.wg = &sync.WaitGroup{}

	return a, nil
}

// Start starting the agent process
func (a Agent) Start() error {
	if a.running {
		return errAlreadyRunning
	}
	a.running = true
	a.log.Info("Starting agent")

	go a.startTaskPullerRoutine()
	go a.startStatusReporterRoutine()

	return nil
}

// Stop stops the agents work and blocks until all leftover tasks are finished
func (a Agent) Stop() error {
	if !a.running {
		return errAlreadyStopped
	}
	a.running = false
	a.log.Info("Agent received graceful termination request stopping tasks...")
	a.reportStatusTicker.Stop()
	a.stoppedChan <- struct{}{} // signal stop
	a.taskPullerTicker.Stop()
	a.stoppedChan <- struct{}{} // signal stop
	a.wg.Wait()
	return nil
}

// Status returns the last knows status of the agent and related runtimes
func (a Agent) Status() Status {
	return a.lastStatus
}

func (a Agent) startTaskPullerRoutine() {
	for {
		select {
		case <-a.stoppedChan:
			return
		case <-a.taskPullerTicker.C:
			a.wg.Add(1)
			go func(client codefresh.Codefresh, runtimes map[string]runtime.Runtime, logger logger.Logger) {
				tasks := pullTasks(client, logger)
				startTasks(tasks, runtimes, logger)
				a.wg.Done()
			}(a.cf, a.runtimes, a.log)
		}
	}
}

func (a Agent) startStatusReporterRoutine() {
	for {
		select {
		case <-a.stoppedChan:
			return
		case <-a.reportStatusTicker.C:
			a.wg.Add(1)
			go func(cf codefresh.Codefresh, log logger.Logger) {
				reportStatus(cf, codefresh.AgentStatus{
					Message: "All good",
				}, log)
				a.wg.Done()
			}(a.cf, a.log)
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
	logger.Info("Received new tasks", "len", len(tasks))
	return tasks
}

func startTasks(tasks []task.Task, runtimes map[string]runtime.Runtime, logger logger.Logger) {
	creationTasks := []task.Task{}
	deletionTasks := []task.Task{}
	for _, t := range tasks {
		logger.Info("Executing tasks", "runtime", t.Metadata.ReName)
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

func checkOptions(opt *Options) error {
	if opt == nil {
		return errOptionsRequired
	}

	if opt.ID == "" {
		return errIDRequired
	}

	if opt.Runtimes == nil {
		return errRuntimesRequired
	}

	if opt.Logger == nil {
		return errLoggerRequired
	}

	if opt.TaskPullingSecondsInterval == 0 {
		return errTaskPullingSecondsIntervalRequired
	}

	if opt.StatusReportingSecondsInterval == 0 {
		return errStatusReportingSecondsIntervalRequired
	}

	return nil
}
