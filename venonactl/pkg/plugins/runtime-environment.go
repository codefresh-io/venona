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

	"github.com/codefresh-io/venona/venonactl/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// runtimeEnvironmentPlugin installs assets on Kubernetes Dind runtimectl Env
type runtimeEnvironmentPlugin struct {
}

const (
	runtimeEnvironmentFilesPattern = ".*.re.yaml"
)

// Install runtimectl environment
func (u *runtimeEnvironmentPlugin) Install(opt *InstallOptions) error {
	s := store.GetStore()
	cs, err := NewKubeClientset(s)
	if err != nil {
		return fmt.Errorf("Cannot create kubernetes clientset: %v\n ", err)
	}

	cfOpt := &codefresh.APIOptions{
		Logger:                logrus.New(),
		CodefreshHost:         opt.CodefreshHost,
		CodefreshToken:        opt.CodefreshToken,
		ClusterName:           opt.ClusterName,
		RegisterWithAgent:     opt.RegisterWithAgent,
		ClusterNamespace:      s.KubernetesAPI.Namespace,
		MarkAsDefault:         opt.MarkAsDefault,
		StorageClass:          opt.StorageClass,
		IsDefaultStorageClass: opt.IsDefaultStorageClass,
	}
	cf := codefresh.NewCodefreshAPI(cfOpt)
	cert, err := cf.Sign()
	if err != nil {
		return err
	}
	s.ServerCert = cert

	if err := cf.Validate(); err != nil {
		return err
	}

	err = install(&installOptions{
		templates:      templates.TemplatesMap(),
		templateValues: s.BuildValues(),
		kubeClientSet:  cs,
		namespace:      s.KubernetesAPI.Namespace,
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
		dryRun:         s.DryRun,
	})
	if err != nil {
		return nil
	}

	re, err := cf.Register()
	if err != nil {
		return err
	}
	s.RuntimeEnvironment = re.Metadata.Name

	return nil
}

func (u *runtimeEnvironmentPlugin) Status(_ *StatusOptions) ([][]string, error) {
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
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
	}
	return status(opt)
}

func (u *runtimeEnvironmentPlugin) Delete(_ *DeleteOptions) error {
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
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
	}
	return delete(opt)
}

func (u *runtimeEnvironmentPlugin) Upgrade(_ *UpgradeOptions) error {
	return nil
}
