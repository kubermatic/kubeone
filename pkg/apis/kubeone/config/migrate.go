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
	"os"

	"github.com/pkg/errors"

	kubeonev1alpha1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1alpha1"
	kubeonev1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeone/v1beta1"
	"github.com/kubermatic/kubeone/pkg/yamled"
)

// MigrateOldConfig migrates KubeOneCluster v1alpha1 object to v1beta1
func MigrateOldConfig(clusterFilePath string) (interface{}, error) {
	var emptyVal struct{}

	oldConfig, err := loadClusterConfig(clusterFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the old config")
	}

	// Check is kubeone.io/v1alpha1 config provided
	apiVersion, apiVersionExists := oldConfig.GetString(yamled.Path{"apiVersion"})
	if !apiVersionExists {
		return nil, errors.New("apiVersion not present in the manifest")
	}
	if apiVersion != kubeonev1alpha1.SchemeGroupVersion.String() {
		return nil, errors.Errorf("migration is available only for %q API, but %q is given", kubeonev1alpha1.SchemeGroupVersion.String(), apiVersion)
	}

	// Ensure kind is KubeOneCluster
	kind, kindExists := oldConfig.GetString(yamled.Path{"kind"})
	if !kindExists {
		return nil, errors.New("kind not present in the manifest")
	}
	if kind != KubeOneClusterKind {
		return nil, errors.Errorf("migration is available only for kind %q, but %q is given", KubeOneClusterKind, kind)
	}

	// The APIVersion has been changed to kubeone.io/v1beta1
	oldConfig.Set(yamled.Path{"apiVersion"}, kubeonev1beta1.SchemeGroupVersion.String())
	oldConfig.Set(yamled.Path{"kind"}, "KubeOneCluster")

	// The .hosts field has been moved to .controlPlane.hosts
	// The .hosts.untaint field is replaced with .controlPlane.hosts.taints field
	migrateHosts(oldConfig, "hosts", "controlPlane")
	// The .staticWorkers field has been renamed to .staticWorkers.hosts
	// The .staticWorkers.untaint field has been replaced with .staticWorkers.hosts.taints
	migrateHosts(oldConfig, "staticWorkers", "staticWorkers")

	// The cloud providers are now defined using typed structs, instead of .cloudProvider.Name
	cloudProviderName, cloudProviderNameExists := oldConfig.GetString(yamled.Path{"cloudProvider", "name"})
	if cloudProviderNameExists {
		oldConfig.Remove(yamled.Path{"cloudProvider", "name"})
		switch kubeonev1alpha1.CloudProviderName(cloudProviderName) {
		case kubeonev1alpha1.CloudProviderNameAWS:
			oldConfig.Set(yamled.Path{"cloudProvider", "aws"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameAzure:
			oldConfig.Set(yamled.Path{"cloudProvider", "azure"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameDigitalOcean:
			oldConfig.Set(yamled.Path{"cloudProvider", "digitalocean"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameGCE:
			oldConfig.Set(yamled.Path{"cloudProvider", "gce"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameHetzner:
			oldConfig.Set(yamled.Path{"cloudProvider", "hetzner"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameOpenStack:
			oldConfig.Set(yamled.Path{"cloudProvider", "openstack"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNamePacket:
			oldConfig.Set(yamled.Path{"cloudProvider", "packet"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameVSphere:
			oldConfig.Set(yamled.Path{"cloudProvider", "vsphere"}, emptyVal)
		case kubeonev1alpha1.CloudProviderNameNone:
			oldConfig.Set(yamled.Path{"cloudProvider", "none"}, emptyVal)
		default:
			return nil, errors.Errorf("invalid cloud provider %q", kubeonev1alpha1.CloudProviderName(cloudProviderName))
		}
	}

	// The .clusterNetwork.networkID field has been moved to the HetznerSpec
	networkID, networkIDExists := oldConfig.GetString(yamled.Path{"clusterNetwork", "networkID"})
	if networkIDExists {
		oldConfig.Remove(yamled.Path{"clusterNetwork", "networkID"})
		// If we have set empty .cloudProvider.hetzner before, remove it, as it's not possible
		// to override the old value
		oldConfig.Remove(yamled.Path{"cloudProvider", "hetzner"})
		oldConfig.Set(yamled.Path{"cloudProvider", "hetzner", "networkID"}, networkID)
	}

	// The CNI plugins are now defined using typed structs, instead of .clusterNetwork.cni.name
	_, cniProviderExists := oldConfig.Get(yamled.Path{"clusterNetwork", "cni"})
	if cniProviderExists {
		cniProviderName, _ := oldConfig.GetString(yamled.Path{"clusterNetwork", "cni", "provider"})
		oldConfig.Remove(yamled.Path{"clusterNetwork", "cni", "provider"})

		encrypted, _ := oldConfig.GetBool(yamled.Path{"clusterNetwork", "cni", "encrypted"})
		oldConfig.Remove(yamled.Path{"clusterNetwork", "cni", "encrypted"})

		switch kubeonev1alpha1.CNIProvider(cniProviderName) {
		case kubeonev1alpha1.CNIProviderCanal:
			oldConfig.Set(yamled.Path{"clusterNetwork", "cni", "canal"}, emptyVal)
		case kubeonev1alpha1.CNIProviderWeaveNet:
			if encrypted {
				oldConfig.Set(yamled.Path{"clusterNetwork", "cni", "weaveNet", "encrypted"}, true)
			} else {
				oldConfig.Set(yamled.Path{"clusterNetwork", "cni", "weaveNet"}, emptyVal)
			}
		case kubeonev1alpha1.CNIProviderExternal:
			oldConfig.Set(yamled.Path{"clusterNetwork", "cni", "external"}, emptyVal)
		default:
			return nil, errors.Errorf("invalid cni provider %q", kubeonev1alpha1.CNIProvider(cniProviderName))
		}
	}

	// The .workers field has been renamed to .dynamicWorkers
	workers, workersExists := oldConfig.Get(yamled.Path{"workers"})
	if workersExists {
		oldConfig.Remove(yamled.Path{"workers"})
		oldConfig.Set(yamled.Path{"dynamicWorkers"}, workers)
	}

	// The .machineController.provider field has been removed
	oldConfig.Remove(yamled.Path{"machineController", "provider"})

	// The .credentials field has been removed
	oldConfig.Remove(yamled.Path{"credentials"})

	return oldConfig.Root(), nil
}

func migrateHosts(doc *yamled.Document, oldKey, newKey string) {
	hosts, hostsExists := doc.GetArray(yamled.Path{oldKey})
	if hostsExists {
		var emptyArr []string

		// The .hosts.untaint field has been replaced with .hosts.taints which takes a list of taints
		// In case .hosts.untaint has been set to true in the old config, set .hosts.taints to empty array.
		total := len(hosts)
		for i := 0; i < total; i++ {
			untaint, _ := doc.GetBool(yamled.Path{oldKey, i, "untaint"})
			doc.Remove(yamled.Path{oldKey, i, "untaint"})
			if untaint {
				doc.Set(yamled.Path{oldKey, i, "taints"}, emptyArr)
			}
		}

		// Rename .hosts to .controlPlane.hosts
		doc.Remove(yamled.Path{oldKey})
		doc.Set(yamled.Path{newKey, "hosts"}, hosts)
	}
}

// loadClusterConfig takes path to the Cluster Config (old API) and returns yamled.Document
func loadClusterConfig(oldConfigPath string) (*yamled.Document, error) {
	f, err := os.Open(oldConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer f.Close()

	return yamled.Load(f)
}
