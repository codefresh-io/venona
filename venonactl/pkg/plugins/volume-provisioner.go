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

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// volumeProvisionerPlugin installs assets on Kubernetes Dind runtimectl Env
type volumeProvisionerPlugin struct {
	logger logger.Logger
}

const (
	volumeProvisionerFilesPattern = ".*.vp.yaml"
)

// Install runtimectl environment
func (u *volumeProvisionerPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("Cannot create kubernetes clientset: %v", err)
	}
	return v, install(&installOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   volumeProvisionerFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   VolumeProvisionerPluginType,
		logger:         u.logger,
	})
}

func (u *volumeProvisionerPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	cs, err := statusOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	opt := &statusOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		logger:         u.logger,
		namespace:      statusOpt.ClusterNamespace,
		matchPattern:   volumeProvisionerFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return status(opt)
}

func (u *volumeProvisionerPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		logger:         u.logger,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   volumeProvisionerFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return uninstall(opt)
}

func (u *volumeProvisionerPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {
	var err error
	kubeClientset, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	kubeObjects, err := getKubeObjectsFromTempalte(v, volumeProvisionerFilesPattern, u.logger)
	if err != nil {
		return nil, err
	}
	for _, local := range kubeObjects {

		_, _, err := kubeobj.ReplaceObject(kubeClientset, local, opt.ClusterNamespace)
		if err != nil {
			return nil, err
		}
	}
	return v, nil

}
