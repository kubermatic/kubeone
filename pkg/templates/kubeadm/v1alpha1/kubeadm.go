// Package v1alpha1 is used to bootstrap Kubernetes 1.10.
// This package mimics upstream kubeadm from
// cmd/kubeadm/app/apis/kubeadm/v1alpha1/types.go.
package v1alpha1

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
)

type api struct {
	AdvertiseAddress     string `yaml:"advertiseAddress"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint"`
}

type etcd struct {
	Endpoints      []string `yaml:"endpoints"`
	CAFile         string   `yaml:"caFile"`
	CertFile       string   `yaml:"certFile"`
	KeyFile        string   `yaml:"keyFile"`
	ServerCertSANs []string `yaml:"serverCertSANs"`
	PeerCertSANs   []string `yaml:"peerCertSANs"`
}

type networking struct {
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

type configuration struct {
	APIVersion                 string            `yaml:"apiVersion"`
	Kind                       string            `yaml:"kind"`
	CloudProvider              string            `yaml:"cloudProvider"`
	KubernetesVersion          string            `yaml:"kubernetesVersion"`
	API                        api               `yaml:"api"`
	Etcd                       etcd              `yaml:"etcd"`
	Networking                 networking        `yaml:"networking"`
	APIServerCertSANs          []string          `yaml:"apiServerCertSANs"`
	APIServerExtraArgs         map[string]string `yaml:"apiServerExtraArgs"`
	ControllerManagerExtraArgs map[string]string `yaml:"controllerManagerExtraArgs"`
}

func NewConfig(cluster *config.Cluster) (*configuration, error) {
	leader := cluster.Leader()
	etcdEndpoints := make([]string, 0)
	etcdSANs := make([]string, 0)
	apiServerCertSANs := make([]string, 0)

	for _, node := range cluster.Hosts {
		etcdEndpoints = append(etcdEndpoints, node.EtcdURL())
		etcdSANs = append(etcdSANs, node.PrivateAddress)

		// TODO: add loadbalancers
		apiServerCertSANs = append(apiServerCertSANs, node.PrivateAddress, node.PublicAddress)
	}

	cfg := &configuration{
		APIVersion:        "kubeadm.k8s.io/v1alpha1",
		Kind:              "MasterConfiguration",
		CloudProvider:     cluster.Provider.Name,
		KubernetesVersion: fmt.Sprintf("v%s", cluster.Versions.Kubernetes),

		API: api{
			AdvertiseAddress:     leader.PrivateAddress,
			ControlPlaneEndpoint: leader.PublicAddress,
		},

		Etcd: etcd{
			CAFile:         "/etc/kubernetes/pki/etcd/ca.crt",
			CertFile:       "/etc/kubernetes/pki/etcd/peer.crt",
			KeyFile:        "/etc/kubernetes/pki/etcd/peer.key",
			Endpoints:      etcdEndpoints,
			ServerCertSANs: etcdSANs,
			PeerCertSANs:   etcdSANs,
		},

		Networking: networking{
			PodSubnet:     cluster.Network.PodSubnet(),
			ServiceSubnet: cluster.Network.ServiceSubnet(),
		},

		APIServerCertSANs: apiServerCertSANs,
		APIServerExtraArgs: map[string]string{
			"endpoint-reconciler-type": "lease",
			"service-node-port-range":  cluster.Network.NodePortRange(),
		},
	}

	if cluster.Provider.CloudConfig != "" {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		cfg.APIServerExtraArgs["cloud-config"] = renderedCloudConfig
		cfg.ControllerManagerExtraArgs = map[string]string{
			"cloud-config": renderedCloudConfig,
		}
	}

	return cfg, nil
}
