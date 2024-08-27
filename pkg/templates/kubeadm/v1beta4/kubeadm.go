/*
Copyright 2024 The KubeOne Authors.

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

package v1beta4

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	bootstraptokenv1 "k8c.io/kubeone/pkg/apis/kubeadm/bootstraptoken/v1"
	kubeadmv1beta4 "k8c.io/kubeone/pkg/apis/kubeadm/v1beta4"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/features"
	"k8c.io/kubeone/pkg/kubeflags"
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

type Config struct {
	InitConfiguration    *kubeadmv1beta4.InitConfiguration
	JoinConfiguration    *kubeadmv1beta4.JoinConfiguration
	ClusterConfiguration *kubeadmv1beta4.ClusterConfiguration

	KubeletConfiguration   runtime.Object
	KubeProxyConfiguration runtime.Object
}

// NewConfig returns all required configs to init a cluster via a set of v1beta4 configs
func NewConfig(s *state.State, host kubeoneapi.HostConfig) (*Config, error) {
	cluster := s.Cluster

	overwriteRegistry := ""
	if cluster.RegistryConfiguration != nil {
		overwriteRegistry = cluster.RegistryConfiguration.OverwriteRegistry
	}

	bootstrapToken, err := bootstraptokenv1.NewBootstrapTokenString(s.JoinToken)
	if err != nil {
		return nil, fail.Runtime(err, "generating kubeadm bootstrap token")
	}

	var advertiseAddress string
	if cluster.ClusterNetwork.IPFamily.IsIPv6Primary() {
		advertiseAddress = host.IPv6Addresses[0]
	} else {
		advertiseAddress = newNodeIP(host)
	}

	controlPlaneEndpoint := net.JoinHostPort(cluster.APIEndpoint.Host, strconv.Itoa(cluster.APIEndpoint.Port))
	certSANS := certificate.GetCertificateSANs(cluster.APIEndpoint.Host, cluster.APIEndpoint.AlternativeNames)

	initConfig := newInitConfiguration(bootstrapToken, advertiseAddress)
	joinConfig := newJoinConfiguration(advertiseAddress, s.JoinToken, controlPlaneEndpoint)
	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--var-lib-etcd",
		"DirAvailable--etc-kubernetes-manifests",
		"ImagePull",
	}

	clusterConfig := &kubeadmv1beta4.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta4",
			Kind:       "ClusterConfiguration",
		},
		ClusterName:          cluster.Name,
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta4.APIServer{
			ControlPlaneComponent: kubeadmv1beta4.ControlPlaneComponent{
				ExtraArgs: []kubeadmv1beta4.Arg{
					{
						Name:  "enable-admission-plugins",
						Value: kubeflags.DefaultAdmissionControllers(),
					},
					{
						Name:  "endpoint-reconciler-type",
						Value: "lease",
					},
					{
						Name:  "kubelet-certificate-authority",
						Value: "/etc/kubernetes/pki/ca.crt",
					},
					{
						Name:  "profiling",
						Value: "false",
					},
					{
						Name:  "request-timeout",
						Value: "1m",
					},
					{
						Name:  "service-node-port-range",
						Value: cluster.ClusterNetwork.NodePortRange,
					},
					{
						Name:  "tls-cipher-suites",
						Value: strings.Join(cluster.TLSCipherSuites.APIServer, ","),
					},
				},
				ExtraVolumes: []kubeadmv1beta4.HostPathMount{},
			},
			CertSANs: certSANS,
		},
		ControllerManager: kubeadmv1beta4.ControlPlaneComponent{
			ExtraArgs: []kubeadmv1beta4.Arg{
				{
					Name:  "flex-volume-plugin-dir",
					Value: "/var/lib/kubelet/volumeplugins",
				},
				{
					Name:  "profiling",
					Value: "false",
				},
				{
					Name:  "terminated-pod-gc-threshold",
					Value: "1000",
				},
			},
			ExtraVolumes: []kubeadmv1beta4.HostPathMount{},
		},
		Scheduler: kubeadmv1beta4.ControlPlaneComponent{
			ExtraArgs: []kubeadmv1beta4.Arg{
				{
					Name:  "profiling",
					Value: "false",
				},
			},
			ExtraVolumes: []kubeadmv1beta4.HostPathMount{},
		},
		Etcd: kubeadmv1beta4.Etcd{
			Local: &kubeadmv1beta4.LocalEtcd{
				ImageMeta: kubeadmv1beta4.ImageMeta{
					ImageRepository: overwriteRegistry,
				},
				ExtraArgs: etcdVersionCorruptCheckExtraArgs(cluster.TLSCipherSuites.Etcd),
			},
		},
		DNS: kubeadmv1beta4.DNS{
			ImageMeta: kubeadmv1beta4.ImageMeta{
				ImageRepository: defaults(s.Cluster.Features.CoreDNS.ImageRepository, overwriteRegistry),
			},
		},
		ImageRepository: overwriteRegistry,
		Networking: kubeadmv1beta4.Networking{
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
	}

	if cluster.ClusterNetwork.KubeProxy != nil && cluster.ClusterNetwork.KubeProxy.SkipInstallation {
		clusterConfig.DNS.Disabled = true
	}

	if err = addFeaturesExtraMounts(s, clusterConfig); err != nil {
		return nil, err
	}
	addControllerManagerNetworkArgs(clusterConfig, cluster.ClusterNetwork)

	args := kubeadmargs.NewFrom(argsToMap(clusterConfig.APIServer.ExtraArgs))
	features.UpdateKubeadmArguments(cluster.Features, args)
	clusterConfig.APIServer.ExtraArgs = stringStringMapToArgs(args.APIServer.ExtraArgs)

	// This function call must be at the very end to ensure flags and feature gates
	// can be overridden.
	addControlPlaneComponentsAdditionalArgs(cluster, clusterConfig)

	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration

	kubeletConfig, err := kubernetesconfigs.NewKubeletConfiguration(cluster, nil)
	if err != nil {
		return nil, err
	}

	kubeProxyConfig, err := kubernetesconfigs.NewKubeProxyConfiguration(cluster)
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

// etcdVersionCorruptCheckExtraArgs provides etcd version and args to be used.
// This is required because:
//   - etcd v3.5.[0-2] has an issue with the data integrity
//     https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ
//   - etcd v3.5.[0-4] has a durability issue affecting single-node (non-HA) etcd clusters
//     https://groups.google.com/a/kubernetes.io/g/dev/c/7q4tB_Vp3Uc/m/MrHalhCIBAAJ
func etcdVersionCorruptCheckExtraArgs(cipherSuites []string) []kubeadmv1beta4.Arg {
	etcdExtraArgs := []kubeadmv1beta4.Arg{
		{
			Name:  "experimental-compact-hash-check-enabled",
			Value: "true",
		},
		{
			Name:  "experimental-initial-corrupt-check",
			Value: "true",
		},
		{
			Name:  "experimental-corrupt-check-time",
			Value: "240m",
		},
	}

	if len(cipherSuites) > 0 {
		etcdExtraArgs = append(etcdExtraArgs, kubeadmv1beta4.Arg{
			Name:  "cipher-suites",
			Value: strings.Join(cipherSuites, ","),
		})
	}

	return etcdExtraArgs
}

func addFeaturesExtraMounts(s *state.State, clusterConfig *kubeadmv1beta4.ClusterConfiguration) error {
	cluster := s.Cluster

	// StaticAuditLog and WebhookAuditLog both share the audit-conf volume and since both
	// can be activated simultaneously, we need to make sure to add the ExtraVolume only once
	if (cluster.Features.StaticAuditLog != nil && cluster.Features.StaticAuditLog.Enable) || (cluster.Features.WebhookAuditLog != nil && cluster.Features.WebhookAuditLog.Enable) {
		auditPolicyVol := kubeadmv1beta4.HostPathMount{
			Name:      "audit-conf",
			HostPath:  "/etc/kubernetes/audit",
			MountPath: "/etc/kubernetes/audit",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, auditPolicyVol)
	}

	if cluster.Features.StaticAuditLog != nil && cluster.Features.StaticAuditLog.Enable {
		logVol := kubeadmv1beta4.HostPathMount{
			Name:      "log",
			HostPath:  filepath.Dir(cluster.Features.StaticAuditLog.Config.LogPath),
			MountPath: filepath.Dir(cluster.Features.StaticAuditLog.Config.LogPath),
			ReadOnly:  false,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, logVol)
	}

	if cluster.Features.PodNodeSelector != nil && cluster.Features.PodNodeSelector.Enable {
		admissionVol := kubeadmv1beta4.HostPathMount{
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
		(s.LiveCluster.EncryptionConfiguration != nil && s.LiveCluster.EncryptionConfiguration.Enable) {
		encryptionProvidersVol := kubeadmv1beta4.HostPathMount{
			Name:      "encryption-providers-conf",
			HostPath:  "/etc/kubernetes/encryption-providers",
			MountPath: "/etc/kubernetes/encryption-providers",
			ReadOnly:  true,
			PathType:  corev1.HostPathDirectoryOrCreate,
		}
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, encryptionProvidersVol)

		// Handle external KMS case.
		if s.LiveCluster.CustomEncryptionEnabled() ||
			cluster.Features.EncryptionProviders != nil && cluster.Features.EncryptionProviders.CustomEncryptionConfiguration != "" {
			ksmSocket, socketErr := s.GetKMSSocketPath()
			if socketErr != nil {
				return socketErr
			}
			if ksmSocket != "" {
				clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, kubeadmv1beta4.HostPathMount{
					Name:      "kms-endpoint",
					HostPath:  ksmSocket,
					MountPath: ksmSocket,
					PathType:  corev1.HostPathSocket,
				})
			}
		}
	}

	return nil
}

func addControllerManagerNetworkArgs(clusterConfig *kubeadmv1beta4.ClusterConfiguration, clusterNetwork kubeoneapi.ClusterNetworkConfig) {
	if clusterNetwork.CNI.Cilium != nil {
		return
	}

	switch clusterNetwork.IPFamily {
	case kubeoneapi.IPFamilyIPv4:
		if clusterNetwork.NodeCIDRMaskSizeIPv4 != nil {
			clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, "node-cidr-mask-size-ipv4", fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv4))
		}
	case kubeoneapi.IPFamilyIPv6:
		if clusterNetwork.NodeCIDRMaskSizeIPv6 != nil {
			clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, "node-cidr-mask-size-ipv6", fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv6))
		}
	case kubeoneapi.IPFamilyIPv4IPv6, kubeoneapi.IPFamilyIPv6IPv4:
		if clusterNetwork.NodeCIDRMaskSizeIPv4 != nil {
			clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, "node-cidr-mask-size-ipv4", fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv4))
		}
		if clusterNetwork.NodeCIDRMaskSizeIPv6 != nil {
			clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, "node-cidr-mask-size-ipv6", fmt.Sprintf("%d", *clusterNetwork.NodeCIDRMaskSizeIPv6))
		}
	}
}

func addControlPlaneComponentsAdditionalArgs(cluster *kubeoneapi.KubeOneCluster, clusterConfig *kubeadmv1beta4.ClusterConfiguration) {
	if cluster.ControlPlaneComponents != nil {
		if cluster.ControlPlaneComponents.ControllerManager != nil {
			if cluster.ControlPlaneComponents.ControllerManager.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.ControllerManager.Flags {
					clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, k, v)
				}
			}
			if cluster.ControlPlaneComponents.ControllerManager.FeatureGates != nil {
				val, _ := kubeadmv1beta4.GetArgValue(clusterConfig.ControllerManager.ExtraArgs, "feature-gates", -1)
				clusterConfig.ControllerManager.ExtraArgs = setAllArgsValue(clusterConfig.ControllerManager.ExtraArgs, "feature-gates", mergeFeatureGates(val, cluster.ControlPlaneComponents.ControllerManager.FeatureGates))
			}
		}
		if cluster.ControlPlaneComponents.Scheduler != nil {
			if cluster.ControlPlaneComponents.Scheduler.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.Scheduler.Flags {
					clusterConfig.Scheduler.ExtraArgs = setAllArgsValue(clusterConfig.Scheduler.ExtraArgs, k, v)
				}
			}
			if cluster.ControlPlaneComponents.Scheduler.FeatureGates != nil {
				val, _ := kubeadmv1beta4.GetArgValue(clusterConfig.Scheduler.ExtraArgs, "feature-gates", -1)
				clusterConfig.Scheduler.ExtraArgs = setAllArgsValue(clusterConfig.Scheduler.ExtraArgs, "feature-gates", mergeFeatureGates(val, cluster.ControlPlaneComponents.Scheduler.FeatureGates))
			}
		}
		if cluster.ControlPlaneComponents.APIServer != nil {
			if cluster.ControlPlaneComponents.APIServer.Flags != nil {
				for k, v := range cluster.ControlPlaneComponents.APIServer.Flags {
					clusterConfig.APIServer.ExtraArgs = setAllArgsValue(clusterConfig.APIServer.ExtraArgs, k, v)
				}
			}
			if cluster.ControlPlaneComponents.APIServer.FeatureGates != nil {
				val, _ := kubeadmv1beta4.GetArgValue(clusterConfig.APIServer.ExtraArgs, "feature-gates", -1)
				clusterConfig.APIServer.ExtraArgs = setAllArgsValue(clusterConfig.APIServer.ExtraArgs, "feature-gates", mergeFeatureGates(val, cluster.ControlPlaneComponents.APIServer.FeatureGates))
			}
		}
	}
}

// NewConfig returns all required configs to init a cluster via a set of v13 configs
func NewConfigWorker(s *state.State, host kubeoneapi.HostConfig) (*Config, error) {
	cluster := s.Cluster

	nodeRegistration := newNodeRegistration(s, host)
	nodeRegistration.IgnorePreflightErrors = []string{
		"DirAvailable--etc-kubernetes-manifests",
	}

	controlPlaneEndpoint := net.JoinHostPort(cluster.APIEndpoint.Host, strconv.Itoa(cluster.APIEndpoint.Port))

	joinConfig := &kubeadmv1beta4.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta4",
			Kind:       "JoinConfiguration",
		},
		Discovery: kubeadmv1beta4.Discovery{
			BootstrapToken: &kubeadmv1beta4.BootstrapTokenDiscovery{
				Token:                    s.JoinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}

	if cluster.CloudProvider.External {
		nodeRegistration.KubeletExtraArgs = setAllArgsValue(nodeRegistration.KubeletExtraArgs, "cloud-provider", "external")
	}

	joinConfig.NodeRegistration = nodeRegistration

	return &Config{
		JoinConfiguration: joinConfig,
	}, nil
}

func newInitConfiguration(bootstrapToken *bootstraptokenv1.BootstrapTokenString, advertiseAddress string) *kubeadmv1beta4.InitConfiguration {
	return &kubeadmv1beta4.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta4",
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
		LocalAPIEndpoint: kubeadmv1beta4.APIEndpoint{
			AdvertiseAddress: advertiseAddress,
		},
	}
}

func newJoinConfiguration(advertiseAddress, joinToken, controlPlaneEndpoint string) *kubeadmv1beta4.JoinConfiguration {
	return &kubeadmv1beta4.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta4",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta4.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta4.APIEndpoint{
				AdvertiseAddress: advertiseAddress,
			},
		},
		Discovery: kubeadmv1beta4.Discovery{
			BootstrapToken: &kubeadmv1beta4.BootstrapTokenDiscovery{
				Token:                    joinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}
}

func newNodeRegistration(s *state.State, host kubeoneapi.HostConfig) kubeadmv1beta4.NodeRegistrationOptions {
	kubeletCLIFlags := []kubeadmv1beta4.Arg{
		{
			Name:  "volume-plugin-dir",
			Value: "/var/lib/kubelet/volumeplugins",
		},
	}
	if s.Cluster.CloudProvider.External {
		kubeletCLIFlags = append(kubeletCLIFlags, kubeadmv1beta4.Arg{
			Name:  "cloud-provider",
			Value: "external",
		})
	}

	// --node-ip flag must be set on kubelet when:
	//   - when running IPv6 Dualstack without CCM
	//   - when IPv6 Dualstack is disabled
	if s.Cluster.ClusterNetwork.IPFamily.IsDualstack() {
		if !s.Cluster.CloudProvider.External {
			switch {
			case s.Cluster.ClusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv4IPv6:
				kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "node-ip", newNodeIP(host)+","+host.IPv6Addresses[0])
			case s.Cluster.ClusterNetwork.IPFamily == kubeoneapi.IPFamilyIPv6IPv4:
				kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "node-ip", host.IPv6Addresses[0]+","+newNodeIP(host))
			}
		}
	} else {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "node-ip", newNodeIP(host))
	}

	if m := host.Kubelet.SystemReserved; m != nil {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "system-reserved", kubeoneapi.MapStringStringToString(m, "="))
	}

	if m := host.Kubelet.KubeReserved; m != nil {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "kube-reserved", kubeoneapi.MapStringStringToString(m, "="))
	}

	if m := host.Kubelet.EvictionHard; m != nil {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "eviction-hard", kubeoneapi.MapStringStringToString(m, "<"))
	}
	if m := host.Kubelet.MaxPods; m != nil {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "max-pods", strconv.Itoa(int(*m)))
	}

	if m := host.Kubelet.PodPidsLimit; m != nil {
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "pod-max-pids", strconv.Itoa(int(*m)))
	} else {
		// Set default value if PodsPidsLimits is nil
		// in order to pass the 4.2.13 Check in CIS Benchmark 1.8
		kubeletCLIFlags = setAllArgsValue(kubeletCLIFlags, "pod-max-pids", strconv.Itoa(-1))
	}

	return kubeadmv1beta4.NodeRegistrationOptions{
		Name:             host.Hostname,
		Taints:           host.Taints,
		CRISocket:        fmt.Sprintf("unix://%s", s.Cluster.ContainerRuntime.CRISocket()),
		KubeletExtraArgs: kubeletCLIFlags,
	}
}

func setAllArgsValue(args []kubeadmv1beta4.Arg, name, value string) []kubeadmv1beta4.Arg {
	return kubeadmv1beta4.SetArgValues(args, name, value, -1)
}

func stringStringMapToArgs(m map[string]string) []kubeadmv1beta4.Arg {
	args := []kubeadmv1beta4.Arg{}
	for k, v := range m {
		args = append(args, kubeadmv1beta4.Arg{
			Name:  k,
			Value: v,
		})
	}

	return args
}

func argsToMap(args []kubeadmv1beta4.Arg) map[string]string {
	m := map[string]string{}
	for _, arg := range args {
		m[arg.Name] = arg.Value
	}

	return m
}

func newNodeIP(host kubeoneapi.HostConfig) string {
	return defaults(host.PrivateAddress, host.PublicAddress)
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

func defaults(input, defaultValue string) string {
	if input != "" {
		return input
	}

	return defaultValue
}

func mergeFeatureGates(featureGates string, additionalFeatureGates map[string]bool) string {
	fgs := splitFeatureGates(featureGates)

	for k, v := range additionalFeatureGates {
		fgs[k] = v
	}

	return featureGatesToString(fgs)
}

func splitFeatureGates(featureGates string) map[string]bool {
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

	return featureGatesMap
}

func featureGatesToString(featureGates map[string]bool) string {
	featureGatesKeys := sets.List(sets.KeySet(featureGates))

	var featureGatesStr []string
	for _, k := range featureGatesKeys {
		featureGatesStr = append(featureGatesStr, fmt.Sprintf("%s=%t", k, featureGates[k]))
	}

	return strings.Join(featureGatesStr, ",")
}
