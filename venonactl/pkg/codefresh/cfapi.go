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
	"bytes"
	"fmt"

	"archive/zip"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
)

type (
	API interface {
		RuntimeEnvironmentRegistrator
	}

	APIOptions struct {
		Logger                logger
		CodefreshHost         string
		CodefreshToken        string
		ClusterName           string
		ClusterNamespace      string
		RegisterWithAgent     bool
		MarkAsDefault         bool
		StorageClass          string
		IsDefaultStorageClass bool
		KubernetesRunnerType  bool
		BuildNodeSelector     map[string]string
		Annootaions           map[string]string
	}

	RuntimeEnvironmentRegistrator interface {
		Validate() error
		Sign() (*certs.ServerCert, error)
		Register() (*codefresh.RuntimeEnvironment, error)
	}

	api struct {
		logger                logger
		codefresh             codefresh.Codefresh
		clustername           string
		clusternamespace      string
		registerWithAgent     bool
		markAsDefault         bool
		storageClass          string
		isDefaultStorageClass bool
		kubernetesRunnerType  bool
		buildNodeSelector     map[string]string
		annotations           map[string]string
	}

	logger interface {
		Debug(message string, args ...interface{})
	}
)

// NewCodefreshAPI - creates new codefresh api
func NewCodefreshAPI(opt *APIOptions) API {
	return &api{
		logger: opt.Logger,
		codefresh: codefresh.New(&codefresh.ClientOptions{
			Auth: codefresh.AuthOptions{
				Token: opt.CodefreshToken,
			},
			Host: opt.CodefreshHost,
		}),
		clustername:           opt.ClusterName,
		clusternamespace:      opt.ClusterNamespace,
		registerWithAgent:     opt.RegisterWithAgent,
		storageClass:          opt.StorageClass,
		isDefaultStorageClass: opt.IsDefaultStorageClass,
		kubernetesRunnerType:  opt.KubernetesRunnerType,
		buildNodeSelector:     opt.BuildNodeSelector,
		annotations:           opt.Annootaions,
	}
}

func (a *api) Validate() error {
	a.logger.Debug("Validating runtime-environment")
	opt := codefresh.ValidateRuntimeOptions{
		Cluster:   a.clustername,
		Namespace: a.clusternamespace,
	}
	return a.codefresh.RuntimeEnvironments().Validate(&opt)
}

func (a *api) Sign() (*certs.ServerCert, error) {
	a.logger.Debug("Signing runtime-environment")
	serverCert, err := certs.NewServerCert()
	if err != nil {
		return nil, err
	}
	certExtraSANs := fmt.Sprintf("IP:127.0.0.1,DNS:dind,DNS:*.dind.%s,DNS:*.dind.%s.svc,DNS:*.cf-cd.com,DNS:*.codefresh.io", a.clusternamespace, a.clusternamespace)
	a.logger.Debug(fmt.Sprintf("certExtraSANs = %s", certExtraSANs))

	byteArray, err := a.codefresh.RuntimeEnvironments().SignCertificate(&codefresh.SignCertificatesOptions{
		AltName: certExtraSANs,
		CSR:     serverCert.Csr,
	})
	if err != nil {
		a.logger.Debug("Failed to sign certificate")
		return nil, err
	}

	a.logger.Debug("Certificate was signed")
	respBodyReaderAt := bytes.NewReader(byteArray)
	zipReader, err := zip.NewReader(respBodyReaderAt, int64(len(byteArray)))
	if err != nil {
		a.logger.Debug("Failed to create zip reader from given certificate")
		return nil, err
	}
	for _, zf := range zipReader.File {
		a.logger.Debug("Reading file ", "name", zf.Name)
		buf := new(bytes.Buffer)
		src, _ := zf.Open()
		defer src.Close()
		buf.ReadFrom(src)

		if zf.Name == "cf-ca.pem" {
			serverCert.Ca = buf.String()
		} else if zf.Name == "cf-server-cert.pem" {
			serverCert.Cert = buf.String()
		} else {
			a.logger.Debug(fmt.Sprintf("Warning: Unknown filename in sign responce %s", zf.Name))
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
		return nil, fmt.Errorf("Failed to to generate and sign certificates: %s is missing", missingCerts)
	}

	// update store with certs
	return serverCert, nil
}

func (a *api) Register() (*codefresh.RuntimeEnvironment, error) {
	a.logger.Debug("Registering runtime-environment")
	options := &codefresh.CreateRuntimeOptions{
		Namespace: a.clusternamespace,
		HasAgent:  a.registerWithAgent,
		Cluster:   a.clustername,
	}
	if a.kubernetesRunnerType {
		options.RunnerType = codefresh.KubernetesRunnerType
	}
	if len(a.buildNodeSelector) != 0 {
		options.NodeSelector = a.buildNodeSelector
	}

	options.StorageClass = fmt.Sprintf("%s-%s", a.storageClass, a.clusternamespace)
	if !a.isDefaultStorageClass {
		options.StorageClass = a.storageClass
	}

	if len(a.annotations) != 0 {
		options.Annotations = a.annotations
	}

	re, err := a.codefresh.RuntimeEnvironments().Create(options)
	if err != nil {
		return nil, err
	}

	if a.markAsDefault {
		a.logger.Debug("Setting runtime as deault")
		_, err := a.codefresh.RuntimeEnvironments().Default(re.Metadata.Name)
		if err != nil {
			return nil, err
		}
	}

	return re, nil
}
