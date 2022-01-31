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

import kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

func UpdateDataMap(cluster *kubeoneapi.KubeOneCluster, inputMap map[string]interface{}) error {
	var (
		crConfig string
		err      error
	)

	switch {
	case cluster.ContainerRuntime.Containerd != nil:
		crConfig, err = marshalContainerdConfig(cluster)
	case cluster.ContainerRuntime.Docker != nil:
		crConfig, err = marshalDockerConfig(cluster)
	}

	if err != nil {
		return err
	}

	inputMap["CONTAINER_RUNTIME_CONFIG_PATH"] = cluster.ContainerRuntime.ConfigPath()
	inputMap["CONTAINER_RUNTIME_CONFIG"] = crConfig
	inputMap["CONTAINER_RUNTIME_SOCKET"] = cluster.ContainerRuntime.CRISocket()

	return nil
}
