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

package canal

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// felixConfigurationCRD creates the FelixConfiguration CRD
func felixConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "felixconfigurations.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "FelixConfiguration",
				Plural:   "felixconfigurations",
				Singular: "felixconfiguration",
			},
		},
	}
}

// bgpConfigurationCRD creates the BGPConfiguration CRD
func bgpConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bgpconfigurations.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "BGPConfiguration",
				Plural:   "bgpconfigurations",
				Singular: "bgpconfiguration",
			},
		},
	}
}

// ipPoolsConfigurationCRD creates the IPPool CRD
func ipPoolsConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ippools.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "IPPool",
				Plural:   "ippools",
				Singular: "ippool",
			},
		},
	}
}

// hostEndpointsConfigurationCRD creates the HostEndpoint CRD
func hostEndpointsConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hostendpoints.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "HostEndpoint",
				Plural:   "hostendpoints",
				Singular: "hostendpoint",
			},
		},
	}
}

// clusterInformationsConfigurationCRD creates the ClusterInformation CRD
func clusterInformationsConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusterinformations.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "ClusterInformation",
				Plural:   "clusterinformations",
				Singular: "clusterinformation",
			},
		},
	}
}

// globalNetworkPoliciesConfigurationCRD creates the GlobalNetworkPolicy CRD
func globalNetworkPoliciesConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "globalnetworkpolicies.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "GlobalNetworkPolicy",
				Plural:   "globalnetworkpolicies",
				Singular: "globalnetworkpolicy",
			},
		},
	}
}

// globalNetworksetsConfigurationCRD creates the GlobalNetworkSet CRD
func globalNetworksetsConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "globalnetworksets.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "GlobalNetworkSet",
				Plural:   "globalnetworksets",
				Singular: "globalnetworkset",
			},
		},
	}
}

// networkPoliciesConfigurationCRD creates the NetworkPolicy CRD
func networkPoliciesConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "networkpolicies.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.NamespaceScoped,
			Group: "crd.projectcalico.org",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "NetworkPolicy",
				Plural:   "networkpolicies",
				Singular: "networkpolicy",
			},
		},
	}
}
