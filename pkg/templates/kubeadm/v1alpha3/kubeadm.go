// Package v1alpha3 is used to bootstrap Kubernetes 1.12.
// This package mimics upstream kubeadm from
// cmd/kubeadm/app/apis/kubeadm/v1alpha3/types.go.
package v1alpha3

import (
	"fmt"
	"strings"

	"github.com/kubermatic/kubeone/pkg/config"
)

type api struct {
	AdvertiseAddress     string `yaml:"advertiseAddress,omitempty"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint,omitempty"`
}

type localEtcd struct {
	ServerCertSANs []string      `yaml:"serverCertSANs,omitempty"`
	PeerCertSANs   []string      `yaml:"peerCertSANs,omitempty"`
	ExtraArgs      etcdExtraArgs `yaml:"extraArgs,omitempty"`
}

type externalEtcd struct {
	Endpoints []string `yaml:"endpoints,omitempty"`
	CAFile    string   `yaml:"caFile,omitempty"`
	CertFile  string   `yaml:"certFile,omitempty"`
	KeyFile   string   `yaml:"keyFile,omitempty"`
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
	PodSubnet     string `yaml:"podSubnet,omitempty"`
	ServiceSubnet string `yaml:"serviceSubnet,omitempty"`
}

// Configuration is a simplified version of v1alpha3 kubeadm config
type Configuration struct {
	APIVersion                 string            `yaml:"apiVersion,omitempty"`
	Kind                       string            `yaml:"kind,omitempty"`
	KubernetesVersion          string            `yaml:"kubernetesVersion,omitempty"`
	API                        api               `yaml:"api,omitempty"`
	ControlPlaneEndpoint       string            `yaml:"controlPlaneEndpoint,omitempty"`
	Etcd                       etcd              `yaml:"etcd,omitempty"`
	Networking                 networking        `yaml:"networking,omitempty"`
	APIServerCertSANs          []string          `yaml:"apiServerCertSANs,omitempty"`
	APIServerExtraArgs         map[string]string `yaml:"apiServerExtraArgs,omitempty"`
	ControllerManagerExtraArgs map[string]string `yaml:"controllerManagerExtraArgs,omitempty"`
	FeatureGates               map[string]bool   `yaml:"featureGates,omitempty"`
}

// NewConfig init new v1alpha3 kubeadm config
func NewConfig(cluster *config.Cluster, instance int) (*Configuration, error) {
	leader := cluster.Leader()
	host := cluster.Hosts[instance]

	etcdSANs := []string{host.PrivateAddress, host.Hostname}
	listenClientURLs := fmt.Sprintf("https://127.0.0.1:2379,https://%s:2379", host.PrivateAddress)
	advertiseClientURLs := fmt.Sprintf("https://%s:2379", host.PrivateAddress)
	listenPeerURLs := fmt.Sprintf("https://%s:2380", host.PrivateAddress)
	initialAdvertisePeerURLs := fmt.Sprintf("https://%s:2380", host.PrivateAddress)

	initialClusterAddresses := []string{}
	for idx, host := range cluster.Hosts {
		if idx > instance {
			break
		}

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
		APIVersion:        "kubeadm.k8s.io/v1alpha3",
		Kind:              "ClusterConfiguration",
		KubernetesVersion: fmt.Sprintf("v%s", cluster.Versions.Kubernetes),
		// TODO: use loadbalancer
		APIServerCertSANs:    []string{leader.PublicAddress},
		ControlPlaneEndpoint: fmt.Sprintf("%s:%d", leader.PublicAddress, 6443),

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
			// "runtime-config":           "admissionregistration.k8s.io/v1beta1",
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
