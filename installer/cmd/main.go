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
	// "fmt"

	"github.com/golang/glog"
	"github.com/codefresh-io/Isser/installer/pkg/codefresh"
	"github.com/codefresh-io/Isser/installer/pkg/runtime"
)

var (

	codefreshApiKey = flag.String("api-key", "", "Codefresh api key (token)")
	codefreshUrl = flag.String("url", codefresh.defaultURL, "Codefresh url")

	runtimeType = flag.String("runtime-type", runtime.typeKubernetesDind, "Runtime Environement Type")
	
	kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	kubecontext = flag.String("kubecontext", "", "Kubeconfig context name")
)

func getRuntimeConfig() (runtime.RuntimeConfig, error) {
   return nil, nil
}

func main() {
  
  flag.Parse()
  flag.Set("v", "4")
  flag.Set("alsologtostderr", "true")
  // Validate Flags

  runtimeConfig, err := getRuntimeConfig()
  if err != nil {

  }
  
  cfApi := codefresh.CfApi{
	url: codefreshUrl,
	apiKey: codefreshApiKey, 
  }

  err = cfApi.Validate(runtimeConfig)
  if err != nil {
	  
  }

  err = cfApi.Sign(runtimeConfig)
  if err != nil {
	  
  }

  err = runtime.Install(runtimeConfig)
  if err != nil {
	  
  }

  err = cfApi.Register(runtimeConfig)
  if err != nil {
	  
  }
}