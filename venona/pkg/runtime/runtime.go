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

package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/task"
)

type (
	// Runtime API client
	Runtime interface {
		StartWorkflow([]task.Task) error
		TerminateWorkflow([]task.Task) []error
	}

	// Options for runtime
	Options struct {
		Kubernetes kubernetes.Kubernetes
	}

	runtime struct {
		client kubernetes.Kubernetes
	}
)

// New creates new Runtime client
func New(opt Options) Runtime {
	return &runtime{
		client: opt.Kubernetes,
	}
}

func (r runtime) StartWorkflow(tasks []task.Task) error {
	for _, task := range tasks {
		err := r.client.CreateResource(task.Spec)
		if err != nil {
			return err // TODO: Return already executed tasks in order to terminate them
		}
	}
	return nil
}
func (r runtime) TerminateWorkflow(tasks []task.Task) []error {
	errs := make([]error, 0, 3)
	for _, task := range tasks {
		opt := kubernetes.DeleteOptions{}
		opt.Kind = task.Type
		b, err := json.Marshal(task.Spec)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to marshal task spec"))
			continue
		}
		if err := json.Unmarshal(b, &opt); err != nil {
			errs = append(errs, fmt.Errorf("failed to unmarshal task spec"))
			continue
		}
		if err = r.client.DeleteResource(opt); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errs
}
