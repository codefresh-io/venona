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

package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

const (
	defaultCertCN = "docker.codefresh.io"
)

// ServerCert contains Server Cert pair
type ServerCert struct {
	Key  string
	Csr  string
	Cert string
	Ca   string
}

// NewServerCert - generates ServerCert with csr
func NewServerCert() (*ServerCert, error) {

	serverCert := &ServerCert{}
	err := serverCert.GenerateCsr()
	return serverCert, err
}

// GenerateCsr - generates csr
func (u *ServerCert) GenerateCsr() error {

	certCN := defaultCertCN
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	subj := pkix.Name{
		CommonName: certCN,
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return err
	}

	// Encoding Csr
	csrBuf := bytes.NewBufferString("")
	pem.Encode(csrBuf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	u.Csr = csrBuf.String()

	// Encoding Key
	keyBuf := bytes.NewBufferString("")
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	pem.Encode(keyBuf, keyBlock)
	u.Key = keyBuf.String()
	return nil
}
