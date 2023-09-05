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
	"github.com/codefresh-io/go/venona/pkg/queue"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/objx"
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
		Concurrency                    int
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
		wfQueue            *queue.WorkflowQueue
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

const (
	defaultProxyRequestTimeout = time.Second * 30
	defaultProxyRequestRetries = 3
)

var (
	// internal errors
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
	taskPullerTicker := time.NewTicker(opts.TaskPullingSecondsInterval)
	reportStatusTicker := time.NewTicker(opts.StatusReportingSecondsInterval)
	wg := &sync.WaitGroup{}

	if opts.Monitor == nil {
		opts.Monitor = monitoring.NewEmpty()
	}

	httpClient.HTTPClient.Transport = opts.Monitor.NewRoundTripper(httpClient.HTTPClient.Transport)
	return &Agent{
		id:                 id,
		cf:                 cf,
		log:                log,
		taskPullerTicker:   taskPullerTicker,
		reportStatusTicker: reportStatusTicker,
		wfQueue:            queue.New(runtimes, log, wg, opts.Monitor, opts.Concurrency),
		running:            false,
		lastStatus:         Status{},
		wg:                 wg,
		monitor:            opts.Monitor,
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
	a.wfQueue.Start(ctx)

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
	a.wfQueue.Stop()
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
			agentTasks, workflows := a.getTasks(ctx)

			// perform all agentTasks (in goroutine)
			for i := range agentTasks {
				a.handleAgentTask(&agentTasks[i])
			}

			// send all wfTasks to tasksQueue
			for i := range workflows {
				a.wfQueue.Enqueue(workflows[i])
			}
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

func (a *Agent) getTasks(ctx context.Context) (task.Tasks, []*workflow.Workflow) {
	tasks := a.pullTasks(ctx)
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

func tasksToIds(tasks task.Tasks) []string {
	keys := make(map[string]bool)
	res := []string{}
	for _, t := range tasks {
		workflow := t.Metadata.Workflow
		if _, ok := keys[workflow]; !ok {
			res = append(res, workflow)
			keys[workflow] = true
		}
	}

	return res
}

func (a *Agent) splitTasks(tasks task.Tasks) (task.Tasks, []*workflow.Workflow) {
	agentTasks := task.Tasks{}
	wfMap := map[string]*workflow.Workflow{}

	// divide tasks by types
	for i := range tasks {
		t := tasks[i]
		switch t.Type {
		case task.TypeAgentTask:
			agentTasks = append(agentTasks, t)
		case task.TypeCreatePod, task.TypeCreatePVC, task.TypeDeletePod, task.TypeDeletePVC:
			wf, ok := wfMap[t.Metadata.Workflow]
			if !ok {
				wf = workflow.New(t.Metadata)
				wfMap[t.Metadata.Workflow] = wf
			}

			err := wf.AddTask(&t)
			if err != nil {
				a.log.Error("failed adding task to workflow", "error", err)
			}
		default:
			a.log.Error("unrecognized task type", "type", t.Type, "tid", t.Metadata.Workflow, "runtime", t.Metadata.ReName)
		}
	}

	// sort agentTasks by creationDate
	sort.SliceStable(agentTasks, func(i, j int) bool {
		task1, task2 := agentTasks[i], tasks[j]
		return task.Less(task1, task2)
	})

	workflows := []*workflow.Workflow{}
	ids := []string{}
	for id, wf := range wfMap {
		workflows = append(workflows, wf)
		ids = append(ids, id)
	}

	a.log.Debug("received workflows", "ids", ids)

	// sort workflows by creationDate
	sort.SliceStable(workflows, func(i, j int) bool {
		wf1, wf2 := workflows[i], workflows[j]
		return workflow.Less(*wf1, *wf2)
	})
	return agentTasks, workflows
}

func (a *Agent) handleAgentTask(t *task.Task) {
	a.log.Info("executing agent task", "tid", t.Metadata.Workflow)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		txn := task.NewTaskTransaction(a.monitor, t.Metadata)
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

func init() {
	httpClient.RetryMax = defaultProxyRequestRetries
	httpClient.HTTPClient.Timeout = defaultProxyRequestTimeout
}
