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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/httptunnel"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	healthEndpoint  = "https://%s:2379/health"
	membersEndpoint = "https://127.0.0.1:2379/v2/members"
)

// Status describes status of the etcd cluster
type Status struct {
	Health bool `json:"health,omitempty"`
	Member bool `json:"member,omitempty"`
}

type healthRaw struct {
	Health string `json:"health"`
}

type membersListRaw struct {
	Members []struct {
		ID         string   `json:"id,omitempty"`
		Name       string   `json:"name,omitempty"`
		PeerURLs   []string `json:"peerURLs,omitempty"`
		ClientURLs []string `json:"clientURLs,omitempty"`
	} `json:"members"`
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// EtcdStatus analyzes health of an etcd cluster
func GetStatus(s *state.State, node kubeoneapi.HostConfig) (*Status, error) {
	tlsCfg, err := loadTLSConfig(s)
	if err != nil {
		return nil, err
	}

	tunneler, err := httptunnel.NewHTTPTunnel(s, tlsCfg)
	if err != nil {
		return nil, err
	}

	etcdRing, err := membersList(tunneler)
	if err != nil {
		return nil, err
	}

	// Check etcd member health
	healthStr, err := memberHealth(tunneler, node.PrivateAddress)
	if err != nil {
		return nil, err
	}

	health, err := strconv.ParseBool(healthStr.Health)
	if err != nil {
		return nil, err
	}

	// Check etcd membership
	status := &Status{
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
func memberHealth(t httptunnel.Doer, nodeAddress string) (*healthRaw, error) {
	endpoint := fmt.Sprintf(healthEndpoint, nodeAddress)
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")

	resp, err := t.Do(request)
	if err != nil {
		return &healthRaw{Health: "false"}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	h := &healthRaw{}
	if err = json.Unmarshal(body, &h); err != nil {
		return nil, err
	}

	return h, nil
}

func membersList(t httptunnel.Doer) (*membersListRaw, error) {
	request, err := http.NewRequest("GET", membersEndpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")

	resp, err := t.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := &membersListRaw{}
	if err = json.Unmarshal(body, m); err != nil {
		return nil, err
	}

	return m, nil
}

// loadTLSConfig creates the tls.Config structure used in an http client to securely connect to etcd
func loadTLSConfig(s *state.State) (*tls.Config, error) {
	caBytes, certBytes, keyBytes, err := downloadEtcdCerts(s)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{}

	// Add certificate and key to the TLS config
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Add CA certificate to the TLS config
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caBytes)
	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
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
