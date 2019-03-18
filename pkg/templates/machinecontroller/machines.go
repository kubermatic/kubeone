package machinecontroller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type providerSpec struct {
	SSHPublicKeys       []string            `json:"sshPublicKeys"`
	CloudProvider       config.ProviderName `json:"cloudProvider"`
	CloudProviderSpec   interface{}         `json:"cloudProviderSpec"`
	OperatingSystem     string              `json:"operatingSystem"`
	OperatingSystemSpec interface{}         `json:"operatingSystemSpec"`
}

// DeployMachineDeployments deploys MachineDeployments that create appropriate machines
func DeployMachineDeployments(ctx *util.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes dynamic client in not initialized")
	}

	bgCtx := context.Background()

	// Apply MachineDeployments
	for _, workerset := range ctx.Cluster.Workers {
		machinedeployment, err := createMachineDeployment(ctx.Cluster, workerset)
		if err != nil {
			return errors.Wrap(err, "failed to generate MachineDeployment")
		}

		err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, machinedeployment)
		if err != nil {
			return errors.Wrap(err, "failed to ensure MachineDeployment")
		}
	}

	return nil
}

func createMachineDeployment(cluster *config.Cluster, workerset config.WorkerConfig) (*clusterv1alpha1.MachineDeployment, error) {
	provider := cluster.Provider.Name

	cloudProviderSpec, err := machineSpec(cluster, workerset, provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate machineSpec")
	}

	config := providerSpec{
		CloudProvider:       provider,
		CloudProviderSpec:   cloudProviderSpec,
		OperatingSystem:     workerset.Config.OperatingSystem,
		OperatingSystemSpec: workerset.Config.OperatingSystemSpec,
		SSHPublicKeys:       workerset.Config.SSHPublicKeys,
	}

	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to JSON marshal providerSpec")
	}

	replicas := int32(*workerset.Replicas)
	maxSurge := intstr.FromInt(1)
	maxUnavailable := intstr.FromInt(0)
	minReadySeconds := int32(0)
	workersetNameLabels := map[string]string{
		"workerset": workerset.Name,
	}

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
			Paused:   false,
			Replicas: &replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: workersetNameLabels,
			},
			Strategy: &clusterv1alpha1.MachineDeploymentStrategy{
				Type: clustercommon.RollingUpdateMachineDeploymentStrategyType,
				RollingUpdate: &clusterv1alpha1.MachineRollingUpdateDeployment{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavailable,
				},
			},
			MinReadySeconds: &minReadySeconds,
			Template: clusterv1alpha1.MachineTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: metav1.NamespaceSystem,
					Labels:    labels.Merge(workerset.Config.Labels, workersetNameLabels),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels.Merge(workerset.Config.Labels, workersetNameLabels),
					},
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: cluster.Versions.Kubernetes,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{Raw: encoded},
					},
				},
			},
		},
	}, nil
}

func machineSpec(cluster *config.Cluster, workerset config.WorkerConfig, provider config.ProviderName) (map[string]interface{}, error) {
	var err error

	spec := workerset.Config.CloudProviderSpec
	if spec == nil {
		return nil, errors.New("could't find cloudProviderSpec")
	}

	// We only need this tag for AWS because it is used to coordinate nodes in ASG
	if provider == config.ProviderNameAWS {
		tagName := fmt.Sprintf("kubernetes.io/cluster/%s", cluster.Name)
		tagValue := "shared"
		spec, err = addMapTag(spec, tagName, tagValue)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse tags for worker machines")
		}
	}

	return spec, nil
}

func addMapTag(spec map[string]interface{}, tagName string, tagValue string) (map[string]interface{}, error) {
	tags, ok := spec["tags"]
	if !ok {
		tags = make(map[string]string)
	}

	tagMap, ok := tags.(map[string]string)
	if !ok {
		return nil, errors.New("tags must be a map string->string")
	}

	tagMap[tagName] = tagValue
	spec["tags"] = tagMap

	return spec, nil
}
