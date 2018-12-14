// Package v1alpha3 is used to bootstrap Kubernetes 1.12.
// This package mimics upstream kubeadm from
// cmd/kubeadm/app/apis/kubeadm/v1alpha3/types.go.
package v1alpha3

import (
	"fmt"

	kubeadmv1alpha3 "github.com/kubermatic/kubeone/pkg/apis/kubeadm/v1alpha3"
	"github.com/kubermatic/kubeone/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConfig init new v1alpha3 kubeadm config
func NewConfig(cluster *config.Cluster, host *config.HostConfig) (*kubeadmv1alpha3.InitConfiguration, *kubeadmv1alpha3.ClusterConfiguration, error) {
	leader, err := cluster.Leader()
	if err != nil {
		return nil, nil, err
	}

	endpoints := make([]string, len(cluster.Hosts))
	for i, host := range cluster.Hosts {
		endpoints[i] = fmt.Sprintf("http://%s:2379", host.PrivateAddress)
	}

	clusterCfg := &kubeadmv1alpha3.ClusterConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1alpha3",
			Kind:       "ClusterConfiguration",
		},
		KubernetesVersion:    fmt.Sprintf("v%s", cluster.Versions.Kubernetes),
		APIServerCertSANs:    []string{leader.PublicAddress},
		ControlPlaneEndpoint: fmt.Sprintf("%s:%d", leader.PublicAddress, 6443),
		Etcd: kubeadmv1alpha3.Etcd{
			External: &kubeadmv1alpha3.ExternalEtcd{
				Endpoints: endpoints,
			},
		},
		Networking: kubeadmv1alpha3.Networking{
			PodSubnet:     cluster.Network.PodSubnet(),
			ServiceSubnet: cluster.Network.ServiceSubnet(),
		},

		APIServerExtraArgs: map[string]string{
			"endpoint-reconciler-type": "lease",
			"service-node-port-range":  cluster.Network.NodePortRange(),
		},

		ControllerManagerExtraArgs: map[string]string{},
	}

	initCfg := &kubeadmv1alpha3.InitConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeadm.k8s.io/v1alpha3",
			Kind:       "InitConfiguration",
		},
		NodeRegistration: kubeadmv1alpha3.NodeRegistrationOptions{
			Name: host.Hostname,
			KubeletExtraArgs: map[string]string{
				"hostname-override": host.Hostname,
			},
		},
	}

	if cluster.Provider.Name != "" {
		provider := string(cluster.Provider.Name)

		clusterCfg.APIServerExtraArgs["cloud-provider"] = provider
		clusterCfg.ControllerManagerExtraArgs["cloud-provider"] = provider
		initCfg.NodeRegistration.KubeletExtraArgs["cloud-provider"] = provider
	}

	if cluster.Provider.CloudConfig != "" {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		clusterCfg.APIServerExtraArgs["cloud-config"] = renderedCloudConfig
		clusterCfg.ControllerManagerExtraArgs["cloud-config"] = renderedCloudConfig
		initCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = renderedCloudConfig
	}

	return initCfg, clusterCfg, nil
}
