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

package kubernetesconfigs

import (
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbasev1alpha1 "k8s.io/component-base/config/v1alpha1"
	kubeproxyv1alpha1 "k8s.io/kube-proxy/config/v1alpha1"
)

func NewKubeProxyConfiguration(cluster *kubeoneapi.KubeOneCluster) (runtime.Object, error) {
	kubeProxyConfig := &kubeproxyv1alpha1.KubeProxyConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeProxyConfiguration",
			APIVersion: "kubeproxy.config.k8s.io/v1alpha1",
		},
		ClusterCIDR: cluster.ClusterNetwork.PodSubnet,
		ClientConnection: componentbasev1alpha1.ClientConnectionConfiguration{
			Kubeconfig: "/var/lib/kube-proxy/kubeconfig.conf",
		},
	}

	if kbPrx := cluster.ClusterNetwork.KubeProxy; kbPrx != nil {
		switch {
		case kbPrx.IPVS != nil:
			kubeProxyConfig.Mode = kubeproxyv1alpha1.ProxyMode("ipvs")
			kubeProxyConfig.IPVS = kubeproxyv1alpha1.KubeProxyIPVSConfiguration{
				StrictARP:     kbPrx.IPVS.StrictARP,
				Scheduler:     kbPrx.IPVS.Scheduler,
				ExcludeCIDRs:  kbPrx.IPVS.ExcludeCIDRs,
				TCPTimeout:    kbPrx.IPVS.TCPTimeout,
				TCPFinTimeout: kbPrx.IPVS.TCPFinTimeout,
				UDPTimeout:    kbPrx.IPVS.UDPTimeout,
			}
		case kbPrx.IPTables != nil:
			kubeProxyConfig.Mode = kubeproxyv1alpha1.ProxyMode("iptables")
		}
	}

	return dropFields(kubeProxyConfig, []string{"detectLocal"}, []string{"winkernel"})
}
