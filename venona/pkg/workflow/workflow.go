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

package workflow

import (
	"fmt"

	"github.com/codefresh-io/go/venona/pkg/task"
)

type (
	// Workflow is a collection of one or more tasks related to the same runtime/workflowId (to be executed sequensially by the handlers)
	Workflow struct {
		Metadata task.Metadata
		Tasks    task.Tasks
	}
)


// New creates a new empty Workflow instance
func New(metadata task.Metadata) *Workflow {
	return &Workflow{
		Metadata: metadata,
		Tasks:    task.Tasks{},
	}
}

// AddTask adds a specific task to its matching parent worklow
func (wf *Workflow) AddTask(t task.Task) error {
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
func Less(wf1 Workflow, wf2 Workflow) bool {
	return wf1.Metadata.CreatedAt < wf2.Metadata.CreatedAt
}
