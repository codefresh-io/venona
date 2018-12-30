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
		glog.V(4).Infof("template = %s\n", n)
		tplEx, err := ExecuteTemplate(tpl, config)
		if err != nil {
			fmt.Printf("Cannot parse and execute template %s: %v\n ", n, err)
			return nil, err
		}
		parsedTemplates[n] = tplEx
	}
	return parsedTemplates, nil
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