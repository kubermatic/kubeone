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

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/tunnel"
)

// NewClientConfig returns etcd clientv3 Config configured with TLS certificates
// and tunneled over SSH
func NewClientConfig(s *state.State, host kubeoneapi.HostConfig) (*clientv3.Config, error) {
	conn, err := s.Executor.Open(host)
	if err != nil {
		return nil, fail.Etcd(err, "open connection")
	}

	grpcDialer, err := tunnel.NewGRPCDialer(s.Executor, host)
	if err != nil {
		return nil, fail.Etcd(err, "gRPC dialing")
	}

	tlsConf, err := LoadTLSConfig(conn)
	if err != nil {
		return nil, fail.Etcd(err, "TLS config creating")
	}

	var endpoints []string
	if s.Cluster.ClusterNetwork.IPFamily.IsIPv6Primary() {
		if len(host.IPv6Addresses) == 0 {
			return nil, fmt.Errorf("no ipv6 addresses")
		}
		endpoints = []string{fmt.Sprintf("[%s]:2379", host.IPv6Addresses[0])}
	} else {
		endpoints = []string{fmt.Sprintf("%s:2379", host.PrivateAddress)}
	}

	return &clientv3.Config{
		Endpoints:   endpoints,
		TLS:         tlsConf,
		Context:     s.Context,
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(), //nolint:staticcheck
			grpcDialer,
		},
	}, nil
}

// LoadTLSConfig creates the tls.Config structure used securely connect to etcd,
// certificates and key are downloaded over SSH from the
// /etc/kubernetes/pki/etcd/ directory.
func LoadTLSConfig(conn executor.Interface) (*tls.Config, error) {
	virtfs := executorfs.New(conn)
	// Download CA
	caCertPem, err := fs.ReadFile(virtfs, "/etc/kubernetes/pki/etcd/ca.crt")
	if err != nil {
		return nil, err
	}

	// Download cert
	certPem, err := fs.ReadFile(virtfs, "/etc/kubernetes/pki/etcd/server.crt")
	if err != nil {
		return nil, err
	}

	// Download key
	keyPem, err := fs.ReadFile(virtfs, "/etc/kubernetes/pki/etcd/server.key")
	if err != nil {
		return nil, err
	}

	// Add certificate and key to the TLS config
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, fail.Runtime(err, "x509 certificate keypair parsing")
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
