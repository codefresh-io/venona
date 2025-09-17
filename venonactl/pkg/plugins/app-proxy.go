package plugins

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

import (
	"context"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"github.com/stretchr/objx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type appProxyPlugin struct {
	logger logger.Logger
}

const (
	appProxyFilesPattern = ".*.app-proxy.yaml"
)

func (u *appProxyPlugin) Install(ctx context.Context, opt *InstallOptions, v Values) (Values, error) {

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
	err = install(ctx, &installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   appProxyFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   AppProxyPluginType,
	})
	if err != nil {
		u.logger.Error(fmt.Sprintf("AppProxy installation failed: %v", err))
		return nil, err
	}

	host := objx.New(v["AppProxy"]).Get("Ingress.Host").Str()
	pathPrefix := objx.New(v["AppProxy"]).Get("Ingress.PathPrefix").Str()
	appProxyURL := fmt.Sprintf("https://%v%v", host, pathPrefix)
	u.logger.Info(fmt.Sprintf("app proxy is running at: %v", appProxyURL))
	return v, nil
}

func (u *appProxyPlugin) Status(ctx context.Context, statusOpt *StatusOptions, v Values) ([][]string, error) {
	return [][]string{}, nil
}

func (u *appProxyPlugin) Delete(ctx context.Context, deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   appProxyFilesPattern,
		operatorType:   AppProxyPluginType,
		logger:         u.logger,
	}
	return uninstall(ctx, opt)
}

func (u *appProxyPlugin) Upgrade(ctx context.Context, opt *UpgradeOptions, v Values) (Values, error) {
	kubeClientset, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}

	list, err := kubeClientset.CoreV1().Pods(opt.ClusterNamespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", opt.Name)})
	if err != nil {
		u.logger.Error(fmt.Sprintf("Failed to list app-proxy pods: %v ", err))
		return nil, err
	}
	if len(list.Items) == 0 {
		u.logger.Info("no app-proxy pods found")
		return nil, nil
	}

	for _, pod := range list.Items {
		podName := pod.ObjectMeta.Name
		u.logger.Debug(fmt.Sprintf("Deleting app-proxy pod: %v", podName))
		err = kubeClientset.CoreV1().Pods(opt.ClusterNamespace).Delete(ctx, podName, metav1.DeleteOptions{})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot delete app-proxy pod: %v ", err))
			return nil, err
		}
	}

	return v, nil
}
func (u *appProxyPlugin) Migrate(context.Context, *MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *appProxyPlugin) Test(ctx context.Context, opt *TestOptions, v Values) error {
	return nil
}

func (u *appProxyPlugin) Name() string {
	return AppProxyPluginType
}
