package pkigen

import (
	"crypto/rsa"
	"crypto/x509"

	k8scert "k8s.io/client-go/util/cert"
)

// NewCA generate new CA
func NewCA(cn string, privateKey *rsa.PrivateKey) (*CA, error) {
	cert, err := k8scert.NewSelfSignedCACert(k8scert.Config{CommonName: cn}, privateKey)
	if err != nil {
		return nil, err
	}

	return &CA{
		privateKey: privateKey,
		cert:       cert,
	}, nil
}

// CA hold together certificate and private key
type CA struct {
	privateKey *rsa.PrivateKey
	cert       *x509.Certificate
}

// Certificate return PEM encoded certificate
func (ca *CA) Certificate() string {
	return string(k8scert.EncodeCertPEM(ca.cert))
}

// Key return PEM encode private key
func (ca *CA) Key() string {
	return string(k8scert.EncodePrivateKeyPEM(ca.privateKey))
}
