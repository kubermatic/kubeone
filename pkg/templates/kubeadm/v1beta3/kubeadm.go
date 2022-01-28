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

package v1beta3

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	bootstraptokenv1 "k8c.io/kubeone/pkg/apis/kubeadm/bootstraptoken/v1"
	kubeadmv1beta3 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta3"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeflags"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"k8c.io/kubeone/pkg/templates/resources"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbasev1alpha1 "k8s.io/component-base/config/v1alpha1"
	kubeproxyv1alpha1 "k8s.io/kube-proxy/config/v1alpha1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

const (
	bootstrapTokenTTL = 60 * time.Minute
)

// NewConfig returns all required configs to init a cluster via a set of v1beta3 configs
func NewConfig(s *state.State, host kubeoneapi.HostConfig) ([]runtime.Object, error) {
	cluster := s.Cluster
	kubeSemVer, err := semver.NewVersion(cluster.Versions.Kubernetes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse generate config, wrong kubernetes version %s", cluster.Versions.Kubernetes)
	}

	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--var-lib-etcd",
		"DirAvailable--etc-kubernetes-manifests",
		"ImagePull",
	}

	bootstrapToken, err := bootstraptokenv1.NewBootstrapTokenString(s.JoinToken)
	if err != nil {
		return nil, err
	}

	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoint.Host, cluster.APIEndpoint.Port)

	initConfig := &kubeadmv1beta3.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta3",
			Kind:       "InitConfiguration",
		},
		BootstrapTokens: []bootstraptokenv1.BootstrapToken{
			{
				Token: bootstrapToken,
				Groups: []string{
					"system:bootstrappers:kubeadm:default-node-token",
				},
				TTL: &metav1.Duration{
					Duration: bootstrapTokenTTL,
				},
				Usages: []string{
					"signing",
					"authentication",
				},
			},
		},
		LocalAPIEndpoint: kubeadmv1beta3.APIEndpoint{
			AdvertiseAddress: newNodeIP(host),
		},
	}

	joinConfig := &kubeadmv1beta3.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta3",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta3.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta3.APIEndpoint{
				AdvertiseAddress: newNodeIP(host),
			},
		},
		Discovery: kubeadmv1beta3.Discovery{
			BootstrapToken: &kubeadmv1beta3.BootstrapTokenDiscovery{
				Token:                    s.JoinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}

	certSANS := certificate.GetCertificateSANs(cluster.APIEndpoint.Host, cluster.APIEndpoint.AlternativeNames)
	clusterConfig := &kubeadmv1beta3.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta3",
			Kind:       "ClusterConfiguration",
		},
		Networking: kubeadmv1beta3.Networking{
			PodSubnet:     cluster.ClusterNetwork.PodSubnet,
			ServiceSubnet: cluster.ClusterNetwork.ServiceSubnet,
			DNSDomain:     cluster.ClusterNetwork.ServiceDomainName,
		},
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta3.APIServer{
			ControlPlaneComponent: kubeadmv1beta3.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"endpoint-reconciler-type": "lease",
					"service-node-port-range":  cluster.ClusterNetwork.NodePortRange,
					"enable-admission-plugins": kubeflags.DefaultAdmissionControllers(kubeSemVer),
				},
				ExtraVolumes: []kubeadmv1beta3.HostPathMount{},
			},
			CertSANs: certSANS,
		},
		ControllerManager: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"flex-volume-plugin-dir": "/var/lib/kubelet/volumeplugins",
			},
			ExtraVolumes: []kubeadmv1beta3.HostPathMount{},
		},
		ClusterName:     cluster.Name,
		ImageRepository: cluster.AssetConfiguration.Kubernetes.ImageRepository,
		Etcd: kubeadmv1beta3.Etcd{
			Local: &kubeadmv1beta3.LocalEtcd{
				ImageMeta: kubeadmv1beta3.ImageMeta{
					ImageRepository: cluster.AssetConfiguration.Etcd.ImageRepository,
					ImageTag:        cluster.AssetConfiguration.Etcd.ImageTag,
				},
			},
		},
		DNS: kubeadmv1beta3.DNS{
			ImageMeta: kubeadmv1beta3.ImageMeta{
				ImageRepository: cluster.AssetConfiguration.CoreDNS.ImageRepository,
				ImageTag:        cluster.AssetConfiguration.CoreDNS.ImageTag,
			},
		},
	}

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
		ClusterDNS:           []string{resources.NodeLocalDNSVirtualIP},
		ContainerLogMaxSize:  cluster.LoggingConfig.ContainerLogMaxSize,
		ContainerLogMaxFiles: &cluster.LoggingConfig.ContainerLogMaxFiles,
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
		FeatureGates: map[string]bool{},
	}

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	}

	if s.ShouldEnableInTreeCloudProvider() {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"
		cloudConfigVol := kubeadmv1beta3.HostPathMount{
			Name:      "cloud-config",
			HostPath:  renderedCloudConfig,
			MountPath: renderedCloudConfig,
			ReadOnly:  true,
			PathType:  corev1.HostPathFile,
		}
		provider := cluster.CloudProvider.CloudProviderName()

		clusterConfig.APIServer.ExtraArgs["cloud-provider"] = provider
		clusterConfig.APIServer.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, cloudConfigVol)

		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = provider
		clusterConfig.ControllerManager.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.ControllerManager.ExtraArgs["cluster-name"] = s.Cluster.Name
		clusterConfig.ControllerManager.ExtraVolumes = append(clusterConfig.ControllerManager.ExtraVolumes, cloudConfigVol)

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig

		switch {
		case cluster.CloudProvider.Azure != nil:
			clusterConfig.ControllerManager.ExtraArgs["configure-cloud-routes"] = "false"
		case cluster.CloudProvider.AWS != nil:
			clusterConfig.ControllerManager.ExtraArgs["configure-cloud-routes"] = "false"
		}
	}

	if cluster.CloudProvider.External {
		if !s.ShouldEnableInTreeCloudProvider() {
			delete(clusterConfig.APIServer.ExtraArgs, "cloud-provider")
			delete(clusterConfig.ControllerManager.ExtraArgs, "cloud-provider")
			nodeRegistration.KubeletExtraArgs["cloud-provider"] = "external"
		} else {
			// .cloudProvider.external enabled, but in-tree cloud provider should be enabled
			// means that we're in the CCM migration process.
			// In that case, we should leave cloud-provider flags in place, but explicitly
			// disable CCM-related controllers.
			clusterConfig.ControllerManager.ExtraArgs["controllers"] = "*,bootstrapsigner,tokencleaner,-cloud-node-lifecycle,-route,-service"
		}

		if s.ShouldEnableCSIMigration() {
			featureGates, featureGatesFlag, err := s.Cluster.CSIMigrationFeatureGates(s.ShouldUnregisterInTreeCloudProvider())
			if err != nil {
				return nil, err
			}

			// Kubernetes API server
			if fg, ok := clusterConfig.APIServer.ExtraArgs["feature-gates"]; ok && len(fg) > 0 {
				clusterConfig.APIServer.ExtraArgs["feature-gates"] = fmt.Sprintf("%s,%s", clusterConfig.APIServer.ExtraArgs["feature-gates"], featureGatesFlag)
			} else {
				clusterConfig.APIServer.ExtraArgs["feature-gates"] = featureGatesFlag
			}

			// Kubernetes Controller Manager
			if fg, ok := clusterConfig.ControllerManager.ExtraArgs["feature-gates"]; ok && len(fg) > 0 {
				clusterConfig.ControllerManager.ExtraArgs["feature-gates"] = fmt.Sprintf("%s,%s", clusterConfig.ControllerManager.ExtraArgs["feature-gates"], featureGatesFlag)
			} else {
				clusterConfig.ControllerManager.ExtraArgs["feature-gates"] = featureGatesFlag
			}

			// Kubelet
			for k, v := range featureGates {
				kubeletConfig.FeatureGates[k] = v
			}
		}
	}

	if cluster.Features.StaticAuditLog != nil && cluster.Features.StaticAuditLog.Enable {
		auditPolicyVol := kubeadmv1beta3.HostPathMount{
			Name:      "audit-conf",
			HostPath:  "/etc/kubernetes/audit",
			MountPath: "/etc/kubernetes/audit",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		logVol := kubeadmv1beta3.HostPathMount{
			Name:      "log",
			HostPath:  filepath.Dir(cluster.Features.StaticAuditLog.Config.LogPath),
			MountPath: filepath.Dir(cluster.Features.StaticAuditLog.Config.LogPath),
			ReadOnly:  false,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, auditPolicyVol)
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, logVol)
	}

	if cluster.Features.PodNodeSelector != nil && cluster.Features.PodNodeSelector.Enable {
		admissionVol := kubeadmv1beta3.HostPathMount{
			Name:      "admission-conf",
			HostPath:  "/etc/kubernetes/admission",
			MountPath: "/etc/kubernetes/admission",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, admissionVol)
	}
	// this is not exactly as s.EncryptionEnabled(). We need this to be true during the enable/disable or disable/enable transition.
	if (cluster.Features.EncryptionProviders != nil && cluster.Features.EncryptionProviders.Enable) ||
		s.LiveCluster.EncryptionConfiguration.Enable {
		encryptionProvidersVol := kubeadmv1beta3.HostPathMount{
			Name:      "encryption-providers-conf",
			HostPath:  "/etc/kubernetes/encryption-providers",
			MountPath: "/etc/kubernetes/encryption-providers",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, encryptionProvidersVol)

		// Handle external KMS case.
		if s.LiveCluster.CustomEncryptionEnabled() ||
			s.Cluster.Features.EncryptionProviders != nil && s.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration != "" {
			ksmSocket, err := s.GetKMSSocketPath()
			if err != nil {
				return nil, err
			}
			if ksmSocket != "" {
				clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, kubeadmv1beta3.HostPathMount{
					Name:      "kms-endpoint",
					HostPath:  ksmSocket,
					MountPath: ksmSocket,
					PathType:  corev1.HostPathSocket,
				})
			}
		}
	}

	args := kubeadmargs.NewFrom(clusterConfig.APIServer.ExtraArgs)
	features.UpdateKubeadmClusterConfiguration(cluster.Features, args)

	clusterConfig.APIServer.ExtraArgs = args.APIServer.ExtraArgs
	clusterConfig.FeatureGates = args.FeatureGates

	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration

	kubeproxyConfig := kubeProxyConfiguration(s)

	return []runtime.Object{initConfig, joinConfig, clusterConfig, kubeletConfig, kubeproxyConfig}, nil
}

// NewConfig returns all required configs to init a cluster via a set of v13 configs
func NewConfigWorker(s *state.State, host kubeoneapi.HostConfig) ([]runtime.Object, error) {
	cluster := s.Cluster

	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--etc-kubernetes-manifests",
	}

	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoint.Host, cluster.APIEndpoint.Port)

	joinConfig := &kubeadmv1beta3.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta3",
			Kind:       "JoinConfiguration",
		},
		Discovery: kubeadmv1beta3.Discovery{
			BootstrapToken: &kubeadmv1beta3.BootstrapTokenDiscovery{
				Token:                    s.JoinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}

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
		ClusterDNS:           []string{resources.NodeLocalDNSVirtualIP},
		ContainerLogMaxSize:  cluster.LoggingConfig.ContainerLogMaxSize,
		ContainerLogMaxFiles: &cluster.LoggingConfig.ContainerLogMaxFiles,
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
		FeatureGates: map[string]bool{},
	}

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	}

	if s.ShouldEnableInTreeCloudProvider() {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = cluster.CloudProvider.CloudProviderName()
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	if cluster.CloudProvider.External {
		if !s.ShouldEnableInTreeCloudProvider() {
			nodeRegistration.KubeletExtraArgs["cloud-provider"] = "external"
		}
		if s.ShouldEnableCSIMigration() {
			featureGates, _, err := s.Cluster.CSIMigrationFeatureGates(s.ShouldUnregisterInTreeCloudProvider())
			if err != nil {
				return nil, err
			}
			for k, v := range featureGates {
				kubeletConfig.FeatureGates[k] = v
			}
		}
	}

	joinConfig.NodeRegistration = nodeRegistration

	kubeproxyConfig := kubeProxyConfiguration(s)

	return []runtime.Object{joinConfig, kubeletConfig, kubeproxyConfig}, nil
}

func newNodeIP(host kubeoneapi.HostConfig) string {
	nodeIP := host.PrivateAddress
	if nodeIP == "" {
		nodeIP = host.PublicAddress
	}

	return nodeIP
}

func newNodeRegistration(s *state.State, host kubeoneapi.HostConfig) kubeadmv1beta3.NodeRegistrationOptions {
	kubeletCLIFlags := map[string]string{
		"node-ip":           newNodeIP(host),
		"volume-plugin-dir": "/var/lib/kubelet/volumeplugins",
	}

	if m := host.Kubelet.SystemReserved; m != nil {
		kubeletCLIFlags["system-reserved"] = kubeoneapi.MapStringStringToString(m, "=")
	}

	if m := host.Kubelet.KubeReserved; m != nil {
		kubeletCLIFlags["kube-reserved"] = kubeoneapi.MapStringStringToString(m, "=")
	}

	if m := host.Kubelet.EvictionHard; m != nil {
		kubeletCLIFlags["eviction-hard"] = kubeoneapi.MapStringStringToString(m, "<")
	}

	return kubeadmv1beta3.NodeRegistrationOptions{
		Name:             host.Hostname,
		Taints:           host.Taints,
		CRISocket:        s.Cluster.ContainerRuntime.CRISocket(),
		KubeletExtraArgs: kubeletCLIFlags,
	}
}

func kubeProxyConfiguration(s *state.State) *kubeproxyv1alpha1.KubeProxyConfiguration {
	kubeProxyConfig := &kubeproxyv1alpha1.KubeProxyConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeProxyConfiguration",
			APIVersion: "kubeproxy.config.k8s.io/v1alpha1",
		},
		ClusterCIDR: s.Cluster.ClusterNetwork.PodSubnet,
		ClientConnection: componentbasev1alpha1.ClientConnectionConfiguration{
			Kubeconfig: "/var/lib/kube-proxy/kubeconfig.conf",
		},
	}

	if kbPrx := s.Cluster.ClusterNetwork.KubeProxy; kbPrx != nil {
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

	return kubeProxyConfig
}
