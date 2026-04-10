package metrics

import (
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/task"
	"github.com/codefresh-io/go/venona/pkg/workflow"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeLabelValue(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"lowercase and replace spaces": {
			input: "My Pipeline",
			want:  "my_pipeline",
		},
		"replace special characters": {
			input: "deploy/production (v2.0)",
			want:  "deploy_production_v2_0",
		},
		"collapse consecutive special chars": {
			input: "Team--Service::Build & Test!",
			want:  "team_service_build_test",
		},
		"trim leading and trailing underscores": {
			input: "  --pipeline-- ",
			want:  "pipeline",
		},
		"already clean": {
			input: "simple_pipeline_name",
			want:  "simple_pipeline_name",
		},
		"empty string": {
			input: "",
			want:  "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := sanitizeLabelValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestObserveWorkflowMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	Register(reg)

	ObserveWorkflowMetrics(workflow.Type("create"), "pid-123", "test-pipeline", time.Second, time.Millisecond*500, time.Millisecond*200)

	mfs, err := reg.Gather()
	assert.NoError(t, err)

	found := map[string]bool{}
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}

	assert.True(t, found["runner_wf_duration_since_creation_sec"])
	assert.True(t, found["runner_wf_duration_in_runner_sec"])
	assert.True(t, found["runner_wf_processing_sec"])
}

func TestObserveAgentTaskMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	Register(reg)

	ObserveAgentTaskMetrics("proxy", "pid-456", "agent-pipeline", time.Second*2, time.Second, time.Millisecond*100)

	mfs, err := reg.Gather()
	assert.NoError(t, err)

	found := map[string]bool{}
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}

	assert.True(t, found["runner_wf_duration_since_creation_sec"])
	assert.True(t, found["runner_agent_processing_sec"])
}

func TestObserveK8sMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	Register(reg)

	ObserveK8sMetrics(task.TypeCreatePod, time.Millisecond*300)

	mfs, err := reg.Gather()
	assert.NoError(t, err)

	found := map[string]bool{}
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}

	assert.True(t, found["runner_k8s_processing_sec"])
}
