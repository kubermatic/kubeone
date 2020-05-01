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

package etcdstatus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh/sshtunnel"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	healthEndpointFmt = "https://%s:2379/health"
	clientEndpointFmt = "%s:2379"
)

// Report describes status of the etcd cluster
type Report struct {
	Health bool `json:"health,omitempty"`
	Member bool `json:"member,omitempty"`
}

func MemberList(s *state.State) (*clientv3.MemberListResponse, error) {
	tunnel, err := s.Connector.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return nil, err
	}

	tlsConfig, err := loadTLSConfig(s)
	if err != nil {
		return nil, err
	}

	etcdEndpoints := []string{}
	for _, node := range s.Cluster.Hosts {
		etcdEndpoints = append(etcdEndpoints, fmt.Sprintf(clientEndpointFmt, node.PrivateAddress))
	}

	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		TLS:         tlsConfig,
		Context:     s.Context,
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return tunnel.TunnelTo(ctx, "tcp4", addr)
			}),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to etcd cluster")
	}
	defer etcdcli.Close()

	etcdRing, err := etcdcli.MemberList(s.Context)
	return etcdRing, errors.Wrap(err, "failed etcd/clientv3.MemberList")
}

// Get analyzes health of an etcd cluster member
func Get(s *state.State, node kubeoneapi.HostConfig, etcdRing *clientv3.MemberListResponse) (*Report, error) {
	tlsConfig, err := loadTLSConfig(s)
	if err != nil {
		return nil, err
	}

	roundTripper, err := sshtunnel.NewHTTPTransport(s.Connector, node, tlsConfig)
	if err != nil {
		return nil, err
	}

	// Check etcd member health
	health, err := memberHealth(roundTripper, node.PrivateAddress)
	if err != nil {
		return nil, err
	}

	status := &Report{
		Health: health,
	}

	for _, mem := range etcdRing.Members {
		if mem.Name == node.Hostname {
			status.Member = true
			break
		}
	}

	return status, nil
}

// memberHealth returns health for a requested etcd member
func memberHealth(t http.RoundTripper, nodeAddress string) (bool, error) {
	endpoint := fmt.Sprintf(healthEndpointFmt, nodeAddress)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	request.Header.Set("Content-type", "application/json")

	httpClient := http.Client{Transport: t}
	resp, err := httpClient.Do(request)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	h := &struct {
		Health string `json:"health"`
	}{}

	if err = json.Unmarshal(body, &h); err != nil {
		return false, err
	}

	return strconv.ParseBool(h.Health)
}

// loadTLSConfig creates the tls.Config structure used in an http client to securely connect to etcd
func loadTLSConfig(s *state.State) (*tls.Config, error) {
	caBytes, certBytes, keyBytes, err := downloadEtcdCerts(s)
	if err != nil {
		return nil, err
	}

	// Add certificate and key to the TLS config
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	// Add CA certificate to the TLS config
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caBytes)

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

// downloadEtcdCerts returns CA certificate, certificate, and key used to securely access etcd
func downloadEtcdCerts(s *state.State) ([]byte, []byte, []byte, error) {
	// Connect to the host
	host, err := s.Cluster.Leader()
	if err != nil {
		return nil, nil, nil, err
	}

	conn, err := s.Connector.Connect(host)
	if err != nil {
		return nil, nil, nil, err
	}

	// Download CA
	caCert, _, _, err := conn.Exec("sudo cat /etc/kubernetes/pki/etcd/ca.crt")
	if err != nil {
		return nil, nil, nil, err
	}

	// Download cert
	cert, _, _, err := conn.Exec("sudo cat /etc/kubernetes/pki/etcd/server.crt")
	if err != nil {
		return nil, nil, nil, err
	}

	// Download key
	key, _, _, err := conn.Exec("sudo cat /etc/kubernetes/pki/etcd/server.key")
	if err != nil {
		return nil, nil, nil, err
	}

	return []byte(caCert), []byte(cert), []byte(key), nil
}
