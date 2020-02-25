package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type (
	Kube interface {
		BuildClient() (*kubernetes.Clientset, error)
	}

	kube struct {
		contextName      string
		namespace        string
		pathToKubeConfig string
		inCluster        bool
	}

	Options struct {
		ContextName      string
		Namespace        string
		PathToKubeConfig string
		InCluster        bool
	}
)

func New(o *Options) Kube {
	return &kube{
		contextName:      o.ContextName,
		namespace:        o.Namespace,
		pathToKubeConfig: o.PathToKubeConfig,
		inCluster:        o.InCluster,
	}
}

func (k *kube) BuildClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if k.inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.pathToKubeConfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: k.contextName,
				Context: clientcmdapi.Context{
					Namespace: k.namespace,
				},
			}).ClientConfig()
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
