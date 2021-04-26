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

package v1beta2

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	kubeadmv1beta2 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta2"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeflags"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"k8c.io/kubeone/pkg/templates/nodelocaldns"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

const (
	bootstrapTokenTTL = 60 * time.Minute
)

// NewConfig returns all required configs to init a cluster via a set of v1beta2 configs
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

	bootstrapToken, err := kubeadmv1beta2.NewBootstrapTokenString(s.JoinToken)
	if err != nil {
		return nil, err
	}

	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoint.Host, cluster.APIEndpoint.Port)

	initConfig := &kubeadmv1beta2.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "InitConfiguration",
		},
		BootstrapTokens: []kubeadmv1beta2.BootstrapToken{
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
		LocalAPIEndpoint: kubeadmv1beta2.APIEndpoint{
			AdvertiseAddress: newNodeIP(host),
		},
	}

	joinConfig := &kubeadmv1beta2.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta2.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta2.APIEndpoint{
				AdvertiseAddress: newNodeIP(host),
			},
		},
		Discovery: kubeadmv1beta2.Discovery{
			BootstrapToken: &kubeadmv1beta2.BootstrapTokenDiscovery{
				Token:                    s.JoinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}

	clusterConfig := &kubeadmv1beta2.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "ClusterConfiguration",
		},
		Networking: kubeadmv1beta2.Networking{
			PodSubnet:     cluster.ClusterNetwork.PodSubnet,
			ServiceSubnet: cluster.ClusterNetwork.ServiceSubnet,
			DNSDomain:     cluster.ClusterNetwork.ServiceDomainName,
		},
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta2.APIServer{
			ControlPlaneComponent: kubeadmv1beta2.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"endpoint-reconciler-type": "lease",
					"service-node-port-range":  cluster.ClusterNetwork.NodePortRange,
					"enable-admission-plugins": kubeflags.DefaultAdmissionControllers(kubeSemVer),
				},
				ExtraVolumes: []kubeadmv1beta2.HostPathMount{},
			},
			CertSANs: []string{strings.ToLower(cluster.APIEndpoint.Host)},
		},
		ControllerManager: kubeadmv1beta2.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"flex-volume-plugin-dir": "/var/lib/kubelet/volumeplugins",
			},
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{},
		},
		ClusterName:     cluster.Name,
		ImageRepository: cluster.AssetConfiguration.Kubernetes.ImageRepository,
		Etcd: kubeadmv1beta2.Etcd{
			Local: &kubeadmv1beta2.LocalEtcd{
				ImageMeta: kubeadmv1beta2.ImageMeta{
					ImageRepository: cluster.AssetConfiguration.Etcd.ImageRepository,
					ImageTag:        cluster.AssetConfiguration.Etcd.ImageTag,
				},
			},
		},
		DNS: kubeadmv1beta2.DNS{
			ImageMeta: kubeadmv1beta2.ImageMeta{
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
		CgroupDriver:       "systemd",
		ReadOnlyPort:       0,
		RotateCertificates: true,
		ClusterDNS:         []string{nodelocaldns.VirtualIP},
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
	}

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	}

	if cluster.CloudProvider.CloudProviderInTree() {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"
		cloudConfigVol := kubeadmv1beta2.HostPathMount{
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
		delete(clusterConfig.APIServer.ExtraArgs, "cloud-provider")
		delete(clusterConfig.ControllerManager.ExtraArgs, "cloud-provider")
		nodeRegistration.KubeletExtraArgs["cloud-provider"] = "external"
	}

	if cluster.Features.StaticAuditLog != nil && cluster.Features.StaticAuditLog.Enable {
		auditPolicyVol := kubeadmv1beta2.HostPathMount{
			Name:      "audit-conf",
			HostPath:  "/etc/kubernetes/audit",
			MountPath: "/etc/kubernetes/audit",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		logVol := kubeadmv1beta2.HostPathMount{
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
		admissionVol := kubeadmv1beta2.HostPathMount{
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
		encryptionProvidersVol := kubeadmv1beta2.HostPathMount{
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
				clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, kubeadmv1beta2.HostPathMount{
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

	return []runtime.Object{initConfig, joinConfig, clusterConfig, kubeletConfig}, nil
}

// NewConfig returns all required configs to init a cluster via a set of v1beta2 configs
func NewConfigWorker(s *state.State, host kubeoneapi.HostConfig) ([]runtime.Object, error) {
	cluster := s.Cluster

	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--etc-kubernetes-manifests",
	}

	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoint.Host, cluster.APIEndpoint.Port)

	joinConfig := &kubeadmv1beta2.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "JoinConfiguration",
		},
		Discovery: kubeadmv1beta2.Discovery{
			BootstrapToken: &kubeadmv1beta2.BootstrapTokenDiscovery{
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
		CgroupDriver:       "systemd",
		ReadOnlyPort:       0,
		RotateCertificates: true,
		ClusterDNS:         []string{nodelocaldns.VirtualIP},
		Authentication: kubeletconfigv1beta1.KubeletAuthentication{
			Anonymous: kubeletconfigv1beta1.KubeletAnonymousAuthentication{
				Enabled: &bfalse,
			},
		},
	}

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	}

	if cluster.CloudProvider.CloudProviderInTree() {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = cluster.CloudProvider.CloudProviderName()
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	if cluster.CloudProvider.External {
		nodeRegistration.KubeletExtraArgs["cloud-provider"] = "external"
	}

	joinConfig.NodeRegistration = nodeRegistration

	return []runtime.Object{joinConfig, kubeletConfig}, nil
}

func newNodeIP(host kubeoneapi.HostConfig) string {
	nodeIP := host.PrivateAddress
	if nodeIP == "" {
		nodeIP = host.PublicAddress
	}

	return nodeIP
}

func newNodeRegistration(s *state.State, host kubeoneapi.HostConfig) kubeadmv1beta2.NodeRegistrationOptions {
	return kubeadmv1beta2.NodeRegistrationOptions{
		Name:      host.Hostname,
		Taints:    host.Taints,
		CRISocket: s.Cluster.ContainerRuntime.CRISocket(),
		KubeletExtraArgs: map[string]string{
			"node-ip":           newNodeIP(host),
			"volume-plugin-dir": "/var/lib/kubelet/volumeplugins",
		},
	}
}
