/*
Copyright 2021 The KubeOne Authors.

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

package kubeone

import "testing"

func TestFeatureGatesString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		featureGates map[string]bool
		expected     string
	}{
		{
			name:         "one feature gate",
			featureGates: map[string]bool{"TestFeatureGate": true},
			expected:     "TestFeatureGate=true",
		},
		{
			name: "two feature gates",
			featureGates: map[string]bool{
				"TestFeatureGate":  true,
				"TestDisabledGate": false,
			},
			expected: "TestDisabledGate=false,TestFeatureGate=true",
		},
		{
			name: "three feature gates",
			featureGates: map[string]bool{
				"TestFeatureGate":  true,
				"TestDisabledGate": false,
				"TestThirdGate":    true,
			},
			expected: "TestDisabledGate=false,TestFeatureGate=true,TestThirdGate=true",
		},
		{
			name:         "no feature gates",
			featureGates: map[string]bool{},
			expected:     "",
		},
		{
			name:         "feature gates nil",
			featureGates: nil,
			expected:     "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := marshalFeatureGates(tc.featureGates)
			if got != tc.expected {
				t.Errorf("TestFeatureGatesString() got = %v, expected %v", got, tc.expected)
			}
		})
	}
}
