package codefresh

import (
	"fmt"
	"net/url"
	"time"
)

type (
	// IPipelineAPI declers Codefresh pipeline API
	IPipelineAPI interface {
		List() ([]*Pipeline, error)
		Run(string) (string, error)
	}

	PipelineMetadata struct {
		Name     string `json:"name"`
		IsPublic bool   `json:"isPublic"`
		Labels   struct {
			Tags []string `json:"tags"`
		} `json:"labels"`
		Deprecate struct {
			ApplicationPort string `json:"applicationPort"`
			RepoPipeline    bool   `json:"repoPipeline"`
		} `json:"deprecate"`
		OriginalYamlString string    `json:"originalYamlString"`
		AccountID          string    `json:"accountId"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		Project            string    `json:"project"`
		ID                 string    `json:"id"`
	}

	PipelineSpec struct {
		Triggers []struct {
			Type     string   `json:"type"`
			Repo     string   `json:"repo"`
			Events   []string `json:"events"`
			Provider string   `json:"provider"`
			Context  string   `json:"context"`
		} `json:"triggers"`
		Contexts  []interface{} `json:"contexts"`
		Variables []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"variables"`
		Steps  map[string]interface{} `json:"steps"`
		Stages []interface{}          `json:"stages"`
		Mode   string                 `json:"mode"`
	}

	Pipeline struct {
		Metadata PipelineMetadata `json:"metadata"`
		Spec     PipelineSpec     `json:"spec"`
	}

	getPipelineResponse struct {
		Docs  []*Pipeline `json:"docs"`
		Count int         `json:"count"`
	}

	pipeline struct {
		codefresh Codefresh
	}
)

func newPipelineAPI(codefresh Codefresh) IPipelineAPI {
	return &pipeline{codefresh}
}

// Get - returns pipelines from API
func (p *pipeline) List() ([]*Pipeline, error) {
	r := &getPipelineResponse{}
	resp, err := p.codefresh.requestAPI(&requestOptions{
		path:   "/api/pipelines",
		method: "GET",
	})
	err = p.codefresh.decodeResponseInto(resp, r)
	return r.Docs, err
}

func (p *pipeline) Run(name string) (string, error) {
	resp, err := p.codefresh.requestAPI(&requestOptions{
		path:   fmt.Sprintf("/api/pipelines/run/%s", url.PathEscape(name)),
		method: "POST",
	})
	if err != nil {
		return "", err
	}
	return p.codefresh.getBodyAsString(resp)
}
