package v1beta1

import (
	"fmt"
	"strings"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConfig returns all required configs to init a cluster via a set of v1beta1 configs
func NewConfig(cluster *config.Cluster, host *config.HostConfig) ([]runtime.Object, error) {
	nodeRegistration := kubeadmv1beta1.NodeRegistrationOptions{
		Name:             host.Hostname,
		KubeletExtraArgs: map[string]string{},
	}
	initConfig := &kubeadmv1beta1.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta1",
			Kind:       "InitConfiguration",
		},
		NodeRegistration: nodeRegistration,
	}
	joinConfig := &kubeadmv1beta1.JoinConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1beta1",
			Kind:       "JoinConfiguration",
		},
		NodeRegistration: nodeRegistration,
		ControlPlane: &kubeadmv1beta1.JoinControlPlane{
			LocalAPIEndpoint: kubeadmv1beta1.APIEndpoint{
				AdvertiseAddress: host.PrivateAddress,
			},
		},
	}

	endpoints := []string{}
	for _, host := range cluster.Hosts {
		endpoints = append(endpoints, strings.ToLower(host.PublicAddress))
	}
	endpoints = append(endpoints, strings.ToLower(cluster.APIServer.Address))
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
		ControlPlaneEndpoint: fmt.Sprintf("%s:6443", cluster.APIServer.Address),
		APIServer: kubeadmv1beta1.APIServer{
			ControlPlaneComponent: kubeadmv1beta1.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"endpoint-reconciler-type": "lease",
					"service-node-port-range":  cluster.Network.NodePortRange(),
				},
				ExtraVolumes: []kubeadmv1beta1.HostPathMount{},
			},
			CertSANs: endpoints,
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
		clusterConfig.APIServer.ExtraVolumes = append(clusterConfig.APIServer.ExtraVolumes, cloudConfigVol)
		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = provider
		clusterConfig.ControllerManager.ExtraVolumes = append(clusterConfig.ControllerManager.ExtraVolumes, cloudConfigVol)
		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		clusterConfig.APIServer.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.ControllerManager.ExtraArgs["cloud-config"] = renderedCloudConfig
		initConfig.NodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	return []runtime.Object{initConfig, joinConfig, clusterConfig}, nil
}
