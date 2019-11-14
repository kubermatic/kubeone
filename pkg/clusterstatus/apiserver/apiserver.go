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

package apiserver

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	healthzEndpoint    = "https://%s:6443/healthz"
)

type Status struct {
	Health bool `json:"health,omitempty"`
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// CheckAPIServer uses the /healthz endpoint to check are all API server instances healthy
func GetStatus(node kubeoneapi.HostConfig, tunneler ssh.Tunneler) (*Status, error) {
	client := httpClient(tunneler)
	health, err := apiserverHealth(client, node.PrivateAddress)
	if err != nil {
		return nil, err
	}

	return &Status{
		Health: health,
	}, nil
}

// apiserverHealth checks is API server healthy
func apiserverHealth(c HTTPDoer, nodeAddress string) (bool, error) {
	endpoint := fmt.Sprintf(healthzEndpoint, nodeAddress)
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.Do(request)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return string(body) == "ok", nil
}

// httpClient builds an HTTP client used to access the API server
func httpClient(tunneler ssh.Tunneler) HTTPDoer {
	transport := &http.Transport{
		DialContext: tunneler.TunnelTo,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &http.Client{
		Timeout:   defaultHTTPTimeout,
		Transport: transport,
	}
}
