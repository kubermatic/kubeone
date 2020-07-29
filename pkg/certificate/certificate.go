/*
Copyright 2019 The KubeOne Authors.

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

package certificate

import (
	"crypto/rsa"
	"crypto/x509"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/configupload"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

// CAKeyPair parses generated PKI CA certificate and key
func CAKeyPair(config *configupload.Configuration) (*rsa.PrivateKey, *x509.Certificate, error) {
	caCert, err := config.Get("pki/ca.crt")
	if err != nil {
		return nil, nil, err
	}

	caKey, err := config.Get("pki/ca.key")
	if err != nil {
		return nil, nil, err
	}

	certs, err := cert.ParseCertsPEM([]byte(caCert))
	if err != nil {
		return nil, nil, err
	}

	if len(certs) == 0 {
		return nil, nil, errors.New("ca.crt does not contain at least one valid certificate")
	}

	possibleKey, err := keyutil.ParsePrivateKeyPEM([]byte(caKey))
	if err != nil {
		return nil, nil, err
	}

	rsaKey, ok := possibleKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("private key is not a RSA private key")
	}

	return rsaKey, certs[0], nil
}
