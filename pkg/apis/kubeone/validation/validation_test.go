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

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

func TestValidateCloudProviderSpec(t *testing.T) {
	tests := []struct {
		name           string
		providerConfig kubeone.CloudProviderSpec
		expectedError  bool
	}{
		{
			name: "valid provider config (AWS)",
			providerConfig: kubeone.CloudProviderSpec{
				Name: kubeone.CloudProviderNameAWS,
			},
			expectedError: false,
		},
		{
			name: "valid provider config with external CCM ",
			providerConfig: kubeone.CloudProviderSpec{
				Name:     kubeone.CloudProviderNameAWS,
				External: true,
			},
			expectedError: false,
		},
		{
			name: "valid provider config with external CCM and cloud-config",
			providerConfig: kubeone.CloudProviderSpec{
				Name:        kubeone.CloudProviderNameAWS,
				External:    true,
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "valid openstack provider config with cloud-config",
			providerConfig: kubeone.CloudProviderSpec{
				Name:        kubeone.CloudProviderNameOpenStack,
				CloudConfig: "test",
			},
			expectedError: false,
		},
		{
			name: "invalid provider config (invalid provider name)",
			providerConfig: kubeone.CloudProviderSpec{
				Name: "testProvider",
			},
			expectedError: true,
		},
		{
			name: "invalid openstack provider config (without cloud-config)",
			providerConfig: kubeone.CloudProviderSpec{
				Name: kubeone.CloudProviderNameOpenStack,
			},
			expectedError: true,
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

func TestValidateHostConfig(t *testing.T) {
	tests := []struct {
		name          string
		hostConfig    []kubeone.HostConfig
		expectedError bool
	}{
		{
			name: "valid host config (with ip addresses)",
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
			name: "valid host config (with dns domain)",
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
			name: "invalid host config (no public address)",
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
			name: "invalid host config (no private address)",
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
			name: "invalid host config (no private key file and agent)",
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
			name: "invalid host config (no username)",
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

func TestValidateVersionConfig(t *testing.T) {
	tests := []struct {
		name          string
		versionConfig kubeone.VersionConfig
		expectedError bool
	}{
		{
			name: "valid version config (1.13.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.13.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (v1.13.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "v1.13.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.13.5)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.13.5",
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
			name: "valid version config (v1.14.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "v1.14.0",
			},
			expectedError: false,
		},
		{
			name: "valid version config (1.14.2)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.14.2",
			},
			expectedError: false,
		},
		{
			name: "invalid version config (1.12.0)",
			versionConfig: kubeone.VersionConfig{
				Kubernetes: "1.12.0",
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
func TestValidateMachineControllerConfig(t *testing.T) {
	tests := []struct {
		name                    string
		cloudProvider           kubeone.CloudProviderName
		machineControllerConfig *kubeone.MachineControllerConfig
		expectedError           bool
	}{
		{
			name:          "valid machine-controller config",
			cloudProvider: kubeone.CloudProviderNameAWS,
			machineControllerConfig: &kubeone.MachineControllerConfig{
				Deploy:   true,
				Provider: kubeone.CloudProviderNameAWS,
			},
			expectedError: false,
		},
		{
			name:          "valid machine-controller config (provider none with machine-controller provider set)",
			cloudProvider: kubeone.CloudProviderNameNone,
			machineControllerConfig: &kubeone.MachineControllerConfig{
				Deploy:   true,
				Provider: kubeone.CloudProviderNameAWS,
			},
			expectedError: false,
		},
		{
			name:          "invalid machine-controller config (provider and machine-controller provider different)",
			cloudProvider: kubeone.CloudProviderNameAWS,
			machineControllerConfig: &kubeone.MachineControllerConfig{
				Deploy:   true,
				Provider: kubeone.CloudProviderNameDigitalOcean,
			},
			expectedError: true,
		},
		{
			name:          "invalid machine-controller config (provider set and machine-controller provider not set)",
			cloudProvider: kubeone.CloudProviderNameAWS,
			machineControllerConfig: &kubeone.MachineControllerConfig{
				Deploy:   true,
				Provider: "",
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateMachineControllerConfig(tc.machineControllerConfig, tc.cloudProvider, nil)
			if (len(errs) == 0) == tc.expectedError {
				t.Errorf("test case failed: expected %v, but got %v", tc.expectedError, (len(errs) != 0))
			}
		})
	}
}

func TestValidateWorkerConfig(t *testing.T) {
	tests := []struct {
		name          string
		workerConfig  []kubeone.WorkerConfig
		expectedError bool
	}{
		{
			name: "valid worker config",
			workerConfig: []kubeone.WorkerConfig{
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
			name:          "valid worker config (no worker defined)",
			workerConfig:  []kubeone.WorkerConfig{},
			expectedError: false,
		},
		{
			name: "invalid worker config (replicas not provided)",
			workerConfig: []kubeone.WorkerConfig{
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
			workerConfig: []kubeone.WorkerConfig{
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
			errs := ValidateWorkerConfig(tc.workerConfig, nil)
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
			name: "valid worker config",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: false,
		},
		{
			name:                 "valid worker config (empty config)",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{},
			expectedError:        false,
		},
		{
			name: "invalid worker config (invalid pod subnet)",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0",
				ServiceSubnet: "192.168.0.0/24",
			},
			expectedError: true,
		},
		{
			name: "invalid worker config (invalid service subnet)",
			clusterNetworkConfig: kubeone.ClusterNetworkConfig{
				PodSubnet:     "192.168.1.0/24",
				ServiceSubnet: "192.168.0.0",
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

func TestValidateFeatures(t *testing.T) {
	tests := []struct {
		name          string
		features      kubeone.Features
		expectedError bool
	}{
		{
			name: "valid features config (psp and auditing enabled)",
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
			name: "valid features config (metrics server disabled)",
			features: kubeone.Features{
				MetricsServer: &kubeone.MetricsServer{
					Enable: false,
				},
			},
			expectedError: false,
		},
		{
			name:          "valid features config (no feature configured)",
			features:      kubeone.Features{},
			expectedError: false,
		},
		{
			name: "valid features config (oidc enabled)",
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
			name: "invalid features config (invalid oidc config)",
			features: kubeone.Features{
				OpenIDConnect: &kubeone.OpenIDConnect{
					Enable: true,
					Config: kubeone.OpenIDConnectConfig{},
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
			name: "invalid oidc config (no issuer url)",
			oidcConfig: kubeone.OpenIDConnectConfig{
				ClientID: "test",
			},
			expectedError: true,
		},
		{
			name: "invalid oidc config (no client id)",
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

func intPtr(i int) *int {
	return &i
}
