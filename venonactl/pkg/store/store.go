package store

import (
	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
)

const (
	ModeInCluster   = "InCluster"
	ApplicationName = "venona"
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

		DryRun bool

		RuntimeEnvironment string
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
	latestVersion := getLatestVersion()
	return map[string]interface{}{
		"ServerCert": map[string]string{
			"Cert": s.ServerCert.Cert,
			"Key":  s.ServerCert.Key,
			"Ca":   s.ServerCert.Ca,
		},
		"AppName":       ApplicationName,
		"Version":       latestVersion,
		"CodefreshHost": s.CodefreshAPI.Host,
		"Mode":          ModeInCluster,
		"Image": map[string]string{
			"Name": "codefresh/venona",
			"Tag":  latestVersion, // TODO calculate the latest tag
		},
		"Namespace":  s.KubernetesAPI.Namespace,
		"AgentToken": s.AgentToken,
	}
}
