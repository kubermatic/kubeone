package certificate

import (
	"crypto/rsa"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/installer/util"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/cert/triple"
)

// CAKeyPair parses generated PKI CA certificate and key
func CAKeyPair(config *util.Configuration) (*triple.KeyPair, error) {
	caCert, err := config.Get("pki/ca.crt")
	if err != nil {
		return nil, err
	}

	caKey, err := config.Get("pki/ca.key")
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCertsPEM([]byte(caCert))
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, errors.New("ca.crt does not contain at least one valid certificate")
	}

	possibleKey, err := cert.ParsePrivateKeyPEM([]byte(caKey))
	if err != nil {
		return nil, err
	}

	rsaKey, ok := possibleKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not a RSA private key")
	}

	return &triple.KeyPair{
		Key:  rsaKey,
		Cert: certs[0],
	}, nil
}
