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
func (u *volumeProvisionerPlugin) Migrate(opt *MigrateOptions, v Values) error {
	return u.Delete(&DeleteOptions{
		ClusterNamespace: opt.ClusterNamespace,
		KubeBuilder:      opt.KubeBuilder,
	}, v)
}

func (u *volumeProvisionerPlugin) Test(opt TestOptions) error {
	validationRequest := validationRequest{
		rbac: []rbacValidation{
			{
				Resource:  "persistentvolumes",
				Verbs:     []string{"get", "list", "watch", "create", "delete", "patch"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource: "persistentvolumeclaims",
				Verbs:    []string{"get", "list", "watch", "update"},
			},
			{
				Resource: "storageclasses",
				Group:    "storage.k8s.io",
				Verbs:    []string{"get", "list", "watch"},
			},
			{
				Resource:  "events",
				Group:     "",
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "secrets",
				Group:     "",
				Verbs:     []string{"get", "list"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource: "nodes",
				Group:    "",
				Verbs:    []string{"get", "list", "watch"},
			},
			{
				Resource:  "pods",
				Group:     "",
				Verbs:     []string{"get", "list", "watch", "create", "delete", "patch"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "endpoints",
				Group:     "",
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "DaemonSet",
				Group:     "",
				Verbs:     []string{"get", "create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "CronJob",
				Group:     "batch",
				Verbs:     []string{"get", "create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "ServiceAccount",
				Group:     "",
				Verbs:     []string{"get", "create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
		},
	}
	return test(testOptions{
		logger:            u.logger,
		kubeBuilder:       opt.KubeBuilder,
		namespace:         opt.ClusterNamespace,
		validationRequest: validationRequest,
	})
}

func (u *volumeProvisionerPlugin) Name() string {
	return VolumeProvisionerPluginType
}
