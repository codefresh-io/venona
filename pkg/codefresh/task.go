package codefresh

import "encoding/json"

const (
	TypeCreatePod = "CreatePod"
	TypeCreatePVC = "CreatePvc"
	TypeDeletePod = "DeletePod"
	TypeDeletePVC = "DeletePvc"
)

func UnmarshalTasks(data []byte) ([]Task, error) {
	var r []Task
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Tasks) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Tasks []Task

type Task struct {
	Type     string      `json:"type"`
	Spec     interface{} `json:"spec"`
	Metadata Metadata    `json:"metadata"`
}

type Metadata struct {
	CreatedAt string `json:"createdAt"`
	Account   string `json:"account"`
	ReName    string `json:"reName"`
	Workflow  string `json:"workflow"`
}
