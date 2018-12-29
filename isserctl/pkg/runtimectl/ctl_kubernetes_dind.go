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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes/scheme"

    templates "github.com/codefresh-io/Isser/isserctl/templates/kubernetes_dind"
)

// KubernetesDindCtl installs assets on Kubernetes Dind runtimectl Env
type KubernetesDindCtl struct {

}


// NewClientsetConfig Returns rest.Config 
func (u *KubernetesDindCtl) NewClientsetConfig(config *Config) (*rest.Config, error) {
	var restConfig *rest.Config
	var err error
	kubeconfig := config.Client.KubeClient.Kubeconfig
    kubecontext := config.Client.KubeClient.Context
	if *kubeconfig != "" {
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfig},
			&clientcmd.ConfigOverrides{
					CurrentContext: *kubecontext,

			}).ClientConfig()

	} else {
		restConfig, err = rest.InClusterConfig()
	}

	return restConfig, err
}

// Install runtimectl environment
func (u *KubernetesDindCtl) Install(config *Config) error {

	templatesMap := templates.TemplatesMap()
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	// https://github.com/kubernetes/client-go/issues/193
	for n, tpl := range templatesMap {
		fmt.Printf("template = %s\n", n)
		tplEx, err := ExecuteTemplate(tpl, config)
		if err != nil {
			fmt.Printf("Cannot parse and execute template %s: %v\n ", n, err)
			continue
		}
		
		obj, _, _ := kubeDecode([]byte(tplEx), nil, nil)
	    objKind := obj.GetObjectKind()
		fmt.Printf("%++v\n\n", objKind)
		fmt.Printf("%++v\n\n", obj)

	}
	return nil
}

// GetStatus of runtimectl environment
func (u *KubernetesDindCtl) GetStatus(config *Config) (Status, error) {

	runtimectlStatus := Status{
		Status:        StatusRunning,
		StatusMessage: "",
	}
	return runtimectlStatus, nil
}

// Delete runtimectl environment
func (u *KubernetesDindCtl) Delete(config *Config) error {

	return nil
}
