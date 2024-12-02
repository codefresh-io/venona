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

package agent

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_groupTasks(t *testing.T) {
	tests := map[string]struct {
		tasks task.Tasks
		want  map[string]task.Tasks
	}{
		"should group by workflow name": {
			tasks: task.Tasks{
				{
					Metadata: task.Metadata{
						WorkflowId: "1",
					},
				},
				{
					Metadata: task.Metadata{
						WorkflowId: "2",
					},
				},
				{
					Metadata: task.Metadata{
						WorkflowId: "1",
					},
				},
			},
			want: map[string]task.Tasks{
				"1": {
					{
						Metadata: task.Metadata{
							WorkflowId: "1",
						},
					},
					{
						Metadata: task.Metadata{
							WorkflowId: "1",
						},
					},
				},
				"2": {
					{
						Metadata: task.Metadata{
							WorkflowId: "2",
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			groupedTasks := groupTasks(tt.tasks)
			assert.Equal(t, tt.want, groupedTasks)
		})
	}
}

func Test_reportStatus(t *testing.T) {
	tests := map[string]struct {
		status   codefresh.AgentStatus
		beforeFn func(cf *codefresh.MockCodefresh, log *logger.MockLogger)
	}{
		"should report status": {
			status: codefresh.AgentStatus{
				Message: "OK",
			},
			beforeFn: func(cf *codefresh.MockCodefresh, _ *logger.MockLogger) {
				cf.EXPECT().ReportStatus(mock.Anything, codefresh.AgentStatus{
					Message: "OK",
				}).Return(nil)
			},
		},
		"should log error": {
			status: codefresh.AgentStatus{
				Message: "OK",
			},
			beforeFn: func(cf *codefresh.MockCodefresh, log *logger.MockLogger) {
				cf.EXPECT().ReportStatus(mock.Anything, codefresh.AgentStatus{
					Message: "OK",
				}).Return(errors.New("some error"))
				log.EXPECT().Error("Failed reporting status", "error", errors.New("some error"))
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cf := codefresh.NewMockCodefresh(t)
			log := logger.NewMockLogger(t)
			tt.beforeFn(cf, log)
			a := &Agent{
				cf:  cf,
				log: log,
			}
			a.reportStatus(context.Background(), tt.status)
		})
	}
}

func TestNew(t *testing.T) {
	tests := map[string]struct {
		opts    *Options
		want    *Agent
		wantErr string
	}{
		"should throw error if options is nil": {
			opts:    nil,
			want:    nil,
			wantErr: errOptionsRequired.Error(),
		},
		"should throw error if ID is not provided": {
			opts: &Options{
				ID:        "",
				Codefresh: &codefresh.MockCodefresh{},
				Runtimes: map[string]runtime.Runtime{
					"x": runtime.New(runtime.Options{}),
				},
				Logger: logger.New(logger.Options{}),
			},
			want:    nil,
			wantErr: errIDRequired.Error(),
		},
		"should throw error if runtimes are not provided": {
			opts: &Options{
				ID:        "foobar",
				Codefresh: &codefresh.MockCodefresh{},
				Runtimes:  nil,
				Logger:    logger.New(logger.Options{}),
			},
			want:    nil,
			wantErr: errRuntimesRequired.Error(),
		},
		"should throw error if logger is nil": {
			opts: &Options{
				ID:        "foobar",
				Codefresh: &codefresh.MockCodefresh{},
				Runtimes: map[string]runtime.Runtime{
					"x": runtime.New(runtime.Options{}),
				},
				Logger: nil,
			},
			want:    nil,
			wantErr: errLoggerRequired.Error(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.opts)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_executeAgentTask(t *testing.T) {
	executorCalled := false
	tests := map[string]struct {
		executorName string
		executorFunc func(*task.AgentTask, logger.Logger) error
		task         *task.Task
		wantErr      string
	}{
		"should successfully run executor and return nil": {
			executorName: "test",
			executorFunc: func(_ *task.AgentTask, _ logger.Logger) error {
				executorCalled = true
				return nil
			},
			task: &task.Task{
				Type:     task.TypeAgentTask,
				Metadata: task.Metadata{},
				Spec: task.AgentTask{
					Type:   "test",
					Params: nil,
				},
			},
		},
		"should call an executor and return an error": {
			executorName: "test",
			executorFunc: func(_ *task.AgentTask, _ logger.Logger) error {
				executorCalled = true
				return errProxyTaskWithoutURL
			},
			task: &task.Task{
				Type:     task.TypeAgentTask,
				Metadata: task.Metadata{},
				Spec: task.AgentTask{
					Type:   "test",
					Params: nil,
				},
			},
			wantErr: errProxyTaskWithoutURL.Error(),
		},
		"should pass the agent task spec to the executor": {
			executorName: "test",
			executorFunc: func(t *task.AgentTask, _ logger.Logger) error {
				executorCalled = true
				data, ok := t.Params["data"].(float64)
				if !ok {
					return fmt.Errorf("expected data to be of type int")
				}

				if data != 3 {
					return fmt.Errorf("expected data to equal 3 but data=%v", data)
				}

				return nil
			},
			task: &task.Task{
				Type:     task.TypeAgentTask,
				Metadata: task.Metadata{},
				Spec: task.AgentTask{
					Type: "test",
					Params: map[string]interface{}{
						"data": 3,
					},
				},
			},
		},
	}

	for name, tt := range tests {
		executorCalled = false
		agentTaskExecutors[tt.executorName] = tt.executorFunc
		t.Run(name, func(t *testing.T) {
			a := &Agent{
				log: logger.New(logger.Options{}),
			}
			err := a.executeAgentTask(context.Background(), tt.task)
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			}

			if !executorCalled {
				t.Errorf("executor function hasn't been called")
			}
		})
		delete(agentTaskExecutors, tt.executorName)
	}
}

func Test_splitTasks(t *testing.T) {
	tests := map[string]struct {
		tasks task.Tasks
		want  []*task.Task
	}{
		"should split tasks to correct order": {
			tasks: task.Tasks{
				{
					Type: task.TypeDeletePod,
				},
				{
					Type: task.TypeCreatePod,
				},
				{
					Type: task.TypeCreatePVC,
				},
			},
			want: []*task.Task{
				{
					Type: task.TypeCreatePVC,
				},
				{
					Type: task.TypeCreatePod,
				},
				{
					Type: task.TypeDeletePod,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := &Agent{
				cf:  &codefresh.MockCodefresh{},
				log: logger.New(logger.Options{}),
			}
			_, workflows := a.splitTasks(tt.tasks)
			assert.Equal(t, tt.want, workflows[0].Tasks)
		})
	}
}
