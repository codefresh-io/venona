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
	"regexp"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
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

	store.GetStore().AgentToken = token.Value
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
func (u *venonaPlugin) Status(_ *StatusOptions) ([][]string, error) {
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
		matchPattern:   venonaFilesPattern,
		operatorType:   VenonaPluginType,
	}
	return status(opt)
}

func (u *venonaPlugin) Delete(_ *DeleteOptions) error {
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
		matchPattern:   venonaFilesPattern,
		operatorType:   VolumeProvisionerPluginType,
	}
	return delete(opt)
}

func (u *venonaPlugin) Upgrade(_ *UpgradeOptions) error {

	// replace of sa creates new secert with sa creds
	// avoid it till patch fully implemented
	var skipUpgradeFor = map[string]interface{}{
		"service-account.venona.yaml": nil,
	}

	var err error
	s := store.GetStore()

	kubeClientset, err := NewKubeClientset(s)
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return err
	}

	namespace := s.KubernetesAPI.Namespace

	// special case when we need to get the token from the remote to no regenrate it
	// whole flow should be more like kubectl apply that build a patch
	// based on remote object and candidate object
	secret, err := kubeClientset.CoreV1().Secrets(namespace).Get(s.AppName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	token := secret.Data["codefresh.token"]
	s.AgentToken = string(token)

	kubeObjects, err := getKubeObjectsFromTempalte(s.BuildValues())
	if err != nil {
		return err
	}

	for fileName, local := range kubeObjects {
		match, _ := regexp.MatchString(venonaFilesPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaPluginType,
				"Pattern":  venonaFilesPattern,
			}).Debugf("Skipping upgrade of %s: pattern not match", fileName)
			continue
		}

		if _, ok := skipUpgradeFor[fileName]; ok {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaPluginType,
			}).Debugf("Skipping upgrade of %s: should be ignored", fileName)
			continue
		}

		_, _, err := kubeobj.ReplaceObject(kubeClientset, local, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}
