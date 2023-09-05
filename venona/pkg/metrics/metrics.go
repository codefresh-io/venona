// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	// Metrics interface to expose duration of various operations
	Metrics interface {
		UpdateQueueSize(agentTasks, wfTasks, queue int)
		ObserveGetTasks(start time.Time)
		ObserveAgentTaskMetrics(t *task.Task, agentType string)
		ObserveWorkflowMetrics(wf *workflow.Workflow)
		ObserveK8sMetrics(taskType task.Type, namespace, name string, start time.Time)
	}

	metrics struct {
		log                       logger.Logger
		agentTasks                prometheus.Counter
		wfTasks                   prometheus.Counter
		queueSize                 prometheus.Counter
		getTasksDuration          prometheus.Histogram
		handlingTimeSinceCreation *prometheus.HistogramVec
		handlingTimeInRunner      *prometheus.HistogramVec
		wfProcessingTime          *prometheus.HistogramVec
		agentProcessingTime       *prometheus.HistogramVec
		k8sProcessingTime         *prometheus.HistogramVec
	}
)

// New creates a new Metrics instnace
func New(reg *prometheus.Registry, log logger.Logger) Metrics {
	m := &metrics{
		log: log,
		agentTasks: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "runner_agent_tasks",
			Help: "Incoming agent tasks",
		}),
		wfTasks: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "runner_wf_tasks",
			Help: "Incoming workflow tasks",
		}),
		queueSize: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "runner_queue_size",
			Help: "Current number of waiting tasks",
		}),
		getTasksDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "runner_get_tasks_duration_sec",
			Help:    "How long each GetTasks request takes (seconds)",
			Buckets: []float64{0.25, 0.5, 1, 2, 3, 6},
		}),
		handlingTimeSinceCreation: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "runner_wf_duration_since_creation_sec",
			Help:    "Time since the creation of each workflow batch in the platform",
			Buckets: []float64{0.5, 1, 1.5, 2, 3, 6, 12, 30, 60},
		}, []string{"workflow_type"}),
		handlingTimeInRunner: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "runner_wf_duration_in_runner_sec",
			Help:    "Time each workflow batch has spent in the runner",
			Buckets: []float64{0.5, 1, 1.5, 2, 3, 6, 12, 30, 60},
		}, []string{"workflow_type"}),
		agentProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "runner_agent_processing_sec",
			Help:    "Net time to process an agent task in the runner",
			Buckets: []float64{0.25, 0.5, 1, 1.5, 2, 3, 6},
		}, []string{"agent_type"}),
		wfProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "runner_wf_processing_sec",
			Help:    "Net time to process each workflow batch in the runner",
			Buckets: []float64{0.5, 1, 1.5, 2, 3, 6, 12},
		}, []string{"workflow_type"}),
		k8sProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "runner_k8s_processing_sec",
			Help:    "Net time to process each workflow k8s task in the runner",
			Buckets: []float64{0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 3, 6, 12},
		}, []string{"k8s_type"}),
	}
	reg.MustRegister(m.agentTasks)
	reg.MustRegister(m.wfTasks)
	reg.MustRegister(m.queueSize)
	reg.MustRegister(m.getTasksDuration)
	reg.MustRegister(m.handlingTimeSinceCreation)
	reg.MustRegister(m.handlingTimeInRunner)
	reg.MustRegister(m.agentProcessingTime)
	reg.MustRegister(m.wfProcessingTime)
	reg.MustRegister(m.k8sProcessingTime)
	return m
}

func (m *metrics) UpdateQueueSize(agentTasks, wfTasks, queue int) {
	m.agentTasks.Add(float64(agentTasks))
	m.wfTasks.Add(float64(wfTasks))
	m.queueSize.Add(float64(queue))
	if agentTasks > 0 || wfTasks > 0 || queue > 0 {
		m.log.Info("done pulling tasks",
			"agentTasks", agentTasks,
			"workflows", wfTasks,
			"queueSize", queue,
		)
	}
}

func (m *metrics) ObserveGetTasks(start time.Time) {
	end := time.Now()
	diff := end.Sub(start)
	m.getTasksDuration.Observe(diff.Seconds())
}

func (m *metrics) ObserveAgentTaskMetrics(t *task.Task, agentType string) {
	end := time.Now()
	created, _ := time.Parse(time.RFC3339, t.Metadata.CreatedAt)
	sinceCreation := end.Sub(created)
	inRunner := end.Sub(t.Timeline.Pulled)
	processed := end.Sub(t.Timeline.Started)

	m.log.Info("Done handling agent task",
		"tid", t.Metadata.Workflow,
		"time since creation", sinceCreation,
		"time in runner", inRunner,
		"processing time", processed,
	)
	labels := prometheus.Labels{"agent_type": agentType}
	m.handlingTimeSinceCreation.With(labels).Observe(sinceCreation.Seconds())
	m.handlingTimeInRunner.With(labels).Observe(inRunner.Seconds())
	m.agentProcessingTime.With(labels).Observe(processed.Seconds())
}

func (m *metrics) ObserveWorkflowMetrics(wf *workflow.Workflow) {
	end := time.Now()
	created, _ := time.Parse(time.RFC3339, wf.Metadata.CreatedAt)
	sinceCreation := end.Sub(created)
	inRunner := end.Sub(wf.Timeline.Pulled)
	processed := end.Sub(wf.Timeline.Started)

	m.log.Info("Done handling workflow",
		"workflow", wf.Metadata.Workflow,
		"runtime", wf.Metadata.ReName,
		"time since creation", sinceCreation,
		"time in runner", inRunner,
		"processing time", processed,
	)
	labels := prometheus.Labels{"workflow_type": string(wf.Type)}
	m.handlingTimeSinceCreation.With(labels).Observe(sinceCreation.Seconds())
	m.handlingTimeInRunner.With(labels).Observe(inRunner.Seconds())
	m.wfProcessingTime.With(labels).Observe(processed.Seconds())
}

func (m *metrics) ObserveK8sMetrics(taskType task.Type, namespace, name string, start time.Time) {
	end := time.Now()
	processed := end.Sub(start)
	m.log.Info("Done handling k8s task",
		"type", taskType,
		"namespace", namespace,
		"name", name,
		"processing time", processed,
	)
	labels := prometheus.Labels{"k8s_type": string(taskType)}
	m.k8sProcessingTime.With(labels).Observe(processed.Seconds())
}
