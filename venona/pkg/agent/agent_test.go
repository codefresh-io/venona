package agent

import (
	"testing"

	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/stretchr/testify/assert"
)

func Test_groupTasks(t *testing.T) {
	type args struct {
		tasks []codefresh.Task
	}
	tests := []struct {
		name string
		args args
		want map[string][]codefresh.Task
	}{
		{
			name: "should group by workflow name",
			args: args{
				tasks: []codefresh.Task{
					{
						Metadata: codefresh.Metadata{
							Workflow: "1",
						},
					},
					{
						Metadata: codefresh.Metadata{
							Workflow: "2",
						},
					},
					{
						Metadata: codefresh.Metadata{
							Workflow: "1",
						},
					},
				},
			},
			want: map[string][]codefresh.Task{
				"1": {
					{

						Metadata: codefresh.Metadata{
							Workflow: "1",
						},
					},
					{
						Metadata: codefresh.Metadata{
							Workflow: "1",
						},
					},
				},
			"2": {
				{
					Metadata: codefresh.Metadata{
						Workflow: "2",
					},
				},
			},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupedTasks  := groupTasks(tt.args.tasks)
			assert.Equal(t, tt.want, groupedTasks)
		})
	}
}
