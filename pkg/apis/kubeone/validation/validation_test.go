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

package validation

import (
	"testing"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

func TestValidateKubeOneCluster(t *testing.T) {
	tests := []struct {
		name          string
		cluster       kubeone.KubeOneCluster
		expectedError bool
	}{
		{
			name: "valid KubeOneCluster config",
			cluster: kubeone.KubeOneCluster{
				Name: "test",
				ControlPlane: kubeone.ControlPlaneConfig{
					Hosts: []kubeone.HostConfig{
						{
							PublicAddress:  "1.1.1.1",
							PrivateAddress: "10.0.0.1",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
						{
							PublicAddress:  "1.1.1.2",
							PrivateAddress: "10.0.0.2",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
					},
				},
				APIEndpoint: kubeone.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeone.CloudProviderSpec{
					AWS: &kubeone.AWSSpec{},
				},
				Versions: kubeone.VersionConfig{
					Kubernetes: "1.18.2",
				},
				MachineController: &kubeone.MachineControllerConfig{
					Deploy: true,
				},
				DynamicWorkers: []kubeone.DynamicWorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
					},
					{
						Name:     "test-2",
						Replicas: intPtr(5),
					},
					{
						Name:     "test-3",
						Replicas: intPtr(0),
					},
				},
			},
			expectedError: false,
		},
		{
			name: "MachineDeployment provided without machine-controller deployed",
			cluster: kubeone.KubeOneCluster{
				Name: "test",
				ControlPlane: kubeone.ControlPlaneConfig{
					Hosts: []kubeone.HostConfig{
						{
							PublicAddress:  "1.1.1.1",
							PrivateAddress: "10.0.0.1",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
						{
							PublicAddress:  "1.1.1.2",
							PrivateAddress: "10.0.0.2",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
					},
				},
				APIEndpoint: kubeone.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeone.CloudProviderSpec{
					AWS: &kubeone.AWSSpec{},
				},
				Versions: kubeone.VersionConfig{
					Kubernetes: "1.18.2",
				},
				MachineController: &kubeone.MachineControllerConfig{
					Deploy: false,
				},
				DynamicWorkers: []kubeone.DynamicWorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
					},
					{
						Name:     "test-2",
						Replicas: intPtr(5),
					},
					{
						Name:     "test-3",
						Replicas: intPtr(0),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "cluster name missing",
			cluster: kubeone.KubeOneCluster{
				Name: "",
				ControlPlane: kubeone.ControlPlaneConfig{
					Hosts: []kubeone.HostConfig{
						{
							PublicAddress:  "1.1.1.1",
							PrivateAddress: "10.0.0.1",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
						{
							PublicAddress:  "1.1.1.2",
							PrivateAddress: "10.0.0.2",
							SSHAgentSocket: "env:SSH_AUTH_SOCK",
							SSHUsername:    "ubuntu",
						},
					},
				},
				APIEndpoint: kubeone.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeone.CloudProviderSpec{
					AWS: &kubeone.AWSSpec{},
				},
				Versions: kubeone.VersionConfig{
					Kubernetes: "1.18.2",
				},
				MachineController: &kubeone.MachineControllerConfig{
					Deploy: true,
				},
				DynamicWorkers: []kubeone.DynamicWorkerConfig{
					{
						Name:     "test-1",
						Replicas: intPtr(3),
					},
					{
						Name:     "test-2",
						Replicas: intPtr(5),
					},
					{
						Name:     "test-3",
						Replicas: intPtr(0),
					},
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateKubeOneCluster(tc.cluster)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateControlPlaneConfig(t *testing.T) {
	tests := []struct {
		name               string
		controlPlaneConfig kubeone.ControlPlaneConfig
		expectedError      bool
	}{
		{
			name: "valid ControlPlane config",
			controlPlaneConfig: kubeone.ControlPlaneConfig{
				Hosts: []kubeone.HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
					{
						PublicAddress:  "1.1.1.2",
						PrivateAddress: "10.0.0.2",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid host config",
			controlPlaneConfig: kubeone.ControlPlaneConfig{
				Hosts: []kubeone.HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
					{
						PublicAddress:  "1.1.1.2",
						PrivateAddress: "10.0.0.2",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "no hosts provided",
			controlPlaneConfig: kubeone.ControlPlaneConfig{
				Hosts: []kubeone.HostConfig{},
			},
			expectedError: true,
		},
		{
			name:               "no hosts field present",
			controlPlaneConfig: kubeone.ControlPlaneConfig{},
			expectedError:      true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateControlPlaneConfig(tc.controlPlaneConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateAPIEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		apiEndpoint   kubeone.APIEndpoint
		expectedError bool
	}{
		{
			name: "valid apiEndpoint config (localhost:6443)",
			apiEndpoint: kubeone.APIEndpoint{
				Host: "localhost",
				Port: 6443,
			},
			expectedError: false,
		},
		{
			name: "valid apiEndpoint config (example.com:443)",
			apiEndpoint: kubeone.APIEndpoint{
				Host: "example.com",
				Port: 443,
			},
			expectedError: false,
		},
		{
			name: "no host specified",
			apiEndpoint: kubeone.APIEndpoint{
				Port: 6443,
			},
			expectedError: true,
		},
		{
			name: "no port specified",
			apiEndpoint: kubeone.APIEndpoint{
				Host: "localhost",
			},
			expectedError: true,
		},
		{
			name: "port lower than 0",
			apiEndpoint: kubeone.APIEndpoint{
				Host: "localhost",
				Port: -1,
			},
			expectedError: true,
		},
		{
			name: "port greater than 65535",
			apiEndpoint: kubeone.APIEndpoint{
				Host: "localhost",
				Port: 65536,
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateAPIEndpoint(tc.apiEndpoint, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateCloudProviderSpec(t *testing.T) {
	tests := []struct {
		name           string
		providerConfig kubeone.CloudProviderSpec
		expectedError  bool
	}{
		{
			name: "valid AWS provider config",
			providerConfig: kubeone.CloudProviderSpec{
				AWS: &kubeone.AWSSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid Azure provider config",
			providerConfig: kubeone.CloudProviderSpec{
				Azure:       &kubeone.AzureSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid DigitalOcean provider config",
			providerConfig: kubeone.CloudProviderSpec{
				DigitalOcean: &kubeone.DigitalOceanSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid GCE provider config",
			providerConfig: kubeone.CloudProviderSpec{
				GCE: &kubeone.GCESpec{},
			},
			expectedError: false,
		},
		{
			name: "valid Hetzner provider config",
			providerConfig: kubeone.CloudProviderSpec{
				Hetzner: &kubeone.HetznerSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid OpenStack provider config",
			providerConfig: kubeone.CloudProviderSpec{
				Openstack:   &kubeone.OpenstackSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid Packet provider config",
			providerConfig: kubeone.CloudProviderSpec{
				Packet: &kubeone.PacketSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid vSphere provider config",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere:     &kubeone.VsphereSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid None provider config",
			providerConfig: kubeone.CloudProviderSpec{
				None: &kubeone.NoneSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid OpenStack provider config with external CCM and cloudConfig",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:         &kubeone.AWSSpec{},
				CloudConfig: "cloud-config",
				External:    true,
			},
			expectedError: false,
		},
		{
			name: "valid DigitalOcean provider config with external CCM",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:      &kubeone.AWSSpec{},
				External: true,
			},
			expectedError: false,
		},
		{
			name: "AWS and Azure specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:   &kubeone.AWSSpec{},
				Azure: &kubeone.AzureSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and DigitalOcean specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:          &kubeone.AWSSpec{},
				DigitalOcean: &kubeone.DigitalOceanSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and GCE specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS: &kubeone.AWSSpec{},
				GCE: &kubeone.GCESpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and Hetzner specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:     &kubeone.AWSSpec{},
				Hetzner: &kubeone.HetznerSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and OpenStack specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:       &kubeone.AWSSpec{},
				Openstack: &kubeone.OpenstackSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and Packet specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:    &kubeone.AWSSpec{},
				Packet: &kubeone.PacketSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and vSphere specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:     &kubeone.AWSSpec{},
				Vsphere: &kubeone.VsphereSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and None specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:  &kubeone.AWSSpec{},
				None: &kubeone.NoneSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS, Azure, and DigitalOcean specified at the same time",
			providerConfig: kubeone.CloudProviderSpec{
				AWS:          &kubeone.AWSSpec{},
				Azure:        &kubeone.AzureSpec{},
				DigitalOcean: &kubeone.DigitalOceanSpec{},
			},
			expectedError: true,
		},
		{
			name: "Azure provider config without cloudConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Azure: &kubeone.AzureSpec{},
			},
			expectedError: true,
		},
		{
			name: "OpenStack provider config without cloudConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Openstack: &kubeone.OpenstackSpec{},
			},
			expectedError: true,
		},
		{
			name: "vSphere provider config without cloudConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere: &kubeone.VsphereSpec{},
			},
			expectedError: true,
		},
		{
			name:           "no provider specified",
			providerConfig: kubeone.CloudProviderSpec{},
			expectedError:  true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateCloudProviderSpec(tc.providerConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateVersionConfig(t *testing.T) {
	tests := []struct {
		name          string
		versionConfig kubeone.VersionConfig
		expectedError bool
	}{
		{
			name: "valid version config (1.18.2)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.18.2",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.18.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.18.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.17.5)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.17.5",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.17.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.17.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.16.9)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.16.9",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.16.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.16.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.14.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.14.0",
			},
			expectedError: false,
		},
		{
			name: "not supported kubernetes version (1.13.5)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.13.5",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.13.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.13.0",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.12.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.12.0",
			},
			expectedError: true,
		},
		{
			name: "invalid kubernetes version (2.0.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "2.0.0",
			},
			expectedError: true,
		},
		{
			name: "kubernetes version with a leading 'v'",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "v1.18.2",
			},
			expectedError: true,
		},
		{
			name: "invalid semver kubernetes version",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "version-1.19.0",
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateVersionConfig(tc.versionConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateClusterNetworkConfig(t *testing.T) {
	tests := []struct {
		name                 string
		clusterNetworkConfig kubeone.ClusterNetworkConfig
		expectedError        bool
	}{
		{
			name: "valid network config",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: false,
		},
		{
			name: "valid network config with cni config",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0/24",
				CNI: &kubeone.CNI{
					Canal: &kubeone.CanalSpec{MTU: 1500},
				},
			},
			expectedError: false,
		},
		{
			name:                 "empty network config",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{},
			expectedError:        false,
		},
		{
			name: "invalid pod subnet",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: true,
		},
		{
			name: "invalid service subnet (non-CIDR)",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0",
			},
			expectedError: true,
		},
		{
			name: "invalid cni config",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				CNI: &kubeone.CNI{
					Canal:    &kubeone.CanalSpec{},
					WeaveNet: &kubeone.WeaveNetSpec{},
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateClusterNetworkConfig(tc.clusterNetworkConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateCNIConfig(t *testing.T) {
	tests := []struct {
		name          string
		cniConfig     *kubeone.CNI
		expectedError bool
	}{
		{
			name: "valid Canal CNI config",
			cniConfig: &kubeone.CNI{
				Canal: &kubeone.CanalSpec{MTU: 1500},
			},
			expectedError: false,
		},
		{
			name: "valid WeaveNet CNI config",
			cniConfig: &kubeone.CNI{
				WeaveNet: &kubeone.WeaveNetSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid WeaveNet CNI config with encryption enabled",
			cniConfig: &kubeone.CNI{
				WeaveNet: &kubeone.WeaveNetSpec{
					Encrypted: true,
				},
			},
			expectedError: false,
		},
		{
			name: "valid External CNI config",
			cniConfig: &kubeone.CNI{
				External: &kubeone.ExternalCNISpec{},
			},
			expectedError: false,
		},
		{
			name: "Canal and WeaveNet specified at the same time",
			cniConfig: &kubeone.CNI{
				Canal:    &kubeone.CanalSpec{},
				WeaveNet: &kubeone.WeaveNetSpec{},
			},
			expectedError: true,
		},
		{
			name: "Canal and External specified at the same time",
			cniConfig: &kubeone.CNI{
				Canal:    &kubeone.CanalSpec{},
				External: &kubeone.ExternalCNISpec{},
			},
			expectedError: true,
		},
		{
			name: "WeaveNet and External specified at the same time",
			cniConfig: &kubeone.CNI{
				WeaveNet: &kubeone.WeaveNetSpec{},
				External: &kubeone.ExternalCNISpec{},
			},
			expectedError: true,
		},
		{
			name:          "no CNI config specified",
			cniConfig:     &kubeone.CNI{},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateCNI(tc.cniConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateStaticWorkersConfig(t *testing.T) {
	tests := []struct {
		name                string
		staticWorkersConfig kubeone.StaticWorkersConfig
		expectedError       bool
	}{
		{
			name: "valid StaticWorkers config",
			staticWorkersConfig: kubeone.StaticWorkersConfig{
				Hosts: []kubeone.HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
					{
						PublicAddress:  "1.1.1.2",
						PrivateAddress: "10.0.0.2",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "no hosts provided",
			staticWorkersConfig: kubeone.StaticWorkersConfig{
				Hosts: []kubeone.HostConfig{},
			},
			expectedError: false,
		},
		{
			name:                "no hosts field present",
			staticWorkersConfig: kubeone.StaticWorkersConfig{},
			expectedError:       false,
		},
		{
			name: "invalid host config",
			staticWorkersConfig: kubeone.StaticWorkersConfig{
				Hosts: []kubeone.HostConfig{
					{
						PublicAddress:  "1.1.1.1",
						PrivateAddress: "10.0.0.1",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "ubuntu",
					},
					{
						PublicAddress:  "1.1.1.2",
						PrivateAddress: "10.0.0.2",
						SSHAgentSocket: "env:SSH_AUTH_SOCK",
						SSHUsername:    "",
					},
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateStaticWorkersConfig(tc.staticWorkersConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateDynamicWorkerConfig(t *testing.T) {
	tests := []struct {
		name                string
		dynamicWorkerConfig []kubeone.DynamicWorkerConfig
		expectedError       bool
	}{
		{
			name: "valid worker config",
			dynamicWorkerConfig: []kubeone.DynamicWorkerConfig{
				{
					Name:     "test-1",
					Replicas: intPtr(3),
				},
				{
					Name:     "test-2",
					Replicas: intPtr(5),
				},
				{
					Name:     "test-3",
					Replicas: intPtr(0),
				},
			},
			expectedError: false,
		},
		{
			name:                "valid worker config (no worker defined)",
			dynamicWorkerConfig: []kubeone.DynamicWorkerConfig{},
			expectedError:       false,
		},
		{
			name: "invalid worker config (replicas not provided)",
			dynamicWorkerConfig: []kubeone.DynamicWorkerConfig{
				{
					Name:     "test-1",
					Replicas: intPtr(3),
				},
				{
					Name: "test-2",
				},
			},
			expectedError: true,
		},
		{
			name: "invalid worker config (no name given)",
			dynamicWorkerConfig: []kubeone.DynamicWorkerConfig{
				{
					Replicas: intPtr(3),
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateDynamicWorkerConfig(tc.dynamicWorkerConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateFeatures(t *testing.T) {
	tests := []struct {
		name          string
		features      kubeone.Features
		expectedError bool
	}{
		{
			name: "psp and auditing enabled",
			features: kubeone.Features{
				PodSecurityPolicy: &kubeone.PodSecurityPolicy{
					Enable: true,
				},
				DynamicAuditLog: &kubeone.DynamicAuditLog{
					Enable: true,
				},
			},
			expectedError: false,
		},
		{
			name: "metrics server disabled",
			features: kubeone.Features{
				MetricsServer: &kubeone.MetricsServer{
					Enable: false,
				},
			},
			expectedError: false,
		},
		{
			name:          "no feature configured",
			features:      kubeone.Features{},
			expectedError: false,
		},
		{
			name: "oidc enabled",
			features: kubeone.Features{
				OpenIDConnect: &kubeone.OpenIDConnect{
					Enable: true,
					Config: kubeone.OpenIDConnectConfig{
						IssuerURL:     "test.cluster.local",
						ClientID:      "123",
						RequiredClaim: "test",
					},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid staticAudit config",
			features: kubeone.Features{
				StaticAuditLog: &kubeone.StaticAuditLog{
					Enable: true,
					Config: kubeone.StaticAuditLogConfig{},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid oidc config",
			features: kubeone.Features{
				OpenIDConnect: &kubeone.OpenIDConnect{
					Enable: true,
					Config: kubeone.OpenIDConnectConfig{},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid podNodeSelector config",
			features: kubeone.Features{
				PodNodeSelector: &kubeone.PodNodeSelector{
					Enable: true,
					Config: kubeone.PodNodeSelectorConfig{},
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateFeatures(tc.features, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidatePodNodeSelectorConfig(t *testing.T) {
	tests := []struct {
		name                  string
		podNodeSelectorConfig kubeone.PodNodeSelectorConfig
		expectedError         bool
	}{
		{
			name: "valid podNodeSelector config",
			podNodeSelectorConfig: kubeone.PodNodeSelectorConfig{
				ConfigFilePath: "./podnodeselector.yaml",
			},
			expectedError: false,
		},
		{
			name:                  "invalid podNodeSelector config",
			podNodeSelectorConfig: kubeone.PodNodeSelectorConfig{},
			expectedError:         true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidatePodNodeSelectorConfig(tc.podNodeSelectorConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateStaticAuditLogConfig(t *testing.T) {
	tests := []struct {
		name                 string
		staticAuditLogConfig kubeone.StaticAuditLogConfig
		expectedError        bool
	}{
		{
			name: "valid staticAuditLog config",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogPath:        "/var/log/kubernetes",
				LogMaxAge:      10,
				LogMaxBackup:   10,
				LogMaxSize:     100,
			},
			expectedError: false,
		},
		{
			name: "policy file path missing",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				LogPath:      "/var/log/kubernetes",
				LogMaxAge:    10,
				LogMaxBackup: 10,
				LogMaxSize:   100,
			},
			expectedError: true,
		},
		{
			name: "log file path missing",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogMaxAge:      10,
				LogMaxBackup:   10,
				LogMaxSize:     100,
			},
			expectedError: true,
		},
		{
			name: "log max age set to 0",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogPath:        "/var/log/kubernetes",
				LogMaxAge:      0,
				LogMaxBackup:   10,
				LogMaxSize:     100,
			},
			expectedError: true,
		},
		{
			name: "log max backup set to 0",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogPath:        "/var/log/kubernetes",
				LogMaxAge:      10,
				LogMaxBackup:   0,
				LogMaxSize:     100,
			},
			expectedError: true,
		},
		{
			name: "log max size set to 0",
			staticAuditLogConfig: kubeone.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogPath:        "/var/log/kubernetes",
				LogMaxAge:      10,
				LogMaxBackup:   10,
				LogMaxSize:     0,
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateStaticAuditLogConfig(tc.staticAuditLogConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateOIDCConfig(t *testing.T) {
	tests := []struct {
		name          string
		oidcConfig    kubeone.OpenIDConnectConfig
		expectedError bool
	}{
		{
			name: "valid oidc config",
			oidcConfig: kubeone.OpenIDConnectConfig{
				IssuerURL: "test.cluster.local",
				ClientID:  "test",
			},
			expectedError: false,
		},
		{
			name: "no issuer url",
			oidcConfig: kubeone.OpenIDConnectConfig{
				ClientID: "test",
			},
			expectedError: true,
		},
		{
			name: "no client id",
			oidcConfig: kubeone.OpenIDConnectConfig{
				IssuerURL: "test.cluster.local",
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateOIDCConfig(tc.oidcConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateAddons(t *testing.T) {
	tests := []struct {
		name          string
		addons        *kubeone.Addons
		expectedError bool
	}{
		{
			name: "valid addons config (enabled)",
			addons: &kubeone.Addons{
				Enable: true,
				Path:   "./addons",
			},
			expectedError: false,
		},
		{
			name: "valid addons config (disabled)",
			addons: &kubeone.Addons{
				Enable: false,
			},
			expectedError: false,
		},
		{
			name:          "valid addons config (empty)",
			addons:        &kubeone.Addons{},
			expectedError: false,
		},
		{
			name:          "valid addons config (nil)",
			addons:        nil,
			expectedError: false,
		},
		{
			name: "invalid addons config (enabled without path)",
			addons: &kubeone.Addons{
				Enable: true,
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateAddons(tc.addons, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Log(errs[0])
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateHostConfig(t *testing.T) {
	tests := []struct {
		name          string
		hostConfig    []kubeone.HostConfig
		expectedError bool
	}{
		{
			name: "host config with ip addresses",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
				},
			},
			expectedError: false,
		},
		{
			name: "host config with dns domain",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
				},
			},
			expectedError: false,
		},
		{
			name: "no public address provided",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
				},
			},
			expectedError: true,
		},
		{
			name: "no private address provided",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
				},
			},
			expectedError: true,
		},
		{
			name: "no private key file and agent provided",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "",
					SSHAgentSocket:    "",
					SSHUsername:       "root",
				},
			},
			expectedError: true,
		},
		{
			name: "no username provided",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "",
				},
			},
			expectedError: true,
		},
		{
			name: "one valid host config and one invalid host config (no username)",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
				},
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "",
				},
			},
			expectedError: true,
		},
		{
			name: "two leaders at the same time",
			hostConfig: []kubeone.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					IsLeader:          true,
				},
				{
					PublicAddress:     "cluster-test.public.local",
					PrivateAddress:    "cluster-test.private.local",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					IsLeader:          true,
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateHostConfig(tc.hostConfig, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
