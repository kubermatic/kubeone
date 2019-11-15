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

package httptunnel

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/state"
)

type HTTPTunnel struct {
	*http.Client
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

func NewHTTPTunnel(s *state.State, tlsConfig *tls.Config) (*HTTPTunnel, error) {
	tunn, err := s.Connector.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get SSH tunnel")
	}

	transport := &http.Transport{
		DialContext:           tunn.TunnelTo,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &HTTPTunnel{
		Client: &http.Client{
			Transport: transport,
		},
	}, nil
}
