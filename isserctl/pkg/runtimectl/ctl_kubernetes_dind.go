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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//	"k8s.io/api/core/v1"

	"github.com/codefresh-io/isser/isserctl/obj/kubeobj"
	"github.com/codefresh-io/isser/isserctl/pkg/store"
	templates "github.com/codefresh-io/isser/isserctl/templates/kubernetes_dind"
)

// KubernetesDindCtl installs assets on Kubernetes Dind runtimectl Env
type KubernetesDindCtl struct {
}

// Install runtimectl environment
func (u *KubernetesDindCtl) Install() error {
	s := store.GetStore()
	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, s.BuildValues())
	if err != nil {
		return err
	}

	kubeClientset, err := NewKubeClientset(s)
	if err != nil {
		fmt.Printf("Cannot create kubernetes clientset: %v\n ", err)
		return err
	}
	namespace := s.KubernetesAPI.Namespace
	var createErr error
	var kind, name string
	for _, obj := range kubeObjects {
		name, kind, createErr = kubeobj.CreateObject(kubeClientset, obj, namespace)

		if createErr == nil {
			fmt.Printf("%s \"%s\" created\n ", kind, name)
		} else if statusError, errIsStatusError := createErr.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				fmt.Printf("%s \"%s\" already exists\n", kind, name)
			} else {
				fmt.Printf("%s \"%s\" failed: %v ", kind, name, statusError)
				return statusError
			}
		} else {
			fmt.Printf("%s \"%s\" failed: %v ", kind, name, createErr)
			return createErr
		}
	}

	return nil
}

// // GetStatus of runtimectl environment
// func (u *KubernetesDindCtl) GetStatus(config *Config) (*Status, error) {
// 	templatesMap := templates.TemplatesMap()
// 	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	kubeClientset, err := NewKubeClientset(config)
// 	if err != nil {
// 		fmt.Printf("Cannot create kubernetes clientset: %v\n ", err)
// 		return nil, err
// 	}
// 	namespace := config.Client.KubeClient.Namespace
// 	var getErr error
// 	var kind, name, status, statusMessage string

// 	status = StatusInstalled
// 	for _, obj := range kubeObjects {
// 		name, kind, getErr = kubeobj.CheckObject(kubeClientset, obj, *namespace)
// 		if getErr == nil {
// 			statusMessage += fmt.Sprintf("%s \"%s\" installed\n", kind, name)
// 		} else if statusError, errIsStatusError := getErr.(*errors.StatusError); errIsStatusError {
// 			statusMessage += fmt.Sprintf("%s\n", statusError.ErrStatus.Message)
// 			status = StatusNotInstalled
// 		} else {
// 			fmt.Printf("%s \"%s\" failed: %v ", kind, name, getErr)
// 			return nil, getErr
// 		}
// 	}
// 	runtimectlStatus := &Status{
// 		Status:        status,
// 		StatusMessage: statusMessage,
// 	}
// 	return runtimectlStatus, nil
// }

// // Delete runtimectl environment
// func (u *KubernetesDindCtl) Delete(config *Config) error {
// 	fmt.Printf("To delete isser delete all the object printed by status\n")
// 	return nil
// }
