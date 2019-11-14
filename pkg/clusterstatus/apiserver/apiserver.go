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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/httptunnel"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	healthzEndpoint = "https://%s:6443/healthz"
)

type Status struct {
	Health bool `json:"health,omitempty"`
}

// CheckAPIServer uses the /healthz endpoint to check are all API server instances healthy
func GetStatus(s *state.State, node kubeoneapi.HostConfig) (*Status, error) {
	client, err := httptunnel.NewHTTPTunnel(s, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	health, err := apiserverHealth(client, node.PrivateAddress)
	if err != nil {
		return nil, err
	}

	return &Status{
		Health: health,
	}, nil
}

// apiserverHealth checks is API server healthy
func apiserverHealth(t *httptunnel.HTTPTunnel, nodeAddress string) (bool, error) {
	endpoint := fmt.Sprintf(healthzEndpoint, nodeAddress)
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	resp, err := t.Client.Do(request)
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
