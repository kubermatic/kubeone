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

package config

import (
	"fmt"
	"os"

	kubeonev1beta1 "k8c.io/kubeone/pkg/apis/kubeone/v1beta1"
	kubeonev1beta2 "k8c.io/kubeone/pkg/apis/kubeone/v1beta2"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/yamled"
)

// MigrateOldConfig migrates KubeOneCluster v1beta1 object to v1beta2
func MigrateOldConfig(clusterFilePath string) (interface{}, error) {
	oldConfig, err := loadClusterConfig(clusterFilePath)
	if err != nil {
		return nil, fail.Runtime(err, "loading cluster config to migrate")
	}

	// Check is kubeone.io/v1beta1 config provided
	apiVersion, apiVersionExists := oldConfig.GetString(yamled.Path{"apiVersion"})
	if !apiVersionExists {
		return nil, fail.Config(fmt.Errorf("apiVersion not present in the manifest"), "checking apiVersion presence")
	}

	if apiVersion != kubeonev1beta1.SchemeGroupVersion.String() {
		return nil, fail.Config(fmt.Errorf("migration is available only for %q API, but %q is given", kubeonev1beta1.SchemeGroupVersion.String(), apiVersion), "checking apiVersion compatibility")
	}

	// Ensure kind is KubeOneCluster
	kind, kindExists := oldConfig.GetString(yamled.Path{"kind"})
	if !kindExists {
		return nil, fail.ConfigValidation(fmt.Errorf("kind not present in the manifest"))
	}
	if kind != KubeOneClusterKind {
		return nil, fail.ConfigValidation(fmt.Errorf("migration is available only for kind %q, but %q is given", KubeOneClusterKind, kind))
	}

	// The APIVersion has been changed to kubeone.k8c.io/v1beta2
	oldConfig.Set(yamled.Path{"apiVersion"}, kubeonev1beta2.SchemeGroupVersion.String())

	// AssetConfiguration API has been removed from the v1beta2 API.
	// We are not able to automatically migrate manifests using the AssetConfiguration API
	// because it has multiple use cases:
	//   * EKS-D clusters -- support for EKS-D cluster has been entirely removed in KubeOne 1.4
	//   * Problem with CoreDNS image when using overwriteRegistry -- can be mitigated by using the latest image-loader
	//     script or by using the RegistryConfiguration API (registry mirrors)
	_, assetConfigExists := oldConfig.Get(yamled.Path{"assetConfiguration"})
	if assetConfigExists {
		return nil, fail.ConfigValidation(fmt.Errorf("the AssetConfiguration API has been removed from the v1beta2 API, please check the docs for information on how to migrate"))
	}

	// Packet has been renamed to Equinix Metal and as a result of this change
	// .cloudProvider.packet field has been renamed to .cloudProvider.equinixmetal
	packetSpec, cloudProviderPacketExists := oldConfig.Get(yamled.Path{"cloudProvider", "packet"})
	if cloudProviderPacketExists {
		oldConfig.Remove(yamled.Path{"cloudProvider", "packet"})
		oldConfig.Set(yamled.Path{"cloudProvider", "equinixmetal"}, packetSpec)
	}

	// The PodPresets feature has been removed from the v1beta2 API because Kubernetes doesn't support it starting
	// with Kubernetes 1.20.
	_, podPresetsExists := oldConfig.Get(yamled.Path{"features", "podPresets"})
	if podPresetsExists {
		oldConfig.Remove(yamled.Path{"features", "podPresets"})
	}

	// The addons path is not defaulted to "./addons" any longer to better support embedded addons.
	// To keep the backwards compatibility, migration will set the addons path to "./addons" if it's
	// empty or unset. The user can remove it if it's not needed.
	_, addonsExists := oldConfig.Get(yamled.Path{"addons"})
	if addonsExists {
		addonsPath, addonsPathExists := oldConfig.Get(yamled.Path{"addons", "path"})
		if !addonsPathExists || addonsPath == "" {
			oldConfig.Set(yamled.Path{"addons", "path"}, "./addons")
		}
	}

	return oldConfig.Root(), nil
}

// loadClusterConfig takes path to the Cluster Config (old API) and returns yamled.Document
func loadClusterConfig(oldConfigPath string) (*yamled.Document, error) {
	f, err := os.Open(oldConfigPath)
	if err != nil {
		return nil, fail.Runtime(err, "open manifest file")
	}
	defer f.Close()

	return yamled.Load(f)
}
