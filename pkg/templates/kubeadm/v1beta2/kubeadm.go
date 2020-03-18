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

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	kubeadmv1beta2 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta2"
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/kubeflags"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm/kubeadmargs"
	"github.com/kubermatic/kubeone/pkg/templates/nodelocaldns"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	bootstrapTokenTTL = 60 * time.Minute
)

// NewConfig returns all required configs to init a cluster via a set of v1beta2 configs
func NewConfig(s *state.State, host kubeoneapi.HostConfig, isWorker bool) ([]runtime.Object, error) {
	cluster := s.Cluster
	kubeSemVer, err := semver.NewVersion(cluster.Versions.Kubernetes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse generate config, wrong kubernetes version %s", cluster.Versions.Kubernetes)
	}

	nodeIP := host.PrivateAddress
	if nodeIP == "" {
		nodeIP = host.PublicAddress
	}

	taints := []corev1.Taint{
		{
			Effect: corev1.TaintEffectNoSchedule,
			Key:    "node-role.kubernetes.io/master",
		},
	}
	if host.Untaint || isWorker {
		taints = nil
	}

	nodeRegistration := kubeadmv1beta2.NodeRegistrationOptions{
		Name:   host.Hostname,
		Taints: taints,
		KubeletExtraArgs: map[string]string{
			"anonymous-auth":      "false",
			"node-ip":             nodeIP,
			"read-only-port":      "0",
			"rotate-certificates": "true",
			"cluster-dns":         nodelocaldns.VirtualIP,
		},
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
			AdvertiseAddress: nodeIP,
		},
	}

	joinConfig := &kubeadmv1beta2.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta2.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta2.APIEndpoint{
				AdvertiseAddress: nodeIP,
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
			ExtraArgs:    map[string]string{},
			ExtraVolumes: []kubeadmv1beta2.HostPathMount{},
		},
		ClusterName: cluster.Name,
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
		provider := string(cluster.CloudProvider.Name)

		clusterConfig.APIServer.ExtraArgs["cloud-provider"] = provider
		clusterConfig.APIServer.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, cloudConfigVol)

		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = provider
		clusterConfig.ControllerManager.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.ControllerManager.ExtraArgs["cluster-name"] = s.Cluster.Name
		clusterConfig.ControllerManager.ExtraVolumes = append(clusterConfig.ControllerManager.ExtraVolumes, cloudConfigVol)

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig

		switch cluster.CloudProvider.Name {
		case kubeoneapi.CloudProviderNameAzure:
			clusterConfig.ControllerManager.ExtraArgs["configure-cloud-routes"] = "false"
		case kubeoneapi.CloudProviderNameAWS:
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

	args := kubeadmargs.NewFrom(clusterConfig.APIServer.ExtraArgs)
	features.UpdateKubeadmClusterConfiguration(cluster.Features, args)

	clusterConfig.APIServer.ExtraArgs = args.APIServer.ExtraArgs
	clusterConfig.FeatureGates = args.FeatureGates

	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration
	if isWorker {
		joinConfig.ControlPlane = nil
		return []runtime.Object{joinConfig}, nil
	}
	return []runtime.Object{initConfig, joinConfig, clusterConfig}, nil
}
