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

package task

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/codefresh-io/go/venona/pkg/monitoring"
)

// Const for task types
const (
	TypeCreatePod Type = "CreatePod"
	TypeCreatePVC Type = "CreatePvc"
	TypeDeletePod Type = "DeletePod"
	TypeDeletePVC Type = "DeletePvc"
	TypeAgentTask Type = "AgentTask"
)

const (
	StatusSuccess Status = "Success"
	StatusError   Status = "Error"
)

type (
	// Tasks array
	Tasks []Task

	// Type of the task
	Type string

	// Task Id
	Id string

	// Task status
	Status string

	// Task options
	Task struct {
		Id       Id          `json:"_id"`
		Type     Type        `json:"type"`
		Metadata Metadata    `json:"metadata"`
		Spec     interface{} `json:"spec"`

		// only used in AgentTasks
		Timeline Timeline
	}

	// Metadata options
	Metadata struct {
		CreatedAt             string `json:"createdAt"`
		ReName                string `json:"reName"`
		WorkflowId            string `json:"workflowId"`
		CurrentStatusRevision int    `json:"currentStatusRevision"`
		ShouldReportStatus    bool   `json:"shouldReportStatus"`
	}

	// Timeline values
	Timeline struct {
		Pulled  time.Time
		Started time.Time
	}

	// AgentTask describes a task of type "AgentTask"
	AgentTask struct {
		Type   string                 `json:"type"`
		Params map[string]interface{} `json:"params"`
	}

	TaskStatus struct {
		Status         Status    `json:"status"`
		OccurredAt     time.Time `json:"occurredAt"`
		StatusRevision int       `json:"statusRevision"`
		IsRetriable    bool      `json:"isRetriable"`
		Reason         string    `json:"reason,omitempty"`
	}
)

// UnmarshalTasks with json
func UnmarshalTasks(data []byte) (Tasks, error) {
	var r Tasks
	err := json.Unmarshal(data, &r)
	return r, err
}

// Marshal tasks
func (r *Tasks) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *TaskStatus) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Less compares two tasks by their CreatedAt values
func Less(task1 Task, task2 Task) bool {
	return task1.Metadata.CreatedAt < task2.Metadata.CreatedAt
}

// NewTaskTransaction creates a new transaction with task-specific attributes
func NewTaskTransaction(monitor monitoring.Monitor, m Metadata) monitoring.Transaction {
	txn := monitor.NewTransaction("runner-tasks-execution")
	txn.AddAttribute("tid", m.WorkflowId)
	txn.AddAttribute("runtime-environment", m.ReName)
	return txn
}

// SortByType sorts the tasks in the specified order: TypeCreatePVC, TypeCreatePod, TypeDeletePod, TypeDeletePVC
func SortByType(tasks []*Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		order := map[Type]int{
			TypeCreatePVC: 1,
			TypeCreatePod: 2,
			TypeDeletePod: 3,
			TypeDeletePVC: 4,
		}
		return order[tasks[i].Type] < order[tasks[j].Type]
	})
}

func (t *Task) GetLatency() (sinceCreation, inRunner, processed time.Duration) {
	end := time.Now()
	created, _ := time.Parse(time.RFC3339, t.Metadata.CreatedAt)
	sinceCreation = end.Sub(created)
	inRunner, processed = t.Timeline.GetLatency(end)
	return
}

func (t *Timeline) GetLatency(end time.Time) (inRunner, processed time.Duration) {
	inRunner = end.Sub(t.Pulled)
	processed = end.Sub(t.Started)
	return
}
