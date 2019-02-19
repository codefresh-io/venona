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

		Mode       string
		Image      *Image
		AgentToken string

		ServerCert *certs.ServerCert

		CodefreshAPI *CodefreshAPI

		KubernetesAPI *KubernetesAPI

		ClusterInCodefresh string

		DryRun bool

		RuntimeEnvironment string

		Version *Version
	}

	KubernetesAPI struct {
		ConfigPath  string
		Namespace   string
		ContextName string
		InCluster   bool
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

	Version struct {
		Current *CurrentVersion
		Latest  *LatestVersion
	}

	CurrentVersion struct {
		Version string
		Commit  string
		Date    string
	}
	LatestVersion struct {
		Version   string
		Commit    string
		Date      string
		IsDefault bool
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
		"AppName":       ApplicationName,
		"Version":       s.Version.Latest.Version,
		"CodefreshHost": s.CodefreshAPI.Host,
		"Mode":          ModeInCluster,
		"Image": map[string]string{
			"Name": "codefresh/venona",
			"Tag":  s.Version.Latest.Version,
		},
		"Namespace":  s.KubernetesAPI.Namespace,
		"AgentToken": "",
		"ServerCert": map[string]string{
			"Cert": "",
			"Key":  "",
			"Ca":   "",
		},
	}
}
