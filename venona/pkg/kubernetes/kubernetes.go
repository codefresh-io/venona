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
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/metrics"
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
		CreateResource(ctx context.Context, taskType task.Type, spec interface{}) error
		DeleteResource(ctx context.Context, opts DeleteOptions) error
	}

	// Options for Kubernetes
	Options struct {
		Logger   logger.Logger
		Type     string
		Cert     string
		Token    string
		Host     string
		Insecure bool
		QPS      float32
		Burst    int
	}

	// DeleteOptions to delete resource from the cluster
	DeleteOptions struct {
		Name      string
		Namespace string
		Kind      task.Type
	}

	kube struct {
		client kubernetes.Interface
		log    logger.Logger
	}
)

// NewInCluster build Kubernetes API based on local in cluster runtime
func NewInCluster(log logger.Logger, qps float32, burst int) (Kubernetes, error) {
	client, err := buildKubeInCluster(qps, burst)
	return &kube{
		client: client,
		log:    log,
	}, err
}

// New build Kubernetes API
func New(opts Options) (Kubernetes, error) {
	if opts.Type != "runtime" {
		return nil, errNotValidType
	}

	client, err := buildKubeClient(opts.Host, opts.Token, opts.Cert, opts.Insecure, opts.QPS, opts.Burst)
	return &kube{
		client: client,
		log:    opts.Logger,
	}, err
}

func (k kube) CreateResource(ctx context.Context, taskType task.Type, spec interface{}) error {
	start := time.Now()
	bytes, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed marshalling when creating resource: %w", err)
	}

	obj, _, err := kubeDecode(bytes, nil, nil)
	if err != nil {
		return fmt.Errorf("failed decoding when creating resource: %w", err)
	}

	var namespace, name string
	switch obj := obj.(type) {
	case *v1.PersistentVolumeClaim:
		namespace, name = obj.Namespace, obj.Name
		_, err = k.client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed creating persistent volume claims \"%s\\%s\": %w", namespace, obj.Name, err)
		}
	case *v1.Pod:
		namespace, name = obj.Namespace, obj.Name
		_, err = k.client.CoreV1().Pods(namespace).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed creating pod \"%s\\%s\": %w", namespace, obj.Name, err)
		}

		metrics.IncWorkflowRetries(name)
	default:
		return fmt.Errorf("failed creating resource of type %s", obj.GetObjectKind().GroupVersionKind())
	}

	processed := time.Since(start)
	k.log.Info("Done handling k8s task",
		"type", taskType,
		"namespace", namespace,
		"name", name,
		"processing time", processed,
	)
	metrics.ObserveK8sMetrics(taskType, processed)
	return nil
}

func (k kube) DeleteResource(ctx context.Context, opts DeleteOptions) error {
	start := time.Now()
	switch opts.Kind {
	case task.TypeDeletePVC:
		err := k.client.CoreV1().PersistentVolumeClaims(opts.Namespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed deleting persistent volume claim \"%s\\%s\": %w", opts.Namespace, opts.Name, err)
		}
	case task.TypeDeletePod:
		err := k.client.CoreV1().Pods(opts.Namespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed deleting pod \"%s\\%s\": %w", opts.Namespace, opts.Name, err)
		}
	default:
		return fmt.Errorf("failed deleting resource of type %s", opts.Kind)
	}

	processed := time.Since(start)
	k.log.Info("Done handling k8s task",
		"type", opts.Kind,
		"namespace", opts.Namespace,
		"name", opts.Name,
		"processing time", processed,
	)
	metrics.ObserveK8sMetrics(opts.Kind, processed)
	return nil
}

func buildKubeClient(host string, token string, crt string, insecure bool, qps float32, burst int) (kubernetes.Interface, error) {
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
		QPS:             qps,
		Burst:           burst,
	})
}

func buildKubeInCluster(qps float32, burst int) (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	config.QPS = qps
	config.Burst = burst
	return kubernetes.NewForConfig(config)
}
