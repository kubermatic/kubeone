// Package v1alpha2 is used to bootstrap Kubernetes 1.11.
// This package mimics upstream kubeadm from
// cmd/kubeadm/app/apis/kubeadm/v1alpha2/types.go.
package v1alpha2

import (
	"fmt"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
)

type api struct {
	AdvertiseAddress     string `yaml:"advertiseAddress"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint"`
}

type localEtcd struct {
	ServerCertSANs []string      `yaml:"serverCertSANs"`
	PeerCertSANs   []string      `yaml:"peerCertSANs"`
	ExtraArgs      etcdExtraArgs `yaml:"extraArgs,omitempty"`
}

type externalEtcd struct {
	Endpoints []string `yaml:"endpoints"`
	CAFile    string   `yaml:"caFile"`
	CertFile  string   `yaml:"certFile"`
	KeyFile   string   `yaml:"keyFile"`
}

type etcdExtraArgs struct {
	ListenClientURLs         string `yaml:"listen-client-urls,omitempty"`
	AdvertiseClientURLs      string `yaml:"advertise-client-urls,omitempty"`
	ListenPeerURLs           string `yaml:"listen-peer-urls,omitempty"`
	InitialAdvertisePeerURLs string `yaml:"initial-advertise-peer-urls,omitempty"`
	InitialCluster           string `yaml:"initial-cluster,omitempty"`
	InitialClusterState      string `yaml:"initial-cluster-state,omitempty"`
}

type etcd struct {
	Local    *localEtcd    `json:"local,omitempty"`
	External *externalEtcd `json:"external,omitempty"`
}

type networking struct {
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

// Configuration is a simplified version of v1alpha2 kubeadm config
type Configuration struct {
	APIVersion                 string            `yaml:"apiVersion"`
	Kind                       string            `yaml:"kind"`
	KubernetesVersion          string            `yaml:"kubernetesVersion"`
	API                        api               `yaml:"api"`
	Etcd                       etcd              `yaml:"etcd"`
	Networking                 networking        `yaml:"networking"`
	APIServerCertSANs          []string          `yaml:"apiServerCertSANs"`
	APIServerExtraArgs         map[string]string `yaml:"apiServerExtraArgs"`
	ControllerManagerExtraArgs map[string]string `yaml:"controllerManagerExtraArgs"`
}

// NewConfig init new v1alpha2 kubeadm config
func NewConfig(cluster *config.Cluster, instance int) (*Configuration, error) {
	leader := cluster.Leader()
	host := cluster.Hosts[instance]

	etcdSANs := []string{host.PrivateAddress, host.Hostname}
	listenClientURLs := fmt.Sprintf("https://127.0.0.1:2379,https://%s:2379", host.PrivateAddress)
	advertiseClientURLs := fmt.Sprintf("https://%s:2379", host.PrivateAddress)
	listenPeerURLs := fmt.Sprintf("https://%s:2380", host.PrivateAddress)
	initialAdvertisePeerURLs := fmt.Sprintf("https://%s:2380", host.PrivateAddress)

	initialClusterAddresses := []string{}
	for _, host := range cluster.Hosts {
		initialClusterAddresses = append(
			initialClusterAddresses,
			fmt.Sprintf("%s=https://%s:2380", host.Hostname, host.PrivateAddress),
		)
	}
	initialCluster := strings.Join(initialClusterAddresses, ",")

	initialClusterState := "new"
	if instance > 0 {
		initialClusterState = "existing"
	}

	cfg := &Configuration{
		APIVersion:        "kubeadm.k8s.io/v1alpha2",
		Kind:              "MasterConfiguration",
		KubernetesVersion: fmt.Sprintf("v%s", cluster.Versions.Kubernetes),
		// TODO: use loadbalancer
		APIServerCertSANs: []string{leader.PublicAddress},

		API: api{
			AdvertiseAddress:     leader.PrivateAddress,
			ControlPlaneEndpoint: fmt.Sprintf("%s:%d", leader.PublicAddress, 6443),
		},

		Etcd: etcd{
			Local: &localEtcd{
				ServerCertSANs: etcdSANs,
				PeerCertSANs:   etcdSANs,
				ExtraArgs: etcdExtraArgs{
					ListenClientURLs:         listenClientURLs,
					AdvertiseClientURLs:      advertiseClientURLs,
					ListenPeerURLs:           listenPeerURLs,
					InitialAdvertisePeerURLs: initialAdvertisePeerURLs,
					InitialCluster:           initialCluster,
					InitialClusterState:      initialClusterState,
				},
			},
		},

		Networking: networking{
			PodSubnet:     cluster.Network.PodSubnet(),
			ServiceSubnet: cluster.Network.ServiceSubnet(),
		},

		APIServerExtraArgs: map[string]string{
			"endpoint-reconciler-type": "lease",
			"service-node-port-range":  cluster.Network.NodePortRange(),
		},
	}

	if cluster.Provider.CloudConfig != "" {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		cfg.APIServerExtraArgs["cloud-config"] = renderedCloudConfig
		cfg.APIServerExtraArgs["cloud-provider"] = string(cluster.Provider.Name)

		cfg.ControllerManagerExtraArgs = map[string]string{
			"cloud-provider": string(cluster.Provider.Name),
			"cloud-config":   renderedCloudConfig,
		}
	}

	return cfg, nil
}
