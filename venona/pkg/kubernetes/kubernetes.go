// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"encoding/json"
	"errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var errNotValidType = errors.New("not a valid type")

type (
	// Kubernetes API client
	Kubernetes interface {
		CreateResource(spec interface{}) error
		DeleteResource(spec interface{}) error
	}
	// Options for Kubernetes
	Options struct {
		Type  string
		Cert  string
		Token string
		Host  string
	}

	kube struct {
		client *kubernetes.Clientset
	}
)

// New build Kubernetes API
func New(opt Options) (Kubernetes, error) {
	if opt.Type != "runtime" {
		return nil, errNotValidType
	}
	client, err := buildKubeClient(opt.Host, opt.Token, opt.Cert)
	return &kube{
		client: client,
	}, err
}

func (k kube) CreateResource(spec interface{}) error {

	bytes, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := kubeDecode([]byte(string(bytes)), nil, nil)
	if err != nil {
		return err
	}

	var namespace string
	switch objT := obj.(type) {
	case *v1.PersistentVolumeClaim:
		namespace = objT.ObjectMeta.Namespace
		_, err = k.client.CoreV1().PersistentVolumeClaims(namespace).Create(obj.(*v1.PersistentVolumeClaim))
		if err != nil {
			return err
		}

	case *v1.Pod:
		namespace = objT.ObjectMeta.Namespace
		_, err = k.client.CoreV1().Pods(namespace).Create(obj.(*v1.Pod))
		if err != nil {
			return err
		}

	}
	return err
}

func (k kube) DeleteResource(spec interface{}) error {
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := kubeDecode([]byte(spec.(string)), nil, nil)
	if err != nil {
		return err
	}

	var namespace string
	switch objT := obj.(type) {
	case *v1.PersistentVolumeClaim:
		namespace = objT.ObjectMeta.Namespace
		err = k.client.CoreV1().PersistentVolumeClaims(namespace).Delete(objT.Name, &metav1.DeleteOptions{})

	case *v1.Pod:
		namespace = objT.ObjectMeta.Namespace
		err = k.client.CoreV1().Pods(namespace).Delete(objT.Name, &metav1.DeleteOptions{})
	}
	return err
}

func buildKubeClient(host string, token string, crt string) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(&rest.Config{
		Host:        host,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(crt),
		},
	})
}
