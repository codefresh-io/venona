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
//	"github.com/golang/glog"
//	"k8s.io/client-go/rest"
//	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes"

//	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/schema"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/client-go/rest"

	"k8s.io/api/core/v1"

    templates "github.com/codefresh-io/Isser/isserctl/templates/kubernetes_dind"
)

// KubernetesDindCtl installs assets on Kubernetes Dind runtimectl Env
type KubernetesDindCtl struct {

}

// Install runtimectl environment
func (u *KubernetesDindCtl) Install(config *Config) error {

	templatesMap := templates.TemplatesMap()
	parsedTemplates, err := ParseTemplates(templatesMap, config)
	if err != nil {
		return err	
	}

	// Deserializing all kube objects from parsedTemplates
	// see https://github.com/kubernetes/client-go/issues/193 for examples	
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	kubeObjects := make(map[string]KubeRuntimeObject)
	for n, objStr := range parsedTemplates {
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
		if groupVersionKind.Group == "" {
			groupVersionKind.Group = "api"
		}
        if err != nil {
			fmt.Printf("Cannot deserialize kuberentes object %s: %v\n ", n, err)
			return err	
		}
		kubeObjects[n] = KubeRuntimeObject{
			Obj: obj,
			GroupVersion: &schema.GroupVersion{
				Group: groupVersionKind.Group,
				Version: groupVersionKind.Version,
			},
		}
	    //objKind := obj.GetObjectKind()
	    //fmt.Printf("%++v\n\n", objKind)
		// fmt.Printf("%++v\n\n", obj)
	}

	kubeClientConfig, err := NewKubeRESTClientConfig(config)
	if err != nil {
		fmt.Printf("Cannot get kubernetes client config: %v\n ", err)
		return err	
	}
	kubeClientset, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		fmt.Printf("Cannot create kubernetes clientset: %v\n ", err)
		return err	
	}
	namespace := config.Client.KubeClient.Namespace
	for n, obj := range kubeObjects {
		switch o := obj.Obj.(type) {

		case *v1.Secret:
			kubeClientset.CoreV1().Secrets(*namespace).Create(obj.Obj.(*v1.Secret))
		case *v1.ConfigMap:
			kubeClientset.CoreV1().ConfigMaps(*namespace).Create(obj.Obj.(*v1.ConfigMap))
		case *v1.Service:
			kubeClientset.CoreV1().Services(*namespace).Create(obj.Obj.(*v1.Service))
		// case *v1beta1.Role:
		// 	// o is the actual role Object with all fields etc
		// case *v1beta1.RoleBinding:
		// case *v1beta1.ClusterRole:
		// case *v1beta1.ClusterRoleBinding:
		// case *v1.ServiceAccount:
		default:
			fmt.Printf("Unknown object type in %s: %v\n ", n, o)
		}
	}

	// for n, obj := range kubeObjects {
	// 	restConfig := rest.CopyConfig(kubeClientConfig)
	// 	//restConfig.APIPath = "/apis"
	// 	restConfig.ContentConfig.GroupVersion = obj.GroupVersion
	// 	restConfig.ContentConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	// 	restConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	// 	restClient, err := rest.UnversionedRESTClientFor(restConfig)
	// 	if err != nil {
	// 		fmt.Printf("Cannot get kubernetes rest client for %s: %v\n ", n, err)
	// 		return err	
	// 	}		
        
	// 	req := restClient.Post().
	// 	    Body(obj.Obj).
	// 		Resource("secrets").
	// 		Namespace("tst1")
	// 	result := req.Do()
	// 	resultRaw, err := result.Raw() 
	// 	if err != nil {
	// 		fmt.Printf("Cannot get request result for %s: %v\n ", n, err)
	// 		return err	
	// 	}	
	// 	glog.V(4).Infof("result for %s : %v", n, string(resultRaw))
	// }

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
