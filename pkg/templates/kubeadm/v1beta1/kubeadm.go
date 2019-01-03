package v1beta1

import (
	"fmt"

	kubeadmv1beta1 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1beta1"
	"github.com/kubermatic/kubeone/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"

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
		endpoints = append(endpoints, host.PublicAddress)
	}
	endpoints = append(endpoints, cluster.APIServer.Address)
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
			},
			CertSANs: endpoints,
		},
		ControllerManager: kubeadmv1beta1.ControlPlaneComponent{
			ExtraArgs: map[string]string{},
		},
		ClusterName: cluster.Name,
	}
	if cluster.Provider.Name != "" {
		provider := string(cluster.Provider.Name)
		clusterConfig.APIServer.ExtraArgs["cloud-provider"] = provider
		clusterConfig.ControllerManager.ExtraArgs["cloud-provider"] = provider
		nodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
		renderedCloudConfig := "/etc/kubernetes/cloud-config"
		clusterConfig.APIServer.ExtraArgs["cloud-config"] = renderedCloudConfig
		clusterConfig.ControllerManager.ExtraArgs["cloud-config"] = renderedCloudConfig
		initConfig.NodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	return []runtime.Object{initConfig, joinConfig, clusterConfig}, nil
}
