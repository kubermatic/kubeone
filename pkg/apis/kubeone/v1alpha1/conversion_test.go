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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	corev1 "k8s.io/api/core/v1"
)

func TestCNIRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                string
		versionedCNI        *CNI
		expectedInternalCNI *kubeoneapi.CNI
	}{
		{
			name: "canal CNI",
			versionedCNI: &CNI{
				Provider: CNIProviderCanal,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				Canal: &kubeoneapi.CanalSpec{
					MTU: 1450,
				},
			},
		},
		{
			name: "weave-net CNI",
			versionedCNI: &CNI{
				Provider: CNIProviderWeaveNet,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				WeaveNet: &kubeoneapi.WeaveNetSpec{},
			},
		},
		{
			name: "weave-net CNI with encryption",
			versionedCNI: &CNI{
				Provider:  CNIProviderWeaveNet,
				Encrypted: true,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				WeaveNet: &kubeoneapi.WeaveNetSpec{
					Encrypted: true,
				},
			},
		},
		{
			name: "external CNI",
			versionedCNI: &CNI{
				Provider: CNIProviderExternal,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				External: &kubeoneapi.ExternalCNISpec{},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalCNI := &kubeoneapi.CNI{}
			if err := Convert_v1alpha1_CNI_To_kubeone_CNI(tc.versionedCNI, convertedInternalCNI, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalCNI, convertedInternalCNI) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalCNI, convertedInternalCNI))
			}

			// Converted internal back to versioned and compare
			convertedVersionedCNI := &CNI{}
			if err := Convert_kubeone_CNI_To_v1alpha1_CNI(convertedInternalCNI, convertedVersionedCNI, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.versionedCNI, convertedVersionedCNI) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.versionedCNI, convertedVersionedCNI))
			}
		})
	}
}

func TestCNIWithEncryptionRoundTripConversion(t *testing.T) {
	// It's expected that Encrypted will be dropped for Canal and External,
	// because those providers don't support encryption.
	tests := []struct {
		name                 string
		versionedCNI         *CNI
		expectedInternalCNI  *kubeoneapi.CNI
		expectedVersionedCNI *CNI
	}{
		{
			name: "canal CNI with encryption",
			versionedCNI: &CNI{
				Provider:  CNIProviderCanal,
				Encrypted: true,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				Canal: &kubeoneapi.CanalSpec{
					MTU: 1450,
				},
			},
			expectedVersionedCNI: &CNI{
				Provider: CNIProviderCanal,
			},
		},
		{
			name: "canal CNI with encryption",
			versionedCNI: &CNI{
				Provider:  CNIProviderExternal,
				Encrypted: true,
			},
			expectedInternalCNI: &kubeoneapi.CNI{
				External: &kubeoneapi.ExternalCNISpec{},
			},
			expectedVersionedCNI: &CNI{
				Provider: CNIProviderExternal,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalCNI := &kubeoneapi.CNI{}
			if err := Convert_v1alpha1_CNI_To_kubeone_CNI(tc.versionedCNI, convertedInternalCNI, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalCNI, convertedInternalCNI) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalCNI, convertedInternalCNI))
			}

			// Converted internal back to versioned and compare
			convertedVersionedCNI := &CNI{}
			if err := Convert_kubeone_CNI_To_v1alpha1_CNI(convertedInternalCNI, convertedVersionedCNI, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.expectedVersionedCNI, convertedVersionedCNI) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.expectedVersionedCNI, convertedVersionedCNI))
			}
		})
	}
}

func TestCNIInvalidConversion(t *testing.T) {
	tests := []struct {
		name         string
		versionedCNI *CNI
	}{
		{
			name: "provider name not specified",
			versionedCNI: &CNI{
				Provider: "",
			},
		},
		{
			name: "provider name not exist",
			versionedCNI: &CNI{
				Provider: "test",
			},
		},
		{
			name:         "empty CNI struct",
			versionedCNI: &CNI{},
		},
		{
			name: "CNI struct without provider field",
			versionedCNI: &CNI{
				Encrypted: true,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			convertedCNI := &kubeoneapi.CNI{}
			if err := Convert_v1alpha1_CNI_To_kubeone_CNI(tc.versionedCNI, convertedCNI, nil); err == nil {
				t.Errorf("expected error but got nil instead")
			}
		})
	}
}

func TestCloudProviderRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                          string
		versionedCloudProvider        *CloudProviderSpec
		expectedInternalCloudProvider *kubeoneapi.CloudProviderSpec
	}{
		{
			name: "aws cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameAWS,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				AWS: &kubeoneapi.AWSSpec{},
			},
		},
		{
			name: "azure cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameAzure,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Azure: &kubeoneapi.AzureSpec{},
			},
		},
		{
			name: "digitalocean cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameDigitalOcean,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
			},
		},
		{
			name: "gce cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameGCE,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				GCE: &kubeoneapi.GCESpec{},
			},
		},
		{
			name: "hetzner cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameHetzner,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Hetzner: &kubeoneapi.HetznerSpec{},
			},
		},
		{
			name: "openstack cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameOpenStack,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Openstack: &kubeoneapi.OpenstackSpec{},
			},
		},
		{
			name: "packet cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNamePacket,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Packet: &kubeoneapi.PacketSpec{},
			},
		},
		{
			name: "vsphere cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameVSphere,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Vsphere: &kubeoneapi.VsphereSpec{},
			},
		},
		{
			name: "none cloud provider",
			versionedCloudProvider: &CloudProviderSpec{
				Name: CloudProviderNameNone,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				None: &kubeoneapi.NoneSpec{},
			},
		},
		{
			name: "vsphere cloud provider with cloud-config",
			versionedCloudProvider: &CloudProviderSpec{
				Name:        CloudProviderNameVSphere,
				CloudConfig: "test-cloud-config",
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				CloudConfig: "test-cloud-config",
			},
		},
		{
			name: "vsphere cloud provider with external",
			versionedCloudProvider: &CloudProviderSpec{
				Name:     CloudProviderNameVSphere,
				External: true,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Vsphere:  &kubeoneapi.VsphereSpec{},
				External: true,
			},
		},
		{
			name: "openstack cloud provider with cloud-config and external",
			versionedCloudProvider: &CloudProviderSpec{
				Name:        CloudProviderNameVSphere,
				CloudConfig: "test-cloud-config",
				External:    true,
			},
			expectedInternalCloudProvider: &kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				CloudConfig: "test-cloud-config",
				External:    true,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalCloudProvider := &kubeoneapi.CloudProviderSpec{}
			if err := Convert_v1alpha1_CloudProviderSpec_To_kubeone_CloudProviderSpec(tc.versionedCloudProvider, convertedInternalCloudProvider, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalCloudProvider, convertedInternalCloudProvider) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalCloudProvider, convertedInternalCloudProvider))
			}

			// Converted internal back to versioned and compare
			convertedVersionedCloudProvider := &CloudProviderSpec{}
			if err := Convert_kubeone_CloudProviderSpec_To_v1alpha1_CloudProviderSpec(convertedInternalCloudProvider, convertedVersionedCloudProvider, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.versionedCloudProvider, convertedVersionedCloudProvider) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.versionedCloudProvider, convertedVersionedCloudProvider))
			}
		})
	}
}

func TestCloudProviderInvalidConversion(t *testing.T) {
	tests := []struct {
		name                   string
		versionedCloudProvider *CloudProviderSpec
	}{
		{
			name: "provider name not specified",
			versionedCloudProvider: &CloudProviderSpec{
				Name: "",
			},
		},
		{
			name: "provider name not exist",
			versionedCloudProvider: &CloudProviderSpec{
				Name: "test",
			},
		},
		{
			name:                   "empty CloudProvider struct",
			versionedCloudProvider: &CloudProviderSpec{},
		},
		{
			name: "CloudProvider struct without name field",
			versionedCloudProvider: &CloudProviderSpec{
				CloudConfig: "test-cloud-config",
				External:    true,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			convertedCloudProvider := &kubeoneapi.CloudProviderSpec{}
			if err := Convert_v1alpha1_CloudProviderSpec_To_kubeone_CloudProviderSpec(tc.versionedCloudProvider, convertedCloudProvider, nil); err == nil {
				t.Errorf("expected error but got nil instead")
			}
		})
	}
}

func TestClusterNetworkRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                           string
		versionedClusterNetwork        *ClusterNetworkConfig
		expectedInternalClusterNetwork *kubeoneapi.ClusterNetworkConfig
	}{
		{
			name: "cluster network without network id and cni",
			versionedClusterNetwork: &ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
			},
			expectedInternalClusterNetwork: &kubeoneapi.ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
			},
		},
		{
			name: "cluster network with cni and without network id",
			versionedClusterNetwork: &ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
				CNI: &CNI{
					Provider: CNIProviderCanal,
				},
			},
			expectedInternalClusterNetwork: &kubeoneapi.ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
				CNI: &kubeoneapi.CNI{
					Canal: &kubeoneapi.CanalSpec{
						MTU: 1450,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalClusterNetwork := &kubeoneapi.ClusterNetworkConfig{}
			if err := Convert_v1alpha1_ClusterNetworkConfig_To_kubeone_ClusterNetworkConfig(tc.versionedClusterNetwork, convertedInternalClusterNetwork, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalClusterNetwork, convertedInternalClusterNetwork) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalClusterNetwork, convertedInternalClusterNetwork))
			}

			// Converted internal back to versioned and compare
			convertedVersionedClusterNetwork := &ClusterNetworkConfig{}
			if err := Convert_kubeone_ClusterNetworkConfig_To_v1alpha1_ClusterNetworkConfig(convertedInternalClusterNetwork, convertedVersionedClusterNetwork, nil); err != nil {
				t.Errorf("error converting from internal to versioned: %v", err)
			}
			if !cmp.Equal(tc.versionedClusterNetwork, convertedVersionedClusterNetwork) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.versionedClusterNetwork, convertedVersionedClusterNetwork))
			}
		})
	}
}

func TestClusterNetworkWithNetworkIDRoundTripConversion(t *testing.T) {
	// The NetworkID conversion is properly handled at the KubeOneCluster level.
	// When converting ClusterNetwork from versioned to internal, the NetworkID
	// field will be dropped, because it doesn't exist in the internal struct.
	tests := []struct {
		name                            string
		versionedClusterNetwork         *ClusterNetworkConfig
		expectedInternalClusterNetwork  *kubeoneapi.ClusterNetworkConfig
		expectedVersionedClusterNetwork *ClusterNetworkConfig
	}{
		{
			name: "cluster network with network id",
			versionedClusterNetwork: &ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
				NetworkID:         "test-network-id",
			},
			expectedInternalClusterNetwork: &kubeoneapi.ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
			},
			expectedVersionedClusterNetwork: &ClusterNetworkConfig{
				PodSubnet:         "10.0.0.0/16",
				ServiceSubnet:     "192.168.1.0/24",
				ServiceDomainName: "cluster.local",
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalClusterNetwork := &kubeoneapi.ClusterNetworkConfig{}
			if err := Convert_v1alpha1_ClusterNetworkConfig_To_kubeone_ClusterNetworkConfig(tc.versionedClusterNetwork, convertedInternalClusterNetwork, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalClusterNetwork, convertedInternalClusterNetwork) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalClusterNetwork, convertedInternalClusterNetwork))
			}

			// Converted internal back to versioned and compare
			convertedVersionedClusterNetwork := &ClusterNetworkConfig{}
			if err := Convert_kubeone_ClusterNetworkConfig_To_v1alpha1_ClusterNetworkConfig(convertedInternalClusterNetwork, convertedVersionedClusterNetwork, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.expectedVersionedClusterNetwork, convertedVersionedClusterNetwork) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.expectedVersionedClusterNetwork, convertedVersionedClusterNetwork))
			}
		})
	}
}

func TestHostConfigRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                       string
		versionedHostConfig        *HostConfig
		expectedInternalHostConfig *kubeoneapi.HostConfig
	}{
		{
			name: "host config without untaint",
			versionedHostConfig: &HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
			},
			expectedInternalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
			},
		},
		{
			name: "host config with untaint set to false",
			versionedHostConfig: &HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Untaint:        false,
			},
			expectedInternalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
			},
		},
		{
			name: "host config with untaint set to true",
			versionedHostConfig: &HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Untaint:        true,
			},
			expectedInternalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Taints:         []corev1.Taint{},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalHostConfig := &kubeoneapi.HostConfig{}
			if err := Convert_v1alpha1_HostConfig_To_kubeone_HostConfig(tc.versionedHostConfig, convertedInternalHostConfig, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalHostConfig, convertedInternalHostConfig) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalHostConfig, convertedInternalHostConfig))
			}

			// Converted internal back to versioned and compare
			convertedVersionedHostConfig := &HostConfig{}
			if err := Convert_kubeone_HostConfig_To_v1alpha1_HostConfig(convertedInternalHostConfig, convertedVersionedHostConfig, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.versionedHostConfig, convertedVersionedHostConfig) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.versionedHostConfig, convertedVersionedHostConfig))
			}
		})
	}
}

func TestHostConfigCustomTaintsConversion(t *testing.T) {
	// If there are custom taints specified in the internal struct,
	// conversion can't be done automatically and losslessly.
	// The exception to this is if the custom taint is
	// node-role.kubernetes.io/master with NoSchedule effect,
	// which can be converted to untaint: true.
	tests := []struct {
		name                        string
		internalHostConfig          *kubeoneapi.HostConfig
		expectedVersionedHostConfig *HostConfig
		expectedError               bool
	}{
		{
			name: "host config with node-role.kubernetes.io/master taint",
			internalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Taints: []corev1.Taint{
					{
						Effect: corev1.TaintEffectNoSchedule,
						Key:    "node-role.kubernetes.io/master",
					},
				},
			},
			expectedVersionedHostConfig: &HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Untaint:        false,
			},
			expectedError: false,
		},
		{
			name: "host config with empty taints",
			internalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Taints:         []corev1.Taint{},
			},
			expectedVersionedHostConfig: &HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Untaint:        true,
			},
			expectedError: false,
		},
		{
			name: "host config with two custom taints",
			internalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Taints: []corev1.Taint{
					{
						Effect: corev1.TaintEffectNoSchedule,
						Key:    "node-role.kubernetes.io/master",
					},
					{
						Effect: corev1.TaintEffectNoSchedule,
						Key:    "test-key/test",
					},
				},
			},
			expectedVersionedHostConfig: nil,
			expectedError:               true,
		},
		{
			name: "host config with one custom taint",
			internalHostConfig: &kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				SSHPort:        22,
				SSHUsername:    "ubuntu",
				Hostname:       "test",
				Taints: []corev1.Taint{
					{
						Effect: corev1.TaintEffectNoExecute,
						Key:    "node-role.kubernetes.io/master",
					},
				},
			},
			expectedVersionedHostConfig: nil,
			expectedError:               true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert internal to versioned and compare
			convertedVersionedHostConfig := &HostConfig{}
			err := Convert_kubeone_HostConfig_To_v1alpha1_HostConfig(tc.internalHostConfig, convertedVersionedHostConfig, nil)
			if err != nil && !tc.expectedError {
				t.Errorf("error converting from internal to versioned: %v", err)
			}
			if err == nil && !tc.expectedError {
				if !cmp.Equal(tc.expectedVersionedHostConfig, convertedVersionedHostConfig) {
					t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedVersionedHostConfig, convertedVersionedHostConfig))
				}
			}
		})
	}
}

func TestKubeOneClusterRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                           string
		versionedKubeOneCluster        *KubeOneCluster
		expectedInternalKubeOneCluster *kubeoneapi.KubeOneCluster
	}{
		{
			name: "aws cluster",
			versionedKubeOneCluster: &KubeOneCluster{
				Versions: VersionConfig{
					Kubernetes: "1.18.2",
				},
				CloudProvider: CloudProviderSpec{
					Name: CloudProviderNameAWS,
				},
				Hosts: []HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
					},
					{
						PublicAddress:  "1.1.1.2",
						PrivateAddress: "10.0.0.2",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
						Untaint:        true,
					},
					{
						PublicAddress:  "1.1.1.3",
						PrivateAddress: "10.0.0.3",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
					},
				},
				StaticWorkers: []HostConfig{
					{
						PublicAddress:  "1.1.1.4",
						PrivateAddress: "10.0.0.4",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
					},
					{
						PublicAddress:  "1.1.1.5",
						PrivateAddress: "10.0.0.5",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
						Untaint:        true,
					},
				},
				Workers: []WorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
						Config: ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-1\""),
							Labels: map[string]string{
								"test-label": "test-val-1",
							},
							OperatingSystem: "ubuntu",
						},
					},
					{
						Name:     "test-2",
						Replicas: intPtr(3),
						Config: ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-2\""),
							Labels: map[string]string{
								"test-label": "test-val-2",
							},
							OperatingSystem: "ubuntu",
						},
					},
					{
						Name:     "test-3",
						Replicas: intPtr(3),
						Config: ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-3\""),
							Labels: map[string]string{
								"test-label": "test-val-3",
							},
							OperatingSystem: "ubuntu",
						},
					},
				},
			},
			expectedInternalKubeOneCluster: &kubeoneapi.KubeOneCluster{
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.18.2",
				},
				ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
					Docker: &kubeoneapi.ContainerRuntimeDocker{},
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					AWS: &kubeoneapi.AWSSpec{},
				},
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
						{
							PublicAddress:  "1.1.1.1",
							PrivateAddress: "10.0.0.1",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
						},
						{
							PublicAddress:  "1.1.1.2",
							PrivateAddress: "10.0.0.2",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
							Taints:         []corev1.Taint{},
						},
						{
							PublicAddress:  "1.1.1.3",
							PrivateAddress: "10.0.0.3",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
						},
					},
				},
				StaticWorkers: kubeoneapi.StaticWorkersConfig{
					Hosts: []kubeoneapi.HostConfig{
						{
							PublicAddress:  "1.1.1.4",
							PrivateAddress: "10.0.0.4",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
						},
						{
							PublicAddress:  "1.1.1.5",
							PrivateAddress: "10.0.0.5",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
							Taints:         []corev1.Taint{},
						},
					},
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
						Config: kubeoneapi.ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-1\""),
							Labels: map[string]string{
								"test-label": "test-val-1",
							},
							OperatingSystem: "ubuntu",
						},
					},
					{
						Name:     "test-2",
						Replicas: intPtr(3),
						Config: kubeoneapi.ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-2\""),
							Labels: map[string]string{
								"test-label": "test-val-2",
							},
							OperatingSystem: "ubuntu",
						},
					},
					{
						Name:     "test-3",
						Replicas: intPtr(3),
						Config: kubeoneapi.ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-3\""),
							Labels: map[string]string{
								"test-label": "test-val-3",
							},
							OperatingSystem: "ubuntu",
						},
					},
				},
			},
		},
		{
			name: "hetzner cluster with network id",
			versionedKubeOneCluster: &KubeOneCluster{
				Versions: VersionConfig{
					Kubernetes: "1.18.2",
				},
				CloudProvider: CloudProviderSpec{
					Name: CloudProviderNameHetzner,
				},
				ClusterNetwork: ClusterNetworkConfig{
					NetworkID: "test-network-id",
				},
				Hosts: []HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHPort:        22,
						SSHUsername:    "ubuntu",
						Hostname:       "test",
					},
				},
				Workers: []WorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
						Config: ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-1\""),
							Labels: map[string]string{
								"test-label": "test-val-1",
							},
							OperatingSystem: "ubuntu",
						},
					},
				},
			},
			expectedInternalKubeOneCluster: &kubeoneapi.KubeOneCluster{
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.18.2",
				},
				ContainerRuntime: kubeoneapi.ContainerRuntimeConfig{
					Docker: &kubeoneapi.ContainerRuntimeDocker{},
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Hetzner: &kubeoneapi.HetznerSpec{
						NetworkID: "test-network-id",
					},
				},
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
						{
							PublicAddress:  "1.1.1.1",
							PrivateAddress: "10.0.0.1",
							SSHPort:        22,
							SSHUsername:    "ubuntu",
							Hostname:       "test",
						},
					},
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
						Config: kubeoneapi.ProviderSpec{
							CloudProviderSpec: []byte("{\"test\": \"test-val-1\""),
							Labels: map[string]string{
								"test-label": "test-val-1",
							},
							OperatingSystem: "ubuntu",
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalKubeOneCluster := &kubeoneapi.KubeOneCluster{}
			if err := Convert_v1alpha1_KubeOneCluster_To_kubeone_KubeOneCluster(tc.versionedKubeOneCluster, convertedInternalKubeOneCluster, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalKubeOneCluster, convertedInternalKubeOneCluster) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalKubeOneCluster, convertedInternalKubeOneCluster))
			}

			// Converted internal back to versioned and compare
			convertedVersionedKubeOneCluster := &KubeOneCluster{}
			if err := Convert_kubeone_KubeOneCluster_To_v1alpha1_KubeOneCluster(convertedInternalKubeOneCluster, convertedVersionedKubeOneCluster, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.versionedKubeOneCluster, convertedVersionedKubeOneCluster) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.versionedKubeOneCluster, convertedVersionedKubeOneCluster))
			}
		})
	}
}

func TestMachineControllerRoundTripConversion(t *testing.T) {
	tests := []struct {
		name                               string
		versionedMachineController         *MachineControllerConfig
		expectedInternalMachineController  *kubeoneapi.MachineControllerConfig
		expectedVersionedMachineController *MachineControllerConfig
	}{
		{
			name: "enabled machine controller",
			versionedMachineController: &MachineControllerConfig{
				Deploy: true,
			},
			expectedInternalMachineController: &kubeoneapi.MachineControllerConfig{
				Deploy: true,
			},
			expectedVersionedMachineController: &MachineControllerConfig{
				Deploy: true,
			},
		},
		{
			name: "disabled machine controller",
			versionedMachineController: &MachineControllerConfig{
				Deploy: false,
			},
			expectedInternalMachineController: &kubeoneapi.MachineControllerConfig{
				Deploy: false,
			},
			expectedVersionedMachineController: &MachineControllerConfig{
				Deploy: false,
			},
		},
		{
			name: "enabled machine controller with provider set",
			versionedMachineController: &MachineControllerConfig{
				Deploy:   true,
				Provider: CloudProviderNameAWS,
			},
			expectedInternalMachineController: &kubeoneapi.MachineControllerConfig{
				Deploy: true,
			},
			expectedVersionedMachineController: &MachineControllerConfig{
				Deploy: true,
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Convert versioned to internal and compare
			convertedInternalMachineController := &kubeoneapi.MachineControllerConfig{}
			if err := Convert_v1alpha1_MachineControllerConfig_To_kubeone_MachineControllerConfig(tc.versionedMachineController, convertedInternalMachineController, nil); err != nil {
				t.Errorf("error converting from versioned to internal: %v", err)
			}
			if !cmp.Equal(tc.expectedInternalMachineController, convertedInternalMachineController) {
				t.Errorf("invalid conversion between versioned and internal: %v", cmp.Diff(tc.expectedInternalMachineController, convertedInternalMachineController))
			}

			// Converted internal back to versioned and compare
			convertedVersionedMachineController := &MachineControllerConfig{}
			if err := Convert_kubeone_MachineControllerConfig_To_v1alpha1_MachineControllerConfig(convertedInternalMachineController, convertedVersionedMachineController, nil); err != nil {
				t.Errorf("error converting from internal to to versioned: %v", err)
			}
			if !cmp.Equal(tc.expectedVersionedMachineController, convertedVersionedMachineController) {
				t.Errorf("invalid conversion between internal and versioned: %v", cmp.Diff(tc.expectedVersionedMachineController, convertedVersionedMachineController))
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
