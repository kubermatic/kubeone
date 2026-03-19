/*
Copyright 2026 The KubeOne Authors.

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

package kubeflags

import (
	"slices"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestDefaultAdmissionControllers(t *testing.T) {
	tests := []struct {
		name     string
		version  *semver.Version
		expected []string
	}{
		{
			name:    "pre v135 includes NodeRestriction",
			version: semver.MustParse("1.34.0"),
			expected: []string{
				"NamespaceLifecycle",
				"LimitRanger",
				"ServiceAccount",
				"TaintNodesByCondition",
				"NodeRestriction",
				"PodSecurity",
				"Priority",
				"DefaultTolerationSeconds",
				"DefaultStorageClass",
				"StorageObjectInUseProtection",
				"PersistentVolumeClaimResize",
				"RuntimeClass",
				"CertificateApproval",
				"CertificateSigning",
				"ClusterTrustBundleAttest",
				"CertificateSubjectRestriction",
				"DefaultIngressClass",
				"PodTopologyLabels",
				"MutatingAdmissionPolicy",
				"MutatingAdmissionWebhook",
				"ValidatingAdmissionPolicy",
				"ValidatingAdmissionWebhook",
				"ResourceQuota",
			},
		},
		{
			name:    "v135 includes NodeRestriction",
			version: semver.MustParse("1.35.0"),
			expected: []string{
				"CertificateApproval",
				"CertificateSigning",
				"CertificateSubjectRestriction",
				"DefaultIngressClass",
				"DefaultStorageClass",
				"DefaultTolerationSeconds",
				"LimitRanger",
				"MutatingAdmissionWebhook",
				"NamespaceLifecycle",
				"NodeRestriction",
				"PersistentVolumeClaimResize",
				"PodSecurity",
				"Priority",
				"ResourceQuota",
				"RuntimeClass",
				"ServiceAccount",
				"StorageObjectInUseProtection",
				"TaintNodesByCondition",
				"ValidatingAdmissionPolicy",
				"ValidatingAdmissionWebhook",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := strings.Split(DefaultAdmissionControllers(test.version), ",")
			if !slices.Equal(result, test.expected) {
				t.Fatalf("unexpected default admission controllers for %s: got %v, want %v", test.version, result, test.expected)
			}
		})
	}
}
