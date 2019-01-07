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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"archive/zip"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
)

const (
	// DefaultURL - by default it is Codefresh Production
	DefaultURL = "https://g.codefresh.io"

	codefreshAgent = "venona-ctl"
	userAgent      = "venona-ctl"
)

// CfAPI struct to call Codefresh API
type CfAPI struct {
}

// New - constructs CfAPI
func New() *CfAPI {
	return &CfAPI{}
}

// Validate calls codefresh API to validate runtimectlConfig
func (u *CfAPI) Validate() error {
	logrus.Debug("Calling codefresh.Validate")
	cf := store.GetStore().CodefreshAPI.Client
	s := store.GetStore()
	opt := &codefresh.ValidateRuntimeOptions{
		Namespace: s.KubernetesAPI.Namespace,
	}
	if s.ClusterInCodefresh != "" {
		opt.Cluster = s.ClusterInCodefresh
	} else {
		opt.Cluster = s.KubernetesAPI.ContextName
	}
	err := cf.ValidateRuntimeEnvironment(opt)

	if err != nil {
		return fmt.Errorf("Validation failed with error: %s", err.Error())
	}

	logrus.Debug("Finished validation")
	return nil
}

// Sign calls codefresh API to sign certificates
func (u *CfAPI) Sign() error {
	logrus.Debug("Entering codefresh.Sign")
	s := store.GetStore()
	serverCert, err := certs.NewServerCert()
	if err != nil {
		return err
	}

	logrus.Debug("Generated ServerCerts Csr")

	var certExtraSANs string
	if "kubernetesDind" == runtimectl.TypeKubernetesDind {
		namespace := s.KubernetesAPI.Namespace
		certExtraSANs = fmt.Sprintf("IP:127.0.0.1,DNS:dind,DNS:*.dind.%s,DNS:*.dind.%s.svc,DNS:*.cf-cd.com,DNS:*.codefresh.io", namespace, namespace)
	} else {
		certExtraSANs = "IP:127.0.0.1,DNS:*.cf-cd.com,DNS:*.codefresh.io"
	}
	logrus.Debugf("certExtraSANs = %s", certExtraSANs)
	byteArray, err := store.GetStore().CodefreshAPI.Client.SignRuntimeEnvironmentCertificate(&codefresh.SignCertificatesOptions{
		AltName: certExtraSANs,
		CSR:     serverCert.Csr,
	})

	respBodyReaderAt := bytes.NewReader(byteArray)
	zipReader, err := zip.NewReader(respBodyReaderAt, int64(len(byteArray)))
	if err != nil {
		return err
	}
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
			logrus.Debugf("Warning: Unknown filename in sign responce %s", zf.Name)
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
	s.ServerCert = serverCert
	return nil
}

// Register calls codefresh API to register runtimectl environment
func (u *CfAPI) Register() error {
	logrus.Debug("Entering codefresh.Register")
	s := store.GetStore()
	options := &codefresh.CreateRuntimeOptions{
		Namespace: s.KubernetesAPI.Namespace,
	}

	if s.ClusterInCodefresh == "" {
		options.HasAgent = true
		options.Cluster = s.KubernetesAPI.ContextName
	} else {
		options.HasAgent = false
		options.Cluster = s.ClusterInCodefresh
	}
	cf := s.CodefreshAPI.Client
	logrus.WithFields(logrus.Fields{
		"Options-Has-Agent": options.HasAgent,
		"Options-Cluster":   options.Cluster,
		"Options-Namespace": options.Namespace,
	}).Debug("Registering runtime environmnet")
	re, err := cf.CreateRuntimeEnvironment(options)

	if err != nil {
		return err
	}

	s.RuntimeEnvironment = re.Name
	logrus.Debugf("Created with name: %s", re.Name)

	return nil
}

func (u *CfAPI) GenerateToken(name string) (string, error) {
	logrus.Debug("Entering codefresh.GenerateToken")
	cf := store.GetStore().CodefreshAPI.Client
	tokenName := fmt.Sprintf("generated-%s", time.Now().Format("20060102150405"))
	re, err := cf.GenerateToken(tokenName, name)
	if err != nil {
		return "", err
	}
	logrus.Debugf(fmt.Sprintf("Created token: %s", re.Value))
	return re.Value, nil
}
