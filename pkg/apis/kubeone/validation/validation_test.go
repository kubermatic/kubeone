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

	"github.com/MakeNowJust/heredoc/v2"

	"k8c.io/kubeone/pkg/apis/kubeone"

	"k8s.io/apimachinery/pkg/util/validation/field"
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
					Kubernetes: "1.22.1",
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
					Kubernetes: "1.22.1",
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
					Kubernetes: "1.22.1",
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
		{
			name: "vSphere 1.22.0 cluster",
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
					Vsphere: &kubeone.VsphereSpec{},
				},
				Versions: kubeone.VersionConfig{
					Kubernetes: "1.22.1",
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
			name: "vSphere provider config without csiConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere:     &kubeone.VsphereSpec{},
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "vSphere provider config with csiConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere:     &kubeone.VsphereSpec{},
				External:    true,
				CloudConfig: "test",
				CSIConfig:   "test",
			},
			expectedError: false,
		},
		{
			name: "vSphere provider config with csiConfig (external disabled)",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere:     &kubeone.VsphereSpec{},
				External:    false,
				CloudConfig: "test",
				CSIConfig:   "test",
			},
			expectedError: true,
		},
		{
			name: "OpenStack provider config without csiConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Openstack:   &kubeone.OpenstackSpec{},
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "OpenStack provider config with csiConfig",
			providerConfig: kubeone.CloudProviderSpec{
				Openstack:   &kubeone.OpenstackSpec{},
				CloudConfig: "test",
				CSIConfig:   "test",
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
			name: "valid version config (1.22.1)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.22.1",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.22.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.22.2",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.21.4)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.21.4",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.21.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.21.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.20.10)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.20.10",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.20.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.20.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.19.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.19.0",
			},
			expectedError: false,
		},
		{
			name: "not supported kubernetes version (1.18.19)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.18.19",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.18.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.18.0",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.17.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.17.0",
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
				Kubernetes: "v1.22.1",
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

func TestValidateCloudProviderSupportsKubernetes(t *testing.T) {
	tests := []struct {
		name           string
		providerConfig kubeone.CloudProviderSpec
		versionConfig  kubeone.VersionConfig
		expectedError  bool
	}{
		{
			name: "AWS 1.21.4 cluster",
			providerConfig: kubeone.CloudProviderSpec{
				AWS: &kubeone.AWSSpec{},
			},
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.21.4",
			},
			expectedError: false,
		},
		{
			name: "AWS 1.22.1 cluster",
			providerConfig: kubeone.CloudProviderSpec{
				AWS: &kubeone.AWSSpec{},
			},
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.22.1",
			},
			expectedError: false,
		},
		{
			name: "vSphere 1.21.4 cluster",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere: &kubeone.VsphereSpec{},
			},
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.21.4",
			},
			expectedError: false,
		},
		{
			name: "vSphere 1.22.1 cluster",
			providerConfig: kubeone.CloudProviderSpec{
				Vsphere: &kubeone.VsphereSpec{},
			},
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.22.1",
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := kubeone.KubeOneCluster{
				CloudProvider: tc.providerConfig,
				Versions:      tc.versionConfig,
			}

			errs := ValidateCloudProviderSupportsKubernetes(c, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateContainerRuntimeConfig(t *testing.T) {
	tests := []struct {
		name             string
		containerRuntime kubeone.ContainerRuntimeConfig
		versions         kubeone.VersionConfig
		expectedError    bool
	}{
		{
			name:             "only docker defined",
			containerRuntime: kubeone.ContainerRuntimeConfig{Docker: &kubeone.ContainerRuntimeDocker{}},
			versions:         kubeone.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name:             "docker with kubernetes 1.22+",
			containerRuntime: kubeone.ContainerRuntimeConfig{Docker: &kubeone.ContainerRuntimeDocker{}},
			versions:         kubeone.VersionConfig{Kubernetes: "1.22"},
			expectedError:    true,
		},
		{
			name:             "only containerd defined",
			containerRuntime: kubeone.ContainerRuntimeConfig{Containerd: &kubeone.ContainerRuntimeContainerd{}},
			versions:         kubeone.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name: "both defined",
			containerRuntime: kubeone.ContainerRuntimeConfig{
				Docker:     &kubeone.ContainerRuntimeDocker{},
				Containerd: &kubeone.ContainerRuntimeContainerd{},
			},
			versions:      kubeone.VersionConfig{Kubernetes: "1.20"},
			expectedError: true,
		},
		{
			name:             "non defined",
			containerRuntime: kubeone.ContainerRuntimeConfig{},
			versions:         kubeone.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name:             "non defined, 1.21+",
			containerRuntime: kubeone.ContainerRuntimeConfig{},
			versions:         kubeone.VersionConfig{Kubernetes: "1.21"},
			expectedError:    false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateContainerRuntimeConfig(tc.containerRuntime, tc.versions, &field.Path{})
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

func TestValidateCABundle(t *testing.T) {
	tests := []struct {
		name          string
		caBundle      string
		expectedError bool
	}{
		{
			name:          "empty",
			caBundle:      "",
			expectedError: false,
		},
		{
			name: "correct",
			caBundle: heredoc.Doc(`
				## some comments

				GlobalSign Root CA
				==================
				-----BEGIN CERTIFICATE-----
				MIIDdTCCAl2gAwIBAgILBAAAAAABFUtaw5QwDQYJKoZIhvcNAQEFBQAwVzELMAkGA1UEBhMCQkUx
				GTAXBgNVBAoTEEdsb2JhbFNpZ24gbnYtc2ExEDAOBgNVBAsTB1Jvb3QgQ0ExGzAZBgNVBAMTEkds
				b2JhbFNpZ24gUm9vdCBDQTAeFw05ODA5MDExMjAwMDBaFw0yODAxMjgxMjAwMDBaMFcxCzAJBgNV
				BAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWduIG52LXNhMRAwDgYDVQQLEwdSb290IENBMRswGQYD
				VQQDExJHbG9iYWxTaWduIFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDa
				DuaZjc6j40+Kfvvxi4Mla+pIH/EqsLmVEQS98GPR4mdmzxzdzxtIK+6NiY6arymAZavpxy0Sy6sc
				THAHoT0KMM0VjU/43dSMUBUc71DuxC73/OlS8pF94G3VNTCOXkNz8kHp1Wrjsok6Vjk4bwY8iGlb
				Kk3Fp1S4bInMm/k8yuX9ifUSPJJ4ltbcdG6TRGHRjcdGsnUOhugZitVtbNV4FpWi6cgKOOvyJBNP
				c1STE4U6G7weNLWLBYy5d4ux2x8gkasJU26Qzns3dLlwR5EiUWMWea6xrkEmCMgZK9FGqkjWZCrX
				gzT/LCrBbBlDSgeF59N89iFo7+ryUp9/k5DPAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNV
				HRMBAf8EBTADAQH/MB0GA1UdDgQWBBRge2YaRQ2XyolQL30EzTSo//z9SzANBgkqhkiG9w0BAQUF
				AAOCAQEA1nPnfE920I2/7LqivjTFKDK1fPxsnCwrvQmeU79rXqoRSLblCKOzyj1hTdNGCbM+w6Dj
				Y1Ub8rrvrTnhQ7k4o+YviiY776BQVvnGCv04zcQLcFGUl5gE38NflNUVyRRBnMRddWQVDf9VMOyG
				j/8N7yy5Y0b2qvzfvGn9LhJIZJrglfCm7ymPAbEVtQwdpf5pLGkkeB6zpxxxYu7KyJesF12KwvhH
				hm4qxFYxldBniYUr+WymXUadDKqC5JlR3XC321Y9YeRq4VzW9v493kHMB65jUr9TU/Qr6cf9tveC
				X4XSQRjbgbMEHMUfpIBvFSDJ3gyICh3WZlXi/EjJKSZp4A==
				-----END CERTIFICATE-----
			`),
			expectedError: false,
		},
		{
			name: "no certs but with comments",
			caBundle: heredoc.Doc(`
				# leading comment
				## additional comment
			`),
			expectedError: true,
		},
		{
			name:          "incorrect",
			caBundle:      "garbadge",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateCABundle(tc.caBundle, field.NewPath("caBundle"))
			if (len(errs) == 0) == tc.expectedError {
				t.Logf("failed value:\n%q", tc.caBundle)
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateFeatures(t *testing.T) {
	tests := []struct {
		name          string
		features      kubeone.Features
		versions      kubeone.VersionConfig
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: false,
		},
		{
			name:     "no feature configured",
			features: kubeone.Features{},
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
			},
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
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
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: true,
		},
		{
			name: "podPresets enabled on 1.19 cluster",
			features: kubeone.Features{
				PodPresets: &kubeone.PodPresets{
					Enable: true,
				},
			},
			versions: kubeone.VersionConfig{
				Kubernetes: "1.19.7",
			},
			expectedError: false,
		},
		{
			name: "podPresets enabled on 1.20 cluster",
			features: kubeone.Features{
				PodPresets: &kubeone.PodPresets{
					Enable: true,
				},
			},
			versions: kubeone.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: true,
		},
		{
			name: "podPresets enabled on 1.21 cluster",
			features: kubeone.Features{
				PodPresets: &kubeone.PodPresets{
					Enable: true,
				},
			},
			versions: kubeone.VersionConfig{
				Kubernetes: "1.21.0",
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateFeatures(tc.features, tc.versions, nil)
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
			name: "addons enabled, no path set",
			addons: &kubeone.Addons{
				Enable: true,
				Path:   "",
			},
			expectedError: true,
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

func TestValidateRegistryConfiguration(t *testing.T) {
	tests := []struct {
		name                  string
		registryConfiguration *kubeone.RegistryConfiguration
		expectedError         bool
	}{
		{
			name: "valid registry config (overwrite registry)",
			registryConfiguration: &kubeone.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
			},
			expectedError: false,
		},
		{
			name: "valid registry config (overwrite registry and insecure)",
			registryConfiguration: &kubeone.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
				InsecureRegistry:  true,
			},
			expectedError: false,
		},
		{
			name:                  "valid registry config (empty)",
			registryConfiguration: &kubeone.RegistryConfiguration{},
			expectedError:         false,
		},
		{
			name:                  "valid registry config (nil)",
			registryConfiguration: nil,
			expectedError:         false,
		},
		{
			name: "invalid registry config (insecure registry without overwrite registry)",
			registryConfiguration: &kubeone.RegistryConfiguration{
				InsecureRegistry: true,
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateRegistryConfiguration(tc.registryConfiguration, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Log(errs[0])
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateAssetConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		assetConfiguration *kubeone.AssetConfiguration
		expectedError      bool
	}{
		{
			name:               "empty asset configuration",
			assetConfiguration: &kubeone.AssetConfiguration{},
			expectedError:      false,
		},
		{
			name: "kubernetes image configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Kubernetes: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "kubernetes image and tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Kubernetes: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: true,
		},
		{
			name: "kubernetes tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Kubernetes: kubeone.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: true,
		},
		{
			name: "pause image configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Pause: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "3.2",
				},
			},
			expectedError: false,
		},
		{
			name: "pause image configured (repository missing)",
			assetConfiguration: &kubeone.AssetConfiguration{
				Pause: kubeone.ImageAsset{
					ImageTag: "3.2",
				},
			},
			expectedError: true,
		},
		{
			name: "pause image configured (tag missing)",
			assetConfiguration: &kubeone.AssetConfiguration{
				Pause: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: true,
		},
		{
			name: "coredns image and tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				CoreDNS: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "coredns image configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				CoreDNS: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "coredns tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				CoreDNS: kubeone.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd image and tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Etcd: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd image configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Etcd: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				Etcd: kubeone.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server image and tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				MetricsServer: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server image configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				MetricsServer: kubeone.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server tag configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				MetricsServer: kubeone.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "cni, node binaries, and kubectl configured",
			assetConfiguration: &kubeone.AssetConfiguration{
				CNI: kubeone.BinaryAsset{
					URL: "https://127.0.0.1/cni",
				},
				NodeBinaries: kubeone.BinaryAsset{
					URL: "https://127.0.0.1/kubernetes-node-linux-amd64.tar.gz",
				},
				Kubectl: kubeone.BinaryAsset{
					URL: "https://127.0.0.1/kubectl",
				},
			},
			expectedError: false,
		},
		{
			name: "binary assets configured (node binaries missing)",
			assetConfiguration: &kubeone.AssetConfiguration{
				CNI: kubeone.BinaryAsset{
					URL: "https://127.0.0.1/cni",
				},
				Kubectl: kubeone.BinaryAsset{
					URL: "https://127.0.0.1/kubectl",
				},
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateAssetConfiguration(tc.assetConfiguration, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Log(errs[0])
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
