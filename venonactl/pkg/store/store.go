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

		AgentAPI *AgentAPI

		ClusterInCodefresh string

		DryRun bool

		RuntimeEnvironment string

		Version *Version
	}

	KubernetesAPI struct {
		ConfigPath   string
		Namespace    string
		ContextName  string
		InCluster    bool
		NodeSelector string
		Tolerations  string
	}

	CodefreshAPI struct {
		Host              string
		Token             string
		Client            codefresh.Codefresh
		BuildNodeSelector map[string]string
	}

	AgentAPI struct {
		Token string
		Id    string
	}

	Image struct {
		Name string
		Tag  string
	}

	Version struct {
		Current *CurrentVersion
	}

	CurrentVersion struct {
		Version string
		Commit  string
		Date    string
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
		"Version":       s.Version.Current.Version,
		"CodefreshHost": s.CodefreshAPI.Host,
		"Mode":          ModeInCluster,
		"Image": map[string]string{
			"Name": "codefresh/venona",
			"Tag":  s.Version.Current.Version,
		},
		"Namespace":    s.KubernetesAPI.Namespace,
		"ConfigPath":   s.KubernetesAPI.ConfigPath,
		"Context":      s.KubernetesAPI.ContextName,
		"NodeSelector": s.KubernetesAPI.NodeSelector,
		"Tolerations":  s.KubernetesAPI.Tolerations,
		"AgentToken":   s.AgentAPI.Token,
		"AgentId":      s.AgentAPI.Id,
		"ServerCert": map[string]string{
			"Cert": "",
			"Key":  "",
			"Ca":   "",
		},
		"Storage": map[string]interface{}{
			"Backend": "local",
			"LocalVolumeParentDir": "/var/lib/codefresh/dind-volumes",
			"AvailabilityZone": "",
			"GoogleServiceAccount": "",
			"AwsAccessKeyId": "",
			"AwsSecretAccessKey": "",
			"VolumeProvisioner": map[string]interface{}{
				"Image": "codefresh/dind-volume-provisioner:v20",
				"NodeSelector": s.KubernetesAPI.NodeSelector,
				"Tolerations":  s.KubernetesAPI.Tolerations,
			},
		},
	}
}
