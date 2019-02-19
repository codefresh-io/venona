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

// volumeProvisionerPlugin installs assets on Kubernetes Dind runtimectl Env
type volumeProvisionerPlugin struct {
}

const (
	volumeProvisionerFilesPattern = ".*.vp.yaml"
)

// Install runtimectl environment
func (u *volumeProvisionerPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("Cannot create kubernetes clientset: %v\n ", err)
	}
	return v, install(&installOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   volumeProvisionerFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   VolumeProvisionerPluginType,
	})
}

func (u *volumeProvisionerPlugin) Status(_ *StatusOptions) ([][]string, error) {
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
		matchPattern:   volumeProvisionerFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return status(opt)
}

func (u *volumeProvisionerPlugin) Delete(_ *DeleteOptions) error {
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
		matchPattern:   volumeProvisionerFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return delete(opt)
}

func (u *volumeProvisionerPlugin) Upgrade(_ *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}
