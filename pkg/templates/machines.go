package templates

import (
	"encoding/json"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type providerConfig struct {
	CloudProvider       string      `json:"cloudProvider"`
	CloudProviderSpec   interface{} `json:"cloudProviderSpec"`
	OperatingSystem     string      `json:"operatingSystem"`
	OperatingSystemSpec interface{} `json:"operatingSystemSpec"`
}

func MachineConfigurations(cluster *config.Cluster) (string, error) {
	deployments := make([]interface{}, 0)

	for _, workerset := range cluster.Workers {
		deployment, err := createMachineDeployment(cluster, workerset)
		if err != nil {
			return "", err
		}

		deployments = append(deployments, deployment)
	}

	return kubernetesToYAML(deployments)
}

func createMachineDeployment(cluster *config.Cluster, workerset config.WorkerConfig) (*clusterv1alpha1.MachineDeployment, error) {
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

	replicas := int32(workerset.Replicas)

	return &clusterv1alpha1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.k8s.io/v1alpha1",
			Kind:       "MachineDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceSystem,
			Name:      fmt.Sprintf("%s-deployment", workerset.Name),
		},
		Spec: clusterv1alpha1.MachineDeploymentSpec{
			Replicas: &replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"workerset": workerset.Name,
				},
			},
			Strategy: &clusterv1alpha1.MachineDeploymentStrategy{
				Type: clustercommon.RollingUpdateMachineDeploymentStrategyType,
			},
			Template: clusterv1alpha1.MachineTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: metav1.NamespaceSystem,
					Labels: map[string]string{
						"workerset": workerset.Name,
					},
				},
				Spec: clusterv1alpha1.MachineSpec{
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: cluster.Versions.Kubernetes,
					},
					ProviderConfig: clusterv1alpha1.ProviderConfig{
						Value: &runtime.RawExtension{Raw: encoded},
					},
				},
			},
		},
	}, nil
}
