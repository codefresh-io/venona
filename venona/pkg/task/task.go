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
)

// Const for task types
const (
	TypeCreatePod = "CreatePod"
	TypeCreatePVC = "CreatePvc"
	TypeDeletePod = "DeletePod"
	TypeDeletePVC = "DeletePvc"
	TypeAgentTask = "AgentTask"
)

// UnmarshalTasks with json
func UnmarshalTasks(data []byte) ([]Task, error) {
	var r []Task
	err := json.Unmarshal(data, &r)
	return r, err
}

// Marshal tasks
func (r *Tasks) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Tasks array
type Tasks []Task

// Task options
type Task struct {
	Type     string      `json:"type"`
	Spec     interface{} `json:"spec"`
	Metadata Metadata    `json:"metadata"`
}

// Metadata options
type Metadata struct {
	CreatedAt string `json:"createdAt"`
	Account   string `json:"account"`
	ReName    string `json:"reName"`
	Workflow  string `json:"workflow"`
}

// AgentTask describes a task of type "AgentTask"
type AgentTask struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}
