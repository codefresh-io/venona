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
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	// import all cloud providers auth clients
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// ExecuteTemplate - executes templates in tpl str with config as values
func ExecuteTemplate(tplStr string, data interface{}) (string, error) {

	template, err := template.New("").Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = template.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ParseTemplates - parses and exexute templates and return map of strings with obj data
func ParseTemplates(templatesMap map[string]string, data interface{}, pattern string) (map[string]string, error) {
	parsedTemplates := make(map[string]string)
	for n, tpl := range templatesMap {
		match, _ := regexp.MatchString(pattern, n)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Pattern": pattern,
			}).Debugf("Skipping parsing of %s: pattern not match", n)
			continue
		}
		logrus.Debugf("parsing template = %s: ", n)
		tplEx, err := ExecuteTemplate(tpl, data)
		if err != nil {
			logrus.Errorf("Cannot parse and execute template %s: %v\n ", n, err)
			return nil, err
		}
		logrus.Debugf("parsing template Success\n")
		parsedTemplates[n] = tplEx
	}
	return parsedTemplates, nil
}

// KubeObjectsFromTemplates return map of runtime.Objects from templateMap
// see https://github.com/kubernetes/client-go/issues/193 for examples
func KubeObjectsFromTemplates(templatesMap map[string]string, data interface{}, pattern string) (map[string]runtime.Object, error) {
	parsedTemplates, err := ParseTemplates(templatesMap, data, pattern)
	if err != nil {
		return nil, err
	}

	// Deserializing all kube objects from parsedTemplates
	// see https://github.com/kubernetes/client-go/issues/193 for examples
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	kubeObjects := make(map[string]runtime.Object)
	for n, objStr := range parsedTemplates {
		logrus.Debugf("Deserializing template = %s: \n", n)
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
		if err != nil {
			logrus.Errorf("Error: %v \n", err)
			fmt.Printf("Cannot deserialize kuberentes object %s: %v\n ", n, err)
			return nil, err
		}
		logrus.Debugf("deserializing template %s Success: %v\n", n, groupVersionKind)
		kubeObjects[n] = obj
	}
	return kubeObjects, nil
}

func getKubeObjectsFromTempalte(values map[string]interface{}, pattern string) (map[string]runtime.Object, error) {
	templatesMap := templates.TemplatesMap()
	return KubeObjectsFromTemplates(templatesMap, values, pattern)
}
