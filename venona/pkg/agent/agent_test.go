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
	"github.com/codefresh-io/go/venona/pkg/mocks"
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
						Workflow: "1",
					},
				},
				{
					Metadata: task.Metadata{
						Workflow: "2",
					},
				},
				{
					Metadata: task.Metadata{
						Workflow: "1",
					},
				},
			},
			want: map[string]task.Tasks{
				"1": {
					{
						Metadata: task.Metadata{
							Workflow: "1",
						},
					},
					{
						Metadata: task.Metadata{
							Workflow: "1",
						},
					},
				},
				"2": {
					{
						Metadata: task.Metadata{
							Workflow: "2",
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
		beforeFn func(cf *codefresh.MockCodefresh, log *mocks.Logger)
	}{
		"should report status": {
			status: codefresh.AgentStatus{
				Message: "OK",
			},
			beforeFn: func(cf *codefresh.MockCodefresh, _ *mocks.Logger) {
				cf.On("ReportStatus", mock.Anything, codefresh.AgentStatus{
					Message: "OK",
				}).Return(nil)
			},
		},
		"should log error": {
			status: codefresh.AgentStatus{
				Message: "OK",
			},
			beforeFn: func(cf *codefresh.MockCodefresh, log *mocks.Logger) {
				cf.On("ReportStatus", mock.Anything, codefresh.AgentStatus{
					Message: "OK",
				}).Return(errors.New("some error"))
				log.On("Error", "some error")
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(_ *testing.T) {
			cf := &codefresh.MockCodefresh{}
			log := &mocks.Logger{}
			tt.beforeFn(cf, log)
			a := &Agent{
				cf:  cf,
				log: log,
			}
			a.reportStatus(context.Background(), tt.status)
		})
	}
}

func getCodefreshMock() codefresh.Codefresh {
	cf := codefresh.MockCodefresh{}

	// cf.On("ReportStatus", mock.Anything, mock.Anything).Return(fmt.Errorf("bad"))

	return &cf
}

func getLoggerMock() *mocks.Logger {
	l := mocks.Logger{}

	l.On("Error", mock.Anything)

	return &l
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
				Codefresh: getCodefreshMock(),
				Runtimes: map[string]runtime.Runtime{
					"x": runtime.New(runtime.Options{}),
				},
				Logger: &mocks.Logger{},
			},
			want:    nil,
			wantErr: errIDRequired.Error(),
		},
		"should throw error if runtimes are not provided": {
			opts: &Options{
				ID:        "foobar",
				Codefresh: getCodefreshMock(),
				Runtimes:  nil,
				Logger:    &mocks.Logger{},
			},
			want:    nil,
			wantErr: errRuntimesRequired.Error(),
		},
		"should throw error if logger is nil": {
			opts: &Options{
				ID:        "foobar",
				Codefresh: getCodefreshMock(),
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
	okExecutor := func(_ *task.AgentTask, _ logger.Logger) error {
		executorCalled = true
		return nil
	}

	badExecutor := func(_ *task.AgentTask, _ logger.Logger) error {
		executorCalled = true
		return errProxyTaskWithoutURL
	}

	type args struct {
		executorName string
		executorFunc func(*task.AgentTask, logger.Logger) error
		task         *task.Task
	}

	tests := []struct {
		name    string
		args    *args
		wantErr error
	}{
		{
			name: "should successfully run executor and return nil",
			args: &args{
				executorName: "test",
				executorFunc: okExecutor,
				task: &task.Task{
					Type:     task.TypeAgentTask,
					Metadata: task.Metadata{},
					Spec: task.AgentTask{
						Type:   "test",
						Params: nil,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should call an executor and return an error",
			args: &args{
				executorName: "test",
				executorFunc: badExecutor,
				task: &task.Task{
					Type:     task.TypeAgentTask,
					Metadata: task.Metadata{},
					Spec: task.AgentTask{
						Type:   "test",
						Params: nil,
					},
				},
			},
			wantErr: errProxyTaskWithoutURL,
		},
		{
			name: "should pass the agent task spec to the executor",
			args: &args{
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
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		executorCalled = false
		agentTaskExecutors[tt.args.executorName] = tt.args.executorFunc
		t.Run(tt.name, func(t *testing.T) {
			ret := executeAgentTask(tt.args.task, getLoggerMock())
			if !executorCalled {
				t.Errorf("executor function hasn't been called")
			}
			if ret == nil && tt.wantErr != nil {
				t.Errorf("expected error %v but got nil", tt.wantErr)
			}
			if ret != nil && tt.wantErr == nil {
				t.Errorf("expected nil but got an error: %v", ret)
			}
			if ret != nil && ret.Error() != tt.wantErr.Error() {
				t.Errorf("expected error: %v but got error: %v", tt.wantErr.Error(), ret.Error())
			}

		})
		delete(agentTaskExecutors, tt.args.executorName)
	}
}
