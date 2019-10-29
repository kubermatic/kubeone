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

// ipamBlockCRD creates the IPAMBlocks CRD
func ipamBlockCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ipamblocks.crd.projectcalico.org",
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
				Kind:     "IPAMBlock",
				Plural:   "ipamblocks",
				Singular: "ipamblock",
			},
		},
	}
}

// blockAffinityCRD creates the BlockAffinity CRD
func blockAffinityCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "blockaffinities.crd.projectcalico.org",
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
				Kind:     "BlockAffinity",
				Plural:   "blockaffinities",
				Singular: "blockaffinity",
			},
		},
	}
}

// ipamHandleCRD creates the IPAMHandle CRD
func ipamHandleCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ipamhandles.crd.projectcalico.org",
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
				Kind:     "IPAMHandle",
				Plural:   "ipamhandles",
				Singular: "ipamhandle",
			},
		},
	}
}

// ipamConfigCRD creates the IPAMConfig CRD
func ipamConfigCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ipamconfigs.crd.projectcalico.org",
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
				Kind:     "IPAMConfig",
				Plural:   "ipamconfigs",
				Singular: "ipamconfig",
			},
		},
	}
}

// bgpPeerCRD creates the BGPPeer CRD
func bgpPeerCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bgppeers.crd.projectcalico.org",
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
				Kind:     "BGPPeer",
				Plural:   "bgppeers",
				Singular: "bgppeer",
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

// ipPoolCRD creates the IPPool CRD
func ipPoolCRD() *apiextensions.CustomResourceDefinition {
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

// hostEndpointCRD creates the HostEndpoint CRD
func hostEndpointCRD() *apiextensions.CustomResourceDefinition {
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

// clusterInformationCRD creates the ClusterInformation CRD
func clusterInformationCRD() *apiextensions.CustomResourceDefinition {
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

// globalNetworkPolicyCRD creates the GlobalNetworkPolicy CRD
func globalNetworkPolicyCRD() *apiextensions.CustomResourceDefinition {
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

// globalNetworksetCRD creates the GlobalNetworkSet CRD
func globalNetworksetCRD() *apiextensions.CustomResourceDefinition {
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

// networkPolicyCRD creates the NetworkPolicy CRD
func networkPolicyCRD() *apiextensions.CustomResourceDefinition {
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

// networkSetCRD creates the NetworkSet CRD
func networkSetCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "networksets.crd.projectcalico.org",
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
				Kind:     "NetworkSet",
				Plural:   "networksets",
				Singular: "networkset",
			},
		},
	}
}
