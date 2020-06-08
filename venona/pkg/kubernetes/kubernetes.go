package kubernetes

import (
	"errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var errNotValidType = errors.New("not a valid type")

type (
	// Kubernetes API client
	Kubernetes interface {
		CreateResource() error
		DeleteResource() error
	}
	Options struct {
		Type  string
		Cert  string
		Token string
		Host  string
		Name  string
	}

	kube struct {
		client *kubernetes.Clientset
	}
)

// New build Kubernetes API
func New(opt Options) (Kubernetes, error) {
	if opt.Type != "runtime" {
		return nil, errNotValidType
	}
	client, err := buildKubeClient(opt.Host, opt.Token, opt.Cert)
	return &kube{
		client: client,
	}, err
}

func (k kube) CreateResource() error {
	return nil
}

func (k kube) DeleteResource() error {
	return nil
}

func buildKubeClient(host string, token string, crt string) (*kubernetes.Clientset, error) {

	return kubernetes.NewForConfig(&rest.Config{
		Host:        host,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(crt),
		},
	})
}
