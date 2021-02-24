package store

import (
	"fmt"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	k8sApi "k8s.io/api/core/v1"
)

const (
	ModeInCluster           = "InCluster"
	ApplicationName         = "runner"
	MonitorApplicationName  = "monitor"
	AppProxyApplicationName = "app-proxy"
	EngineAppName           = "codefresh-engine"
	NetworkTesterName       = "cf-venona-network-tester"
)

var (
	store *Values
)

type (
	Values struct {
		AppName string

		Mode           string
		Image          *Image
		DockerRegistry string
		AgentToken     string

		ServerCert *certs.ServerCert

		CodefreshAPI *CodefreshAPI

		KubernetesAPI *KubernetesAPI

		Runner Runner

		VolumeProvisioner VolumeProvisioner

		LocalVolumeMonitor LocalVolumeMonitor

		Monitor Monitor

		AppProxy AppProxy

		AgentAPI *AgentAPI

		ClusterInCodefresh string

		DryRun bool

		Verbose bool

		Insecure bool

		RuntimeEnvironment string

		Version *Version

		ClusterId string

		Helm3 bool

		// need for define if monitor use cluster role or just role
		UseNamespaceWithRole bool

		AdditionalEnvVars map[string]string
	}

	KubernetesAPI struct {
		ConfigPath   string
		Namespace    string
		ContextName  string
		InCluster    bool
		NodeSelector string
		Tolerations  []k8sApi.Toleration
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

	Runner struct {
		Resources map[string]interface{}
	}
	VolumeProvisioner struct {
		Resources map[string]interface{}
	}

	LocalVolumeMonitor struct {
		Resources map[string]interface{}
	}
	Monitor struct {
		Resources map[string]interface{}
	}
	AppProxy struct {
		Resources map[string]interface{}
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
		"ClusterId":     s.ClusterId,
		"Version":       s.Version.Current.Version,
		"CodefreshHost": s.CodefreshAPI.Host,
		"Token":         s.CodefreshAPI.Token,
		"Mode":          ModeInCluster,
		"Verbose":       s.Verbose,
		"Insecure":      s.Insecure,
		"Image": map[string]string{
			"Name": "codefresh/venona",
			"Tag":  s.Version.Current.Version,
		},
		"AdditionalEnvVars": s.AdditionalEnvVars,
		"Namespace":         s.KubernetesAPI.Namespace,
		"ConfigPath":        s.KubernetesAPI.ConfigPath,
		"Context":           s.KubernetesAPI.ContextName,
		"NodeSelector":      s.KubernetesAPI.NodeSelector,
		"DockerRegistry":    s.DockerRegistry,
		"Tolerations":       s.KubernetesAPI.Tolerations,
		"AgentToken":        s.AgentAPI.Token,
		"AgentId":           s.AgentAPI.Id,
		"ServerCert": map[string]string{
			"Cert": "",
			"Key":  "",
			"Ca":   "",
		},
		"Runner": map[string]interface{}{
			"Resources": s.Runner.Resources,
		},
		"CreateRbac": true,
		"Storage": map[string]interface{}{
			"Backend":              "local",
			"CreateStorageClass":   true,
			"StorageClassName":     fmt.Sprintf("dind-local-volumes-%s-%s", ApplicationName, s.KubernetesAPI.Namespace),
			"LocalVolumeParentDir": "/var/lib/codefresh/dind-volumes",
			"AvailabilityZone":     "",
			"GoogleServiceAccount": "",
			"AwsAccessKeyId":       "",
			"AwsSecretAccessKey":   "",
			"VolumeProvisioner": map[string]interface{}{
				"Image":          "codefresh/dind-volume-provisioner:1.29.1",
				"NodeSelector":   s.KubernetesAPI.NodeSelector,
				"Resources":      s.VolumeProvisioner.Resources,
				"MountAzureJson": false,
			},
			"LocalVolumeMonitor": s.LocalVolumeMonitor.Resources,
		},
		"Monitor": map[string]interface{}{
			"Enabled":              true,
			"UseNamespaceWithRole": s.UseNamespaceWithRole,
			//TODO: need verify it on cluster level
			"RbacEnabled": true,
			"Helm3":       s.Helm3,
			"AppName":     MonitorApplicationName,
			"Image": map[string]string{
				"Name": "codefresh/agent",
				"Tag":  "stable",
			},
			"Resources": s.Monitor.Resources,
		},
		"AppProxy": map[string]interface{}{
			"AppName": AppProxyApplicationName,
			"Image": map[string]string{
				"Name": "codefresh/cf-app-proxy",
				"Tag":  "latest",
			},
			"Resources": s.AppProxy.Resources,
			"Ingress": map[string]interface{}{
				"Host":         "",
				"IngressClass": "",
			},
		},
		"Runtime": map[string]interface{}{
			"EngineAppName": EngineAppName,
		},
		"NetworkTester": map[string]interface{}{
			"PodName": NetworkTesterName,
			"Image": map[string]string{
				"Name": "codefresh/cf-venona-network-tester",
				"Tag":  "latest",
			},
		},
	}
}
