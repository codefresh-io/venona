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

// KubernetesDindInstaller installs assets on Kubernetes Dind Runtime Env
type KubernetesDindInstaller struct {

}

// Install runtime environment  
func(u *KubernetesDindInstaller) Install(*RuntimeConfig) error {

    return nill
}

// GetStatus of runtime environment
func(u *KubernetesDindInstaller) GetStatus(*RuntimeConfig) RuntimeStatus, error{

    runtimeStatus := RuntimeStatus{
        status: statusRunning,
        statusMessage: "",
    }
    return runtimeStatus, nil
}



// Delete runtime environment
func(u *KubernetesDindInstaller) Delete(*RuntimeConfig) error {

    return nil
}