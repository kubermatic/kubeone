/*
Copyright 2021 The KubeOne Authors.

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

package containerruntime

import (
	"encoding/json"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

type dockerConfig struct {
	ExecOpts           []string          `json:"exec-opts,omitempty"`
	StorageDriver      string            `json:"storage-driver,omitempty"`
	LogDriver          string            `json:"log-driver,omitempty"`
	LogOpts            map[string]string `json:"log-opts,omitempty"`
	InsecureRegistries []string          `json:"insecure-registries,omitempty"`
	RegistryMirrors    []string          `json:"registry-mirrors,omitempty"`
}

func marshalDockerConfig(cluster *kubeone.KubeOneCluster) (string, error) {
	cfg := dockerConfig{
		ExecOpts:      []string{"native.cgroupdriver=systemd"},
		StorageDriver: "overlay2",
		LogDriver:     "json-file",
		LogOpts: map[string]string{
			"max-size": "100m",
		},
	}

	insecureRegistry := cluster.RegistryConfiguration.InsecureRegistryAddress()
	if insecureRegistry != "" {
		cfg.InsecureRegistries = []string{insecureRegistry}
	}

	b, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
