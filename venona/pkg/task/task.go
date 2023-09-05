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

type (
	// Tasks array
	Tasks []Task

	// Type of the task
	Type string

	// Task options
	Task struct {
		Type     Type        `json:"type"`
		Metadata Metadata    `json:"metadata"`
		Spec     interface{} `json:"spec"`
	}

	// Metadata options
	Metadata struct {
		CreatedAt string `json:"createdAt"`
		ReName    string `json:"reName"`
		Workflow  string `json:"workflow"`
	}

	// AgentTask describes a task of type "AgentTask"
	AgentTask struct {
		Type   string                 `json:"type"`
		Params map[string]interface{} `json:"params"`
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

// Less compares two tasks by their CreatedAt values
func Less(task1 Task, task2 Task) bool {
	return task1.Metadata.CreatedAt < task2.Metadata.CreatedAt
}

// NewTaskTransaction creates a new transaction with task-specific attributes
func NewTaskTransaction(monitor monitoring.Monitor, t *Task) monitoring.Transaction {
	txn := monitor.NewTransaction("runner-tasks-execution")
	txn.AddAttribute("task-type", t.Type)
	txn.AddAttribute("tid", t.Metadata.Workflow)
	txn.AddAttribute("runtime-environment", t.Metadata.ReName)
	return txn
}
