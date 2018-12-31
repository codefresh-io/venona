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
//	"k8s.io/client-go/kubernetes/scheme"
//	"k8s.io/client-go/kubernetes"

//	"k8s.io/apimachinery/pkg/runtime/serializer"
//	"k8s.io/apimachinery/pkg/runtime/schema"//
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/client-go/rest"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/core/v1"

    templates "github.com/codefresh-io/Isser/isserctl/templates/kubernetes_dind"
)

// KubernetesDindCtl installs assets on Kubernetes Dind runtimectl Env
type KubernetesDindCtl struct {

}

// Install runtimectl environment
func (u *KubernetesDindCtl) Install(config *Config) error {

	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, config)
	if err != nil {
		return err	
	}	

	kubeClientset, err := NewKubeClientset(config)
	if err != nil {
		fmt.Printf("Cannot create kubernetes clientset: %v\n ", err)
		return err	
	}
	namespace := config.Client.KubeClient.Namespace
	var createErr error
	var kind, name string
	for n, obj := range kubeObjects {
		switch objT := obj.(type) {
			case *v1.Secret:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, createErr = kubeClientset.CoreV1().Secrets(*namespace).Create(objT)
			case *v1.ConfigMap:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, createErr = kubeClientset.CoreV1().ConfigMaps(*namespace).Create(objT)
			case *v1.Service:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, createErr = kubeClientset.CoreV1().Services(*namespace).Create(objT)
			// case *v1beta1.Role:
			// 	// o is the actual role Object with all fields etc
			// case *v1beta1.RoleBinding:
			// case *v1beta1.ClusterRole:
			// case *v1beta1.ClusterRoleBinding:
			// case *v1.ServiceAccount:
			default:
				fmt.Printf("Unknown object type in %s: %T\n ", n, objT)
			}
		if createErr == nil {
			fmt.Printf("%s \"%s\" created\n ", kind, name)
		} else if statusError, errIsStatusError := createErr.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				fmt.Printf("%s \"%s\" already exists\n ", kind, name)
			} else {
				fmt.Printf("%s \"%s\" failed: %v ", kind, name, statusError)
				return statusError
			}
		} else {
			fmt.Printf("%s \"%s\" failed: %v ", kind, name, createErr)
			return createErr
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
func (u *KubernetesDindCtl) GetStatus(config *Config) (*Status, error) {
	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, config)
	if err != nil {
		return nil, err	
	}	

	kubeClientset, err := NewKubeClientset(config)
	if err != nil {
		fmt.Printf("Cannot create kubernetes clientset: %v\n ", err)
		return nil, err	
	}
	namespace := config.Client.KubeClient.Namespace
	var getErr error
	var kind, name, status, statusMessage string
	
	status = StatusInstalled
	for n, obj := range kubeObjects {
		switch objT := obj.(type) {
			case *v1.Secret:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, getErr = kubeClientset.CoreV1().Secrets(*namespace).Get(name, metav1.GetOptions{})
			case *v1.ConfigMap:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, getErr = kubeClientset.CoreV1().ConfigMaps(*namespace).Get(name, metav1.GetOptions{})
			case *v1.Service:
				name = objT.ObjectMeta.Name
                kind = objT.TypeMeta.Kind
				_, getErr = kubeClientset.CoreV1().Services(*namespace).Get(name, metav1.GetOptions{})
			// case *v1beta1.Role:
			// 	// o is the actual role Object with all fields etc
			// case *v1beta1.RoleBinding:
			// case *v1beta1.ClusterRole:
			// case *v1beta1.ClusterRoleBinding:
			// case *v1.ServiceAccount:
			default:
				fmt.Printf("Unknown object type in %s: %T\n ", n, objT)
			}
		if getErr == nil {
			statusMessage += fmt.Sprintf("%s \"%s\" installed\n", kind, name)
		} else if statusError, errIsStatusError := getErr.(*errors.StatusError); errIsStatusError {
            statusMessage += fmt.Sprintf("%s\n", statusError.ErrStatus.Message)
            status = StatusNotInstalled
		} else {
			fmt.Printf("%s \"%s\" failed: %v ", kind, name, getErr)
			return nil, getErr
		}
	}
	runtimectlStatus := &Status{
		Status:        status,
		StatusMessage: statusMessage,
	}
	return runtimectlStatus, nil
}

// Delete runtimectl environment
func (u *KubernetesDindCtl) Delete(config *Config) error {
	fmt.Printf("To delete isser delete all the object printed by status\n")
	return nil
}
