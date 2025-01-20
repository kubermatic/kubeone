/*
Copyright 2024 The KubeOne Authors.

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

package v1beta3

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	conversion "k8s.io/apimachinery/pkg/conversion"
)

func Convert_kubeone_KubeOneCluster_To_v1beta3_KubeOneCluster(in *kubeoneapi.KubeOneCluster, out *KubeOneCluster, scope conversion.Scope) error {
	// AssetsConfiguration has been removed in the v1beta3 API
	return autoConvert_kubeone_KubeOneCluster_To_v1beta3_KubeOneCluster(in, out, scope)
}

func Convert_v1beta3_ContainerRuntimeConfig_To_kubeone_ContainerRuntimeConfig(in *ContainerRuntimeConfig, out *kubeoneapi.ContainerRuntimeConfig, scope conversion.Scope) error {
	return autoConvert_v1beta3_ContainerRuntimeConfig_To_kubeone_ContainerRuntimeConfig(in, out, scope)
}

func Convert_v1beta3_CiliumSpec_To_kubeone_CiliumSpec(in *CiliumSpec, out *kubeoneapi.CiliumSpec, _ conversion.Scope) error {
	out.KubeProxyReplacement = in.KubeProxyReplacement
	out.EnableHubble = in.EnableHubble

	return nil
}

func Convert_kubeone_CiliumSpec_To_v1beta3_CiliumSpec(in *kubeoneapi.CiliumSpec, out *CiliumSpec, s conversion.Scope) error {
	out.KubeProxyReplacement = in.KubeProxyReplacement
	return autoConvert_kubeone_CiliumSpec_To_v1beta3_CiliumSpec(in, out, s)
}
