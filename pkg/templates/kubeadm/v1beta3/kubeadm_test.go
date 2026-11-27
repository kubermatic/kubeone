/*
Copyright 2022 The KubeOne Authors.

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
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"

	kubeadmv1beta3 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta3"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/state"
)

func TestEtcdVersionCorruptCheckExtraArgs(t *testing.T) {
	etcdExtraArgs := map[string]string{
		"experimental-compact-hash-check-enabled": "true",
		"experimental-initial-corrupt-check":      "true",
		"experimental-corrupt-check-time":         "240m",
	}

	etcdExtraArgsWithCiphers := maps.Clone(etcdExtraArgs)
	etcdExtraArgsWithCiphers["cipher-suites"] = "cipher1,cipher2"

	tests := []struct {
		name                 string
		kubeVersion          *semver.Version
		etcdImageTag         string
		expectedEtcdImageTag string
		cipherSuites         []string
		expectedEtcdArgs     map[string]string
	}{
		{
			name:                 "any 1.29",
			kubeVersion:          semver.MustParse("1.29"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "any 1.29 high",
			kubeVersion:          semver.MustParse("1.29.999"),
			expectedEtcdImageTag: "",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "unfixed, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.26.12"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "fixed, but tag is overwritten",
			kubeVersion:          semver.MustParse("1.26.13"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			expectedEtcdArgs:     etcdExtraArgs,
		},
		{
			name:                 "tls cipher suites",
			kubeVersion:          semver.MustParse("1.26.13"),
			etcdImageTag:         "9.9.9-0",
			expectedEtcdImageTag: "9.9.9-0",
			cipherSuites:         []string{"cipher1", "cipher2"},
			expectedEtcdArgs:     etcdExtraArgsWithCiphers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, args := etcdVersionCorruptCheckExtraArgs(tt.kubeVersion, tt.etcdImageTag, tt.cipherSuites)
			if ver != tt.expectedEtcdImageTag {
				t.Errorf("got etcd image tag %q, but expected %q", ver, tt.expectedEtcdImageTag)
			}
			if !reflect.DeepEqual(args, tt.expectedEtcdArgs) {
				t.Errorf("got etcd tags %q, but expected %q", args, tt.expectedEtcdArgs)
			}
		})
	}
}

func TestMergeFeatureGates(t *testing.T) {
	tests := []struct {
		name                   string
		featureGates           string
		additionalFeatureGates map[string]bool
		expectedResult         string
	}{
		{
			name:                   "Empty feature gates",
			featureGates:           "",
			additionalFeatureGates: map[string]bool{"foo": true, "bar": false},
			expectedResult:         "foo=true,bar=false",
		},
		{
			name:                   "Existing feature gates",
			featureGates:           "feature1=true",
			additionalFeatureGates: map[string]bool{"foo": true, "bar": false},
			expectedResult:         "feature1=true,bar=false,foo=true",
		},
		{
			name:                   "Overwriting existing feature gates",
			featureGates:           "feature1=true",
			additionalFeatureGates: map[string]bool{"feature1": false, "feature3": true},
			expectedResult:         "feature1=false,feature3=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeFeatureGates(tt.featureGates, tt.additionalFeatureGates)
			if !areFeatureGatesEqual(result, tt.expectedResult) {
				t.Errorf("got %q, but expected %q", result, tt.expectedResult)
			}
		})
	}
}

func TestAddControlPlaneComponentsAdditionalArgs(t *testing.T) {
	cluster := &kubeoneapi.KubeOneCluster{
		ControlPlaneComponents: &kubeoneapi.ControlPlaneComponents{
			ControllerManager: &kubeoneapi.ControlPlaneComponentConfig{
				Flags: map[string]string{
					"flag1": "value1",
					"flag2": "value2",
				},
				FeatureGates: map[string]bool{
					"feature1": true,
					"feature2": false,
				},
			},
			Scheduler: &kubeoneapi.ControlPlaneComponentConfig{
				Flags: map[string]string{
					"flag3": "value3",
					"flag4": "value4",
				},
				FeatureGates: map[string]bool{
					"feature3": true,
					"feature4": false,
				},
			},
			APIServer: &kubeoneapi.ControlPlaneComponentConfig{
				Flags: map[string]string{
					"flag5": "value5",
					"flag6": "value6",
				},
				FeatureGates: map[string]bool{
					"feature5": true,
					"feature6": false,
				},
			},
		},
	}

	clusterConfig := &kubeadmv1beta3.ClusterConfiguration{
		ControllerManager: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"existing-flag": "existing-value",
			},
		},
		Scheduler: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"existing-flag": "existing-value",
			},
		},
		APIServer: kubeadmv1beta3.APIServer{
			ControlPlaneComponent: kubeadmv1beta3.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"existing-flag": "existing-value",
				},
			},
		},
	}

	expectedControllerManagerExtraArgs := map[string]string{
		"existing-flag": "existing-value",
		"flag1":         "value1",
		"flag2":         "value2",
		"feature-gates": "feature1=true,feature2=false",
	}
	expectedSchedulerExtraArgs := map[string]string{
		"existing-flag": "existing-value",
		"flag3":         "value3",
		"flag4":         "value4",
		"feature-gates": "feature3=true,feature4=false",
	}
	expectedAPIServerExtraArgs := map[string]string{
		"existing-flag": "existing-value",
		"flag5":         "value5",
		"flag6":         "value6",
		"feature-gates": "feature5=true,feature6=false",
	}

	expectedControllerManagerFeatureGates := "feature1=true,feature2=false"
	expectedSchedulerFeatureGates := "feature3=true,feature4=false"
	expectedAPIServerFeatureGates := "feature5=true,feature6=false"

	addControlPlaneComponentsAdditionalArgs(cluster, clusterConfig)

	// Check ControllerManager
	if !areArgsEqual(clusterConfig.ControllerManager.ExtraArgs, expectedControllerManagerExtraArgs) {
		t.Errorf("ControllerManager ExtraArgs mismatch, got: %v, want: %v", clusterConfig.ControllerManager.ExtraArgs, expectedControllerManagerExtraArgs)
	}
	if !areFeatureGatesEqual(clusterConfig.ControllerManager.ExtraArgs["feature-gates"], expectedControllerManagerFeatureGates) {
		t.Errorf("ControllerManager FeatureGates mismatch, got: %s, want: %s", clusterConfig.ControllerManager.ExtraArgs["feature-gates"], expectedControllerManagerFeatureGates)
	}

	// Check Scheduler
	if !areArgsEqual(clusterConfig.Scheduler.ExtraArgs, expectedSchedulerExtraArgs) {
		t.Errorf("Scheduler ExtraArgs mismatch, got: %v, want: %v", clusterConfig.Scheduler.ExtraArgs, expectedSchedulerExtraArgs)
	}
	if !areFeatureGatesEqual(clusterConfig.Scheduler.ExtraArgs["feature-gates"], expectedSchedulerFeatureGates) {
		t.Errorf("Scheduler FeatureGates mismatch, got: %s, want: %s", clusterConfig.Scheduler.ExtraArgs["feature-gates"], expectedSchedulerFeatureGates)
	}

	// Check APIServer
	if !areArgsEqual(clusterConfig.APIServer.ExtraArgs, expectedAPIServerExtraArgs) {
		t.Errorf("APIServer ExtraArgs mismatch, got: %v, want: %v", clusterConfig.APIServer.ExtraArgs, expectedAPIServerExtraArgs)
	}
	if !areFeatureGatesEqual(clusterConfig.APIServer.ExtraArgs["feature-gates"], expectedAPIServerFeatureGates) {
		t.Errorf("APIServer FeatureGates mismatch, got: %s, want: %s", clusterConfig.APIServer.ExtraArgs["feature-gates"], expectedAPIServerFeatureGates)
	}
}

// areFeatureGatesEqual compares two feature gates strings irrespective of the order of the feature gates.
func areFeatureGatesEqual(result, expected string) bool {
	resultMap := make(map[string]bool)
	expectedMap := make(map[string]bool)

	// Parse the result string into a map
	for fg := range strings.SplitSeq(result, ",") {
		kv := strings.Split(fg, "=")
		if len(kv) == 2 {
			resultMap[kv[0]] = kv[1] == "true"
		}
	}

	// Parse the expected string into a map
	for fg := range strings.SplitSeq(expected, ",") {
		kv := strings.Split(fg, "=")
		if len(kv) == 2 {
			expectedMap[kv[0]] = kv[1] == "true"
		}
	}

	// Compare the maps
	if len(resultMap) != len(expectedMap) {
		return false
	}

	for k, v := range resultMap {
		if expectedMap[k] != v {
			return false
		}
	}

	return true
}

// areArgsEqual compares two maps irrespective of the order of the keys.
func areArgsEqual(result, expected map[string]string) bool {
	if len(result) != len(expected) {
		return false
	}

	for k, v := range expected {
		if result[k] != v {
			return false
		}
	}

	return true
}

func TestNewConfigWorkerExternalCloudProvider(t *testing.T) {
	tests := []struct {
		name                    string
		external                bool
		noExternal              bool
		ipFamily                kubeoneapi.IPFamily
		expectCloudProviderFlag bool
		expectNodeIPFlag        bool
	}{
		{
			name:                    "external cloud provider is set for workers by default",
			external:                true,
			ipFamily:                kubeoneapi.IPFamilyIPv4IPv6,
			expectCloudProviderFlag: true,
		},
		{
			name:                    "external cloud provider is left out for disabled hosts",
			external:                true,
			noExternal:              true,
			ipFamily:                kubeoneapi.IPFamilyIPv4IPv6,
			expectCloudProviderFlag: false,
			// without CCM the kubelet has to know the node IPs in dualstack clusters
			expectNodeIPFlag: true,
		},
		{
			name:                    "no external cloud provider configured",
			external:                false,
			ipFamily:                kubeoneapi.IPFamilyIPv4,
			expectCloudProviderFlag: false,
			expectNodeIPFlag:        true,
		},
		{
			name:                    "disabled host in a cluster without external cloud provider",
			external:                false,
			noExternal:              true,
			ipFamily:                kubeoneapi.IPFamilyIPv4,
			expectCloudProviderFlag: false,
			expectNodeIPFlag:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &state.State{
				Cluster: &kubeoneapi.KubeOneCluster{
					APIEndpoint:    kubeoneapi.APIEndpoint{Host: "api.example.com", Port: 6443},
					CloudProvider:  kubeoneapi.CloudProviderSpec{External: tc.external},
					ClusterNetwork: kubeoneapi.ClusterNetworkConfig{IPFamily: tc.ipFamily},
				},
				JoinToken: "abcdef.0123456789abcdef",
			}
			host := kubeoneapi.HostConfig{
				PublicAddress:  "1.1.1.1",
				PrivateAddress: "10.0.0.1",
				IPv6Addresses:  []string{"fd00::1"},
				NoExternal:     tc.noExternal,
			}

			cfg, err := NewConfigWorker(s, host)
			if err != nil {
				t.Fatalf("NewConfigWorker() error = %v", err)
			}

			args := cfg.JoinConfiguration.NodeRegistration.KubeletExtraArgs

			value, cloudProviderFlag := args["cloud-provider"]
			if cloudProviderFlag != tc.expectCloudProviderFlag {
				t.Errorf("cloud-provider flag present = %v, want %v", cloudProviderFlag, tc.expectCloudProviderFlag)
			}

			if cloudProviderFlag && value != "external" {
				t.Errorf("cloud-provider = %q, want %q", value, "external")
			}

			if _, nodeIPFlag := args["node-ip"]; nodeIPFlag != tc.expectNodeIPFlag {
				t.Errorf("node-ip flag present = %v, want %v", nodeIPFlag, tc.expectNodeIPFlag)
			}
		})
	}
}
