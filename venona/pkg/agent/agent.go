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

	workflowCandidate struct {
		tasks   task.Tasks
		runtime string
	}
)

var (
	httpClient = retryablehttp.NewClient()

	agentTaskExecutors = map[string]func(t *task.AgentTask, log logger.Logger) error{
		"proxy": proxyRequest,
	}
)

// New creates a new Agent instance
func New(opt *Options) (*Agent, error) {
	if err := checkOptions(opt); err != nil {
		return nil, err
	}
	id := opt.ID
	cf := opt.Codefresh
	runtimes := opt.Runtimes
	log := opt.Logger
	taskPullingInterval := defaultTaskPullingInterval
	if opt.TaskPullingSecondsInterval != time.Duration(0) {
		taskPullingInterval = opt.TaskPullingSecondsInterval
	}

	statusReportingInterval := defaultStatusReportingInterval
	if opt.StatusReportingSecondsInterval != time.Duration(0) {
		statusReportingInterval = opt.StatusReportingSecondsInterval
	}

	taskPullerTicker := time.NewTicker(taskPullingInterval)
	reportStatusTicker := time.NewTicker(statusReportingInterval)
	wg := &sync.WaitGroup{}

	if opt.Monitor == nil {
		opt.Monitor = monitoring.NewEmpty()
	}

	httpClient.HTTPClient.Transport = opt.Monitor.NewRoundTripper(httpClient.HTTPClient.Transport)
	return &Agent{
		id,
		cf,
		runtimes,
		log,
		taskPullerTicker,
		reportStatusTicker,
		false,
		Status{},
		wg,
		opt.Monitor,
	}, nil
}

// Start starting the agent process
func (a *Agent) Start(ctx context.Context) error {
	if a.running {
		return errAlreadyRunning
	}

	a.running = true
	a.log.Info("Starting agent")

	go a.startTaskPullerRoutine(ctx)
	go a.startStatusReporterRoutine(ctx)

	reportStatus(ctx, a.cf, codefresh.AgentStatus{
		Message: "All good",
	}, a.log)

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
			return
		case <-a.taskPullerTicker.C:
			a.wg.Add(1)
			go func(client codefresh.Codefresh, runtimes map[string]runtime.Runtime, wg *sync.WaitGroup, log logger.Logger, monitor monitoring.Monitor) {
				defer wg.Done()
				tasks := pullTasks(ctx, client, log)
				sortTasks(tasks)
				startTasks(ctx, tasks, runtimes, wg, log, monitor)
			}(a.cf, a.runtimes, a.wg, a.log, a.monitor)
		}
	}
}

func (a *Agent) startStatusReporterRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.reportStatusTicker.C:
			a.wg.Add(1)
			go func(cf codefresh.Codefresh, wg *sync.WaitGroup, log logger.Logger) {
				defer wg.Done()
				reportStatus(ctx, cf, codefresh.AgentStatus{
					Message: "All good",
				}, log)
			}(a.cf, a.wg, a.log)
		}
	}
}

func reportStatus(ctx context.Context, client codefresh.Codefresh, status codefresh.AgentStatus, log logger.Logger) {
	err := client.ReportStatus(ctx, status)
	if err != nil {
		log.Error(err.Error())
	}
}

func pullTasks(ctx context.Context, client codefresh.Codefresh, log logger.Logger) task.Tasks {
	log.Debug("Requesting tasks from API server")
	tasks, err := client.Tasks(ctx)
	if err != nil {
		log.Error(err.Error())
		return task.Tasks{}
	}

	if len(tasks) == 0 {
		log.Debug("No new tasks received")
		return task.Tasks{}
	}

	log.Info("Received new tasks", "len", len(tasks))
	return tasks
}

func sortTasks(tasks task.Tasks) {
	sort.SliceStable(tasks, func(i, j int) bool {
		task1, task2 := tasks[i], tasks[j]
		return task.Less(task1, task2)
	})
}

func startTasks(ctx context.Context, tasks task.Tasks, runtimes map[string]runtime.Runtime, wg *sync.WaitGroup, log logger.Logger, monitor monitoring.Monitor) {
	creationTasks := task.Tasks{}
	deletionTasks := task.Tasks{}
	agentTasks := task.Tasks{}

	// divide tasks by types
	for _, t := range tasks {
		log.Debug("Received task", "type", t.Type, "tid", t.Metadata.Workflow, "runtime", t.Metadata.ReName)
		switch t.Type {
		case task.TypeCreatePod, task.TypeCreatePVC:
			creationTasks = append(creationTasks, t)
		case task.TypeDeletePod, task.TypeDeletePVC:
			deletionTasks = append(deletionTasks, t)
		case task.TypeAgentTask:
			agentTasks = append(agentTasks, t)
		default:
			log.Error("unrecognized task type", "type", t.Type, "tid", t.Metadata.Workflow, "runtime", t.Metadata.ReName)
		}
	}

	if len(creationTasks) > 0 || len(deletionTasks) > 0 || len(agentTasks) > 0 {
		log.Info("starting tasks", "creation", len(creationTasks), "deletion", len(deletionTasks), "agent", len(agentTasks))
	}

	// process agent tasks
	for i := range agentTasks {
		t := agentTasks[i]
		log.Info("executing agent task", "tid", t.Metadata.Workflow)
		txn := newTransaction(monitor, t.Type, t.Metadata.Workflow, t.Metadata.ReName)
		wg.Add(1)
		go func(t task.Task, wg *sync.WaitGroup, log logger.Logger) {
			defer wg.Done()
			if err := executeAgentTask(&t, log); err != nil {
				log.Error(err.Error())
				txn.NoticeError(err)
			}

			txn.End()
			log.Info("finished agent task", "tid", t.Metadata.Workflow)
		}(t, wg, log)
	}

	// process creation tasks
	grouped := groupTasks(creationTasks)
	for _, tasks := range grouped {
		reName := tasks[0].Metadata.ReName
		runtime, ok := runtimes[reName]
		txn := newTransaction(monitor, "start-workflow", tasks[0].Metadata.Workflow, reName)

		if !ok {
			log.Error("Runtime not found", "workflow", tasks[0].Metadata.Workflow, "runtime", reName)
			txn.NoticeError(errRuntimeNotFound)
			txn.End()
			continue
		}

		log.Info("Starting workflow", "workflow", tasks[0].Metadata.Workflow, "runtime", reName, "# tasks", len(tasks))
		if err := runtime.StartWorkflow(ctx, tasks); err != nil {
			log.Error(err.Error())
			txn.NoticeError(err)
		}

		time.Sleep(time.Second * 10)
		txn.End()
	}

	// process deletion tasks
	grouped = groupTasks(deletionTasks)
	for _, tasks := range grouped {
		reName := tasks[0].Metadata.ReName
		runtime, ok := runtimes[reName]
		txn := newTransaction(monitor, "terminate-workflow", tasks[0].Metadata.Workflow, reName)

		if !ok {
			log.Error("Runtime not found", "workflow", tasks[0].Metadata.Workflow, "runtime", reName)
			txn.NoticeError(errRuntimeNotFound)
			txn.End()
			continue
		}

		log.Info("Terminating workflow", "workflow", tasks[0].Metadata.Workflow, "runtime", reName, "# tasks", len(tasks))
		if errs := runtime.TerminateWorkflow(ctx, tasks); len(errs) != 0 {
			for _, err := range errs {
				log.Error(err.Error())
				txn.NoticeError(err)
			}
		}

		txn.End()
	}
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

func checkOptions(opt *Options) error {
	if opt == nil {
		return errOptionsRequired
	}

	if opt.ID == "" {
		return errIDRequired
	}

	if opt.Runtimes == nil || len(opt.Runtimes) == 0 {
		return errRuntimesRequired
	}

	if opt.Logger == nil {
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
