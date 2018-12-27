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
	 
	"net/http"
	"net/url"
	"fmt"
	"encoding/json"
	"bytes"
	"io/ioutil"

	"archive/zip"

	"github.com/golang/glog"
	"github.com/codefresh-io/Isser/installer/pkg/certs"
	"github.com/codefresh-io/Isser/installer/pkg/runtime"
)

const (
	// DefaultURL - by default it is Codefresh Production
	DefaultURL = "https://g.codefresh.io"
	
	codefreshAgent = "isser-installer"
	userAgent = "isser-installer"
)

// CfAPI struct to call Codefresh API
type CfAPI struct {
   url string
   apiKey string    
}

// NewCfAPI - constructs CfAPI
func NewCfAPI(url string, apiKey string) (*CfAPI, error) {
	return &CfAPI {
		url: url,
		apiKey: apiKey,
	}, nil
}

func (u *CfAPI) createCfRequest(path string, reqBodyMap map[string]string) (*http.Request, error) {
	
	reqURL := u.url + "/" + path
    _, err := url.Parse(reqURL)
    if err != nil {
        return nil, err
	}
	
	body, err := json.Marshal(reqBodyMap)
    if err != nil {
        return nil, err
	}
	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", u.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Codefresh-Agent", codefreshAgent)
	req.Header.Add("User-Agent", userAgent)

	return req, nil
}

// Validate calls codefresh API to validate runtimeConfig
func (u *CfAPI) Validate (runtimeConfig *runtime.Config) error {
	
	reqPath := "api/custom_clusters/validate"
	glog.V(4).Infof("Entering codefresh.Validate - %s",  reqPath)

	var reqBodyMap map[string]string
	if runtimeConfig.Type == runtime.TypeKubernetesDind {
		reqBodyMap = make(map[string]string)
		reqBodyMap["clusterName"] = runtimeConfig.Name
		reqBodyMap["namespace"] = runtimeConfig.Client.KubeClient.Namespace
	} else {
		return fmt.Errorf("Unknown runtime type %s", runtimeConfig.Type)
	}

	req, err := u.createCfRequest(reqPath, reqBodyMap)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Validation failed: %s", respBody)
	}

	glog.V(4).Infof("Validation Response %s", respBody)
    return nil
}

// Sign calls codefresh API to sign certificates
func (u *CfAPI) Sign (runtimeConfig *runtime.Config) error {
	
	glog.V(4).Infof("Entering codefresh.Sign" )
	serverCert, err := certs.NewServerCert()
	if err != nil {
		return err
	}

	glog.V(4).Infof("Generated ServerCerts Csr")

	var certExtraSANs string
    if runtimeConfig.Type == runtime.TypeKubernetesDind {
       certExtraSANs = fmt.Sprintf("IP:127.0.0.1,DNS:dind,DNS:*.dind.%s,DNS:*.dind.%s.svc,DNS:*.cf-cd.com,DNS:*.codefresh.io", 
           runtimeConfig.Client.KubeClient.Namespace, runtimeConfig.Client.KubeClient.Namespace)
    } else {
        certExtraSANs = "IP:127.0.0.1,DNS:*.cf-cd.com,DNS:*.codefresh.io"
	}
	glog.V(4).Infof("certExtraSANs = %s", certExtraSANs )

	reqPath := "api/custom_clusters/signServerCerts"
	var reqBodyMap map[string]string
	reqBodyMap = make(map[string]string)
	reqBodyMap["reqSubjectAltName"] = certExtraSANs
	reqBodyMap["csr"] = serverCert.Csr

	glog.V(4).Infof("Submitting request to %s", reqPath )
	req, err := u.createCfRequest(reqPath, reqBodyMap)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Sign certificates failed: %s", respBody)
	}
	
	respBodyReaderAt := bytes.NewReader(respBody)
	zipReader, err := zip.NewReader(respBodyReaderAt, resp.ContentLength)
	for _, zf := range zipReader.File {
		buf := new(bytes.Buffer)
		src, _ := zf.Open()
		defer src.Close()
		buf.ReadFrom(src)
		
		if zf.Name == "cf-ca.pem" {
			serverCert.Ca = buf.String()
		} else if zf.Name == "cf-server-cert.pem" {
			serverCert.Cert = buf.String()
		} else {
			glog.V(4).Infof("Warning: Unknown filename in sign responce %s", zf.Name)
		}
	}   

	// Validating serverCert
	var missingCerts string
	if serverCert.Csr == "" {
		missingCerts += " csr"
	}
	if serverCert.Cert == "" {
		missingCerts += " cert"
	}
	if serverCert.Key == "" {
		missingCerts += " key"
	}
	if serverCert.Ca == "" {
		missingCerts += " ca"
	}
	if missingCerts != "" {
		return fmt.Errorf("Failed to to generate and sign certificates: %s is missing", missingCerts)
	}

	runtimeConfig.ServerCert = serverCert
	return nil
}

// Register calls codefresh API to register runtime environment
func (u *CfAPI) Register (runtimeConfig *runtime.Config) error {

	reqPath := "api/custom_clusters/register"
	glog.V(4).Infof("Entering codefresh.Register - %s",  reqPath)

	var reqBodyMap map[string]string
	if runtimeConfig.Type == runtime.TypeKubernetesDind {
		reqBodyMap = make(map[string]string)
		reqBodyMap["clusterName"] = runtimeConfig.Name
		reqBodyMap["namespace"] = runtimeConfig.Client.KubeClient.Namespace
	} else {
		return fmt.Errorf("Unknown runtime type %s", runtimeConfig.Type)
	}

	req, err := u.createCfRequest(reqPath, reqBodyMap)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Register request failed: %s", respBody)
	}

	glog.V(4).Infof("Register Response %s", respBody)
    return nil
}