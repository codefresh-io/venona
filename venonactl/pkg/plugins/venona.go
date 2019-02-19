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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// venonaPlugin installs assets on Kubernetes Dind runtimectl Env
type venonaPlugin struct {
}

const (
	venonaFilesPattern = ".*.venona.yaml"
)

// Install venona agent
func (u *venonaPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	logrus.Debug("Generating token for agent")
	tokenName := fmt.Sprintf("generated-%s", time.Now().Format("20060102150405"))
	logrus.Debugf("Token candidate name: %s", tokenName)

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
	logrus.Debugf(fmt.Sprintf("Created token: %s", token.Value))
	v["AgentToken"] = base64.StdEncoding.EncodeToString([]byte(token.Value))
	if err != nil {
		return nil, err
	}

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("Cannot create kubernetes clientset: %v\n ", err)
	}
	return v, install(&installOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   venonaFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   VolumeProvisionerPluginType,
	})
}

// Status of runtimectl environment
func (u *venonaPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	cs, err := statusOpt.KubeBuilder.BuildClient()
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return nil, err
	}
	opt := &statusOptions{
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
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return nil
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   venonaFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return delete(opt)
}

func (u *venonaPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {

	// replace of sa creates new secert with sa creds
	// avoid it till patch fully implemented
	var skipUpgradeFor = map[string]interface{}{
		"service-account.venona.yaml": nil,
	}

	var err error

	kubeClientset, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
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
	v["AgentToken"] = string(token)

	kubeObjects, err := getKubeObjectsFromTempalte(v, venonaFilesPattern)
	if err != nil {
		return nil, err
	}

	for fileName, local := range kubeObjects {
		if _, ok := skipUpgradeFor[fileName]; ok {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaPluginType,
			}).Debugf("Skipping upgrade of %s: should be ignored", fileName)
			continue
		}

		_, _, err := kubeobj.ReplaceObject(kubeClientset, local, opt.ClusterNamespace)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}
