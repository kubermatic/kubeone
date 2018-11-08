// Package v1alpha2 is used to bootstrap Kubernetes 1.11.
// This package mimics upstream kubeadm from
// cmd/kubeadm/app/apis/kubeadm/v1alpha2/types.go.
package v1alpha2

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

type api struct {
	AdvertiseAddress     string `yaml:"advertiseAddress"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint"`
}

type localEtcd struct {
	ServerCertSANs []string `yaml:"serverCertSANs"`
	PeerCertSANs   []string `yaml:"peerCertSANs"`
}

type externalEtcd struct {
	Endpoints []string `yaml:"endpoints"`
	CAFile    string   `yaml:"caFile"`
	CertFile  string   `yaml:"certFile"`
	KeyFile   string   `yaml:"keyFile"`
}

type etcd struct {
	Local    *localEtcd    `json:"local,omitempty"`
	External *externalEtcd `json:"external,omitempty"`
}

type networking struct {
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

type configuration struct {
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

func NewConfig(manifest *manifest.Manifest) (*configuration, error) {
	firstMaster := manifest.Hosts[0]
	etcdEndpoints := make([]string, 0)
	etcdSANs := make([]string, 0)
	apiServerCertSANs := make([]string, 0)

	for _, node := range manifest.Hosts {
		etcdEndpoints = append(etcdEndpoints, node.EtcdURL())
		etcdSANs = append(etcdSANs, node.PrivateAddress)

		// TODO: add loadbalancers
		apiServerCertSANs = append(apiServerCertSANs, node.PrivateAddress, node.PublicAddress)
	}

	cfg := &configuration{
		APIVersion:        "kubeadm.k8s.io/v1alpha2",
		Kind:              "MasterConfiguration",
		KubernetesVersion: fmt.Sprintf("v%s", manifest.Versions.Kubernetes),

		API: api{
			AdvertiseAddress:     firstMaster.PrivateAddress,
			ControlPlaneEndpoint: firstMaster.PublicAddress,
		},

		Etcd: etcd{
			Local: &localEtcd{
				ServerCertSANs: etcdSANs,
				PeerCertSANs:   etcdSANs,
			},
			// External: &externalEtcd{
			// 	CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
			// 	CertFile:  "/etc/kubernetes/pki/etcd/peer.crt",
			// 	KeyFile:   "/etc/kubernetes/pki/etcd/peer.key",
			// 	Endpoints: etcdEndpoints,
			// },
		},

		Networking: networking{
			PodSubnet:     manifest.Network.PodSubnet,
			ServiceSubnet: manifest.Network.ServiceSubnet,
		},

		APIServerCertSANs: apiServerCertSANs,
		APIServerExtraArgs: map[string]string{
			"endpoint-reconciler-type": "lease",
			"service-node-port-range":  manifest.Network.NodePortRange,
		},
	}

	if manifest.Provider.CloudConfig != "" {
		renderedCloudConfig := "/etc/kubernetes/cloud-config"

		cfg.APIServerExtraArgs["cloud-config"] = renderedCloudConfig
		cfg.APIServerExtraArgs["cloud-provider"] = manifest.Provider.Name

		cfg.ControllerManagerExtraArgs = map[string]string{
			"cloud-provider": manifest.Provider.Name,
			"cloud-config":   renderedCloudConfig,
		}
	}

	return cfg, nil
}
