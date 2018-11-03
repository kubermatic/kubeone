package templates

import (
	"errors"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/kubermatic/kubeone/pkg/manifest"
)

type kubeadmMasterConfigurationAPI struct {
	AdvertiseAddress     string `yaml:"advertiseAddress"`
	ControlPlaneEndpoint string `yaml:"controlPlaneEndpoint"`
}

type kubeadmMasterConfigurationEtcd struct {
	Endpoints      []string `yaml:"endpoints"`
	CAFile         string   `yaml:"caFile"`
	CertFile       string   `yaml:"certFile"`
	KeyFile        string   `yaml:"keyFile"`
	ServerCertSANs []string `yaml:"serverCertSANs"`
	PeerCertSANs   []string `yaml:"peerCertSANs"`
}

type kubeadmMasterConfigurationNetworking struct {
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

type kubeadmMasterConfigurationExtras struct {
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

type kubeadmMasterConfiguration struct {
	APIVersion                 string                               `yaml:"apiVersion"`
	Kind                       string                               `yaml:"kind"`
	CloudProvider              string                               `yaml:"cloudProvider"`
	KubernetesVersion          string                               `yaml:"kubernetesVersion"`
	API                        kubeadmMasterConfigurationAPI        `yaml:"api"`
	Etcd                       kubeadmMasterConfigurationEtcd       `yaml:"etcd"`
	Networking                 kubeadmMasterConfigurationNetworking `yaml:"networking"`
	APIServerCertSANs          []string                             `yaml:"apiServerCertSANs"`
	APIServerExtraArgs         map[string]string                    `yaml:"apiServerExtraArgs"`
	ControllerManagerExtraArgs map[string]string                    `yaml:"controllerManagerExtraArgs"`
}

func KubeadmConfig(manifest *manifest.Manifest) (string, error) {
	masterNodes := manifest.Hosts
	if len(masterNodes) == 0 {
		return "", errors.New("manifest does not contain at least one master node")
	}

	etcdEndpoints := make([]string, 0)
	etcdSANs := make([]string, 0)
	apiServerCertSANs := make([]string, 0)

	for _, node := range masterNodes {
		etcdEndpoints = append(etcdEndpoints, node.EtcdUrl())
		etcdSANs = append(etcdSANs, node.PrivateAddress)

		// TODO: add loadbalancers
		apiServerCertSANs = append(apiServerCertSANs, node.PrivateAddress, node.Address)
	}

	cfg := kubeadmMasterConfiguration{
		APIVersion:        "kubeadm.k8s.io/v1alpha1",
		Kind:              "MasterConfiguration",
		CloudProvider:     manifest.Provider.Name,
		KubernetesVersion: fmt.Sprintf("v%s", manifest.Versions.Kubernetes),

		API: kubeadmMasterConfigurationAPI{
			AdvertiseAddress:     masterNodes[0].PrivateAddress,
			ControlPlaneEndpoint: masterNodes[0].Address,
		},

		Etcd: kubeadmMasterConfigurationEtcd{
			CAFile:         "/etc/kubernetes/pki/etcd/ca.crt",
			CertFile:       "/etc/kubernetes/pki/etcd/peer.crt",
			KeyFile:        "/etc/kubernetes/pki/etcd/peer.key",
			Endpoints:      etcdEndpoints,
			ServerCertSANs: etcdSANs,
			PeerCertSANs:   etcdSANs,
		},

		Networking: kubeadmMasterConfigurationNetworking{
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
		cfg.ControllerManagerExtraArgs = map[string]string{
			"cloud-config": renderedCloudConfig,
		}
	}

	encoded, err := yaml.Marshal(cfg)

	return string(encoded), err
}
