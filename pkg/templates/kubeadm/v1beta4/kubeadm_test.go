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

package v1beta4

import (
	"reflect"
	"slices"
	"testing"

	"github.com/Masterminds/semver/v3"

	kubeadmv1beta4 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta4"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
)

func TestEtcdVersionCorruptCheckExtraArgs(t *testing.T) {
	legacyEtcdExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "experimental-compact-hash-check-enabled",
			Value: "true",
		},
		{
			Name:  "experimental-corrupt-check-time",
			Value: "240m",
		},
	}

	etcdExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "feature-gates",
			Value: "CompactHashCheck=true",
		},
		{
			Name:  "corrupt-check-time",
			Value: "240m",
		},
	}

	cipherSuitesArg := kubeadmv1beta4.Arg{
		Name:  "cipher-suites",
		Value: "cipher1,cipher2",
	}

	tests := []struct {
		name             string
		kubeVersion      string
		cipherSuites     []string
		expectedEtcdArgs []kubeadmv1beta4.Arg
	}{
		{
			name:             "legacy flags before 1.37",
			kubeVersion:      "1.36.4",
			expectedEtcdArgs: legacyEtcdExtraArgs,
		},
		{
			name:             "legacy flags before 1.37 with tls cipher suites",
			kubeVersion:      "1.36.4",
			cipherSuites:     []string{"cipher1", "cipher2"},
			expectedEtcdArgs: append(slices.Clone(legacyEtcdExtraArgs), cipherSuitesArg),
		},
		{
			name:             "new flags since 1.37",
			kubeVersion:      "1.37.0",
			expectedEtcdArgs: etcdExtraArgs,
		},
		{
			name:             "new flags for 1.37 pre-release",
			kubeVersion:      "1.37.0-rc.0",
			expectedEtcdArgs: etcdExtraArgs,
		},
		{
			name:             "new flags since 1.37 with tls cipher suites",
			kubeVersion:      "1.37.0",
			cipherSuites:     []string{"cipher1", "cipher2"},
			expectedEtcdArgs: append(slices.Clone(etcdExtraArgs), cipherSuitesArg),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := etcdVersionCorruptCheckExtraArgs(semver.MustParse(tt.kubeVersion), tt.cipherSuites)
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

	clusterConfig := &kubeadmv1beta4.ClusterConfiguration{
		ControllerManager: kubeadmv1beta4.ControlPlaneComponent{
			ExtraArgs: []kubeadmv1beta4.Arg{
				{
					Name:  "existing-flag",
					Value: "existing-value",
				},
			},
		},
		Scheduler: kubeadmv1beta4.ControlPlaneComponent{
			ExtraArgs: []kubeadmv1beta4.Arg{
				{
					Name:  "existing-flag",
					Value: "existing-value",
				},
			},
		},
		APIServer: kubeadmv1beta4.APIServer{
			ControlPlaneComponent: kubeadmv1beta4.ControlPlaneComponent{
				ExtraArgs: []kubeadmv1beta4.Arg{
					{
						Name:  "existing-flag",
						Value: "existing-value",
					},
				},
			},
		},
	}

	expectedControllerManagerExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "existing-flag",
			Value: "existing-value",
		},
		{
			Name:  "flag1",
			Value: "value1",
		},
		{
			Name:  "flag2",
			Value: "value2",
		},
		{
			Name:  "feature-gates",
			Value: "feature1=true,feature2=false",
		},
	}
	expectedSchedulerExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "existing-flag",
			Value: "existing-value",
		},
		{
			Name:  "flag3",
			Value: "value3",
		},
		{
			Name:  "flag4",
			Value: "value4",
		},
		{
			Name:  "feature-gates",
			Value: "feature3=true,feature4=false",
		},
	}
	expectedAPIServerExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "existing-flag",
			Value: "existing-value",
		},
		{
			Name:  "flag5",
			Value: "value5",
		},
		{
			Name:  "flag6",
			Value: "value6",
		},
		{
			Name:  "feature-gates",
			Value: "feature5=true,feature6=false",
		},
	}
	expectedControllerManagerFeatureGates := "feature1=true,feature2=false"
	expectedSchedulerFeatureGates := "feature3=true,feature4=false"
	expectedAPIServerFeatureGates := "feature5=true,feature6=false"

	addControlPlaneComponentsAdditionalArgs(cluster, clusterConfig)

	// Check ControllerManager
	if !areArgsEqual(clusterConfig.ControllerManager.ExtraArgs, expectedControllerManagerExtraArgs) {
		t.Errorf("ControllerManager ExtraArgs mismatch, got: %v, want: %v", clusterConfig.ControllerManager.ExtraArgs, expectedControllerManagerExtraArgs)
	}
	fgs, _ := kubeadmv1beta4.GetArgValue(clusterConfig.ControllerManager.ExtraArgs, "feature-gates", -1)
	if !areFeatureGatesEqual(fgs, expectedControllerManagerFeatureGates) {
		t.Errorf("ControllerManager FeatureGates mismatch, got: %s, want: %s", fgs, expectedControllerManagerFeatureGates)
	}

	// Check Scheduler
	if !areArgsEqual(clusterConfig.Scheduler.ExtraArgs, expectedSchedulerExtraArgs) {
		t.Errorf("Scheduler ExtraArgs mismatch, got: %v, want: %v", clusterConfig.Scheduler.ExtraArgs, expectedSchedulerExtraArgs)
	}
	fgs, _ = kubeadmv1beta4.GetArgValue(clusterConfig.Scheduler.ExtraArgs, "feature-gates", -1)
	if !areFeatureGatesEqual(fgs, expectedSchedulerFeatureGates) {
		t.Errorf("Scheduler FeatureGates mismatch, got: %s, want: %s", fgs, expectedSchedulerFeatureGates)
	}

	// Check APIServer
	if !areArgsEqual(clusterConfig.APIServer.ExtraArgs, expectedAPIServerExtraArgs) {
		t.Errorf("APIServer ExtraArgs mismatch, got: %v, want: %v", clusterConfig.APIServer.ExtraArgs, expectedAPIServerExtraArgs)
	}
	fgs, _ = kubeadmv1beta4.GetArgValue(clusterConfig.APIServer.ExtraArgs, "feature-gates", -1)
	if !areFeatureGatesEqual(fgs, expectedAPIServerFeatureGates) {
		t.Errorf("APIServer FeatureGates mismatch, got: %s, want: %s", fgs, expectedAPIServerFeatureGates)
	}
}

// areFeatureGatesEqual compares two feature gates strings irrespective of the order of the feature gates.
func areFeatureGatesEqual(result, expected string) bool {
	resultMap := splitFeatureGates(result)
	expectedMap := splitFeatureGates(expected)

	// Compare the maps
	if len(resultMap) != len(expectedMap) {
		return false
	}

	return featureGatesToString(resultMap) == featureGatesToString(expectedMap)
}

// areArgsEqual compares two maps irrespective of the order of the keys.
func areArgsEqual(result, expected []kubeadmv1beta4.Arg) bool {
	if len(result) != len(expected) {
		return false
	}

	for _, arg := range expected {
		val, i := kubeadmv1beta4.GetArgValue(result, arg.Name, -1)
		if i == -1 || arg.Value != val {
			return false
		}
	}

	return true
}
