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

package v1beta1

import (
	"fmt"
	"strings"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
)

// NewConfig returns all required configs to init a cluster via a set of v1beta1 configs
func NewConfig(ctx *util.Context, host kubeoneapi.HostConfig) ([]runtime.Object, error) {
	cluster := ctx.Cluster

	nodeRegistration := kubeadmv1beta1.NodeRegistrationOptions{
		Name:             host.Hostname,
		KubeletExtraArgs: map[string]string{"node-ip": host.PrivateAddress},
	}

	if ctx.JoinToken == "" {
		tokenStr, err := bootstraputil.GenerateBootstrapToken()
		if err != nil {
			return nil, err
		}
		ctx.JoinToken = tokenStr
	}

	bootstrapToken, err := kubeadmv1beta1.NewBootstrapTokenString(ctx.JoinToken)
	if err != nil {
		return nil, err
	}

	// TODO(xmudrii): Support more than one API endpoint
	controlPlaneEndpoint := fmt.Sprintf("%s:%d", cluster.APIEndpoints[0].Host, cluster.APIEndpoints[0].Port)
	hostAdvertiseAddress := host.PrivateAddress
	if hostAdvertiseAddress == "" {
		hostAdvertiseAddress = host.PublicAddress
	}

	initConfig := &kubeadmv1beta1.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta1",
			Kind:       "InitConfiguration",
		},
		BootstrapTokens: []kubeadmv1beta1.BootstrapToken{{Token: bootstrapToken}},
		LocalAPIEndpoint: kubeadmv1beta1.APIEndpoint{
			AdvertiseAddress: hostAdvertiseAddress,
		},
	}

	joinConfig := &kubeadmv1beta1.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta1",
			Kind:       "JoinConfiguration",
		},
		ControlPlane: &kubeadmv1beta1.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta1.APIEndpoint{
				AdvertiseAddress: hostAdvertiseAddress,
			},
		},
		Discovery: kubeadmv1beta1.Discovery{
			BootstrapToken: &kubeadmv1beta1.BootstrapTokenDiscovery{
				Token:                    ctx.JoinToken,
				APIServerEndpoint:        controlPlaneEndpoint,
				UnsafeSkipCAVerification: true,
			},
		},
	}

	clusterConfig := &kubeadmv1beta1.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta1",
			Kind:       "ClusterConfiguration",
		},
		Networking: kubeadmv1beta1.Networking{
			PodSubnet:     cluster.ClusterNetwork.PodSubnet,
			ServiceSubnet: cluster.ClusterNetwork.ServiceSubnet,
			DNSDomain:     cluster.ClusterNetwork.ServiceDomainName,
		},
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta1.APIServer{
			ControlPlaneComponent: kubeadmv1beta1.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"endpoint-reconciler-type": "lease",
					"service-node-port-range":  cluster.ClusterNetwork.NodePortRange,
				},
				ExtraVolumes: []kubeadmv1beta1.HostPathMount{},
			},
			// TODO(xmudrii): Support more than one API endpoint
			CertSANs: []string{strings.ToLower(cluster.APIEndpoints[0].Host)},
		},
		ControllerManager: kubeadmv1beta1.ControlPlaneComponent{
			ExtraArgs:    map[string]string{},
			ExtraVolumes: []kubeadmv1beta1.HostPathMount{},
		},
		ClusterName: cluster.Name,
	}

	if cluster.CloudProvider.CloudProviderInTree() {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"
		cloudConfigVol := kubeadmv1beta1.HostPathMount{
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
		clusterConfig.ControllerManager.ExtraVolumes = append(clusterConfig.ControllerManager.ExtraVolumes, cloudConfigVol)

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	if cluster.CloudProvider.External {
		clusterConfig.APIServer.ExtraArgs["cloud-provider"] = ""
		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = ""
		nodeRegistration.KubeletExtraArgs["cloud-provider"] = "external"
	}

	features.UpdateKubeadmClusterConfiguration(cluster.Features, clusterConfig)

	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration

	return []runtime.Object{initConfig, joinConfig, clusterConfig}, nil
}
