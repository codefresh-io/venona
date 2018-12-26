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
)

// Installer Interface to implement
type Installer interface {

    // Install runtime environment  
    Install(*Config) error

    // GetStatus of runtime environment
    GetStatus(*Config) (Status, error)

    // Delete runtime environment
    Delete(*Config) error
}

// GetInstaller Returns right installer based on Config object
func GetInstaller(runtimeConfig *Config) (Installer, error) {
   var installer Installer
   var err error
   if runtimeConfig.RuntimeType == TypeKubernetesDind {
      installer = &KubernetesDindInstaller{}
   } else {
      err = fmt.Errorf("Unknown runtime type %s", runtimeConfig.RuntimeType)
   }
   return installer, err
}