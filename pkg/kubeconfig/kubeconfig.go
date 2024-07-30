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

package kubeconfig

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/fs"
	"net"
	"os"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	clientapipublic "k8s.io/client-go/tools/clientcmd/api/latest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	pkiCAcrt = "/etc/kubernetes/pki/ca.crt"
	pkiCAkey = "/etc/kubernetes/pki/ca.key"
)

func GenerateSuperAdmin(st *state.State, cn string, on []string, ttl time.Duration) ([]byte, error) {
	host, err := st.Cluster.Leader()
	if err != nil {
		return nil, err
	}

	conn, err := st.Executor.Open(host)
	if err != nil {
		return nil, err
	}

	caCertsPEM, err := fs.ReadFile(executorfs.New(conn), pkiCAcrt)
	if err != nil {
		return nil, err
	}

	caCerts, err := certutil.ParseCertsPEM(caCertsPEM)
	if err != nil {
		return nil, err
	}
	if len(caCerts) == 0 {
		return nil, fail.NewRuntimeError("no certificates found in %s", pkiCAcrt)
	}

	caKeyPEM, err := fs.ReadFile(executorfs.New(conn), pkiCAkey)
	if err != nil {
		return nil, err
	}

	possibleCAKey, err := keyutil.ParsePrivateKeyPEM(caKeyPEM)
	if err != nil {
		return nil, fail.Runtime(err, "parsing private key %s PEM", pkiCAkey)
	}

	caRSAKey, ok := possibleCAKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fail.NewRuntimeError("type asserting rsa.PrivateKey type", "private key is not a RSA private key")
	}

	superAdminUserKey, err := certificate.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	certCfg := certutil.Config{
		CommonName:   cn,
		Organization: on,
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	superAdminUserCert, err := certificate.NewSignedCert(&certCfg, superAdminUserKey, caCerts[0], caRSAKey, time.Now().Add(ttl))
	if err != nil {
		return nil, err
	}

	superAdminUserKeyPEM, err := keyutil.MarshalPrivateKeyToPEM(superAdminUserKey)
	if err != nil {
		return nil, fail.Runtime(err, "marshal generated private key to PEM")
	}

	superAdminUserCertPEM, err := certutil.EncodeCertificates(superAdminUserCert)
	if err != nil {
		return nil, fail.Runtime(err, "marshal generated certificate to PEM")
	}

	contextName := fmt.Sprintf("%s@%s", cn, st.Cluster.Name)
	kubeconfig := clientapi.NewConfig()
	kubeconfig.Clusters[st.Cluster.Name] = &clientapi.Cluster{
		Server:                   fmt.Sprintf("https://%s:%d", st.Cluster.APIEndpoint.Host, st.Cluster.APIEndpoint.Port),
		CertificateAuthorityData: caCertsPEM,
	}
	kubeconfig.Contexts[contextName] = &clientapi.Context{
		Cluster:  st.Cluster.Name,
		AuthInfo: contextName,
	}
	kubeconfig.AuthInfos[contextName] = &clientapi.AuthInfo{
		ClientCertificateData: superAdminUserCertPEM,
		ClientKeyData:         superAdminUserKeyPEM,
	}
	kubeconfig.CurrentContext = contextName

	var buf bytes.Buffer
	err = clientapipublic.Codec.Encode(kubeconfig, &buf)

	return buf.Bytes(), fail.Runtime(err, "marshalling client kubeconfig")
}

// Download downloads Kubeconfig over SSH
func Download(s *state.State) ([]byte, error) {
	// connect to host
	host, err := s.Cluster.Leader()
	if err != nil {
		return nil, err
	}

	conn, err := s.Executor.Open(host)
	if err != nil {
		return nil, err
	}

	return catKubernetesAdminConf(conn)
}

func catKubernetesAdminConf(conn executor.Interface) ([]byte, error) {
	return fs.ReadFile(executorfs.New(conn), "/etc/kubernetes/admin.conf")
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(s *state.State) error {
	s.Logger.Infoln("Building Kubernetes clientset...")

	kubeconfig, err := Download(s)
	if err != nil {
		return err
	}

	s.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return fail.KubeClient(err, "building config from kubeconfig")
	}

	err = TunnelRestConfig(s, s.RESTConfig)
	if err != nil {
		return err
	}

	dynamicClient, err := client.New(s.RESTConfig, client.Options{})
	if err != nil {
		return fail.KubeClient(err, "building dynamic kubernetes client")
	}

	s.DynamicClient = dynamicClient

	return nil
}

func TunnelRestConfig(s *state.State, rc *rest.Config) error {
	rc.WarningHandler = rest.NewWarningWriter(os.Stderr, rest.WarningWriterOptions{
		Deduplicate: true,
	})

	rc.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		dial := TunnelDialerFactory(s.Executor, s.Cluster.RandomHost())

		return dial(ctx, network, address)
	}

	return nil
}

func TunnelDialerFactory(adapter executor.Adapter, host kubeoneapi.HostConfig) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		tunn, err := adapter.Tunnel(host)
		if err != nil {
			return nil, fail.KubeClient(err, "getting SSH tunnel")
		}

		netConn, err := tunn.TunnelTo(ctx, network, address)
		if err != nil {
			tunn.Close()

			return nil, err
		}

		return netConn, err
	}
}
