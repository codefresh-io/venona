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
	"context"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	// import all cloud providers auth clients
	"gopkg.in/yaml.v2"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/Masterminds/sprig"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"

	"github.com/Masterminds/semver"
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
		rbac                 []rbacValidation
	}

	rbacValidation struct {
		Namespace string
		Resource  string
		Verbs     []string
		Group     string
	}
)

var requiredK8sVersion, _ = semver.NewConstraint(">= 1.10.0")

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

func toYAML(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		if len(v.(map[string]interface{})) == 0 {
			return ""
		}
	case []v1.Toleration:
		if len(v.([]v1.Toleration)) == 0 {
			return ""
		}
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func nodeSelectorToString(nodeSelectors map[string]string) string {
	str := ""
	for key, value := range nodeSelectors {

		//str = strings.Join([]string {str, fmt.Sprintf("%s=%s", key, value)} , ",")
		str = fmt.Sprintf("%s,%s=%s", str, key, value)

	}
	return strings.TrimPrefix(str, ",")
}
func tolerationsToSring(tolerations []v1.Toleration) string {
	// [{\"effect\":\"NoSchedule\",\"key\":\"dedicated\",\"value\":\"codefresh\"},{\"effect\":\"NoSchedule\",\"key\":\"dedicated\",\"value\":\"codefresh\"}]
	y, err := yaml.Marshal(&tolerations)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\n%s", string(y))

}

// ExecuteTemplate - executes templates in tpl str with config as values
func ExecuteTemplate(tplStr string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"unescape":                unescape,
		"nodeSelectorParamToYaml": nodeSelectorParamToYaml,
		"toYaml":                  toYAML,
		"isString":                isString,
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
			logger.Debug(fmt.Sprintf("Skipping parsing, pattern does not match %s - %s", pattern, n))
			continue
		}
		logger.Debug(fmt.Sprintf("parsing template %s", n))
		tplEx, err := ExecuteTemplate(tpl, data)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse and execute template %s", n))
			return nil, err
		}

		// we add only non-empty parsedTemplates
		if nonEmptyParsedTemplateFunc(tplEx) {
			parsedTemplates[n] = tplEx
		}
<<<<<<< HEAD
=======
		
>>>>>>> master
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
		logger.Debug(fmt.Sprintf("Deserializing template %s %s", n, objStr))
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Cannot deserialize kuberentes object %s: %v", n, err))
			return nil, err
		}
		logger.Debug(fmt.Sprintf("deserializing template success %s group=%s", n, groupVersionKind.Group))
		kubeObjects[n] = obj
	}
	return kubeObjects, nil
}

func getKubeObjectsFromTempalte(values map[string]interface{}, pattern string, logger logger.Logger) (map[string]runtime.Object, error) {
	templatesMap := templates.TemplatesMap()
	return KubeObjectsFromTemplates(templatesMap, values, pattern, logger)
}

func ensureClusterRequirements(ctx context.Context, client *kubernetes.Clientset, req validationRequest, logger logger.Logger) (validationResult, error) {
	result := validationResult{true, nil}
	specs := []*authv1.SelfSubjectAccessReview{}
	for _, rbac := range req.rbac {
		for _, verb := range rbac.Verbs {
			attr := &authv1.ResourceAttributes{
				Resource: rbac.Resource,
				Verb:     verb,
				Group:    rbac.Group,
			}
			if rbac.Namespace != "" {
				attr.Namespace = rbac.Namespace
			}
			specs = append(specs, &authv1.SelfSubjectAccessReview{
				Spec: authv1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: attr,
				},
			})
		}
	}
	rbacres := testRBAC(ctx, client, specs)
	if len(rbacres) > 0 {
		result.isValid = false
		for _, res := range rbacres {
			result.message = append(result.message, res)
		}
		return result, nil
	}

	v, err := client.ServerVersion()
	if err != nil {
		// should not fail if can't get version
		logger.Warn("Failed to validate kubernetes version", "cause", err)
	} else if res, err := testKubernetesVersion(v); !res {
		if err != nil {
			logger.Warn("Failed to validate kubernetes version", "cause", err)
		} else {
			result.isValid = false
			result.message = append(result.message, fmt.Sprintf("Cluster does not meet the version requirements, minimum supported version is: '1.10.0' found version: '%v'", v))
		}
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
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

func handleValidationResult(res validationResult, logger logger.Logger) error {
	if !res.isValid {
		for _, m := range res.message {
			logger.Error(m)
		}
		return errors.New("Failed to run acceptance test on cluster")
	}

	for _, m := range res.message {
		logger.Warn(m)
	}
	return nil
}

func testKubernetesVersion(version *version.Info) (bool, error) {
	v, err := semver.NewVersion(version.String())
	if err != nil {
		return false, err
	}
	// extract only major, minor and patch
	verStr := fmt.Sprintf("%v.%v.%v", v.Major(), v.Minor(), v.Patch())
	v, err = semver.NewVersion(verStr)
	if err != nil {
		return false, err
	}
	return requiredK8sVersion.Check(v), nil
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

func testRBAC(ctx context.Context, client *kubernetes.Clientset, specs []*authv1.SelfSubjectAccessReview) []string {
	res := []string{}
	for _, sar := range specs {
		resp, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
		if err != nil {
			res = append(res, err.Error())
			continue
		}
		if !resp.Status.Allowed {
			verb := sar.Spec.ResourceAttributes.Verb
			namespace := sar.Spec.ResourceAttributes.Namespace
			resource := sar.Spec.ResourceAttributes.Resource
			group := sar.Spec.ResourceAttributes.Group
			msg := strings.Builder{}
			msg.WriteString(fmt.Sprintf("Insufficient permission, %s %s/%s is not allowed", verb, group, resource))
			if namespace != "" {
				msg.WriteString(fmt.Sprintf(" on namespace %s", namespace))
			}
			res = append(res, msg.String())
		}
	}
	return res
}
