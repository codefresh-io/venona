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
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	// import all cloud providers auth clients
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/Masterminds/sprig"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"

	ver "github.com/hashicorp/go-version"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes/scheme"
)

type (
	validationResult struct {
		isValid bool
		message []string
	}

	validationRequest struct {
		cpu                  string
		localDiskMinimumSize string
		momorySize           string
	}
)

var requiredK8sVersion, _ = ver.NewConstraint(">= 1.17")

func unescape(s string) template.HTML {
	return template.HTML(s)
}

// template function to parse values for nodeSelector in form "key1=value1,key2=value2"
func nodeSelectorParamToYaml(ns string) string {
	nodeSelectorParts := strings.Split(ns, ",")
	var nodeSelectorYaml string
	for _, p := range nodeSelectorParts {
		pSplit := strings.Split(p, "=")
		if len(pSplit) != 2 {
			continue
		}

		if len(nodeSelectorYaml) > 0 {
			nodeSelectorYaml += "\n"
		}
		nodeSelectorYaml += fmt.Sprintf("%s: %q", pSplit[0], pSplit[1])
	}
	return nodeSelectorYaml
}

// ExecuteTemplate - executes templates in tpl str with config as values
func ExecuteTemplate(tplStr string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"unescape":                unescape,
		"nodeSelectorParamToYaml": nodeSelectorParamToYaml,
	}
	template, err := template.New("base").Funcs(sprig.FuncMap()).Funcs(funcMap).Parse(tplStr)
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
func ParseTemplates(templatesMap map[string]string, data interface{}, pattern string, logger logger.Logger) (map[string]string, error) {
	parsedTemplates := make(map[string]string)
	nonEmptyParsedTemplateFunc := regexp.MustCompile(`[a-zA-Z0-9]`).MatchString
	for n, tpl := range templatesMap {
		match, _ := regexp.MatchString(pattern, n)
		if match != true {
			logger.Debug("Skipping parsing, pattern does not match", "Pattern", pattern, "Name", n)
			continue
		}
		logger.Debug("parsing template", "Name", n)
		tplEx, err := ExecuteTemplate(tpl, data)
		if err != nil {
			logger.Error("Failed to parse and execute template", "Name", n)
			return nil, err
		}

		// we add only non-empty parsedTemplates
		if nonEmptyParsedTemplateFunc(tplEx) {
			parsedTemplates[n] = tplEx
		}
	}
	return parsedTemplates, nil
}

// KubeObjectsFromTemplates return map of runtime.Objects from templateMap
// see https://github.com/kubernetes/client-go/issues/193 for examples
func KubeObjectsFromTemplates(templatesMap map[string]string, data interface{}, pattern string, logger logger.Logger) (map[string]runtime.Object, error) {
	parsedTemplates, err := ParseTemplates(templatesMap, data, pattern, logger)
	if err != nil {
		return nil, err
	}

	// Deserializing all kube objects from parsedTemplates
	// see https://github.com/kubernetes/client-go/issues/193 for examples
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	kubeObjects := make(map[string]runtime.Object)
	for n, objStr := range parsedTemplates {
		logger.Debug("Deserializing template", "Name", n)
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Cannot deserialize kuberentes object %s: %v", n, err))
			return nil, err
		}
		logger.Debug("deserializing template success", "Name", n, "Group", groupVersionKind.Group)
		kubeObjects[n] = obj
	}
	return kubeObjects, nil
}

func getKubeObjectsFromTempalte(values map[string]interface{}, pattern string, logger logger.Logger) (map[string]runtime.Object, error) {
	templatesMap := templates.TemplatesMap()
	return KubeObjectsFromTemplates(templatesMap, values, pattern, logger)
}

func ensureClusterRequirements(client *kubernetes.Clientset, req validationRequest, logger logger.Logger) (validationResult, error) {
	result := validationResult{}
	result.isValid = true

	v, err := client.ServerVersion()
	if err != nil {
		// should not fail if can't validate version
		logger.Warn("Failed to validate kubernetes version", "cause", err)
	}
	res := testKubernetesVersion(v)
	if !res {
		result.isValid = false
		result.message = append(result.message, "Cluster does not meet the kubernetes version requirements")
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return result, err
	}
	if nodes == nil {
		return result, errors.New("Nodes not found")
	}

	if len(nodes.Items) == 0 {
		result.message = append(result.message, "No nodes in cluster")
		result.isValid = false
	}

	atLeastOneMet := false
	for _, n := range nodes.Items {
		res := testNode(n, req)
		if len(res) > 0 {
			result.message = append(result.message, res...)
		} else {
			atLeastOneMet = true
		}
	}
	if !atLeastOneMet {
		result.isValid = false
		return result, nil
	}
	return result, nil
}

func testKubernetesVersion(version *version.Info) bool {
	v, _ := ver.NewVersion(version.String())
	return requiredK8sVersion.Check(v)
}

func testNode(n v1.Node, req validationRequest) []string {
	result := []string{}

	if req.cpu != "" {
		requiredCPU, err := resource.ParseQuantity(req.cpu)
		if err != nil {
			result = append(result, err.Error())
			return result
		}
		cpu := n.Status.Capacity.Cpu()

		if cpu != nil && cpu.Cmp(requiredCPU) == -1 {
			msg := fmt.Sprintf("Insufficiant CPU on node %s, current: %s - required: %s", n.GetObjectMeta().GetName(), cpu.String(), requiredCPU.String())
			result = append(result, msg)
		}
	}

	if req.momorySize != "" {
		requiredMemory, err := resource.ParseQuantity(req.momorySize)
		if err != nil {
			result = append(result, err.Error())
			return result
		}
		memory := n.Status.Capacity.Memory()
		if memory != nil && memory.Cmp(requiredMemory) == -1 {
			msg := fmt.Sprintf("Insufficiant Memory on node %s, current: %s - required: %s", n.GetObjectMeta().GetName(), memory.String(), requiredMemory.String())
			result = append(result, msg)
		}
	}

	return result
}
