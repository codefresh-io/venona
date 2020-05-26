package runtime

import "github.com/codefresh-io/venona/pkg/codefresh"

type (
	// Runtime API client
	Runtime interface {
		StartWorkflow([]codefresh.Task) error
		TerminateWorkflow([]codefresh.Task) error
	}

	Options struct{}

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
