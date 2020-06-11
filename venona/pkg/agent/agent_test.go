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
	"testing"

	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/stretchr/testify/assert"
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
