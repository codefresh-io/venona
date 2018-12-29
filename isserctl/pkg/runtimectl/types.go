/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
We are using generated template.go for serialized kubernetes assets
*/
//go:generate go run github.com/codefresh-io/Isser/isserctl/templates kubernetes_dind

package runtimectl

import (
	"github.com/codefresh-io/Isser/isserctl/pkg/certs"
)

const (
	// TypeKubernetesDind - name for Kubernetes  Dind runtimectl
	TypeKubernetesDind = "kubernetesDind"
	//typeDockerd = "dockerd"

	// StatusInstalled - status installed
	StatusInstalled = "Installed"
	// StatusNotIntalled - status installed
	StatusNotIntalled = "NotInstalled"
	// StatusRunning - status failed
	StatusRunning = "Running"
	// StatusFailed - status failed
	StatusFailed = "Failed"
)

// Config - contains all data needed to config specific runtimectl
type Config struct {
	// runtimectl Env Type
	Type string
	// runtimectl Env Name
	Name string

	Client ClientConfig

	ServerCert *certs.ServerCert
}

// ClientConfig - structy for client of runtimectl env (kube client config)
type ClientConfig struct {
	KubeClient KubernetesClientConfig
}

// KubernetesClientConfig - kube client config struct
type KubernetesClientConfig struct {
	Kubeconfig *string
	Context    *string
	Namespace  *string
}

// Status - status of runtimectl env configuration
type Status struct {
	Status        string
	StatusMessage string
}
