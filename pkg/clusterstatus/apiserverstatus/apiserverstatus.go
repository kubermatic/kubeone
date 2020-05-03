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

package apiserverstatus

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh/sshtunnel"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	healthzEndpoint = "https://%s:6443/healthz"
)

type Report struct {
	Health bool `json:"health,omitempty"`
}

// Get uses the /healthz endpoint to check are all API server instances healthy
func Get(s *state.State, node kubeoneapi.HostConfig) (*Report, error) {
	roundTripper, err := sshtunnel.NewHTTPTransport(s.Connector, node, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	health, err := apiserverHealth(roundTripper, node.PrivateAddress)
	if err != nil {
		return &Report{
			Health: false,
		}, err
	}

	return &Report{Health: health}, nil
}

// apiserverHealth checks is API server healthy
func apiserverHealth(t http.RoundTripper, nodeAddress string) (bool, error) {
	endpoint := fmt.Sprintf(healthzEndpoint, nodeAddress)
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

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

	return string(body) == "ok", nil
}
