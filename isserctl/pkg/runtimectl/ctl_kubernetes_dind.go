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

package runtimectl

import (
	"fmt"

	"k8s.io/client-go/kubernetes/scheme"
	templates "github.com/codefresh-io/Isser/isserctl/pkg/runtimectl/templates/kubernetes_dind"
)

// KubernetesDindCtl installs assets on Kubernetes Dind runtimectl Env
type KubernetesDindCtl struct {
}

// Install runtimectl environment
func (u *KubernetesDindCtl) Install(*Config) error {

	templatesMap := templates.TemplatesMap()
	// https://github.com/kubernetes/client-go/issues/193
	for n, tpl := range templatesMap {
	
		fmt.Printf("template = %s\n", n)
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, _ := decode([]byte(tpl), nil, nil)
	
		fmt.Printf("%++v\n\n", obj.GetObjectKind())
		fmt.Printf("%++v\n\n", obj)

	}
	return nil
}

// GetStatus of runtimectl environment
func (u *KubernetesDindCtl) GetStatus(*Config) (Status, error) {

	runtimectlStatus := Status{
		Status:        StatusRunning,
		StatusMessage: "",
	}
	return runtimectlStatus, nil
}

// Delete runtimectl environment
func (u *KubernetesDindCtl) Delete(*Config) error {

	return nil
}
