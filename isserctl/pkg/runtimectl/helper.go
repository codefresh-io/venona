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
	"bytes"
	"text/template"
//	"encoding/base64"
	"github.com/hairyhenderson/gomplate"
	dataPkg "github.com/hairyhenderson/gomplate/data"
)


// func b64(s string) string {
// 	data := []byte(s)
// 	return base64.StdEncoding.EncodeToString(data)
// }

// ExecuteTemplate - executes templates in tpl str with config as values 
func ExecuteTemplate(tplStr string, data interface{}) (string, error){

	// gomplate func initializing
	dataSources := []string{}
	dataSourceHeaders := []string{}
	d, _ := dataPkg.NewData(dataSources, dataSourceHeaders)

	template, err := template.New("").Funcs(gomplate.Funcs(d)).Parse(tplStr)
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