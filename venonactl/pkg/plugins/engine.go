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
	"context"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// enginePlugin installs assets on Kubernetes Dind runtimectl Env
type enginePlugin struct {
	logger logger.Logger
}

const (
	engineFilesPattern = ".*.engine.yaml"
)

// Install venona agent
func (u *enginePlugin) Install(ctx context.Context, opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(ctx, cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}
	return v, install(ctx, &installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   engineFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   EnginePluginType,
	})
}

// Status of runtimectl environment
func (u *enginePlugin) Status(ctx context.Context, statusOpt *StatusOptions, v Values) ([][]string, error) {
	cs, err := statusOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	opt := &statusOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      statusOpt.ClusterNamespace,
		matchPattern:   engineFilesPattern,
		operatorType:   EnginePluginType,
	}
	return status(ctx, opt)
}

func (u *enginePlugin) Delete(ctx context.Context, deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil
	}
	opt := &deleteOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   engineFilesPattern,
		operatorType:   EnginePluginType,
	}
	return uninstall(ctx, opt)
}

func (u *enginePlugin) Upgrade(ctx context.Context, opt *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}

func (u *enginePlugin) Migrate(context.Context, *MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *enginePlugin) Test(ctx context.Context, opt *TestOptions, v Values) error {
	return nil
}

func (u *enginePlugin) Name() string {
	return EnginePluginType
}
