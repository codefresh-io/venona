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

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/codefresh-io/Isser/isserctl/pkg/codefresh"
	"github.com/codefresh-io/Isser/isserctl/pkg/runtimectl"
	"github.com/golang/glog"
)

var (
	codefreshAPIKey = flag.String("api-key", "", "Codefresh api key (token)")
	codefreshURL    = flag.String("url", codefresh.DefaultURL, "Codefresh url")

	//runtimectlType = flag.String("runtimectl-type", runtimectl.TypeKubernetesDind, "runtimectl Environement Type")
	runtimectlType = runtimectl.TypeKubernetesDind
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	kubecontext = flag.String("kubecontext", "", "Kubeconfig context name")

	namespace   = flag.String("namespace", "default", "Kubernetes namespace")
	clusterName = flag.String("cluster-name", "", "Cluster Name registered in Codefresh")
)

func dieIfError(err error) {
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)		
	}
}

func getruntimectlConfig() (*runtimectl.Config, error) {

	var clientConfig runtimectl.ClientConfig
	if *kubeconfig == "" {
		currentUser, _ := user.Current()
		*kubeconfig = path.Join(currentUser.HomeDir, ".kube", "config")
	}
	if runtimectlType == runtimectl.TypeKubernetesDind {
		clientConfig = runtimectl.ClientConfig{
			KubeClient: runtimectl.KubernetesClientConfig{
				Kubeconfig: kubeconfig,
				Context:    kubecontext,
				Namespace:  namespace,
			},
		}
	} else {
		return nil, fmt.Errorf("Unknown runtimectl type %s", runtimectlType)
	}

	runtimectlConfig := &runtimectl.Config{
		Type:   runtimectlType,
		Name:   *clusterName,
		Client: clientConfig,
	}
	return runtimectlConfig, nil
}

func main() {
	flag.Parse()
	flag.Set("v", "4")
	flag.Set("alsologtostderr", "true")
	glog.V(4).Infof("Entering\n codefreshUrl = %s \n clusterName = %s ", *codefreshURL, *clusterName)
	// Validate Flags

	runtimectlConfig, err := getruntimectlConfig()
    dieIfError(err)

	cfAPI, err := codefresh.NewCfAPI(*codefreshURL, *codefreshAPIKey)
	dieIfError(err)
	
	err = cfAPI.Validate(runtimectlConfig)
    dieIfError(err)

	err = cfAPI.Sign(runtimectlConfig)
    dieIfError(err)

	ctl, err := runtimectl.GetCtl(runtimectlConfig)
    dieIfError(err)

	err = ctl.Install(runtimectlConfig)
    dieIfError(err)

	err = cfAPI.Register(runtimectlConfig)
    dieIfError(err)

	fmt.Printf("Installation completed Successfully")
}
