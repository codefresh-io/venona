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


package runtime

import (
    "fmt"
    templates "github.com/codefresh-io/Isser/installer/pkg/runtime/templates/kubernetes_dind"
)

// KubernetesDindInstaller installs assets on Kubernetes Dind Runtime Env
type KubernetesDindInstaller struct {

}

// Install runtime environment  
func(u *KubernetesDindInstaller) Install(*Config) error {

    templatesMap := templates.TemplatesMap()
// https://github.com/kubernetes/client-go/issues/193
    for n, _ := range templatesMap {
       fmt.Printf("template = %s\n", n)
    }
    return nil
}

// GetStatus of runtime environment
func(u *KubernetesDindInstaller) GetStatus(*Config) (Status, error){

    runtimeStatus := Status{
        Status: StatusRunning,
        StatusMessage: "",
    }
    return runtimeStatus, nil
}



// Delete runtime environment
func(u *KubernetesDindInstaller) Delete(*Config) error {

    return nil
}