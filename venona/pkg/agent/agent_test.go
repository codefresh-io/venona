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
	type args struct {
		tasks []task.Task
	}
	tests := []struct {
		name string
		args args
		want map[string][]task.Task
	}{
		{
			name: "should group by workflow name",
			args: args{
				tasks: []task.Task{
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
			},
			want: map[string][]task.Task{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupedTasks := groupTasks(tt.args.tasks)
			assert.Equal(t, tt.want, groupedTasks)
		})
	}
}

func Test_reportStatus(t *testing.T) {
	type args struct {
		client codefresh.Codefresh
		status codefresh.AgentStatus
		logger logger.Logger
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should report status",
			args: args{
				client: getCodefreshMock(),
				logger: getLoggerMock(),
				status: codefresh.AgentStatus{
					Message: "OK",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportStatus(tt.args.client, tt.args.status, tt.args.logger)
		})
	}
}

func getCodefreshMock() *mocks.Codefresh {
	cf := mocks.Codefresh{}

	cf.On("ReportStatus", mock.Anything).Return(fmt.Errorf("bad"))

	return &cf
}

func getLoggerMock() *mocks.Logger {
	l := mocks.Logger{}

	l.On("Error", mock.Anything)

	return &l
}

func TestNew(t *testing.T) {
	runtimes := make(map[string]runtime.Runtime)
	runtimes["x"] = runtime.New(runtime.Options{})
	type args struct {
		opt *Options
	}
	tests := []struct {
		name    string
		args    args
		want    *Agent
		wantErr error
	}{
		{
			"should throw error if options is nil",
			args{
				nil,
			},
			nil,
			errOptionsRequired,
		},
		{
			"should throw error if ID is not provided",
			args{
				&Options{
					ID:        "",
					Codefresh: &mocks.Codefresh{},
					Runtimes:  runtimes,
					Logger:    &mocks.Logger{},
				},
			},
			nil,
			errIDRequired,
		},
		{
			"should throw error if runtimes is not provided",
			args{
				&Options{
					ID:        "foobar",
					Codefresh: &mocks.Codefresh{},
					Runtimes:  nil,
					Logger:    &mocks.Logger{},
				},
			},
			nil,
			errRuntimesRequired,
		},
		{
			"should throw error if runtimes is empty",
			args{
				&Options{
					ID:        "foobar",
					Codefresh: &mocks.Codefresh{},
					Runtimes:  make(map[string]runtime.Runtime),
					Logger:    &mocks.Logger{},
				},
			},
			nil,
			errRuntimesRequired,
		},
		{
			"should throw error if logger is nil",
			args{
				&Options{
					ID:        "foobar",
					Codefresh: &mocks.Codefresh{},
					Runtimes:  runtimes,
					Logger:    nil,
				},
			},
			nil,
			errLoggerRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opt)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error \"%v\" but got no error", tt.wantErr)
				} else if err != tt.wantErr {
					t.Errorf("expected error \"%v\" but got error \"%v\"", tt.wantErr, err)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createMockAgent() *Agent {
	runtimes := make(map[string]runtime.Runtime)
	runtimes["x"] = runtime.New(runtime.Options{})
	a, _ := New(&Options{
		ID:        "foobar",
		Codefresh: &mocks.Codefresh{},
		Logger:    &mocks.Logger{},
		Runtimes:  runtimes,
	})

	return a
}
