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

package kubeflags

// Up to date list of default admission plugins
// is: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#which-plugins-are-enabled-by-default

var (
	defaultAdmissionControllersv135 = []string{
		"CertificateApproval",
		"CertificateSigning",
		"CertificateSubjectRestriction",
		"DefaultIngressClass",
		"DefaultStorageClass",
		"DefaultTolerationSeconds",
		"LimitRanger",
		"MutatingAdmissionWebhook",
		"NamespaceLifecycle",
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
	}

	defaultAdmissionControllersPreV135 = []string{
		"NamespaceLifecycle",
		"LimitRanger",
		"ServiceAccount",
		"TaintNodesByCondition",
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
	}
)
