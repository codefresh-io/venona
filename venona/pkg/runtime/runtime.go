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
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/kubernetes"
)

type (
	// Runtime API client
	Runtime interface {
		StartWorkflow([]codefresh.Task) error
		TerminateWorkflow([]codefresh.Task) error
	}

	// Options for runtime 
	Options struct {
		Kubernetes kubernetes.Kubernetes
	}

	runtime struct{}
)

// New creates new Runtime client
func New(opt Options) Runtime {
	return &runtime{}
}

func (r runtime) StartWorkflow(tasks []codefresh.Task) error {
	return nil
}
func (r runtime) TerminateWorkflow(tasks []codefresh.Task) error {
	return nil
}
