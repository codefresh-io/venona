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
//go:generate go run github.com/codefresh-io/venona/venonactl/templates kubernetes
//go:generate go run github.com/codefresh-io/venona/venonactl/obj kubernetes
package operators

const (
	// AppName - app name for config
	AppName = "venona"

	// TypeKubernetesDind - name for Kubernetes  Dind runtimectl
	TypeKubernetesDind = "kubernetesDind"
	//typeDockerd = "dockerd"

	// StatusInstalled - status installed
	StatusInstalled = "Installed"
	// StatusNotInstalled - status installed
	StatusNotInstalled = "Not Installed"
)

// TableRows - array of string arrays
type TableRows [][]string
