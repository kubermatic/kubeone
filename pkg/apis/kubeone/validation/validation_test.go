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

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/resources"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
)

func TestValidateKubeOneCluster(t *testing.T) {
	tests := []struct {
		name          string
		cluster       kubeoneapi.KubeOneCluster
		expectedError bool
	}{
		{
			name: "valid KubeOneCluster config",
			cluster: kubeoneapi.KubeOneCluster{
				Name: "test",
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
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
				APIEndpoint: kubeoneapi.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					AWS: &kubeoneapi.AWSSpec{},
				},
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.22.1",
				},
				MachineController: &kubeoneapi.MachineControllerConfig{
					Deploy: true,
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
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
			cluster: kubeoneapi.KubeOneCluster{
				Name: "test",
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
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
				APIEndpoint: kubeoneapi.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					AWS: &kubeoneapi.AWSSpec{},
				},
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.22.1",
				},
				MachineController: &kubeoneapi.MachineControllerConfig{
					Deploy: false,
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
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
			cluster: kubeoneapi.KubeOneCluster{
				Name: "",
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
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
				APIEndpoint: kubeoneapi.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					AWS: &kubeoneapi.AWSSpec{},
				},
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.22.1",
				},
				MachineController: &kubeoneapi.MachineControllerConfig{
					Deploy: true,
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
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
			cluster: kubeoneapi.KubeOneCluster{
				Name: "test",
				ControlPlane: kubeoneapi.ControlPlaneConfig{
					Hosts: []kubeoneapi.HostConfig{
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
				APIEndpoint: kubeoneapi.APIEndpoint{
					Host: "localhost",
					Port: 6443,
				},
				CloudProvider: kubeoneapi.CloudProviderSpec{
					Vsphere: &kubeoneapi.VsphereSpec{},
				},
				Versions: kubeoneapi.VersionConfig{
					Kubernetes: "1.22.1",
				},
				MachineController: &kubeoneapi.MachineControllerConfig{
					Deploy: true,
				},
				DynamicWorkers: []kubeoneapi.DynamicWorkerConfig{
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

func TestValdiateName(t *testing.T) {
	tests := []struct {
		name          string
		clusterName   string
		expectedError bool
	}{
		{
			name:          "valid cluster name",
			clusterName:   "test",
			expectedError: false,
		},
		{
			name:          "valid cluster name (with periods)",
			clusterName:   "test-1",
			expectedError: false,
		},
		{
			name:          "valid cluster name (with dots)",
			clusterName:   "test.example.com",
			expectedError: false,
		},
		{
			name:          "valid cluster name (with periods and dots)",
			clusterName:   "test-1.example.com",
			expectedError: false,
		},
		{
			name:          "valid cluster name (starts with number)",
			clusterName:   "1test",
			expectedError: false,
		},
		{
			name:          "invalid cluster name (empty)",
			clusterName:   "",
			expectedError: true,
		},
		{
			name:          "invalid cluster name (underscore)",
			clusterName:   "test_1.example.com",
			expectedError: true,
		},
		{
			name:          "invalid cluster name (uppercase)",
			clusterName:   "Test",
			expectedError: true,
		},
		{
			name:          "invalid cluster name (starts with dot)",
			clusterName:   ".test",
			expectedError: true,
		},
		{
			name:          "invalid cluster name (ends with dot)",
			clusterName:   "test.",
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateName(tc.clusterName, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateControlPlaneConfig(t *testing.T) {
	tests := []struct {
		name               string
		controlPlaneConfig kubeoneapi.ControlPlaneConfig
		expectedError      bool
	}{
		{
			name: "valid ControlPlane config",
			controlPlaneConfig: kubeoneapi.ControlPlaneConfig{
				Hosts: []kubeoneapi.HostConfig{
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
			controlPlaneConfig: kubeoneapi.ControlPlaneConfig{
				Hosts: []kubeoneapi.HostConfig{
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
			controlPlaneConfig: kubeoneapi.ControlPlaneConfig{
				Hosts: []kubeoneapi.HostConfig{},
			},
			expectedError: true,
		},
		{
			name:               "no hosts field present",
			controlPlaneConfig: kubeoneapi.ControlPlaneConfig{},
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
		apiEndpoint   kubeoneapi.APIEndpoint
		expectedError bool
	}{
		{
			name: "valid apiEndpoint config (localhost:6443)",
			apiEndpoint: kubeoneapi.APIEndpoint{
				Host: "localhost",
				Port: 6443,
			},
			expectedError: false,
		},
		{
			name: "valid apiEndpoint config (example.com:443)",
			apiEndpoint: kubeoneapi.APIEndpoint{
				Host: "example.com",
				Port: 443,
			},
			expectedError: false,
		},
		{
			name: "no host specified",
			apiEndpoint: kubeoneapi.APIEndpoint{
				Port: 6443,
			},
			expectedError: true,
		},
		{
			name: "no port specified",
			apiEndpoint: kubeoneapi.APIEndpoint{
				Host: "localhost",
			},
			expectedError: true,
		},
		{
			name: "port lower than 0",
			apiEndpoint: kubeoneapi.APIEndpoint{
				Host: "localhost",
				Port: -1,
			},
			expectedError: true,
		},
		{
			name: "port greater than 65535",
			apiEndpoint: kubeoneapi.APIEndpoint{
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
		providerConfig kubeoneapi.CloudProviderSpec
		expectedError  bool
	}{
		{
			name: "valid AWS provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS: &kubeoneapi.AWSSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid Azure provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Azure:       &kubeoneapi.AzureSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid DigitalOcean provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid GCE provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				GCE: &kubeoneapi.GCESpec{},
			},
			expectedError: false,
		},
		{
			name: "valid Hetzner provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Hetzner: &kubeoneapi.HetznerSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid Nutanix provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Nutanix: &kubeoneapi.NutanixSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid OpenStack provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Openstack:   &kubeoneapi.OpenstackSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid Equinix Metal provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				EquinixMetal: &kubeoneapi.EquinixMetalSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid VMware Cloud Director provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				VMwareCloudDirector: &kubeoneapi.VMwareCloudDirectorSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid vSphere provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				CloudConfig: "cloud-config",
			},
			expectedError: false,
		},
		{
			name: "valid None provider config",
			providerConfig: kubeoneapi.CloudProviderSpec{
				None: &kubeoneapi.NoneSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid OpenStack provider config with external CCM and cloudConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:         &kubeoneapi.AWSSpec{},
				CloudConfig: "cloud-config",
				External:    true,
			},
			expectedError: false,
		},
		{
			name: "valid DigitalOcean provider config with external CCM",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:      &kubeoneapi.AWSSpec{},
				External: true,
			},
			expectedError: false,
		},
		{
			name: "AWS and Azure specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:   &kubeoneapi.AWSSpec{},
				Azure: &kubeoneapi.AzureSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and DigitalOcean specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:          &kubeoneapi.AWSSpec{},
				DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and GCE specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS: &kubeoneapi.AWSSpec{},
				GCE: &kubeoneapi.GCESpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and Hetzner specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:     &kubeoneapi.AWSSpec{},
				Hetzner: &kubeoneapi.HetznerSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and OpenStack specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:       &kubeoneapi.AWSSpec{},
				Openstack: &kubeoneapi.OpenstackSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and Equinix Metal specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:          &kubeoneapi.AWSSpec{},
				EquinixMetal: &kubeoneapi.EquinixMetalSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and vSphere specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:     &kubeoneapi.AWSSpec{},
				Vsphere: &kubeoneapi.VsphereSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS and None specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:  &kubeoneapi.AWSSpec{},
				None: &kubeoneapi.NoneSpec{},
			},
			expectedError: true,
		},
		{
			name: "AWS, Azure, and DigitalOcean specified at the same time",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS:          &kubeoneapi.AWSSpec{},
				Azure:        &kubeoneapi.AzureSpec{},
				DigitalOcean: &kubeoneapi.DigitalOceanSpec{},
			},
			expectedError: true,
		},
		{
			name: "Azure provider config without cloudConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Azure: &kubeoneapi.AzureSpec{},
			},
			expectedError: true,
		},
		{
			name: "Nutanix provider config with external enabled",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Nutanix:  &kubeoneapi.NutanixSpec{},
				External: true,
			},
			expectedError: true,
		},
		{
			name: "OpenStack provider config without cloudConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Openstack: &kubeoneapi.OpenstackSpec{},
			},
			expectedError: true,
		},
		{
			name: "vSphere provider config without cloudConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere: &kubeoneapi.VsphereSpec{},
			},
			expectedError: true,
		},
		{
			name: "vSphere provider config without csiConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "vSphere provider config with csiConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				External:    true,
				CloudConfig: "test",
				CSIConfig:   "test",
			},
			expectedError: false,
		},
		{
			name: "vSphere provider config with csiConfig (external disabled)",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere:     &kubeoneapi.VsphereSpec{},
				External:    false,
				CloudConfig: "test",
				CSIConfig:   "test",
			},
			expectedError: true,
		},
		{
			name: "OpenStack provider config without csiConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Openstack:   &kubeoneapi.OpenstackSpec{},
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "OpenStack provider config with csiConfig",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Openstack:   &kubeoneapi.OpenstackSpec{},
				CloudConfig: "test",
				CSIConfig:   "test",
			},
			expectedError: true,
		},
		{
			name:           "no provider specified",
			providerConfig: kubeoneapi.CloudProviderSpec{},
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
		versionConfig kubeoneapi.VersionConfig
		expectedError bool
	}{
		{
			name: "valid version config (1.23.1)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.23.1",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.22.1)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.22.1",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.22.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.22.2",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.21.4)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.21.4",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.21.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.21.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.20.10)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.10",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.20.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.0",
			},
			expectedError: false,
		},
		{
			name: "not supported kubernetes version (1.99.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.99.0",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.19.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.19.0",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.18.19)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.18.19",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.18.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.18.0",
			},
			expectedError: true,
		},
		{
			name: "not supported kubernetes version (1.17.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.17.0",
			},
			expectedError: true,
		},
		{
			name: "invalid kubernetes version (2.0.0)",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "2.0.0",
			},
			expectedError: true,
		},
		{
			name: "kubernetes version with a leading 'v'",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "v1.22.1",
			},
			expectedError: true,
		},
		{
			name: "invalid semver kubernetes version",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "version-1.19.0",
			},
			expectedError: true,
		},
		{
			name: "not supported eks-d cluster",
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "v1.19.9-eks-1-18-1",
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

func TestValidateKubernetesSupport(t *testing.T) {
	tests := []struct {
		name           string
		providerConfig kubeoneapi.CloudProviderSpec
		networkConfig  kubeoneapi.ClusterNetworkConfig
		versionConfig  kubeoneapi.VersionConfig
		addonsConfig   *kubeoneapi.Addons
		expectedError  bool
	}{
		{
			name: "AWS 1.21.4 cluster",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS: &kubeoneapi.AWSSpec{},
			},
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.21.4",
			},
			expectedError: false,
		},
		{
			name: "AWS 1.22.1 cluster",
			providerConfig: kubeoneapi.CloudProviderSpec{
				AWS: &kubeoneapi.AWSSpec{},
			},
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.22.1",
			},
			expectedError: false,
		},
		{
			name: "vSphere 1.22.4 cluster",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere: &kubeoneapi.VsphereSpec{},
			},
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.22.4",
			},
			expectedError: false,
		},
		{
			name: "vSphere 1.25.0 cluster",
			providerConfig: kubeoneapi.CloudProviderSpec{
				Vsphere: &kubeoneapi.VsphereSpec{},
			},
			versionConfig: kubeoneapi.VersionConfig{
				Kubernetes: "1.25.0",
			},
			expectedError: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := kubeoneapi.KubeOneCluster{
				CloudProvider:  tc.providerConfig,
				Versions:       tc.versionConfig,
				ClusterNetwork: tc.networkConfig,
				Addons:         tc.addonsConfig,
			}

			errs := ValidateKubernetesSupport(c, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateContainerRuntimeConfig(t *testing.T) {
	tests := []struct {
		name             string
		containerRuntime kubeoneapi.ContainerRuntimeConfig
		versions         kubeoneapi.VersionConfig
		expectedError    bool
	}{
		{
			name:             "only docker defined",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{Docker: &kubeoneapi.ContainerRuntimeDocker{}},
			versions:         kubeoneapi.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name:             "docker with kubernetes 1.24+",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{Docker: &kubeoneapi.ContainerRuntimeDocker{}},
			versions:         kubeoneapi.VersionConfig{Kubernetes: "1.24"},
			expectedError:    true,
		},
		{
			name:             "only containerd defined",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{Containerd: &kubeoneapi.ContainerRuntimeContainerd{}},
			versions:         kubeoneapi.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name: "both defined",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{
				Docker:     &kubeoneapi.ContainerRuntimeDocker{},
				Containerd: &kubeoneapi.ContainerRuntimeContainerd{},
			},
			versions:      kubeoneapi.VersionConfig{Kubernetes: "1.20"},
			expectedError: true,
		},
		{
			name:             "non defined",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{},
			versions:         kubeoneapi.VersionConfig{Kubernetes: "1.20"},
			expectedError:    false,
		},
		{
			name:             "non defined, 1.21+",
			containerRuntime: kubeoneapi.ContainerRuntimeConfig{},
			versions:         kubeoneapi.VersionConfig{Kubernetes: "1.21"},
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
		clusterNetworkConfig kubeoneapi.ClusterNetworkConfig
		expectedError        bool
	}{
		{
			name: "valid network config",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: false,
		},
		{
			name: "valid network config with cni config",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0/24",
				CNI: &kubeoneapi.CNI{
					Canal: &kubeoneapi.CanalSpec{MTU: 1500},
				},
			},
			expectedError: false,
		},
		{
			name:                 "empty network config",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{},
			expectedError:        false,
		},
		{
			name: "invalid pod subnet",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: true,
		},
		{
			name: "invalid service subnet (non-CIDR)",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0",
			},
			expectedError: true,
		},
		{
			name: "invalid cni config",
			clusterNetworkConfig: kubeoneapi.ClusterNetworkConfig{
				CNI: &kubeoneapi.CNI{
					Canal:    &kubeoneapi.CanalSpec{},
					WeaveNet: &kubeoneapi.WeaveNetSpec{},
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
		cniConfig     *kubeoneapi.CNI
		expectedError bool
	}{
		{
			name: "valid Canal CNI config",
			cniConfig: &kubeoneapi.CNI{
				Canal: &kubeoneapi.CanalSpec{MTU: 1500},
			},
			expectedError: false,
		},
		{
			name: "valid WeaveNet CNI config",
			cniConfig: &kubeoneapi.CNI{
				WeaveNet: &kubeoneapi.WeaveNetSpec{},
			},
			expectedError: false,
		},
		{
			name: "valid WeaveNet CNI config with encryption enabled",
			cniConfig: &kubeoneapi.CNI{
				WeaveNet: &kubeoneapi.WeaveNetSpec{
					Encrypted: true,
				},
			},
			expectedError: false,
		},
		{
			name: "valid External CNI config",
			cniConfig: &kubeoneapi.CNI{
				External: &kubeoneapi.ExternalCNISpec{},
			},
			expectedError: false,
		},
		{
			name: "Canal and WeaveNet specified at the same time",
			cniConfig: &kubeoneapi.CNI{
				Canal:    &kubeoneapi.CanalSpec{},
				WeaveNet: &kubeoneapi.WeaveNetSpec{},
			},
			expectedError: true,
		},
		{
			name: "Canal and External specified at the same time",
			cniConfig: &kubeoneapi.CNI{
				Canal:    &kubeoneapi.CanalSpec{},
				External: &kubeoneapi.ExternalCNISpec{},
			},
			expectedError: true,
		},
		{
			name: "WeaveNet and External specified at the same time",
			cniConfig: &kubeoneapi.CNI{
				WeaveNet: &kubeoneapi.WeaveNetSpec{},
				External: &kubeoneapi.ExternalCNISpec{},
			},
			expectedError: true,
		},
		{
			name:          "no CNI config specified",
			cniConfig:     &kubeoneapi.CNI{},
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
		staticWorkersConfig kubeoneapi.StaticWorkersConfig
		expectedError       bool
	}{
		{
			name: "valid StaticWorkers config",
			staticWorkersConfig: kubeoneapi.StaticWorkersConfig{
				Hosts: []kubeoneapi.HostConfig{
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
			staticWorkersConfig: kubeoneapi.StaticWorkersConfig{
				Hosts: []kubeoneapi.HostConfig{},
			},
			expectedError: false,
		},
		{
			name:                "no hosts field present",
			staticWorkersConfig: kubeoneapi.StaticWorkersConfig{},
			expectedError:       false,
		},
		{
			name: "invalid host config",
			staticWorkersConfig: kubeoneapi.StaticWorkersConfig{
				Hosts: []kubeoneapi.HostConfig{
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
		dynamicWorkerConfig []kubeoneapi.DynamicWorkerConfig
		expectedError       bool
	}{
		{
			name: "valid worker config",
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
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
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{},
			expectedError:       false,
		},
		{
			name: "invalid worker config (replicas not provided)",
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
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
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
				{
					Replicas: intPtr(3),
				},
			},
			expectedError: true,
		},
		{
			name: "only machineAnnotations set",
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
				{
					Name:     "test-1",
					Replicas: intPtr(3),
					Config: kubeoneapi.ProviderSpec{
						MachineAnnotations: map[string]string{"test": "test"},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "only nodeAnnotations set",
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
				{
					Name:     "test-1",
					Replicas: intPtr(3),
					Config: kubeoneapi.ProviderSpec{
						NodeAnnotations: map[string]string{"test": "test"},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "both machineAnnotations and nodeAnnotations set",
			dynamicWorkerConfig: []kubeoneapi.DynamicWorkerConfig{
				{
					Name:     "test-1",
					Replicas: intPtr(3),
					Config: kubeoneapi.ProviderSpec{
						MachineAnnotations: map[string]string{"test": "test"},
						NodeAnnotations:    map[string]string{"test": "test"},
					},
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
		features      kubeoneapi.Features
		versions      kubeoneapi.VersionConfig
		expectedError bool
	}{
		{
			name: "psp and auditing enabled",
			features: kubeoneapi.Features{
				PodSecurityPolicy: &kubeoneapi.PodSecurityPolicy{
					Enable: true,
				},
				DynamicAuditLog: &kubeoneapi.DynamicAuditLog{
					Enable: true,
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: false,
		},
		{
			name: "metrics server disabled",
			features: kubeoneapi.Features{
				MetricsServer: &kubeoneapi.MetricsServer{
					Enable: false,
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: false,
		},
		{
			name:     "no feature configured",
			features: kubeoneapi.Features{},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: false,
		},
		{
			name: "oidc enabled",
			features: kubeoneapi.Features{
				OpenIDConnect: &kubeoneapi.OpenIDConnect{
					Enable: true,
					Config: kubeoneapi.OpenIDConnectConfig{
						IssuerURL:     "test.cluster.local",
						ClientID:      "123",
						RequiredClaim: "test",
					},
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: false,
		},
		{
			name: "invalid staticAudit config",
			features: kubeoneapi.Features{
				StaticAuditLog: &kubeoneapi.StaticAuditLog{
					Enable: true,
					Config: kubeoneapi.StaticAuditLogConfig{},
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: true,
		},
		{
			name: "invalid oidc config",
			features: kubeoneapi.Features{
				OpenIDConnect: &kubeoneapi.OpenIDConnect{
					Enable: true,
					Config: kubeoneapi.OpenIDConnectConfig{},
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
			},
			expectedError: true,
		},
		{
			name: "invalid podNodeSelector config",
			features: kubeoneapi.Features{
				PodNodeSelector: &kubeoneapi.PodNodeSelector{
					Enable: true,
					Config: kubeoneapi.PodNodeSelectorConfig{},
				},
			},
			versions: kubeoneapi.VersionConfig{
				Kubernetes: "1.20.2",
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
		podNodeSelectorConfig kubeoneapi.PodNodeSelectorConfig
		expectedError         bool
	}{
		{
			name: "valid podNodeSelector config",
			podNodeSelectorConfig: kubeoneapi.PodNodeSelectorConfig{
				ConfigFilePath: "./podnodeselector.yaml",
			},
			expectedError: false,
		},
		{
			name:                  "invalid podNodeSelector config",
			podNodeSelectorConfig: kubeoneapi.PodNodeSelectorConfig{},
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
		staticAuditLogConfig kubeoneapi.StaticAuditLogConfig
		expectedError        bool
	}{
		{
			name: "valid staticAuditLog config",
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
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
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
				LogPath:      "/var/log/kubernetes",
				LogMaxAge:    10,
				LogMaxBackup: 10,
				LogMaxSize:   100,
			},
			expectedError: true,
		},
		{
			name: "log file path missing",
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
				PolicyFilePath: "/etc/kubernetes/policy.yaml",
				LogMaxAge:      10,
				LogMaxBackup:   10,
				LogMaxSize:     100,
			},
			expectedError: true,
		},
		{
			name: "log max age set to 0",
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
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
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
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
			staticAuditLogConfig: kubeoneapi.StaticAuditLogConfig{
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
		oidcConfig    kubeoneapi.OpenIDConnectConfig
		expectedError bool
	}{
		{
			name: "valid oidc config",
			oidcConfig: kubeoneapi.OpenIDConnectConfig{
				IssuerURL: "test.cluster.local",
				ClientID:  "test",
			},
			expectedError: false,
		},
		{
			name: "no issuer url",
			oidcConfig: kubeoneapi.OpenIDConnectConfig{
				ClientID: "test",
			},
			expectedError: true,
		},
		{
			name: "no client id",
			oidcConfig: kubeoneapi.OpenIDConnectConfig{
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
		addons        *kubeoneapi.Addons
		expectedError bool
	}{
		{
			name: "valid addons config (enabled)",
			addons: &kubeoneapi.Addons{
				Enable: true,
				Path:   "./addons",
			},
			expectedError: false,
		},
		{
			name: "addons enabled, no path set and no embedded addons specified",
			addons: &kubeoneapi.Addons{
				Enable: true,
				Path:   "",
			},
			expectedError: true,
		},
		{
			name: "embedded addon enabled, no path set",
			addons: &kubeoneapi.Addons{
				Enable: true,
				Path:   "",
				Addons: []kubeoneapi.Addon{
					{
						Name: resources.AddonMachineController,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "valid addons config (disabled)",
			addons: &kubeoneapi.Addons{
				Enable: false,
			},
			expectedError: false,
		},
		{
			name:          "valid addons config (empty)",
			addons:        &kubeoneapi.Addons{},
			expectedError: false,
		},
		{
			name:          "valid addons config (nil)",
			addons:        nil,
			expectedError: false,
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
		hostConfig    []kubeoneapi.HostConfig
		expectedError bool
	}{
		{
			name: "host config with ip addresses",
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
			hostConfig: []kubeoneapi.HostConfig{
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
		{
			name: "valid OS",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					OperatingSystem:   kubeoneapi.OperatingSystemNameCentOS,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid OS",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					OperatingSystem:   kubeoneapi.OperatingSystemName("non-existing"),
				},
			},
			expectedError: true,
		},
		{
			name: "kubelet.maxPods valid",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					Kubelet: kubeoneapi.KubeletConfig{
						MaxPods: pointer.Int32Ptr(110),
					},
				},
			},
			expectedError: false,
		},
		{
			name: "kubelet.maxPods zero (invalid)",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					Kubelet: kubeoneapi.KubeletConfig{
						MaxPods: pointer.Int32Ptr(0),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "kubelet.maxPods negative (invalid)",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					Kubelet: kubeoneapi.KubeletConfig{
						MaxPods: pointer.Int32Ptr(-10),
					},
				},
			},
			expectedError: true,
		},
		{
			name: "incorrect label marked to remove",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					Labels: map[string]string{
						"label-to-remove-": "this values has to be empty",
					},
				},
			},
			expectedError: true,
		},
		{
			name: "correct label marked to remove",
			hostConfig: []kubeoneapi.HostConfig{
				{
					PublicAddress:     "192.168.1.1",
					PrivateAddress:    "192.168.0.1",
					SSHPrivateKeyFile: "test",
					SSHAgentSocket:    "test",
					SSHUsername:       "root",
					Labels: map[string]string{
						"label-to-remove-": "",
					},
				},
			},
			expectedError: false,
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
		registryConfiguration *kubeoneapi.RegistryConfiguration
		expectedError         bool
	}{
		{
			name: "valid registry config (overwrite registry)",
			registryConfiguration: &kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
			},
			expectedError: false,
		},
		{
			name: "valid registry config (overwrite registry and insecure)",
			registryConfiguration: &kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
				InsecureRegistry:  true,
			},
			expectedError: false,
		},
		{
			name:                  "valid registry config (empty)",
			registryConfiguration: &kubeoneapi.RegistryConfiguration{},
			expectedError:         false,
		},
		{
			name:                  "valid registry config (nil)",
			registryConfiguration: nil,
			expectedError:         false,
		},
		{
			name: "invalid registry config (insecure registry without overwrite registry)",
			registryConfiguration: &kubeoneapi.RegistryConfiguration{
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
		assetConfiguration *kubeoneapi.AssetConfiguration
		expectedError      bool
	}{
		{
			name:               "empty asset configuration",
			assetConfiguration: &kubeoneapi.AssetConfiguration{},
			expectedError:      false,
		},
		{
			name: "kubernetes image configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Kubernetes: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "kubernetes image and tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Kubernetes: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: true,
		},
		{
			name: "kubernetes tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Kubernetes: kubeoneapi.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: true,
		},
		{
			name: "pause image configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Pause: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "3.2",
				},
			},
			expectedError: false,
		},
		{
			name: "pause image configured (repository missing)",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Pause: kubeoneapi.ImageAsset{
					ImageTag: "3.2",
				},
			},
			expectedError: true,
		},
		{
			name: "pause image configured (tag missing)",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Pause: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: true,
		},
		{
			name: "coredns image and tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				CoreDNS: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "coredns image configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				CoreDNS: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "coredns tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				CoreDNS: kubeoneapi.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd image and tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Etcd: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd image configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Etcd: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "etcd tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				Etcd: kubeoneapi.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server image and tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				MetricsServer: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
					ImageTag:        "test",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server image configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				MetricsServer: kubeoneapi.ImageAsset{
					ImageRepository: "127.0.0.1:5000",
				},
			},
			expectedError: false,
		},
		{
			name: "metrics-server tag configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				MetricsServer: kubeoneapi.ImageAsset{
					ImageTag: "test",
				},
			},
			expectedError: false,
		},
		{
			name: "cni, node binaries, and kubectl configured",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				CNI: kubeoneapi.BinaryAsset{
					URL: "https://127.0.0.1/cni",
				},
				NodeBinaries: kubeoneapi.BinaryAsset{
					URL: "https://127.0.0.1/kubernetes-node-linux-amd64.tar.gz",
				},
				Kubectl: kubeoneapi.BinaryAsset{
					URL: "https://127.0.0.1/kubectl",
				},
			},
			expectedError: false,
		},
		{
			name: "binary assets configured (node binaries missing)",
			assetConfiguration: &kubeoneapi.AssetConfiguration{
				CNI: kubeoneapi.BinaryAsset{
					URL: "https://127.0.0.1/cni",
				},
				Kubectl: kubeoneapi.BinaryAsset{
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
