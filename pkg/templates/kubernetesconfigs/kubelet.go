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
	"crypto/tls"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/templates/resources"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

func SafeTLSCiphers() []string {
	return []string{
		tls.CipherSuiteName(tls.TLS_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_CHACHA20_POLY1305_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305),
		tls.CipherSuiteName(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_128_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_128_GCM_SHA256),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_256_CBC_SHA),
		tls.CipherSuiteName(tls.TLS_RSA_WITH_AES_256_GCM_SHA384),
	}
}

func NewKubeletConfiguration(cluster *kubeoneapi.KubeOneCluster, featureGates map[string]bool) (runtime.Object, error) {
	bfalse := false
	kubeletConfig := &kubeletconfigv1beta1.KubeletConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubelet.config.k8s.io/v1beta1",
			Kind:       "KubeletConfiguration",
		},
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
		CgroupDriver:         "systemd",
		ContainerLogMaxFiles: &cluster.LoggingConfig.ContainerLogMaxFiles,
		ContainerLogMaxSize:  cluster.LoggingConfig.ContainerLogMaxSize,
		FeatureGates:         featureGates,
		ReadOnlyPort:         0,
		RotateCertificates:   true,
		ServerTLSBootstrap:   true,
		TLSCipherSuites:      SafeTLSCiphers(),
	}

	if cluster.Features.NodeLocalDNS.Deploy {
		kubeletConfig.ClusterDNS = []string{resources.NodeLocalDNSVirtualIP}
	}

	return dropFields(kubeletConfig, []string{"logging"}, []string{"containerRuntimeEndpoint"})
}
