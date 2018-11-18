package templates

import (
	"encoding/json"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type providerConfig struct {
	CloudProvider       string      `json:"cloudProvider"`
	CloudProviderSpec   interface{} `json:"cloudProviderSpec"`
	OperatingSystem     string      `json:"operatingSystem"`
	OperatingSystemSpec interface{} `json:"operatingSystemSpec"`
}

func MachineConfigurations(cluster *config.Cluster) (string, error) {
	machines := make([]interface{}, 0)

	for _, workerset := range cluster.Workers {
		for i := 0; i < workerset.Replicas; i++ {
			machine, err := createMachine(cluster, workerset, i)
			if err != nil {
				return "", err
			}

			machines = append(machines, machine)
		}
	}

	return kubernetesToYAML(machines)
}

func createMachine(cluster *config.Cluster, workerset config.WorkerConfig, index int) (*clusterv1alpha1.Machine, error) {
	provider := workerset.Provider
	if len(provider) == 0 {
		provider = cluster.Provider.Name
	}

	config := providerConfig{
		CloudProvider:       provider,
		CloudProviderSpec:   workerset.Spec,
		OperatingSystem:     workerset.OperatingSystem.Name,
		OperatingSystemSpec: workerset.OperatingSystem.Spec,
	}

	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s-%d-", workerset.Name, index)
	name := fmt.Sprintf("%s%s", prefix, rand.String(10))

	return &clusterv1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.k8s.io/v1alpha1",
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    metav1.NamespaceSystem,
			Name:         name,
			GenerateName: prefix,
		},
		Spec: clusterv1alpha1.MachineSpec{
			Versions: clusterv1alpha1.MachineVersionInfo{
				Kubelet: cluster.Versions.Kubernetes,
			},
			ProviderConfig: clusterv1alpha1.ProviderConfig{
				Value: &runtime.RawExtension{Raw: encoded},
			},
		},
	}, nil
}
