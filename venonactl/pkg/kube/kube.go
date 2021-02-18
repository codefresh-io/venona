package kube

import (
	v1Core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type (
	Kube interface {
		BuildClient() (*kubernetes.Clientset, error)
		BuildConfig() (*rest.Config, error)
		EnsureNamespaceExists(cs *kubernetes.Clientset) error
	}

	kube struct {
		contextName      string
		namespace        string
		pathToKubeConfig string
		inCluster        bool
		dryRun           bool
	}

	Options struct {
		ContextName      string
		Namespace        string
		PathToKubeConfig string
		InCluster        bool
		DryRun           bool
	}
)

func New(o *Options) Kube {
	return &kube{
		contextName:      o.ContextName,
		namespace:        o.Namespace,
		pathToKubeConfig: o.PathToKubeConfig,
		inCluster:        o.InCluster,
		dryRun:           o.DryRun,
	}
}

func (k *kube) BuildClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if k.inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = k.BuildConfig()
		if err != nil { // if cannot create from kubeConfigPath, try in-cluster config
			config, err = rest.InClusterConfig()
		}
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (k *kube) EnsureNamespaceExists(cs *kubernetes.Clientset) error {
	if k.dryRun {
		return nil
	}
	_, err := cs.CoreV1().Namespaces().Get(k.namespace, v1.GetOptions{})
	if err != nil {
		nsSpec := &v1Core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: k.namespace}}
		_, err := cs.CoreV1().Namespaces().Create(nsSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *kube) BuildConfig() (*rest.Config, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.pathToKubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: k.contextName,
			Context: clientcmdapi.Context{
				Namespace: k.namespace,
			},
		})
	cc, err := config.ClientConfig()

	if err != nil { // if cannot create from kubeConfigPath, try in-cluster config
		return rest.InClusterConfig()
	}

	return cc, nil

}
