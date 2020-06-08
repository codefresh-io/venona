// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		CreateResource(spec string) error
		DeleteResource(spec string) error
	}
	// Options for Kubernetes
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

func (k kube) CreateResource(spec string) error {
	return nil
}

func (k kube) DeleteResource(spec string) error {
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
