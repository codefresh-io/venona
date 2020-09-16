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

	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// venonaPlugin installs assets on Kubernetes Dind runtimectl Env
type venonaPlugin struct {
	logger logger.Logger
}

type migrationData struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
}

const (
	venonaFilesPattern = ".*.venona.yaml"
)

// Install venona agent
func (u *venonaPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	if v["AgentToken"] == "" {
		u.logger.Debug("Generating token for agent")
		tokenName := fmt.Sprintf("generated-%s", time.Now().Format("20060102150405"))
		u.logger.Debug(fmt.Sprintf("Token candidate name: %s", tokenName))

		client := codefresh.New(&codefresh.ClientOptions{
			Auth: codefresh.AuthOptions{
				Token: opt.CodefreshToken,
			},
			Host: opt.CodefreshHost,
		})

		token, err := client.Tokens().Create(tokenName, v["RuntimeEnvironment"].(string))
		if err != nil {
			return nil, err
		}
		u.logger.Debug("Token created")
		v["AgentToken"] = token.Value
		if err != nil {
			return nil, err
		}
	}

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}

	return v, install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   venonaFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   VenonaPluginType,
	})
}

// Status of runtimectl environment
func (u *venonaPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
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
		matchPattern:   venonaFilesPattern,
		operatorType:   VenonaPluginType,
	}
	return status(opt)
}

func (u *venonaPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
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
		matchPattern:   venonaFilesPattern,
		operatorType:   VenonaPluginType,
	}
	return uninstall(opt)
}

func (u *venonaPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {

	// replace of sa creates new secert with sa creds
	// avoid it till patch fully implemented
	var skipUpgradeFor = map[string]interface{}{
		"service-account.venona.yaml": nil,
		"deployment.venona.yaml":      nil,
		"venonaconf.secret.venona.yaml": nil,
	}

	var deletePriorUpgrade = map[string]interface{}{
		"deployment.venona.yaml": nil,
	}

	var err error

	kubeClientset, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}

	// special case when we need to get the token from the remote to no regenrate it
	// whole flow should be more like kubectl apply that build a patch
	// based on remote object and candidate object

	secret, err := kubeClientset.CoreV1().Secrets(opt.ClusterNamespace).Get(opt.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	token := secret.Data["codefresh.token"]
	v["AgentToken"] = token

	kubeObjects, err := getKubeObjectsFromTempalte(v, venonaFilesPattern, u.logger)
	if err != nil {
		return nil, err
	}
	v, err = updateValuesBasedOnPreviousDeployment(opt.ClusterNamespace, kubeClientset, v)

	for fileName, local := range kubeObjects {
		if _, ok := deletePriorUpgrade[fileName]; ok {
			u.logger.Debug(fmt.Sprintf("Deleting previous deplopyment of %s", fileName))
			delOpt := &deleteOptions{
				logger:         u.logger,
				templates:      templates.TemplatesMap(),
				templateValues: v,
				kubeClientSet:  kubeClientset,
				namespace:      opt.ClusterNamespace,
				matchPattern:   fileName,
				operatorType:   VenonaPluginType,
			}
			err := uninstall(delOpt)
			if err != nil {
				return nil, err
			}
			installOpt := &installOptions{
				logger:         u.logger,
				templates:      templates.TemplatesMap(),
				templateValues: v,
				kubeClientSet:  kubeClientset,
				namespace:      opt.ClusterNamespace,
				matchPattern:   fileName,
				operatorType:   VenonaPluginType,
			}
			err = install(installOpt)
			if err != nil {
				return nil, err
			}
		}

		if _, ok := skipUpgradeFor[fileName]; ok {
			u.logger.Debug(fmt.Sprintf("Skipping upgrade of %s: should be ignored", fileName))
			continue
		}

		_, _, err := kubeobj.ReplaceObject(kubeClientset, local, opt.ClusterNamespace)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func updateValuesBasedOnPreviousDeployment(ns string, kubeClientset *kubernetes.Clientset, v Values) (Values, error) {

	runnerDeployment, err := kubeClientset.AppsV1().Deployments(ns).Get(AppName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// Update the values with existing deployment values
	if runnerDeployment.Spec.Template.Spec.NodeSelector != nil {

	}
	if runnerDeployment.Spec.Template.Spec.NodeSelector != nil {
		v["NodeSelector"] = nodeSelectorToString(runnerDeployment.Spec.Template.Spec.NodeSelector)
	}
	if runnerDeployment.Spec.Template.Spec.Tolerations != nil {
		v["Tolerations"] = tolerationsToSring(runnerDeployment.Spec.Template.Spec.Tolerations)
	}

	for _, envVar := range runnerDeployment.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "DOCKER_REGISTRY" {
			v["DockerRegistry"] = envVar.Value
		} else if envVar.Name == "AGENT_ID" {
			v["AgentId"] = envVar.Value
		}
	}

	v["AdditionalEnvVars"] = getEnvVarsFromDeployment(runnerDeployment.Spec.Template.Spec.Containers)
	return v, nil

}

func getEnvVarsFromDeployment(containers []v1.Container) map[string]string {
	// Get env for containers
	preDefinedEnvVars := map[string]interface{}{}
	newEnvVars := map[string]string{}
	preDefinedEnvVars["SELF_DEPLOYMENT_NAME"] = "SELF_DEPLOYMENT_NAME"
	preDefinedEnvVars["CODEFRESH_TOKEN"] = "CODEFRESH_TOKEN"
	preDefinedEnvVars["CODEFRESH_HOST"] = "CODEFRESH_HOST"
	preDefinedEnvVars["AGENT_MODE"] = "AGENT_MODE"
	preDefinedEnvVars["AGENT_NAME"] = "AGENT_NAME"
	preDefinedEnvVars["AGENT_ID"] = "AGENT_ID"
	preDefinedEnvVars["VENONA_CONFIG_DIR"] = "VENONA_CONFIG_DIR"

	for _, container := range containers {
		for _, envVar := range container.Env {
			if preDefinedEnvVars[envVar.Name] == nil {
				newEnvVars[envVar.Name] = envVar.Value
			}
		}
	}
	return newEnvVars
}

func (u *venonaPlugin) Migrate(opt *MigrateOptions, v Values) error {
	var deletePriorUpgrade = map[string]interface{}{
		"deployment.venona.yaml": nil,
		"secret.venona.yaml":     nil,
	}

	kubeClientset, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	kubeObjects, err := getKubeObjectsFromTempalte(v, venonaFilesPattern, u.logger)
	if err != nil {
		return err
	}
	list, err := kubeClientset.CoreV1().Pods(opt.ClusterNamespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", v["AppName"])})
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot find agent pod: %v ", err))
		return err
	}
	if len(list.Items) == 0 {
		u.logger.Debug("Runner pod not found , existing migration")
		return nil
	}
	migrationData := migrationData{
		Tolerations:  list.Items[0].Spec.Tolerations,
		NodeSelector: list.Items[0].Spec.NodeSelector,
		Env:          getEnvVarsFromDeployment(list.Items[0].Spec.Containers),
	}
	var jsonData []byte
	jsonData, err = json.Marshal(migrationData)
	err = ioutil.WriteFile("migration.json", jsonData, 0644)
	if err != nil {
		u.logger.Error("Cannot write migration json")
	}

	podName := list.Items[0].ObjectMeta.Name
	for fileName := range kubeObjects {
		if _, ok := deletePriorUpgrade[fileName]; ok {
			u.logger.Debug(fmt.Sprintf("Deleting previous deplopyment of %s", fileName))
			delOpt := &deleteOptions{
				logger:         u.logger,
				templates:      templates.TemplatesMap(),
				templateValues: v,
				kubeClientSet:  kubeClientset,
				namespace:      opt.ClusterNamespace,
				matchPattern:   fileName,
				operatorType:   VenonaPluginType,
			}
			err := uninstall(delOpt)
			if err != nil {
				return err
			}
		}
	}
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			u.logger.Debug("Validating old runner pod termination")
			_, err = kubeClientset.CoreV1().Pods(opt.ClusterNamespace).Get(podName, metav1.GetOptions{})
			if err != nil {
				if statusError, errIsStatusError := err.(*kerrors.StatusError); errIsStatusError {
					if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
						return nil
					}
				}
			}
		case <-time.After(60 * time.Second):
			u.logger.Error("Failed to validate old venona pod termination")
			return fmt.Errorf("Failed to validate old venona pod termination")
		}
	}
}

func (u *venonaPlugin) Test(opt TestOptions) error {
	validationRequest := validationRequest{
		cpu:        "500m",
		momorySize: "1Gi",
		rbac: []rbacValidation{
			{
				Resource:  "deployment",
				Verbs:     []string{"create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource:  "secret",
				Verbs:     []string{"create", "update", "delete"},
				Namespace: opt.ClusterNamespace,
			},
			{
				Resource: "ClusterRoleBinding",
				Group:    "rbac.authorization.k8s.io",
				Verbs:    []string{"create", "update", "delete"},
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

func (u *venonaPlugin) Name() string {
	return VenonaPluginType
}
