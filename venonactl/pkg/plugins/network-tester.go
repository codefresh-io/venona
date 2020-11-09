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
	"io"
	"strings"
	"time"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"github.com/stretchr/objx"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type networkTesterPlugin struct {
	logger logger.Logger
}

const (
	networkTesterFilesPattern = ".*.network-tester.yaml"
	networkTestsTimeout       = 120 * time.Second
	defaultRegistry           = "https://docker.io"
	defaultCodefreshHost      = "https://g.codefresh.io"
	defaultFirebaseHost       = "https://codefresh-prod-public-builds-1.firebaseio.com"
)

var (
	errNetworkTestFailed = errors.New("Cluster network tests failed. If you are using a proxy, run again with the correct http proxy environment variables. For more details run again with --verbose")
)

func (u *networkTesterPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	return nil, fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	return nil, fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	return fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {
	return v, fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Migrate(*MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func prepareTestDomains(v map[string]interface{}) []string {
	testDomains := make([]string, 0, 10)

	vObj := objx.New(v)
	// codefresh host
	cfHost := vObj.Get("CodefreshHost").Str(defaultCodefreshHost)

	// registry
	dockerRegistry := vObj.Get("DockerRegistry").Str(defaultRegistry)

	// logging
	firebaseURL := vObj.Get("Logging.FirebaseHost").Str(defaultFirebaseHost)

	// git url
	if gitProviderURL := vObj.Get("GitProviderURL").Str(); gitProviderURL != "" {
		testDomains = append(testDomains, gitProviderURL)
	}
	testDomains = append(testDomains, dockerRegistry, cfHost, firebaseURL)

	return testDomains
}

func (u *networkTesterPlugin) Test(opt TestOptions, v Values) error {
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return err
	}

	testDomains := prepareTestDomains(v)
	urls := strings.Join(testDomains, ",")
	objx.New(v["NetworkTester"]).Set("AdditionalEnvVars.URLS", urls)

	err = install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   networkTesterFilesPattern,
		operatorType:   NetworkTesterPluginType,
	})
	if err != nil {
		u.logger.Error(fmt.Sprintf("Failed to run network-tester pod: %v", err))
		return err
	}
	// defer cleanup
	defer func() {
		err := uninstall(&deleteOptions{
			templates:      templates.TemplatesMap(),
			templateValues: v,
			kubeClientSet:  cs,
			namespace:      opt.ClusterNamespace,
			matchPattern:   networkTesterFilesPattern,
			operatorType:   NetworkTesterPluginType,
			logger:         u.logger,
		})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Failed to cleanup network-tester pod: %v", err))
		}
	}()

	u.logger.Info("Running network tests...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var podLastState *v1.Pod
	timeoutChan := time.After(networkTestsTimeout)
Loop:
	for {
		select {
		case <-ticker.C:
			u.logger.Debug("Waiting for network tester to finish")
			pod, err := cs.CoreV1().Pods(opt.ClusterNamespace).Get(store.NetworkTesterName, metav1.GetOptions{})
			if err != nil {
				if statusError, errIsStatusError := err.(*kerrors.StatusError); errIsStatusError {
					if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
						u.logger.Debug("Network tester pod not found")
					}
				}
			}
			if pod.Status.ContainerStatuses[0].State.Running != nil {
				u.logger.Debug("Network tester pod: running")
			}
			if pod.Status.ContainerStatuses[0].State.Waiting != nil {
				u.logger.Debug("Network tester pod: waiting")
			}
			if pod.Status.ContainerStatuses[0].State.Terminated != nil {
				u.logger.Debug("Network tester pod: terminated")
				podLastState = pod
				break Loop
			}
		case <-timeoutChan:
			u.logger.Error("Network tests timeout reached!")
			return fmt.Errorf("Network tests timeout reached")
		}
	}

	req := cs.CoreV1().Pods(opt.ClusterNamespace).GetLogs(store.NetworkTesterName, &v1.PodLogOptions{})
	podLogs, err := req.Stream()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Failed to get network-tester pod logs: %v", err))
		return err
	}
	defer podLogs.Close()

	logsBuf := new(bytes.Buffer)
	_, err = io.Copy(logsBuf, podLogs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Failed to read network-tester pod logs: %v", err))
		return err
	}
	logs := strings.Trim(logsBuf.String(), "\n")
	u.logger.Debug(fmt.Sprintf("%s", logs))

	if podLastState.Status.ContainerStatuses[0].State.Terminated.ExitCode != 0 {
		terminationMessage := strings.Trim(podLastState.Status.ContainerStatuses[0].State.Terminated.Message, "\n")
		u.logger.Error(fmt.Sprintf("Network tests failed with: %v", terminationMessage))
		return errNetworkTestFailed
	}

	return nil
}

func (u *networkTesterPlugin) Name() string {
	return NetworkTesterPluginType
}
