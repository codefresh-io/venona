package agent

import (
	"errors"
	"time"

	"github.com/codefresh-io/venona/pkg/codefresh"
	"github.com/codefresh-io/venona/pkg/logger"
	"github.com/codefresh-io/venona/pkg/runtime"
)

var (
	errAlreadyStarted = errors.New("Already started")
)

type (
	// Agent holds all the references from Codefresh
	// in order to run the process
	Agent struct {
		ID                 string
		Codefresh          codefresh.Codefresh
		Runtimes           map[string]runtime.Runtime
		Logger             logger.Logger
		TaskPullerTicker   *time.Ticker
		ReportStatusTicker *time.Ticker
		started            bool
	}

	workflowCandidate struct {
		tasks   []codefresh.Task
		runtime string
	}
)

// Start starting the agent process
func (a Agent) Start() error {
	if a.started {
		return errAlreadyStarted
	}
	a.Logger.Debug("Starting agent")
	a.started = true
	go a.fetchTasks()
	go a.reportStatus()
	time.Sleep(30 * time.Second)
	return nil
}

func (a Agent) fetchTasks() {
	for {
		select {
		case <-a.TaskPullerTicker.C:
			a.Logger.Debug("Requesting tasks from API server")
			tasks, err := a.Codefresh.Tasks()
			if err != nil {
				a.Logger.Error(err.Error())
				continue
			}
			a.Logger.Debug("Recieved new tasks", "len", len(tasks))
			creationTasks := []codefresh.Task{}
			deletionTasks := []codefresh.Task{}
			for _, t := range tasks {
				a.Logger.Debug("Starting tasks", "runtime", t.Metadata.ReName)
				if t.Type == codefresh.TypeCreatePod || t.Type == codefresh.TypeCreatePVC {
					creationTasks = append(creationTasks, t)
				}

				if t.Type == codefresh.TypeDeletePod || t.Type == codefresh.TypeDeletePVC {
					deletionTasks = append(deletionTasks, t)
				}
			}

			candidates := []workflowCandidate{}
			candidates = append(candidates, groupTasks(creationTasks)...)
			candidates = append(candidates, groupTasks(deletionTasks)...)

			for _, c := range candidates {
				if err := a.Runtimes[c.runtime].StartWorkflow(c.tasks); err != nil {
					a.Logger.Error(err.Error())
				}
			}

		}
	}
}

func (a Agent) reportStatus() {
	for {
		select {
		case <-a.ReportStatusTicker.C:
			err := a.Codefresh.ReportStatus(codefresh.AgentStatus{
				Message: "All good",
			})
			if err != nil {
				a.Logger.Error(err.Error())
				continue
			}
		}
	}
}

func groupTasks(tasks []codefresh.Task) []workflowCandidate {
	candidates := []workflowCandidate{}
	for _, task := range tasks {
		name := task.Metadata.Workflow
		if name == "" {
			// If for some reason the task is not related to any workflow
			// Might heppen in older versions on Codefresh
			name = "_"
		}
		for _, c := range candidates {
			if c.runtime != name {
				continue
			}
			c.tasks = append(c.tasks, task)
			break
		}
	}
	return candidates
}
