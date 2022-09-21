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
	"k8c.io/kubeone/pkg/templates/resources"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

func NewKubeletConfiguration(cluster *kubeoneapi.KubeOneCluster, featureGates map[string]bool) (runtime.Object, error) {
	bfalse := false
	kubeletConfig := &kubeletconfigv1beta1.KubeletConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubelet.config.k8s.io/v1beta1",
			Kind:       "KubeletConfiguration",
		},
		CgroupDriver:         "systemd",
		ReadOnlyPort:         0,
		RotateCertificates:   true,
		ServerTLSBootstrap:   true,
		ContainerLogMaxSize:  cluster.LoggingConfig.ContainerLogMaxSize,
		ContainerLogMaxFiles: &cluster.LoggingConfig.ContainerLogMaxFiles,
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
		FeatureGates: featureGates,
	}

	if cluster.Features.NodeLocalDNS.Deploy {
		kubeletConfig.ClusterDNS = []string{resources.NodeLocalDNSVirtualIP}
	}

	return dropFields(kubeletConfig, []string{"logging"})
}
