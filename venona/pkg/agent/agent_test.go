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
	"testing"
	"time"

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

func TestAgent_Stop(t *testing.T) {
	type fields struct {
		ID                 string
		Codefresh          codefresh.Codefresh
		Runtimes           map[string]runtime.Runtime
		Logger             logger.Logger
		TaskPullerTicker   *time.Ticker
		ReportStatusTicker *time.Ticker
		running            bool
		lastStatus         Status
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "should stop agent",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Agent{
				ID:                 tt.fields.ID,
				Codefresh:          tt.fields.Codefresh,
				Runtimes:           tt.fields.Runtimes,
				Logger:             tt.fields.Logger,
				TaskPullerTicker:   tt.fields.TaskPullerTicker,
				ReportStatusTicker: tt.fields.ReportStatusTicker,
				running:            tt.fields.running,
				lastStatus:         tt.fields.lastStatus,
			}
			a.Stop()

			assert.Equal(t, false, a.running)
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
