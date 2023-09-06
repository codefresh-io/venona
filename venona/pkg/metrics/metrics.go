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
	"regexp"
	"time"

	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	runnerNamespace = "runner"
	agentSubsystem  = "agent"
	wfSubsystem     = "wf"
)

var (
	engineRegex = regexp.MustCompile(`engine-.*$`)
	retryRegex = regexp.MustCompile(`engine-.*-retry-(\d+)$`)

	agentTasks = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: runnerNamespace,
		Subsystem: agentSubsystem,
		Name:      "tasks",
		Help:      "Incoming agent tasks",
	})
	wfTasks = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: runnerNamespace,
		Subsystem: wfSubsystem,
		Name:      "tasks",
		Help:      "Incoming workflow tasks",
	})
	wfTaskRetries = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: runnerNamespace,
		Subsystem: wfSubsystem,
		Name:      "tasks_retries",
		Help:      "Incoming workflow retry tasks",
	}, []string{"retry"})
	queueSize = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: runnerNamespace,
		Name:      "queue_size",
		Help:      "Current number of waiting tasks",
	})
	getTasksDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Name:      "get_tasks_duration_sec",
		Help:      "How long each GetTasks request takes (seconds)",
		Buckets:   []float64{0.25, 0.5, 1, 2, 3, 6},
	})
	handlingTimeSinceCreation = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Subsystem: wfSubsystem,
		Name:      "duration_since_creation_sec",
		Help:      "Time since the creation of each workflow batch in the platform",
		Buckets:   []float64{0.5, 1, 1.5, 2, 3, 6, 12, 30, 60},
	}, []string{"workflow_type"})
	handlingTimeInRunner = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Subsystem: wfSubsystem,
		Name:      "duration_in_runner_sec",
		Help:      "Time each workflow batch has spent in the runner",
		Buckets:   []float64{0.5, 1, 1.5, 2, 3, 6, 12, 30, 60},
	}, []string{"workflow_type"})
	agentProcessingTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Subsystem: agentSubsystem,
		Name:      "processing_sec",
		Help:      "Net time to process an agent task in the runner",
		Buckets:   []float64{0.25, 0.5, 1, 1.5, 2, 3, 6},
	}, []string{"agent_type"})
	wfProcessingTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Subsystem: wfSubsystem,
		Name:      "processing_sec",
		Help:      "Net time to process each workflow batch in the runner",
		Buckets:   []float64{0.5, 1, 1.5, 2, 3, 6, 12},
	}, []string{"workflow_type"})
	k8sProcessingTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: runnerNamespace,
		Name:      "k8s_processing_sec",
		Help:      "Net time to process each workflow k8s task in the runner",
		Buckets:   []float64{0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 3, 6, 12},
	}, []string{"k8s_type"})
)

// New creates a new Metrics instnace
func Register(reg *prometheus.Registry) {
	reg.MustRegister([]prometheus.Collector{
		agentTasks,
		wfTasks,
		wfTaskRetries,
		queueSize,
		getTasksDuration,
		handlingTimeSinceCreation,
		handlingTimeInRunner,
		agentProcessingTime,
		wfProcessingTime,
		k8sProcessingTime,
	}...)
}

func UpdateQueueSizes(agentTasksValue, wfTasksValue, queue int) {
	agentTasks.Add(float64(agentTasksValue))
	wfTasks.Add(float64(wfTasksValue))
	queueSize.Add(float64(queue))
}

func IncWorkflowRetries(podName string) {
	matches := engineRegex.FindStringSubmatch(podName)
	if matches == nil {
		return
	}

	matches = retryRegex.FindStringSubmatch(podName)
	retry := "0"
	if len(matches) == 2 {
		retry = matches[1]
	}

	labels := prometheus.Labels{"retry": retry}
	wfTaskRetries.With(labels).Inc()
}

func ObserveGetTasks(start time.Time) {
	end := time.Now()
	diff := end.Sub(start)
	getTasksDuration.Observe(diff.Seconds())
}

func ObserveAgentTaskMetrics(agentType string, sinceCreation, inRunner, processed time.Duration) {
	labels := prometheus.Labels{"workflow_type": agentType}
	handlingTimeSinceCreation.With(labels).Observe(sinceCreation.Seconds())
	handlingTimeInRunner.With(labels).Observe(inRunner.Seconds())
	agentProcessingTime.With(prometheus.Labels{"agent_type": agentType}).Observe(processed.Seconds())
}

func ObserveWorkflowMetrics(wfType workflow.Type, sinceCreation, inRunner, processed time.Duration) {
	labels := prometheus.Labels{"workflow_type": string(wfType)}
	handlingTimeSinceCreation.With(labels).Observe(sinceCreation.Seconds())
	handlingTimeInRunner.With(labels).Observe(inRunner.Seconds())
	wfProcessingTime.With(labels).Observe(processed.Seconds())
}

func ObserveK8sMetrics(taskType task.Type, processed time.Duration) {
	labels := prometheus.Labels{"k8s_type": string(taskType)}
	k8sProcessingTime.With(labels).Observe(processed.Seconds())
}
