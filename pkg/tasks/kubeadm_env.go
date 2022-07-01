/*
Copyright 2022 The KubeOne Authors.

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

package tasks

import (
	"strconv"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
)

type kubeadmFlagsModifier func(flags map[string]string)

func updateKubeadmFlagsEnv(s *state.State, node *kubeoneapi.HostConfig) error {
	modifiers := []kubeadmFlagsModifier{
		updateKubeletNodeValues(node),
	}

	return updateKubeadmFlagsEnvFile(s, modifiers...)
}

func updateKubeletNodeValues(node *kubeoneapi.HostConfig) kubeadmFlagsModifier {
	return func(flags map[string]string) {
		if m := node.Kubelet.SystemReserved; m != nil {
			flags["--system-reserved"] = kubeoneapi.MapStringStringToString(m, "=")
		}

		if m := node.Kubelet.KubeReserved; m != nil {
			flags["--kube-reserved"] = kubeoneapi.MapStringStringToString(m, "=")
		}

		if m := node.Kubelet.EvictionHard; m != nil {
			flags["--eviction-hard"] = kubeoneapi.MapStringStringToString(m, "<")
		}
		if m := node.Kubelet.MaxPods; m != nil {
			flags["--max-pods"] = strconv.Itoa(int(*m))
		}
	}
}

func updateKubeadmFlagsEnvFile(s *state.State, modifiers ...kubeadmFlagsModifier) error {
	return updateRemoteFile(s, kubeadmEnvFlagsFile, func(content []byte) ([]byte, error) {
		kubeletFlags, err := unmarshalKubeletFlags(content)
		if err != nil {
			return nil, err
		}

		for _, m := range modifiers {
			m(kubeletFlags)
		}

		buf := marshalKubeletFlags(kubeletFlags)

		return buf, nil
	})
}
