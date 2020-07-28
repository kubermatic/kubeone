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
	"bytes"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// configMap creates a ConfigMap used to configure a self-hosted Canal installation
func configMap(netConf bytes.Buffer, mtu int) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canal-config",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			// Typha is disabled.
			"typha_service_name": "none",

			// The interface used by canal for host <-> host communication.
			// If left blank, then the interface is chosen using the node's
			// default route
			"canal_iface": "",

			// Whether or not to masquerade traffic to destinations not within
			// the pod network
			"masquerade": "true",

			// Configure the MTU to use for workload interfaces and tunnels.
			// - If Wireguard is enabled, set to your network MTU - 60
			// - Otherwise, if VXLAN or BPF mode is enabled, set to your network MTU - 50
			// - Otherwise, if IPIP is enabled, set to your network MTU - 20
			// - Otherwise, if not using any encapsulation, set to your network MTU.
			"veth_mtu": strconv.Itoa(mtu),

			// The CNI network configuration to install on each node.  The special
			// values in this config will be automatically populated.
			"cni_network_config": cniNetworkConfig,

			// Flannel network configuration. Mounted into the flannel container
			"net-conf.json": netConf.String(),
		},
	}
}

// daemonsetServiceAccount creates the canal ServiceAccount
func daemonsetServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canal",
			Namespace: metav1.NamespaceSystem,
		},
	}
}

// serviceAccount creates the canal ServiceAccount
func deploymentServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "calico-kube-controllers",
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func calicoKubeControllersClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "calico-kube-controllers",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "calico-kube-controllers",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "calico-kube-controllers",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

func calicoKubeControllersClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "calico-kube-controllers",
		},
		Rules: []rbacv1.PolicyRule{
			{
				// Nodes are watched to monitor for deletions.
				APIGroups: []string{""},
				Resources: []string{
					"nodes",
				},
				Verbs: []string{
					"watch",
					"list",
					"get",
				},
			},
			{
				// Pods are queried to check for existence.
				APIGroups: []string{""},
				Resources: []string{
					"pods",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				// IPAM resources are manipulated when nodes are deleted.
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"ippools",
				},
				Verbs: []string{
					"list",
				},
			},
			{
				//
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"blockaffinities",
					"ipamblocks",
					"ipamhandles",
				},
				Verbs: []string{
					"get",
					"list",
					"create",
					"update",
					"delete",
				},
			},
			{
				// kube-controllers manages hostendpoints.
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"hostendpoints",
				},
				Verbs: []string{
					"get",
					"list",
					"create",
					"update",
					"delete",
				},
			},
			{
				// Needs access to update clusterinformations.
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"clusterinformations",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
				},
			},
			{
				// KubeControllersConfiguration is where it gets its config
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"kubecontrollersconfigurations",
				},
				Verbs: []string{
					"get",
					"create",
					"update",
					"watch",
				},
			},
		},
	}
}

// calicoNodeClusterRole creates a ClusterRole for the calico-node DaemonSet
func calicoNodeClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "calico-node",
		},
		Rules: []rbacv1.PolicyRule{
			{
				// The CNI plugin needs to get pods, nodes, and namespaces
				APIGroups: []string{""},
				Resources: []string{
					"pods",
					"nodes",
					"namespaces",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{
					"endpoints",
					"services",
				},
				Verbs: []string{
					// Used to discover service IPs for advertisement
					"watch",
					"list",
					// Used to discover Typhas
					"get",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{
					"nodes/status",
				},
				Verbs: []string{
					// Needed for clearing NodeNetworkUnavailable flag
					"patch",
					// Calico stores some configuration information in node annotations
					"update",
				},
			},
			{
				// Watch for changes to Kubernetes NetworkPolicies
				APIGroups: []string{
					"networking.k8s.io",
				},
				Resources: []string{
					"networkpolicies",
				},
				Verbs: []string{
					"watch",
					"list",
				},
			},
			{
				// Used by Calico for policy information
				APIGroups: []string{""},
				Resources: []string{
					"pods",
					"namespaces",
					"serviceaccounts",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				// The CNI plugin patches pods/status
				APIGroups: []string{""},
				Resources: []string{
					"pods/status",
				},
				Verbs: []string{
					"patch",
				},
			},
			{
				// Calico monitors various CRDs for config
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"globalfelixconfigs",
					"felixconfigurations",
					"bgppeers",
					"globalbgpconfigs",
					"bgpconfigurations",
					"ippools",
					"ipamblocks",
					"globalnetworkpolicies",
					"globalnetworksets",
					"networkpolicies",
					"networksets",
					"clusterinformations",
					"hostendpoints",
					"blockaffinities",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				// Calico must create and update some CRDs on startup
				APIGroups: []string{"crd.projectcalico.org"},
				Resources: []string{
					"ippools",
					"felixconfigurations",
					"clusterinformations",
				},
				Verbs: []string{
					"create",
					"update",
				},
			},
			{
				// Calico stores some configuration information on the node
				APIGroups: []string{""},
				Resources: []string{
					"nodes",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				// These permissions are only required for upgrade from v2.6 and can be removed after upgrade
				// or on fresh installations
				APIGroups: []string{""},
				Resources: []string{
					"bgpconfigurations",
					"bgppeers",
				},
				Verbs: []string{
					"create",
					"update",
				},
			},
		},
	}
}

// flannelClusterRole creates a ClusterRole for the Flannel DaemonSet
func flannelClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{
					"pods",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{
					"nodes",
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{""},
				Resources: []string{
					"nodes/status",
				},
				Verbs: []string{
					"patch",
				},
			},
		},
	}
}

// flannelClusterRoleBinding creates a ClusterRoleBinding to bind the Flannel ClusterRole to the Canal ServiceAccount
func flannelClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "canal-flannel",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "flannel",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "canal",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}

// canalClusterRoleBinding creates a ClusterRoleBinding to bind the Calico ClusterRole to the Canal ServiceAccount
func canalClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "canal-calico",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "calico-node",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "canal",
				Namespace: metav1.NamespaceSystem,
			},
		},
	}
}
