package codefresh

import "fmt"

type (
	// IRuntimeEnvironmentAPI declers Codefresh runtime environment API
	IRuntimeEnvironmentAPI interface {
		CreateRuntimeEnvironment(*CreateRuntimeOptions) (*RuntimeEnvironment, error)
		ValidateRuntimeEnvironment(*ValidateRuntimeOptions) error
		SignRuntimeEnvironmentCertificate(*SignCertificatesOptions) ([]byte, error)
	}

	RuntimeEnvironment struct {
		Name string
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
		Name: fmt.Sprintf("%s/%s", opt.Cluster, opt.Namespace),
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
	return resp.Bytes(), nil
}
