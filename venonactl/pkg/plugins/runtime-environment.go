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

package plugins

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/venona/venonactl/pkg/store"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// RuntimeEnvironmentOperator installs assets on Kubernetes Dind runtimectl Env
type RuntimeEnvironmentOperator struct {
}

const (
	RuntimeInstallPattern = ".*.re.yaml"
)

// Install runtimectl environment
func (u *RuntimeEnvironmentOperator) Install() error {
	s := store.GetStore()
	cs, err := NewKubeClientset(s)
	if err != nil {
		return fmt.Errorf("Cannot create kubernetes clientset: %v\n ", err)
	}
	return install(&installOptions{
		templates:      templates.TemplatesMap(),
		templateValues: s.BuildValues(),
		kubeClientSet:  cs,
		namespace:      s.KubernetesAPI.Namespace,
		matchPattern:   RuntimeInstallPattern,
		operatorType:   RuntimeEnvironmentOperatorType,
		dryRun:         s.DryRun,
	})
}

func (u *RuntimeEnvironmentOperator) Status() ([][]string, error) {
	s := store.GetStore()
	cs, err := NewKubeClientset(s)
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return nil, err
	}
	opt := &statusOptions{
		templates:      templates.TemplatesMap(),
		templateValues: s.BuildValues(),
		kubeClientSet:  cs,
		namespace:      s.KubernetesAPI.Namespace,
		matchPattern:   RuntimeInstallPattern,
		operatorType:   RuntimeEnvironmentOperatorType,
	}
	return status(opt)
}

func (u *RuntimeEnvironmentOperator) Delete() error {
	s := store.GetStore()
	cs, err := NewKubeClientset(s)
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return nil
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: s.BuildValues(),
		kubeClientSet:  cs,
		namespace:      s.KubernetesAPI.Namespace,
		matchPattern:   RuntimeInstallPattern,
		operatorType:   RuntimeEnvironmentOperatorType,
	}
	return delete(opt)
}

func (u *RuntimeEnvironmentOperator) Upgrade() error {
	return nil
}
