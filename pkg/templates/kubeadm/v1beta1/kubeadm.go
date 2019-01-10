package v1beta1

import (
	"fmt"
	"strings"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
)

// NewConfig returns all required configs to init a cluster via a set of v1beta1 configs
func NewConfig(ctx *util.Context, host *config.HostConfig) ([]runtime.Object, error) {
	cluster := ctx.Cluster

	nodeRegistration := kubeadmv1beta1.NodeRegistrationOptions{
		Name:             host.Hostname,
		KubeletExtraArgs: map[string]string{},
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

	controlPlaneEndpoint := fmt.Sprintf("%s:6443", cluster.APIServer.Address)
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
			PodSubnet:     cluster.Network.PodSubnet(),
			ServiceSubnet: cluster.Network.ServiceSubnet(),
		},
		KubernetesVersion:    cluster.Versions.Kubernetes,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta1.APIServer{
			ControlPlaneComponent: kubeadmv1beta1.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"endpoint-reconciler-type": "lease",
					"service-node-port-range":  cluster.Network.NodePortRange(),
				},
				ExtraVolumes: []kubeadmv1beta1.HostPathMount{},
			},
			CertSANs: []string{strings.ToLower(cluster.APIServer.Address)},
		},
		ControllerManager: kubeadmv1beta1.ControlPlaneComponent{
			ExtraArgs:    map[string]string{},
			ExtraVolumes: []kubeadmv1beta1.HostPathMount{},
		},
		ClusterName: cluster.Name,
	}
	if cluster.Provider.Name != "" {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"
		cloudConfigVol := kubeadmv1beta1.HostPathMount{
			Name:      "cloud-config",
			HostPath:  renderedCloudConfig,
			MountPath: renderedCloudConfig,
			ReadOnly:  true,
			PathType:  corev1.HostPathFile,
		}
		provider := string(cluster.Provider.Name)

		clusterConfig.APIServer.ExtraArgs["cloud-provider"] = provider
		clusterConfig.APIServer.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, cloudConfigVol)

		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = provider
		clusterConfig.ControllerManager.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.ControllerManager.ExtraVolumes = append(clusterConfig.ControllerManager.ExtraVolumes, cloudConfigVol)

		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		nodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}
	initConfig.NodeRegistration = nodeRegistration
	joinConfig.NodeRegistration = nodeRegistration

	return []runtime.Object{initConfig, joinConfig, clusterConfig}, nil
}
