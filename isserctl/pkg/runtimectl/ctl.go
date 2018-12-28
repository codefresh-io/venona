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

package runtimectl

import (
	"fmt"
)

// Ctl Interface to implement
type Ctl interface {

	// Install runtimectl environment
	Install(*Config) error

	// GetStatus of runtimectl environment
	GetStatus(*Config) (Status, error)

	// Delete runtimectl environment
	Delete(*Config) error
}

// GetCtl Returns right ctl based on Config object
func GetCtl(runtimectlConfig *Config) (Ctl, error) {
	var ctl Ctl
	var err error
	if runtimectlConfig.Type == TypeKubernetesDind {
		ctl = &KubernetesDindCtl{}
	} else {
		err = fmt.Errorf("Unknown runtimectl type %s", runtimectlConfig.Type)
	}
	return ctl, err
}
