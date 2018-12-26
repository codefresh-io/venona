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

package codefresh

import (

	//"github.com/golang/glog"
	//"github.com/codefresh-io/Isser/installer/pkg/certs"
	"github.com/codefresh-io/Isser/installer/pkg/runtime"
)

const (
	// DefaultURL - by default it is Codefresh Production
	DefaultURL = "https://g.codefresh.io"
	
	//runtimeTypeDockerd = "dockerd"
)

// CfAPI struct to call Codefresh API
type CfAPI struct {
   URL string
   APIKey string    
}

// Validate calls codefresh API to validate runtimeConfig
func (u *CfAPI) Validate (runtimeConfig *runtime.Config) error {

    return nil
}

// Sign calls codefresh API to sign certificates
func (u *CfAPI) Sign (runtimeConfig *runtime.Config) error {

    return nil
}

// Register calls codefresh API to register runtime environment
func (u *CfAPI) Register (runtimeConfig *runtime.Config) error {

    return nil
}