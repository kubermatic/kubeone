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

package e2e

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

type containerRuntimeFlag struct {
	crc kubeoneapi.ContainerRuntimeConfig
}

func (crf containerRuntimeFlag) String() string {
	return crf.crc.String()
}

func (crf containerRuntimeFlag) ContainerRuntimeConfig() kubeoneapi.ContainerRuntimeConfig {
	crc := kubeoneapi.ContainerRuntimeConfig{}

	switch {
	case crf.crc.Docker != nil:
		crc.Docker = &kubeoneapi.ContainerRuntimeDocker{}
	case crf.crc.Containerd != nil:
		crc.Containerd = &kubeoneapi.ContainerRuntimeContainerd{}
	}

	return crc
}

func (crf *containerRuntimeFlag) Set(text string) error {
	if text == "" {
		text = "docker"
	}

	return crf.crc.UnmarshalText([]byte(text))
}
