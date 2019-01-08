package codefresh

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type (
	// IRuntimeEnvironmentAPI declers Codefresh runtime environment API
	IRuntimeEnvironmentAPI interface {
		CreateRuntimeEnvironment(*CreateRuntimeOptions) (*RuntimeEnvironment, error)
		ValidateRuntimeEnvironment(*ValidateRuntimeOptions) error
		SignRuntimeEnvironmentCertificate(*SignCertificatesOptions) ([]byte, error)
		GetRuntimeEnvironment(string) (*RuntimeEnvironment, error)
		GetRuntimeEnvironments() ([]*RuntimeEnvironment, error)
	}

	RuntimeEnvironment struct {
		Version               int                   `json:"version"`
		Metadata              RuntimeMetadata       `json:"metadata"`
		Extends               []string              `json:"extends"`
		Description           string                `json:"description"`
		AccountID             string                `json:"accountId"`
		RuntimeScheduler      RuntimeScheduler      `json:"runtimeScheduler"`
		DockerDaemonScheduler DockerDaemonScheduler `json:"dockerDaemonScheduler"`
		Status                struct {
			Message   string    `json:"message"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"status"`
	}

	RuntimeScheduler struct {
		Cluster struct {
			ClusterProvider struct {
				AccountID string `json:"accountId"`
				Selector  string `json:"selector"`
			} `json:"clusterProvider"`
			Namespace string `json:"namespace"`
		} `json:"cluster"`
		UserAccess bool `json:"userAccess"`
	}

	DockerDaemonScheduler struct {
		Cluster struct {
			ClusterProvider struct {
				AccountID string `json:"accountId"`
				Selector  string `json:"selector"`
			} `json:"clusterProvider"`
			Namespace string `json:"namespace"`
		} `json:"cluster"`
		UserAccess bool `json:"userAccess"`
	}

	RuntimeMetadata struct {
		Agent        bool   `json:"agent"`
		Name         string `json:"name"`
		ChangedBy    string `json:"changedBy"`
		CreationTime string `json:"creationTime"`
	}

	CreateRuntimeOptions struct {
		Cluster   string
		Namespace string
		HasAgent  bool
	}

	ValidateRuntimeOptions struct {
		Cluster   string
		Namespace string
	}

	SignCertificatesOptions struct {
		AltName string
		CSR     string
	}

	createRuntimeEnvironmentResponse struct {
		Name string
	}
)

// CreateRuntimeEnvironment - create Runtime-Environment
func (c *codefresh) CreateRuntimeEnvironment(opt *CreateRuntimeOptions) (*RuntimeEnvironment, error) {
	re := &RuntimeEnvironment{
		Metadata: RuntimeMetadata{
			Name: fmt.Sprintf("%s/%s", opt.Cluster, opt.Namespace),
		},
	}
	body := map[string]interface{}{
		"clusterName": opt.Cluster,
		"namespace":   opt.Namespace,
	}
	if opt.HasAgent {
		body["agent"] = true
	}
	resp, err := c.requestAPI(&requestOptions{
		path:   "/api/custom_clusters/register",
		method: "POST",
		body:   body,
	})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 400 {
		return re, nil
	}
	return nil, fmt.Errorf("Error during runtime environment creation")
}

func (c *codefresh) ValidateRuntimeEnvironment(opt *ValidateRuntimeOptions) error {
	body := map[string]interface{}{
		"clusterName": opt.Cluster,
		"namespace":   opt.Namespace,
	}
	_, err := c.requestAPI(&requestOptions{
		path:   "/api/custom_clusters/validate",
		method: "POST",
		body:   body,
	})
	return err
}

func (c *codefresh) SignRuntimeEnvironmentCertificate(opt *SignCertificatesOptions) ([]byte, error) {
	body := map[string]interface{}{
		"reqSubjectAltName": opt.AltName,
		"csr":               opt.CSR,
	}
	resp, err := c.requestAPI(&requestOptions{
		path:   "/api/custom_clusters/signServerCerts",
		method: "POST",
		body:   body,
	})
	if err != nil {
		return nil, err
	}
	return c.getBodyAsBytes(resp)
}

func (c *codefresh) GetRuntimeEnvironment(name string) (*RuntimeEnvironment, error) {
	re := &RuntimeEnvironment{}
	path := fmt.Sprintf("/api/runtime-environments/%s", url.PathEscape(name))
	resp, err := c.requestAPI(&requestOptions{
		path:   path,
		method: "GET",
		qs: map[string]string{
			"extend": "false",
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	c.decodeResponseInto(resp, re)
	return re, nil
}

func (c *codefresh) GetRuntimeEnvironments() ([]*RuntimeEnvironment, error) {
	emptySlice := make([]*RuntimeEnvironment, 0)
	resp, err := c.requestAPI(&requestOptions{
		path:   "/api/runtime-environments",
		method: "GET",
	})
	tokensAsBytes, err := c.getBodyAsBytes(resp)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(tokensAsBytes, &emptySlice)

	return emptySlice, err
}
