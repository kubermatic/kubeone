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

package etcdutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/v3/clientv3"
	"google.golang.org/grpc"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
	"k8c.io/kubeone/pkg/ssh/sshtunnel"
	"k8c.io/kubeone/pkg/state"
)

// NewClientConfig returns etcd clientv3 Config configured with TLS certificates
// and tunneled over SSH
func NewClientConfig(s *state.State, host kubeone.HostConfig) (*clientv3.Config, error) {
	sshconn, err := s.Connector.Connect(host)
	if err != nil {
		return nil, err
	}

	grpcDialer, err := sshtunnel.NewGRPCDialer(s.Connector, host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create grpc tunnel dialer")
	}

	tlsConf, err := LoadTLSConfig(sshconn)
	if err != nil {
		return nil, err
	}

	return &clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", host.PrivateAddress)},
		TLS:         tlsConf,
		Context:     s.Context,
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
			grpcDialer,
		},
	}, nil
}

// LoadTLSConfig creates the tls.Config structure used securely connect to etcd,
// certificates and key are downloaded over SSH from the
// /etc/kubernetes/pki/etcd/ directory.
func LoadTLSConfig(conn ssh.Connection) (*tls.Config, error) {
	sshfs := sshiofs.New(conn)
	// Download CA
	caCertPem, err := fs.ReadFile(sshfs, "/etc/kubernetes/pki/etcd/ca.crt")
	if err != nil {
		return nil, err
	}

	// Download cert
	certPem, err := fs.ReadFile(sshfs, "/etc/kubernetes/pki/etcd/server.crt")
	if err != nil {
		return nil, err
	}

	// Download key
	keyPem, err := fs.ReadFile(sshfs, "/etc/kubernetes/pki/etcd/server.key")
	if err != nil {
		return nil, err
	}

	// Add certificate and key to the TLS config
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, err
	}

	// Add CA certificate to the TLS config
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPem)

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
