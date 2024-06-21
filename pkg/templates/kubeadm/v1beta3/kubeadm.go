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
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	bootstraptokenv1 "k8c.io/kubeone/pkg/apis/kubeadm/bootstraptoken/v1"
	kubeadmv1beta3 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta3"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeflags"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"k8c.io/kubeone/pkg/templates/kubernetesconfigs"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	bootstrapTokenTTL = 60 * time.Minute
)

const (
	// fixedEtcdVersion is an etcd version that doesn't have known data integrity and durability bugs
	// (see etcdVersionCorruptCheckExtraArgs for more details)
	fixedEtcdVersion = "3.5.10-0"
)

const (
	registryK8sio = "registry.k8s.io"
)

var (
	// defines a semver constraint used to check if Kubernetes 1.24 uses fixed etcd version
	fixedEtcd124Constraint = semverutil.MustParseConstraint(">= 1.24, < 1.25")

	// defines a semver constraint used to check if Kubernetes 1.25 uses fixed etcd version
	fixedEtcd125Constraint = semverutil.MustParseConstraint(">= 1.25, < 1.26")

	// defines a semver constraint used to check if Kubernetes 1.26 uses fixed etcd version
	fixedEtcd126Constraint = semverutil.MustParseConstraint(">= 1.26.13, < 1.27")

	// defines a semver constraint used to check if Kubernetes 1.27 uses fixed etcd version
	fixedEtcd127Constraint = semverutil.MustParseConstraint(">= 1.27.9, < 1.28")

	// defines a semver constraint used to check if Kubernetes 1.28 uses fixed etcd version
	fixedEtcd128Constraint = semverutil.MustParseConstraint(">= 1.28.6, < 1.29")

	// defines a semver constraint used to check if Kubernetes 1.29 uses fixed etcd version
	fixedEtcd129Constraint = semverutil.MustParseConstraint(">= 1.29")
)

type Config struct {
	InitConfiguration    *kubeadmv1beta3.InitConfiguration
	JoinConfiguration    *kubeadmv1beta3.JoinConfiguration
	ClusterConfiguration *kubeadmv1beta3.ClusterConfiguration

	KubeletConfiguration   runtime.Object
	KubeProxyConfiguration runtime.Object
}

// NewConfig returns all required configs to init a cluster via a set of v1beta3 configs
func NewConfig(s *state.State, host kubeoneapi.HostConfig) (*Config, error) {
	cluster := s.Cluster
	kubeSemVer, err := semver.NewVersion(cluster.Versions.Kubernetes)
	if err != nil {
		return nil, fail.Config(err, "parsing kubernetes semver")
	}

	etcdImageTag, etcdExtraArgs := etcdVersionCorruptCheckExtraArgs(kubeSemVer, cluster.AssetConfiguration.Etcd.ImageTag, cluster.TLSCipherSuites.Etcd)

	if s.Cluster.ClusterNetwork.HasIPv6() && len(host.IPv6Addresses) == 0 {
		return nil, fmt.Errorf("host must have ipv6 address for %q family", s.Cluster.ClusterNetwork.IPFamily)
	}

	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--var-lib-etcd",
		"DirAvailable--etc-kubernetes-manifests",
		"ImagePull",
	}

	bootstrapToken, err := bootstraptokenv1.NewBootstrapTokenString(s.JoinToken)
	if err != nil {
		return nil, fail.Runtime(err, "generating kubeadm bootstrap token")
	}

	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoint.Host, cluster.APIEndpoint.Port)

	var advertiseAddress string
	if s.Cluster.ClusterNetwork.IPFamily.IsIPv6Primary() {
		advertiseAddress = host.IPv6Addresses[0]
	} else {
		advertiseAddress = newNodeIP(host)
	}

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
			AdvertiseAddress: advertiseAddress,
		},
	}

	joinConfig := &kubeadmv1beta3.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta3",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta3.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta3.APIEndpoint{
				AdvertiseAddress: advertiseAddress,
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
			PodSubnet: join(
				cluster.ClusterNetwork.IPFamily,
				cluster.ClusterNetwork.PodSubnet,
				cluster.ClusterNetwork.PodSubnetIPv6,
			),
			ServiceSubnet: join(
				cluster.ClusterNetwork.IPFamily,
				cluster.ClusterNetwork.ServiceSubnet,
				cluster.ClusterNetwork.ServiceSubnetIPv6,
			),
			DNSDomain: cluster.ClusterNetwork.ServiceDomainName,
		},
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta3.APIServer{
			ControlPlaneComponent: kubeadmv1beta3.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"enable-admission-plugins":      kubeflags.DefaultAdmissionControllers(),
					"endpoint-reconciler-type":      "lease",
					"kubelet-certificate-authority": "/etc/kubernetes/pki/ca.crt",
					"profiling":                     "false",
					"request-timeout":               "1m",
					"service-node-port-range":       cluster.ClusterNetwork.NodePortRange,
					"tls-cipher-suites":             strings.Join(cluster.TLSCipherSuites.APIServer, ","),
				},
				ExtraVolumes: []kubeadmv1beta3.HostPathMount{},
			},
			CertSANs: certSANS,
		},
		ControllerManager: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"flex-volume-plugin-dir":      "/var/lib/kubelet/volumeplugins",
				"profiling":                   "false",
				"terminated-pod-gc-threshold": "1000",
			},
			ExtraVolumes: []kubeadmv1beta3.HostPathMount{},
		},
		Scheduler: kubeadmv1beta3.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"profiling": "false",
			},
			ExtraVolumes: []kubeadmv1beta3.HostPathMount{},
		},
		ClusterName:     cluster.Name,
		ImageRepository: defaults(cluster.AssetConfiguration.Kubernetes.ImageRepository, registryK8sio),
		Etcd: kubeadmv1beta3.Etcd{
			Local: &kubeadmv1beta3.LocalEtcd{
				ImageMeta: kubeadmv1beta3.ImageMeta{
					ImageRepository: defaults(cluster.AssetConfiguration.Etcd.ImageRepository, registryK8sio),
					ImageTag:        etcdImageTag,
				},
				ExtraArgs: etcdExtraArgs,
			},
		},
		DNS: kubeadmv1beta3.DNS{
			ImageMeta: kubeadmv1beta3.ImageMeta{
				ImageRepository: defaults(cluster.AssetConfiguration.CoreDNS.ImageRepository, fmt.Sprintf("%s/coredns", registryK8sio)),
				ImageTag:        cluster.AssetConfiguration.CoreDNS.ImageTag,
			},
		},
	}

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	} else {
		sandboxImage, serr := cluster.Versions.SandboxImage(cluster.RegistryConfiguration.ImageRegistry)
		if serr != nil {
			return nil, serr
		}

		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = sandboxImage
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

	var (
		kubeletFeatureGates map[string]bool
		featureGatesFlag    string
	)

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
			kubeletFeatureGates, featureGatesFlag, err = s.Cluster.CSIMigrationFeatureGates(s.ShouldUnregisterInTreeCloudProvider())
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
			ksmSocket, socketErr := s.GetKMSSocketPath()
			if socketErr != nil {
				return nil, socketErr
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

	addControllerManagerNetworkArgs(clusterConfig.ControllerManager.ExtraArgs, cluster.ClusterNetwork)

	args := kubeadmargs.NewFrom(clusterConfig.APIServer.ExtraArgs)
	features.UpdateKubeadmClusterConfiguration(cluster.Features, args)

	clusterConfig.APIServer.ExtraArgs = args.APIServer.ExtraArgs
	clusterConfig.FeatureGates = args.FeatureGates

	// This function call must be at the very end to ensure flags and feature gates
	// can be overridden.
	addControlPlaneComponentsAdditionalArgs(cluster, clusterConfig)

	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration

	kubeletConfig, err := kubernetesconfigs.NewKubeletConfiguration(s.Cluster, kubeletFeatureGates)
	if err != nil {
		return nil, err
	}

	kubeProxyConfig, err := kubernetesconfigs.NewKubeProxyConfiguration(s.Cluster)
	if err != nil {
		return nil, err
	}

	return &Config{
		InitConfiguration:      initConfig,
		JoinConfiguration:      joinConfig,
		ClusterConfiguration:   clusterConfig,
		KubeletConfiguration:   kubeletConfig,
		KubeProxyConfiguration: kubeProxyConfig,
	}, nil
}

func addControllerManagerNetworkArgs(m map[string]string, clusterNetwork kubeoneapi.ClusterNetworkConfig) {
	if clusterNetwork.CNI.Cilium != nil {
		return
	}

	switch clusterNetwork.IPFamily {
	case kubeoneapi.IPFamilyIPv4:
		if clusterNetwork.NodeCIDRMaskSizeIPv4 != nil {
			m["node-cidr-mask-size-ipv4"] = fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv4)
		}
	case kubeoneapi.IPFamilyIPv6:
		if clusterNetwork.NodeCIDRMaskSizeIPv6 != nil {
			m["node-cidr-mask-size-ipv6"] = fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv6)
		}
	case kubeoneapi.IPFamilyIPv4IPv6, kubeoneapi.IPFamilyIPv6IPv4:
		if clusterNetwork.NodeCIDRMaskSizeIPv4 != nil {
			m["node-cidr-mask-size-ipv4"] = fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv4)
		}
		if clusterNetwork.NodeCIDRMaskSizeIPv6 != nil {
			m["node-cidr-mask-size-ipv6"] = fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv6)
		}
	}
}

func addControlPlaneComponentsAdditionalArgs(cluster *kubeoneapi.KubeOneCluster, clusterCfg *kubeadmv1beta3.ClusterConfiguration) {
	if cluster.ControlPlaneComponents != nil {
		if cluster.ControlPlaneComponents.ControllerManager != nil {
			if cluster.ControlPlaneComponents.ControllerManager.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.ControllerManager.Flags {
					clusterCfg.ControllerManager.ExtraArgs[k] = v
				}
			}
			if cluster.ControlPlaneComponents.ControllerManager.FeatureGates != nil {
				clusterCfg.ControllerManager.ExtraArgs["feature-gates"] = mergeFeatureGates(clusterCfg.ControllerManager.ExtraArgs["feature-gates"], cluster.ControlPlaneComponents.ControllerManager.FeatureGates)
			}
		}
		if cluster.ControlPlaneComponents.Scheduler != nil {
			if cluster.ControlPlaneComponents.Scheduler.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.Scheduler.Flags {
					clusterCfg.Scheduler.ExtraArgs[k] = v
				}
			}
			if cluster.ControlPlaneComponents.Scheduler.FeatureGates != nil {
				clusterCfg.Scheduler.ExtraArgs["feature-gates"] = mergeFeatureGates(clusterCfg.Scheduler.ExtraArgs["feature-gates"], cluster.ControlPlaneComponents.Scheduler.FeatureGates)
			}
		}
		if cluster.ControlPlaneComponents.APIServer != nil {
			if cluster.ControlPlaneComponents.APIServer.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.APIServer.Flags {
					clusterCfg.APIServer.ExtraArgs[k] = v
				}
			}
			if cluster.ControlPlaneComponents.APIServer.FeatureGates != nil {
				clusterCfg.APIServer.ExtraArgs["feature-gates"] = mergeFeatureGates(clusterCfg.APIServer.ExtraArgs["feature-gates"], cluster.ControlPlaneComponents.APIServer.FeatureGates)
			}
		}
	}
}

func join(ipFamily kubeoneapi.IPFamily, ipv4Subnet, ipv6Subnet string) string {
	switch ipFamily {
	case kubeoneapi.IPFamilyIPv4:
		return ipv4Subnet
	case kubeoneapi.IPFamilyIPv6:
		return ipv6Subnet
	case kubeoneapi.IPFamilyIPv4IPv6:
		return strings.Join([]string{ipv4Subnet, ipv6Subnet}, ",")
	case kubeoneapi.IPFamilyIPv6IPv4:
		return strings.Join([]string{ipv6Subnet, ipv4Subnet}, ",")
	default:
		return "unknown IP family"
	}
}

// NewConfig returns all required configs to init a cluster via a set of v13 configs
func NewConfigWorker(s *state.State, host kubeoneapi.HostConfig) (*Config, error) {
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

	if cluster.AssetConfiguration.Pause.ImageRepository != "" {
		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = cluster.AssetConfiguration.Pause.ImageRepository + "/pause:" + cluster.AssetConfiguration.Pause.ImageTag
	} else {
		sandboxImage, serr := cluster.Versions.SandboxImage(s.Cluster.RegistryConfiguration.ImageRegistry)
		if serr != nil {
			return nil, serr
		}

		nodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = sandboxImage
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
	}

	joinConfig.NodeRegistration = nodeRegistration

	return &Config{
		JoinConfiguration: joinConfig,
	}, nil
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
		"volume-plugin-dir": "/var/lib/kubelet/volumeplugins",
	}

	// If external or in-tree CCM is in use we don't need to set --node-ip
	// as the cloud provider will know what IPs to return.
	if s.Cluster.ClusterNetwork.IPFamily.IsDualstack() {
		if !s.Cluster.CloudProvider.External {
			switch {
			case s.Cluster.ClusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv4IPv6:
				kubeletCLIFlags["node-ip"] = newNodeIP(host) + "," + host.IPv6Addresses[0]
			case s.Cluster.ClusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv6IPv4:
				kubeletCLIFlags["node-ip"] = host.IPv6Addresses[0] + "," + newNodeIP(host)
			}
		}
	} else {
		kubeletCLIFlags["node-ip"] = newNodeIP(host)
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
	if m := host.Kubelet.MaxPods; m != nil {
		kubeletCLIFlags["max-pods"] = strconv.Itoa(int(*m))
	}

	return kubeadmv1beta3.NodeRegistrationOptions{
		Name:             host.Hostname,
		Taints:           host.Taints,
		CRISocket:        fmt.Sprintf("unix://%s", s.Cluster.ContainerRuntime.CRISocket()),
		KubeletExtraArgs: kubeletCLIFlags,
	}
}

// etcdVersionCorruptCheckExtraArgs provides etcd version and args to be used.
// This is required because:
//   - etcd v3.5.[0-2] has an issue with the data integrity
//     https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ
//   - etcd v3.5.[0-4] has a durability issue affecting single-node (non-HA) etcd clusters
//     https://groups.google.com/a/kubernetes.io/g/dev/c/7q4tB_Vp3Uc/m/MrHalhCIBAAJ
func etcdVersionCorruptCheckExtraArgs(kubeVersion *semver.Version, etcdImageTag string, cipherSuites []string) (string, map[string]string) {
	etcdExtraArgs := map[string]string{
		"experimental-compact-hash-check-enabled": "true",
		"experimental-initial-corrupt-check":      "true",
		"experimental-corrupt-check-time":         "240m",
	}

	if len(cipherSuites) > 0 {
		etcdExtraArgs["cipher-suites"] = strings.Join(cipherSuites, ",")
	}

	switch {
	case etcdImageTag != "":
		return etcdImageTag, etcdExtraArgs
	case fixedEtcd124Constraint.Check(kubeVersion):
		return fixedEtcdVersion, etcdExtraArgs
	case fixedEtcd125Constraint.Check(kubeVersion):
		return fixedEtcdVersion, etcdExtraArgs
	case fixedEtcd126Constraint.Check(kubeVersion):
		fallthrough
	case fixedEtcd127Constraint.Check(kubeVersion):
		fallthrough
	case fixedEtcd128Constraint.Check(kubeVersion):
		fallthrough
	case fixedEtcd129Constraint.Check(kubeVersion):
		return "", etcdExtraArgs
	default:
		return fixedEtcdVersion, etcdExtraArgs
	}
}

func defaults(input, defaultValue string) string {
	if input != "" {
		return input
	}

	return defaultValue
}

func mergeFeatureGates(featureGates string, additionalFeatureGates map[string]bool) string {
	featureGatesMap := make(map[string]bool)
	featureGatesArr := strings.Split(featureGates, ",")
	for _, fg := range featureGatesArr {
		kv := strings.Split(fg, "=")
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if value == "true" {
				featureGatesMap[key] = true
			} else if value == "false" {
				featureGatesMap[key] = false
			}
		}
	}

	for k, v := range additionalFeatureGates {
		featureGatesMap[k] = v
	}

	featureGatesKeys := sets.List(sets.KeySet(featureGatesMap))

	var featureGatesStr []string
	for _, k := range featureGatesKeys {
		featureGatesStr = append(featureGatesStr, fmt.Sprintf("%s=%t", k, featureGatesMap[k]))
	}

	return strings.Join(featureGatesStr, ",")
}
