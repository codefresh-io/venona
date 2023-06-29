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
	"fmt"

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

	Workflow struct {
		Metadata Metadata
		Tasks    Tasks
	}

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
func TaskLess(task1 Task, task2 Task) bool {
	return task1.Metadata.CreatedAt < task2.Metadata.CreatedAt
}

// NewTaskTransaction creates a new transaction with task-specific attributes
func NewTaskTransaction(monitor monitoring.Monitor, m Metadata) monitoring.Transaction {
	txn := monitor.NewTransaction("runner-tasks-execution")
	txn.AddAttribute("tid", m.Workflow)
	txn.AddAttribute("runtime-environment", m.ReName)
	return txn
}

func NewWorkflow(t Task) *Workflow {
	return &Workflow{
		Metadata: t.Metadata,
		Tasks: Tasks{
			t,
		},
	}
}

func (wf *Workflow) AddTask(t Task) error {
	if wf.Metadata.ReName != t.Metadata.ReName || wf.Metadata.Workflow != t.Metadata.Workflow {
		return fmt.Errorf("mismatch runtime or workflow id, %s/%s is different from %s/%s", wf.Metadata.ReName, wf.Metadata.Workflow, t.Metadata.ReName, t.Metadata.Workflow)
	}

	if wf.Metadata.CreatedAt > t.Metadata.CreatedAt {
		wf.Metadata = t.Metadata
	}

	wf.Tasks = append(wf.Tasks, t)
	return nil
}

// Less compares two workflows by their CreatedAt values
func WorkflowLess(task1 Workflow, task2 Workflow) bool {
	return task1.Metadata.CreatedAt < task2.Metadata.CreatedAt
}