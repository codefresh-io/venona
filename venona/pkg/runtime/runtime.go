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
	"context"
	"encoding/json"
	"fmt"

	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/task"
)

type (
	// Runtime API client
	Runtime interface {
		HandleTask(ctx context.Context, t *task.Task) error
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
func New(opts Options) Runtime {
	return &runtime{
		client: opts.Kubernetes,
	}
}

func (r runtime) HandleTask(ctx context.Context, t *task.Task) error {
	var err error

	switch t.Type {
	case task.TypeCreatePVC, task.TypeCreatePod:
		err = r.client.CreateResource(ctx, t.Spec)
		if err != nil {
			return fmt.Errorf("failed creating resource: %w", err) // TODO: Return already executed tasks in order to terminate them
		}
	case task.TypeDeletePVC, task.TypeDeletePod:
		opts := kubernetes.DeleteOptions{}
		opts.Kind = t.Type
		b, err := json.Marshal(t.Spec)
		if err != nil {
			return fmt.Errorf("failed to marshal task spec: %w", err)
		}

		if err := json.Unmarshal(b, &opts); err != nil {
			return fmt.Errorf("failed to unmarshal task spec: %w", err)
		}

		if err = r.client.DeleteResource(ctx, opts); err != nil {
			return fmt.Errorf("failed deleting resource: %w", err)
		}
	default:
		return fmt.Errorf("unknown task type \"%s\"", t.Type)
	}

	return nil
}
