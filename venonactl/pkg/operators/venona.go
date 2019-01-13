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

package operators

import (
	"fmt"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/venona/venonactl/internal"

	"github.com/codefresh-io/venona/venonactl/obj/kubeobj"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	templates "github.com/codefresh-io/venona/venonactl/templates/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// venonaOperator installs assets on Kubernetes Dind runtimectl Env
type venonaOperator struct {
}

const (
	venonaInstallPattern = ".*.venona.yaml"
)

// Install runtimectl environment
func (u *venonaOperator) Install() error {
	s := store.GetStore()
	logrus.Debug("Generating token for agent")
	tokenName := fmt.Sprintf("generated-%s", time.Now().Format("20060102150405"))
	logrus.Debugf("Token candidate name: %s", tokenName)
	token, err := s.CodefreshAPI.Client.Tokens().Create(tokenName, s.RuntimeEnvironment)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	logrus.Debugf(fmt.Sprintf("Created token: %s", token.Value))

	store.GetStore().AgentToken = token.Value
	if err != nil {
		return err
	}

	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, s.BuildValues())
	if err != nil {
		return err
	}

	kubeClientset, err := NewKubeClientset(s)
	if err != nil {
		internal.DieOnError(fmt.Errorf("Cannot create kubernetes clientset: %v\n ", err))
	}
	namespace := s.KubernetesAPI.Namespace
	var createErr error
	var kind, name string
	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(venonaInstallPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaOperatorType,
				"Pattern":  venonaInstallPattern,
			}).Debugf("Skipping installation of %s: pattern not match", fileName)
			continue
		}
		if store.GetStore().DryRun == true {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaOperatorType,
			}).Debugf("Skipping installation of %s due to dry-run flag", fileName)
			continue
		}
		logrus.WithFields(logrus.Fields{
			"Operator":  VenonaOperatorType,
			"Namespace": namespace,
		}).Infof("Installing %s", fileName)
		name, kind, createErr = kubeobj.CreateObject(kubeClientset, obj, namespace)

		if createErr == nil {
			logrus.WithFields(logrus.Fields{
				"Kind": kind,
				"Name": name,
			}).Info("Created")
		} else if statusError, errIsStatusError := createErr.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logrus.WithFields(logrus.Fields{
					"Kind": kind,
					"Name": name,
				}).Info("Already exists")
			} else {
				logrus.WithFields(logrus.Fields{
					"Kind": kind,
					"Name": name,
				}).Errorf("Failed: %v ", statusError)
				return statusError
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"Kind": kind,
				"Name": name,
			}).Errorf("Failed: %v ", createErr)
			return createErr
		}
	}

	return nil
}

// Status of runtimectl environment
func (u *venonaOperator) Status() ([][]string, error) {
	s := store.GetStore()
	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, s.BuildValues())
	if err != nil {
		return nil, err
	}

	kubeClientset, err := NewKubeClientset(s)
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return nil, err
	}
	namespace := s.KubernetesAPI.Namespace
	var getErr error
	var kind, name string
	var rows [][]string
	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(venonaInstallPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaOperatorType,
				"Pattern":  venonaInstallPattern,
			}).Debugf("Skipping status check of %s: pattern not match", fileName)
			continue
		}
		name, kind, getErr = kubeobj.CheckObject(kubeClientset, obj, namespace)
		if getErr == nil {
			rows = append(rows, []string{kind, name, StatusInstalled})
		} else if statusError, errIsStatusError := getErr.(*errors.StatusError); errIsStatusError {
			rows = append(rows, []string{kind, name, StatusNotInstalled, statusError.ErrStatus.Message})
		} else {
			fmt.Printf("%s \"%s\" failed: %v ", kind, name, getErr)
			return nil, getErr
		}
	}

	return rows, nil
}

func (u *venonaOperator) Delete() error {
	s := store.GetStore()
	templatesMap := templates.TemplatesMap()
	kubeObjects, err := KubeObjectsFromTemplates(templatesMap, s.BuildValues())
	if err != nil {
		return err
	}

	kubeClientset, err := NewKubeClientset(s)
	if err != nil {
		logrus.Errorf("Cannot create kubernetes clientset: %v\n ", err)
		return err
	}
	namespace := s.KubernetesAPI.Namespace
	var kind, name string
	var deleteError error
	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(venonaInstallPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": VenonaOperatorType,
				"Pattern":  venonaInstallPattern,
			}).Debugf("Skipping deletion of %s: pattern not match", fileName)
			continue
		}
		kind, name, deleteError = kubeobj.DeleteObject(kubeClientset, obj, namespace)
		if deleteError == nil {
			logrus.Debugf("%s \"%s\" deleted\n ", kind, name)
		} else if statusError, errIsStatusError := deleteError.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logrus.Debugf("%s \"%s\" already exists\n", kind, name)
			} else if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
				logrus.Debugf("%s \"%s\" not found\n", kind, name)
			} else {
				logrus.Errorf("%s \"%s\" failed: %v ", kind, name, statusError)
				return statusError
			}
		} else {
			logrus.Errorf("%s \"%s\" failed: %v ", kind, name, deleteError)
			return deleteError
		}
	}
	return nil
}
