package store

import (
	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/isser/isserctl/pkg/certs"
)

const (
	ModeInCluster   = "InCluster"
	ApplicationName = "isser"
)

var (
	store *Values
)

type (
	Values struct {
		AppName string

		Version    string
		Mode       string
		Image      *Image
		AgentToken string

		ServerCert *certs.ServerCert

		CodefreshAPI *CodefreshAPI

		KubernetesAPI *KubernetesAPI

		ClusterInCodefresh string
	}

	KubernetesAPI struct {
		ConfigPath  string
		Namespace   string
		ContextName string
	}

	CodefreshAPI struct {
		Host   string
		Token  string
		Client codefresh.Codefresh
	}

	Image struct {
		Name string
		Tag  string
	}
)

func GetStore() *Values {
	if store == nil {
		store = &Values{}
		return store
	}
	return store
}

func (s *Values) BuildValues() map[string]interface{} {
	return map[string]interface{}{
		"ServerCert": map[string]string{
			"Cert": s.ServerCert.Cert,
			"Key":  s.ServerCert.Key,
			"Ca":   s.ServerCert.Ca,
		},
		"AppName":       ApplicationName,
		"Version":       "0.0.1", // TODO calculate the latest version
		"CodefreshHost": s.CodefreshAPI.Host,
		"Mode":          ModeInCluster,
		"Image": map[string]string{
			"Name": "codefresh/isser",
			"Tag":  "master", // TODO calculate the latest tag
		},
		"Namespace":  s.KubernetesAPI.Namespace,
		"AgentToken": s.AgentToken,
	}
}
