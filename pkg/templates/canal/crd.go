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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "FelixConfiguration",
				ListKind: "FelixConfigurationList",
				Plural:   "felixconfigurations",
				Singular: "felixconfiguration",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "IPAMBlock",
				ListKind: "IPAMBlockList",
				Plural:   "ipamblocks",
				Singular: "ipamblock",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "BlockAffinity",
				ListKind: "BlockAffinityList",
				Plural:   "blockaffinities",
				Singular: "blockaffinity",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "IPAMHandle",
				ListKind: "IPAMHandleList",
				Plural:   "ipamhandles",
				Singular: "ipamhandle",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "IPAMConfig",
				ListKind: "IPAMConfigList",
				Plural:   "ipamconfigs",
				Singular: "ipamconfig",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "BGPPeer",
				ListKind: "BGPPeerList",
				Plural:   "bgppeers",
				Singular: "bgppeer",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "BGPConfiguration",
				ListKind: "BGPConfigurationList",
				Plural:   "bgpconfigurations",
				Singular: "bgpconfiguration",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "IPPool",
				ListKind: "IPPoolList",
				Plural:   "ippools",
				Singular: "ippool",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}
}

// kubeControllersConfigurationCRD creates the KubeControllersConfiguration CRD
func kubeControllersConfigurationCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubecontrollersconfigurations.crd.projectcalico.org",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Scope: apiextensions.ClusterScoped,
			Group: "crd.projectcalico.org",
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "KubeControllersConfiguration",
				ListKind: "KubeControllersConfigurationList",
				Plural:   "kubecontrollersconfigurations",
				Singular: "kubecontrollersconfiguration",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "HostEndpoint",
				ListKind: "HostEndpointList",
				Plural:   "hostendpoints",
				Singular: "hostendpoint",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "ClusterInformation",
				ListKind: "ClusterInformationList",
				Plural:   "clusterinformations",
				Singular: "clusterinformation",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "GlobalNetworkPolicy",
				ListKind: "GlobalNetworkPolicyList",
				Plural:   "globalnetworkpolicies",
				Singular: "globalnetworkpolicy",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "GlobalNetworkSet",
				ListKind: "GlobalNetworkSetList",
				Plural:   "globalnetworksets",
				Singular: "globalnetworkset",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "NetworkPolicy",
				ListKind: "NetworkPolicyList",
				Plural:   "networkpolicies",
				Singular: "networkpolicy",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
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
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "NetworkSet",
				ListKind: "NetworkSetList",
				Plural:   "networksets",
				Singular: "networkset",
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}
}
