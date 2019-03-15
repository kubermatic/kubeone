package certificate

import (
	"crypto/rsa"
	"crypto/x509"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"

	"k8s.io/client-go/util/cert"
)

// CAKeyPair parses generated PKI CA certificate and key
func CAKeyPair(config *util.Configuration) (*rsa.PrivateKey, *x509.Certificate, error) {
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

	possibleKey, err := cert.ParsePrivateKeyPEM([]byte(caKey))
	if err != nil {
		return nil, nil, err
	}

	rsaKey, ok := possibleKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("private key is not a RSA private key")
	}

	return rsaKey, certs[0], nil
}
