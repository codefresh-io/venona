package codefresh

import "encoding/json"

type (
	// AgentStatus is the latest status of the agent
	AgentStatus struct {
		Message string `json:"message"`
	}
)

func (r *AgentStatus) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
