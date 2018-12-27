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

package runtime

import (

	"github.com/codefresh-io/Isser/installer/pkg/certs"
)

const (
	// TypeKubernetesDind - name for Kubernetes  Dind runtime 
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

// Config - contains all data needed to config specific runtime
type Config struct {
    // Runtime Env Type
	Type string
	// Runtime Env Name
	Name string
	
	Client ClientConfig
    
    ServerCerts certs.RuntimeServerCert
}

// ClientConfig - structy for client of runtime env (kube client config)
type ClientConfig struct {
	KubeClient KubernetesClientConfig
}

// KubernetesClientConfig - kube client config struct
type KubernetesClientConfig struct {
	Kubeconfig string
	Context string
	Namespace string
}

// Status - status of runtime env configuration
type Status struct {
	Status string
	StatusMessage string
}