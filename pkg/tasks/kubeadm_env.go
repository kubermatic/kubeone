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

package tasks

import (
	"strconv"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/state"
)

var (
	version124 = semverutil.MustParseConstraint("1.24.*")
	version123 = semverutil.MustParseConstraint("1.23.*")
)

type kubeadmFlagsModifier func(flags map[string]string)

func updateKubeadmFlagsEnv(s *state.State, node *kubeoneapi.HostConfig) error {
	modifiers := []kubeadmFlagsModifier{
		updateKubeletNodeValues(s, node),
	}
	if m := removeNetworkPluginFlagKubelet(s, node); m != nil {
		modifiers = append(modifiers, m)
	}

	return updateKubeadmFlagsEnvFile(s, node, modifiers...)
}

// removeNetworkPluginFlagKubelet removes --network-plugin flag from Kubelet
// config because it has been removed in Kubernetes 1.24.
// This function is executed only when upgrading cluster from 1.23 to 1.24.
// TODO: Remove this function after dropping support for Kubernetes 1.23.
func removeNetworkPluginFlagKubelet(s *state.State, node *kubeoneapi.HostConfig) kubeadmFlagsModifier {
	needed := false
	if !version124.Check(s.LiveCluster.ExpectedVersion) {
		return nil
	}
	for _, cp := range s.LiveCluster.ControlPlane {
		if cp.Config.Hostname == node.Hostname && version123.Check(cp.Kubelet.Version) {
			needed = true

			break
		}
	}
	if !needed {
		return nil
	}

	return func(flags map[string]string) {
		// --network-plugin flag is not used with containerd and has been
		// removed in Kubernetes 1.24
		delete(flags, networkPluginFlag)
	}
}

func updateKubeletNodeValues(s *state.State, node *kubeoneapi.HostConfig) kubeadmFlagsModifier {
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

func updateKubeadmFlagsEnvFile(s *state.State, node *kubeoneapi.HostConfig, modifiers ...kubeadmFlagsModifier) error {
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
