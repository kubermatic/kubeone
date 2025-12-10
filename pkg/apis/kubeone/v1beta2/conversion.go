/*
Copyright 2020 The KubeOne Authors.

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

package v1beta2

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	conversion "k8s.io/apimachinery/pkg/conversion"
)

func Convert_v1beta2_KubeOneCluster_To_kubeone_KubeOneCluster(in *KubeOneCluster, out *kubeoneapi.KubeOneCluster, scope conversion.Scope) error {
	if err := autoConvert_v1beta2_KubeOneCluster_To_kubeone_KubeOneCluster(in, out, scope); err != nil {
		return err
	}

	if len(in.HelmReleases) > 0 && out.Addons == nil {
		out.Addons = &kubeoneapi.Addons{}
	}

	for _, hr := range in.HelmReleases {
		hr := hr.DeepCopy()
		hrOut := kubeoneapi.HelmRelease{}

		if err := autoConvert_v1beta2_HelmRelease_To_kubeone_HelmRelease(hr, &hrOut, scope); err != nil {
			return err
		}

		out.Addons.Addons = append(out.Addons.Addons, kubeoneapi.AddonRef{HelmRelease: &hrOut})
	}

	return nil
}

func Convert_kubeone_KubeOneCluster_To_v1beta2_KubeOneCluster(in *kubeoneapi.KubeOneCluster, out *KubeOneCluster, scope conversion.Scope) error {
	// AssetsConfiguration has been removed in the v1beta2 API
	return autoConvert_kubeone_KubeOneCluster_To_v1beta2_KubeOneCluster(in, out, scope)
}

func Convert_v1beta2_ContainerRuntimeConfig_To_kubeone_ContainerRuntimeConfig(in *ContainerRuntimeConfig, out *kubeoneapi.ContainerRuntimeConfig, scope conversion.Scope) error {
	return autoConvert_v1beta2_ContainerRuntimeConfig_To_kubeone_ContainerRuntimeConfig(in, out, scope)
}

func Convert_v1beta2_Addon_To_kubeone_AddonRef(in *Addon, out *kubeoneapi.AddonRef, _ conversion.Scope) error {
	out.Addon = &kubeoneapi.Addon{
		Name:              in.Name,
		DisableTemplating: in.DisableTemplating,
		Params:            in.Params,
		Delete:            in.Delete,
	}

	return nil
}

func Convert_kubeone_KubeletConfig_To_v1beta2_KubeletConfig(in *kubeoneapi.KubeletConfig, out *KubeletConfig, s conversion.Scope) error {
	return autoConvert_kubeone_KubeletConfig_To_v1beta2_KubeletConfig(in, out, s)
}

func Convert_kubeone_AddonRef_To_v1beta2_Addon(*kubeoneapi.AddonRef, *Addon, conversion.Scope) error {
	return nil
}

func Convert_v1beta2_Addons_To_kubeone_Addons(in *Addons, out *kubeoneapi.Addons, scope conversion.Scope) error {
	return autoConvert_v1beta2_Addons_To_kubeone_Addons(in, out, scope)
}

func Convert_kubeone_Features_To_v1beta2_Features(in *kubeoneapi.Features, out *Features, s conversion.Scope) error {
	return autoConvert_kubeone_Features_To_v1beta2_Features(in, out, s)
}

func Convert_v1beta2_Features_To_kubeone_Features(in *Features, out *kubeoneapi.Features, s conversion.Scope) error {
	return autoConvert_v1beta2_Features_To_kubeone_Features(in, out, s)
}

func Convert_v1beta2_CiliumSpec_To_kubeone_CiliumSpec(in *CiliumSpec, out *kubeoneapi.CiliumSpec, _ conversion.Scope) error {
	out.KubeProxyReplacement = in.KubeProxyReplacement == KubeProxyReplacementStrict
	out.EnableHubble = in.EnableHubble

	return nil
}

func Convert_kubeone_CiliumSpec_To_v1beta2_CiliumSpec(in *kubeoneapi.CiliumSpec, out *CiliumSpec, _ conversion.Scope) error {
	out.KubeProxyReplacement = KubeProxyReplacementDisabled
	if in.KubeProxyReplacement {
		out.KubeProxyReplacement = KubeProxyReplacementStrict
	}
	out.EnableHubble = in.EnableHubble

	return nil
}

func Convert_kubeone_ContainerRuntimeContainerd_To_v1beta2_ContainerRuntimeContainerd(in *kubeoneapi.ContainerRuntimeContainerd, out *ContainerRuntimeContainerd, s conversion.Scope) error {
	return autoConvert_kubeone_ContainerRuntimeContainerd_To_v1beta2_ContainerRuntimeContainerd(in, out, s)
}

func Convert_v1beta2_ProviderSpec_To_kubeone_ProviderSpec(in *ProviderSpec, out *kubeoneapi.ProviderSpec, s conversion.Scope) error {
	return autoConvert_v1beta2_ProviderSpec_To_kubeone_ProviderSpec(in, out, s)
}
