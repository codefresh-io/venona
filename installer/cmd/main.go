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

	"github.com/golang/glog"
	"github.com/codefresh-io/Isser/installer/pkg/codefresh"
	"github.com/codefresh-io/Isser/installer/pkg/runtime"
)

var (

	codefreshAPIKey = flag.String("api-key", "", "Codefresh api key (token)")
	codefreshURL = flag.String("url", codefresh.DefaultURL, "Codefresh url")

	//runtimeType = flag.String("runtime-type", runtime.TypeKubernetesDind, "Runtime Environement Type")
	runtimeType = runtime.TypeKubernetesDind
	kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	kubecontext = flag.String("kubecontext", "", "Kubeconfig context name")
	
	namespace = flag.String("namespace", "default", "Kubernetes namespace")
	clusterName = flag.String("cluster-name", "", "Cluster Name registered in Codefresh")
)

func getRuntimeConfig() (*runtime.Config, error) {
   
	var clientConfig runtime.ClientConfig
	if runtimeType == runtime.TypeKubernetesDind {
		clientConfig = runtime.ClientConfig{
			KubeClient: runtime.KubernetesClientConfig{
				Kubeconfig: *kubeconfig,
				Context: *kubecontext,
				Namespace: *namespace,
			},
		}
	} else {
		return nil, fmt.Errorf("Unknown runtime type %s", runtimeType)
	}
	
	runtimeConfig := &runtime.Config{
			Type: runtimeType,
			Name: *clusterName,
			Client: clientConfig,
    }	
    return runtimeConfig, nil
}

func main() {
  flag.Parse()
  flag.Set("v", "4")
  flag.Set("alsologtostderr", "true")
  glog.V(4).Infof("Entering\n codefreshUrl = %s \n clusterName = %s ", *codefreshURL, clusterName)
  // Validate Flags

  runtimeConfig, err := getRuntimeConfig()
  if err != nil {
	 fmt.Printf("Error: %v", err)
	 os.Exit(1)
  }
  
  cfAPI, _ := codefresh.NewCfAPI(*codefreshURL, *codefreshAPIKey)

  err = cfAPI.Validate(runtimeConfig)
  if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)	  
  }

  err = cfAPI.Sign(runtimeConfig)
  if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)	  
  }

	installer, err := runtime.GetInstaller(runtimeConfig)
  if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)	  
	}
		
  err = installer.Install(runtimeConfig)
  if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)	  
  }

  err = cfAPI.Register(runtimeConfig)
  if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
  }

  fmt.Printf("Installation completed Successfully")
}