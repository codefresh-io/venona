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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/objx"
)

// internal errors
var (
	errAlreadyRunning           = errors.New("Agent already running")
	errAlreadyStopped           = errors.New("Agent already stopped")
	errOptionsRequired          = errors.New("Options are required")
	errIDRequired               = errors.New("ID options is required")
	errRuntimesRequired         = errors.New("Runtimes options is required")
	errLoggerRequired           = errors.New("Logger options is required")
	errRuntimeNotFound          = errors.New("Runtime environment not found")
	errFailedToParseAgentTask   = errors.New("Failed to parse agent task spec")
	errUknownAgentTaskType      = errors.New("Agent task has unknown type")
	errAgentTaskMalformedParams = errors.New("failed to marshal agent task params")
	errProxyTaskWithoutURL      = errors.New(`url not provided for task of type "proxy"`)
	errProxyTaskWithoutToken    = errors.New(`token not provided for task of type "proxy"`)
)

const (
	defaultTaskPullingInterval     = time.Second * 3
	defaultStatusReportingInterval = time.Second * 10
	defaultProxyRequestTimeout     = time.Second * 30
	defaultProxyRequestRetries     = 3
	defaultWfTaskBufferSize        = 10
)

type (
	// Options for creating a new Agent instance
	Options struct {
		ID                             string
		Codefresh                      codefresh.Codefresh
		Runtimes                       map[string]runtime.Runtime
		Logger                         logger.Logger
		TaskPullingSecondsInterval     time.Duration
		StatusReportingSecondsInterval time.Duration
		WfTaskBufferSize               int
		Monitor                        monitoring.Monitor
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
		wfTasksChannel     chan task.Tasks
		running            bool
		lastStatus         Status
		wg                 *sync.WaitGroup
		monitor            monitoring.Monitor
	}

	// Status of the agent
	Status struct {
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
	}
)

var (
	httpClient = retryablehttp.NewClient()

	agentTaskExecutors = map[string]func(t *task.AgentTask, log logger.Logger) error{
		"proxy": proxyRequest,
	}
)

// New creates a new Agent instance
func New(opts *Options) (*Agent, error) {
	if err := checkOptions(opts); err != nil {
		return nil, err
	}

	id := opts.ID
	cf := opts.Codefresh
	runtimes := opts.Runtimes
	log := opts.Logger
	taskPullingInterval := defaultTaskPullingInterval
	if opts.TaskPullingSecondsInterval != time.Duration(0) {
		taskPullingInterval = opts.TaskPullingSecondsInterval
	}

	statusReportingInterval := defaultStatusReportingInterval
	if opts.StatusReportingSecondsInterval != time.Duration(0) {
		statusReportingInterval = opts.StatusReportingSecondsInterval
	}

	wfTaskBufferSize := defaultWfTaskBufferSize
	if opts.WfTaskBufferSize != 0 {
		wfTaskBufferSize = opts.WfTaskBufferSize
	}

	taskPullerTicker := time.NewTicker(taskPullingInterval)
	reportStatusTicker := time.NewTicker(statusReportingInterval)
	wfTasksChannel := make(chan task.Tasks, wfTaskBufferSize)
	wg := &sync.WaitGroup{}

	if opts.Monitor == nil {
		opts.Monitor = monitoring.NewEmpty()
	}

	httpClient.HTTPClient.Transport = opts.Monitor.NewRoundTripper(httpClient.HTTPClient.Transport)
	return &Agent{
		id,
		cf,
		runtimes,
		log,
		taskPullerTicker,
		reportStatusTicker,
		wfTasksChannel,
		false,
		Status{},
		wg,
		opts.Monitor,
	}, nil
}

// Start starting the agent process
func (a *Agent) Start(ctx context.Context) error {
	if a.running {
		return errAlreadyRunning
	}

	a.running = true
	a.log.Info("Starting agent")

	// only 1 for the wfTaskHandlerRoutine, the other 2 don't need to be waited on
	a.wg.Add(1)
	go a.startTaskPullerRoutine(ctx)
	go a.startWfTaskHandlerRoutine(ctx)
	go a.startStatusReporterRoutine(ctx)

	a.reportStatus(ctx, codefresh.AgentStatus{
		Message: "All good",
	})

	return nil
}

// Stop stops the agents work and blocks until all leftover tasks are finished
func (a *Agent) Stop() error {
	if !a.running {
		return errAlreadyStopped
	}

	a.running = false
	a.log.Warn("Received graceful termination request, stopping tasks...")
	a.reportStatusTicker.Stop()
	a.taskPullerTicker.Stop()
	close(a.wfTasksChannel)
	a.log.Warn("stopped both tickers")
	a.wg.Wait()
	return nil
}

// Status returns the last knows status of the agent and related runtimes
func (a *Agent) Status() Status {
	return a.lastStatus
}

func (a *Agent) startTaskPullerRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping task puller routine")
			return
		case <-a.taskPullerTicker.C:
			agentTasks, wfTasks := a.getTasks(ctx)

			// perform all agentTasks (in goroutine)
			for i := range agentTasks {
				a.handleAgentTask(&agentTasks[i])
			}

			// send all wfTasks to tasksChannel
			wfGroups := groupTasks(wfTasks)
			for _, wfTask := range wfGroups {
				a.wfTasksChannel <- wfTask
			}
		}
	}
}

func (a *Agent) startWfTaskHandlerRoutine(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping wf task handler routine")
			return
		case tasks, ok := <-a.wfTasksChannel:
			if !ok {
				a.log.Info("Wofkrlow tasks channel closed, stopping task handler")
				return
			}

			a.handleTasks(ctx, tasks)
		}
	}
}

func (a *Agent) handleTasks(ctx context.Context, tasks task.Tasks) {
	reName := tasks[0].Metadata.ReName
	runtime, ok := a.runtimes[reName]
	workflow := tasks[0].Metadata.Workflow
	txn := newTransaction(a.monitor, "workflow-tasks", workflow, reName)
	defer txn.End()
	if !ok {
		a.log.Error("Runtime not found", "runtime", reName, "workflow", workflow)
		txn.NoticeError(errRuntimeNotFound)
		return
	}

	a.log.Info("Handling workflow tasks", "runtime", reName, "workflow", workflow)
	for _, t := range tasks {
		if err := runtime.HandleTask(ctx, t); err != nil {
			a.log.Error(err.Error())
			txn.NoticeError(err)
		}
	}
}

func (a *Agent) startStatusReporterRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping status reporter routine")
			return
		case <-a.reportStatusTicker.C:
			a.wg.Add(1)
			go func() {
				defer a.wg.Done()
				a.reportStatus(ctx, codefresh.AgentStatus{
					Message: "All good",
				})
			}()
		}
	}
}

func (a *Agent) reportStatus(ctx context.Context, status codefresh.AgentStatus) {
	err := a.cf.ReportStatus(ctx, status)
	if err != nil {
		a.log.Error(err.Error())
	}
}

func (a *Agent) getTasks(ctx context.Context) (task.Tasks, task.Tasks) {
	tasks := a.pullTasks(ctx)

	// sort tasks by creationDate
	sortTasks(tasks)
	return a.splitTasks(tasks)
}

func (a *Agent) pullTasks(ctx context.Context) task.Tasks {
	a.log.Debug("Requesting tasks from API server")
	tasks, err := a.cf.Tasks(ctx)
	if err != nil {
		a.log.Error(err.Error())
		return task.Tasks{}
	}

	if len(tasks) == 0 {
		a.log.Debug("No new tasks received")
		return task.Tasks{}
	}

	a.log.Info("Received new tasks", "len", len(tasks))
	return tasks
}

func sortTasks(tasks task.Tasks) {
	sort.SliceStable(tasks, func(i, j int) bool {
		task1, task2 := tasks[i], tasks[j]
		return task.Less(task1, task2)
	})
}

func (a *Agent) splitTasks(tasks task.Tasks) (task.Tasks, task.Tasks) {
	agentTasks := task.Tasks{}
	wfTasks := task.Tasks{}

	// divide tasks by types
	for _, t := range tasks {
		a.log.Debug("Received task", "type", t.Type, "tid", t.Metadata.Workflow, "runtime", t.Metadata.ReName)
		switch t.Type {
		case task.TypeAgentTask:
			agentTasks = append(agentTasks, t)
		case task.TypeCreatePod, task.TypeCreatePVC, task.TypeDeletePod, task.TypeDeletePVC:
			wfTasks = append(wfTasks, t)
		default:
			a.log.Error("unrecognized task type", "type", t.Type, "tid", t.Metadata.Workflow, "runtime", t.Metadata.ReName)
		}
	}

	return agentTasks, wfTasks
}

func (a *Agent) handleAgentTask(t *task.Task) {
	a.log.Info("executing agent task", "tid", t.Metadata.Workflow)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		txn := newTransaction(a.monitor, t.Type, t.Metadata.Workflow, t.Metadata.ReName)
		defer txn.End()
		if err := executeAgentTask(t, a.log); err != nil {
			a.log.Error(err.Error())
			txn.NoticeError(err)
		}

		a.log.Info("finished agent task", "tid", t.Metadata.Workflow)
	}()
}

func executeAgentTask(t *task.Task, log logger.Logger) error {
	specJSON, err := json.Marshal(t.Spec)
	if err != nil {
		return errFailedToParseAgentTask
	}

	spec := task.AgentTask{}
	if err = json.Unmarshal(specJSON, &spec); err != nil {
		return errFailedToParseAgentTask
	}

	e, ok := agentTaskExecutors[spec.Type]
	if !ok {
		return errUknownAgentTaskType
	}

	return e(&spec, log)
}

func proxyRequest(t *task.AgentTask, log logger.Logger) error {
	spec := objx.Map(t.Params)
	vars := objx.Map(spec.Get("runtimeContext.context.variables").MSI())
	token := spec.Get("runtimeContext.context.eventReporting.token").Str()
	if token == "" {
		return errProxyTaskWithoutToken
	}

	url := vars.Get("proxyUrl").Str()
	if url == "" {
		return errProxyTaskWithoutURL
	}

	method := vars.Get("method").Str("POST")

	json, err := json.Marshal(t.Params)
	if err != nil {
		return errAgentTaskMalformedParams
	}

	if json == nil {
		json = []byte{}
	}

	req, err := retryablehttp.NewRequest(method, url, bytes.NewReader(json))
	if err != nil {
		return fmt.Errorf("failed creating new request: %w", err)
	}

	req.Header.Add("x-req-type", "workflow-request")
	req.Header.Add("x-access-token", token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", fmt.Sprintf("%v", len(json)))

	log.Info("executing proxy task", "url", url, "method", method)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending request: %w", err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Info("finished proxy task", "url", url, "method", method, "status", resp.Status, "body", string(body))

	return nil
}

func groupTasks(tasks task.Tasks) map[string]task.Tasks {
	candidates := map[string]task.Tasks{}
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

func checkOptions(opts *Options) error {
	if opts == nil {
		return errOptionsRequired
	}

	if opts.ID == "" {
		return errIDRequired
	}

	if len(opts.Runtimes) == 0 {
		return errRuntimesRequired
	}

	if opts.Logger == nil {
		return errLoggerRequired
	}

	return nil
}

func newTransaction(monitor monitoring.Monitor, taskType task.Type, tid, runtime string) monitoring.Transaction {
	txn := monitor.NewTransaction("runner-tasks-execution")
	txn.AddAttribute("task-type", taskType)
	txn.AddAttribute("tid", tid)
	txn.AddAttribute("runtime-environment", runtime)
	return txn
}

func init() {
	httpClient.RetryMax = defaultProxyRequestRetries
	httpClient.HTTPClient.Timeout = defaultProxyRequestTimeout
}
