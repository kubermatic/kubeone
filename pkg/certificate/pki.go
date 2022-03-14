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

// Upstream PR https://github.com/kubernetes/kubernetes/pull/73198 moved few
// functions from "k8s.io/client-go/util/cert" to
// "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil".
//
// to avoid importing "k8s.io/kubernetes" those functions are copied over here.

package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/fail"

	certutil "k8s.io/client-go/util/cert"
)

const (
	// PrivateKeyBlockType is a possible value for pem.Block.Type.
	PrivateKeyBlockType = "PRIVATE KEY"
	// PublicKeyBlockType is a possible value for pem.Block.Type.
	PublicKeyBlockType = "PUBLIC KEY"
	// CertificateBlockType is a possible value for pem.Block.Type.
	CertificateBlockType = "CERTIFICATE"
	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
	rsaKeySize             = 2048
	duration365d           = time.Hour * 24 * 365
)

// encodeCertPEM returns PEM-endcoded certificate data
func encodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  CertificateBlockType,
		Bytes: cert.Raw,
	}

	return pem.EncodeToMemory(&block)
}

// EncodePublicKeyPEM returns PEM-encoded public data
func EncodePublicKeyPEM(key *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return []byte{}, err
	}
	block := pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: der,
	}

	return pem.EncodeToMemory(&block), nil
}

// encodePrivateKeyPEM returns PEM-encoded private key data
func encodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.EncodeToMemory(&block)
}

// newPrivateKey creates an RSA private key
func newPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

// newSignedCert creates a signed certificate using the given CA certificate and key
func newSignedCert(cfg *certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, fail.RuntimeError{
			Op:  "checking requested CommonName",
			Err: errors.New("must specify a CommonName"),
		}
	}

	if len(cfg.Usages) == 0 {
		return nil, fail.RuntimeError{
			Op:  "checking certificate ExtKeyUsage",
			Err: errors.New("must specify at least one ExtKeyUsage"),
		}
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, fail.Runtime(err, "creating ASN.1 DER x509 certificate")
	}

	cert, err := x509.ParseCertificate(certDERBytes)

	return cert, fail.Runtime(err, "parsing ASN.1 DEP x509 certificate")
}
