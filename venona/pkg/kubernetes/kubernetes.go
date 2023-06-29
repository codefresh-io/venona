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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/task"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var errNotValidType = errors.New("not a valid type")
var kubeDecode = scheme.Codecs.UniversalDeserializer().Decode

type (
	// Kubernetes API client
	Kubernetes interface {
		CreateResource(ctx context.Context, spec interface{}) error
		DeleteResource(ctx context.Context, opt DeleteOptions) error
	}
	// Options for Kubernetes
	Options struct {
		Type     string
		Cert     string
		Token    string
		Host     string
		Insecure bool
	}

	// DeleteOptions to delete resource from the cluster
	DeleteOptions struct {
		Name      string
		Namespace string
		Kind      string
	}

	kube struct {
		client kubernetes.Interface
		logger logger.Logger
	}
)

// NewInCluster build Kubernetes API based on local in cluster runtime
func NewInCluster() (Kubernetes, error) {
	client, err := buildKubeInCluster()
	return &kube{
		client: client,
		logger: logger.New(logger.Options{}),
	}, err
}

// New build Kubernetes API
func New(opt Options) (Kubernetes, error) {
	if opt.Type != "runtime" {
		return nil, errNotValidType
	}
	client, err := buildKubeClient(opt.Host, opt.Token, opt.Cert, opt.Insecure)
	return &kube{
		client: client,
		logger: logger.New(logger.Options{}),
	}, err
}

func (k kube) CreateResource(ctx context.Context, spec interface{}) error {

	bytes, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed marshalling when creating resource: %w", err)
	}

	obj, _, err := kubeDecode(bytes, nil, nil)
	if err != nil {
		return fmt.Errorf("failed decoding when creating resource: %w", err)
	}

	var namespace string
	switch obj := obj.(type) {
	case *v1.PersistentVolumeClaim:
		namespace = obj.Namespace
		_, err = k.client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed creating persistent volume claims: %w", err)
		}
		k.logger.Info("PersistentVolumeClaim has been created", "name", obj.Name)

	case *v1.Pod:
		namespace = obj.Namespace
		_, err = k.client.CoreV1().Pods(namespace).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed creating pod: %w", err)
		}
		k.logger.Info("Pod has been created", "name", obj.Name)

	}
	return err
}

func (k kube) DeleteResource(ctx context.Context, opt DeleteOptions) error {
	switch opt.Kind {
	case task.TypeDeletePVC:
		err := k.client.CoreV1().PersistentVolumeClaims(opt.Namespace).Delete(ctx, opt.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed deleting persistent volume claim: %w", err)
		}
		k.logger.Info("PersistentVolumeClaim has been deleted", "name", opt.Name)

	case task.TypeDeletePod:
		err := k.client.CoreV1().Pods(opt.Namespace).Delete(ctx, opt.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed deleting pod: %w", err)
		}
		k.logger.Info("Pod has been deleted", "name", opt.Name)

	}

	return nil
}

func buildKubeClient(host string, token string, crt string, insecure bool) (kubernetes.Interface, error) {
	var tlsconf rest.TLSClientConfig
	if insecure {
		tlsconf = rest.TLSClientConfig{
			Insecure: true,
		}
	} else {
		tlsconf = rest.TLSClientConfig{
			CAData: []byte(crt),
		}
	}
	return kubernetes.NewForConfig(&rest.Config{
		Host:            host,
		BearerToken:     token,
		TLSClientConfig: tlsconf,
	})
}

func buildKubeInCluster() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
