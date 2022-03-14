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
	"strconv"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
)

type dockerConfig struct {
	ExecOpts           []string          `json:"exec-opts,omitempty"`
	StorageDriver      string            `json:"storage-driver,omitempty"`
	LogDriver          string            `json:"log-driver,omitempty"`
	LogOpts            map[string]string `json:"log-opts,omitempty"`
	InsecureRegistries []string          `json:"insecure-registries,omitempty"`
	RegistryMirrors    []string          `json:"registry-mirrors,omitempty"`
}

func marshalDockerConfig(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	// Parse log max size to ensure that it has the correct units
	logSize := strings.ToLower(cluster.LoggingConfig.ContainerLogMaxSize)
	logSize = strings.ReplaceAll(logSize, "ki", "k")
	logSize = strings.ReplaceAll(logSize, "mi", "m")
	logSize = strings.ReplaceAll(logSize, "gi", "g")

	cfg := dockerConfig{
		ExecOpts:      []string{"native.cgroupdriver=systemd"},
		StorageDriver: "overlay2",
		LogDriver:     "json-file",
		LogOpts: map[string]string{
			"max-size": logSize,
			"max-file": strconv.Itoa(int(cluster.LoggingConfig.ContainerLogMaxFiles)),
		},
	}

	insecureRegistry := cluster.RegistryConfiguration.InsecureRegistryAddress()
	if insecureRegistry != "" {
		cfg.InsecureRegistries = []string{insecureRegistry}
	}

	b, err := json.MarshalIndent(cfg, "", "	")

	return string(b), fail.Runtime(err, "encoding docker config")
}
