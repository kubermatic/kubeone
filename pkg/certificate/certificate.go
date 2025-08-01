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
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/configupload"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/runner"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

// CAKeyPair parses generated PKI CA certificate and key
func CAKeyPair(config *configupload.Configuration) (*rsa.PrivateKey, *x509.Certificate, error) {
	caCert, found := config.KubernetesPKI[KubernetesCACertPath]
	if !found {
		return nil, nil, fail.RuntimeError{
			Op: "getting CA certificate from internal kubernetes PKI",
			Err: errors.WithStack(&os.PathError{
				Op:   "read",
				Path: KubernetesCACertPath,
				Err:  fmt.Errorf("not found"),
			}),
		}
	}

	caKey, found := config.KubernetesPKI[KubernetesCAKeyPath]
	if !found {
		return nil, nil, fail.RuntimeError{
			Op: "getting CA key from internal kubernetes PKI",
			Err: errors.WithStack(&os.PathError{
				Op:   "read",
				Path: KubernetesCAKeyPath,
				Err:  fmt.Errorf("not found"),
			}),
		}
	}

	certs, err := certutil.ParseCertsPEM(caCert)
	if err != nil {
		return nil, nil, fail.Runtime(err, "parsing kubernetes CA certificate PEM")
	}

	if len(certs) == 0 {
		return nil, nil, fail.Runtime(fmt.Errorf("does not contain at least one valid certificate"), "parsing kubernetes CA certificate PEM")
	}

	possibleKey, err := keyutil.ParsePrivateKeyPEM(caKey)
	if err != nil {
		return nil, nil, fail.Runtime(err, "parsing kubernetes CA key")
	}

	rsaKey, ok := possibleKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fail.Runtime(fmt.Errorf("private key is not a RSA private key"), "parsing kubernetes CA key")
	}

	return rsaKey, certs[0], nil
}

func NewSignedKubernetesServiceTLSCert(name, namespace, domain string, caKey crypto.Signer, caCert *x509.Certificate) (map[string]string, error) {
	serviceCommonName := strings.Join([]string{name, namespace, "svc"}, ".")
	serviceFQDNCommonName := strings.Join([]string{serviceCommonName, domain, ""}, ".")

	altdnsNames := []string{
		serviceFQDNCommonName,
		serviceCommonName,
	}

	newKPKey, err := NewPrivateKey()
	if err != nil {
		return nil, fail.Runtime(err, "generating RSA private key")
	}

	certCfg := certutil.Config{
		AltNames: certutil.AltNames{
			DNSNames: altdnsNames,
		},
		CommonName: serviceCommonName,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	newKPCert, err := NewSignedCert(&certCfg, newKPKey, caCert, caKey, time.Now().Add(duration365d))
	if err != nil {
		return nil, fail.Runtime(err, "generating certificate")
	}

	return map[string]string{
		resources.TLSCertName:          string(encodeCertPEM(newKPCert)),
		resources.TLSKeyName:           string(encodePrivateKeyPEM(newKPKey)),
		resources.KubernetesCACertName: string(encodeCertPEM(caCert)),
	}, nil
}

// GetCertificateSANs combines host name and subject alternative names into a list of SANs after transformation
func GetCertificateSANs(host string, alternativeNames []string) []string {
	certSANS := []string{strings.ToLower(host)}
	for _, name := range alternativeNames {
		certSANS = append(certSANS, strings.ToLower(name))
	}

	return certSANS
}

func RenewAll(st *state.State) error {
	return st.RunTaskOnControlPlane(func(ctx *state.State, node *kubeoneapi.HostConfig, _ executor.Interface) error {
		logger := ctx.Logger.WithField("node", node.PublicAddress)
		logger.Infoln("Renew certificates...")

		_, _, err := ctx.Runner.RunRaw("sudo kubeadm certs renew all")
		if err != nil {
			return fail.SSH(err, "renewing certificates")
		}

		pods := []string{
			"etcd",
			"kube-apiserver",
			"kube-controller-manager",
			"kube-scheduler",
		}

		for _, pod := range pods {
			logger.Infof("Restarting %s pod...", pod)
			_, _, err := ctx.Runner.Run(scripts.RestartPodCrictlTemplate, runner.TemplateVariables{
				"NAME": pod,
			})
			if err != nil {
				return fail.SSH(err, "restarting pod %q after renewing certificates", pod)
			}
		}

		return nil
	}, state.RunParallel)
}
