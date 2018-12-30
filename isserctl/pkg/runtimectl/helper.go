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
	"bytes"
	"fmt"
	"text/template"
	"github.com/golang/glog"
	"github.com/hairyhenderson/gomplate"
	gomplateData "github.com/hairyhenderson/gomplate/data"
	
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ExecuteTemplate - executes templates in tpl str with config as values
// Using template Funcs from gomplate - github.com/hairyhenderson/gomplate
func ExecuteTemplate(tplStr string, data interface{}) (string, error){

	// gomplate func initializing
	dataSources := []string{}
	dataSourceHeaders := []string{}
	d, _ := gomplateData.NewData(dataSources, dataSourceHeaders)

	template, err := template.New("").Funcs(gomplate.Funcs(d)).Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = template.Execute(buf, data) 
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ParseTemplates - parses and exexute templates and return map of strings with obj data
func ParseTemplates(templatesMap map[string]string, config *Config) (map[string]string, error) {
	parsedTemplates := make(map[string]string) 
	for n, tpl := range templatesMap {
		glog.V(4).Infof("parsing template = %s: ", n)
		tplEx, err := ExecuteTemplate(tpl, config)
		if err != nil {
			glog.V(4).Infof("Error: %v\n", err)
			fmt.Printf("Cannot parse and execute template %s: %v\n ", n, err)
			return nil, err
		}
		glog.V(4).Infof("parsing template Success\n")
		parsedTemplates[n] = tplEx
	}
	return parsedTemplates, nil
}

// KubeObjectsFromTemplates return map of runtime.Objects from templateMap
// see https://github.com/kubernetes/client-go/issues/193 for examples
func KubeObjectsFromTemplates(templatesMap map[string]string, config *Config) (map[string]runtime.Object, error) {
	parsedTemplates, err := ParseTemplates(templatesMap, config)
	if err != nil {
		return nil, err	
	}

	// Deserializing all kube objects from parsedTemplates
	// see https://github.com/kubernetes/client-go/issues/193 for examples	
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	kubeObjects := make(map[string]runtime.Object)
	for n, objStr := range parsedTemplates {
		glog.V(4).Infof("Deserializing template = %s: \n", n)
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
        if err != nil {
			glog.V(4).Infof("Error: %v \n", err)
			fmt.Printf("Cannot deserialize kuberentes object %s: %v\n ", n, err)
			return nil, err	
		}
		glog.V(4).Infof("deserializing template %s Success: %v\n", n, groupVersionKind)
		kubeObjects[n] = obj
	}
	return kubeObjects, nil
}

// NewKubeRESTClientConfig Returns rest.Config 
func NewKubeRESTClientConfig(config *Config) (*rest.Config, error) {
	var restConfig *rest.Config
	var err error
	kubeconfig := config.Client.KubeClient.Kubeconfig
	kubecontext := config.Client.KubeClient.Context
	namespace := config.Client.KubeClient.Namespace
	if *kubeconfig != "" {
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfig},
			&clientcmd.ConfigOverrides{
					CurrentContext: *kubecontext,
					Context: clientcmdapi.Context{
						Namespace:  *namespace,
					},

			}).ClientConfig()

	} else {
		restConfig, err = rest.InClusterConfig()
	}

	return restConfig, err
}

// NewKubeClientset - returns clientset
func NewKubeClientset(config *Config) (*kubernetes.Clientset, error) {
	kubeClientConfig, err := NewKubeRESTClientConfig(config)
	if err != nil {
		return nil, err	
	}
	return kubernetes.NewForConfig(kubeClientConfig)
}