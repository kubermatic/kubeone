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

package tunnel

import (
	"crypto/tls"
	"net/http"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
)

// NewHTTPTransport initialize net/http Transport that will use SSH tunnel as
// transport
func NewHTTPTransport(tunneler executor.Adapter, target kubeoneapi.HostConfig, tlsConfig *tls.Config) (http.RoundTripper, error) {
	tunn, err := tunneler.Tunnel(target)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		DialContext:           tunn.TunnelTo,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}, nil
}
