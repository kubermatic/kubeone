package machinecontroller

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	clustertypes "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
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
	if ctx.Clientset == nil {
		return errors.New("kubernetes clientset not initialized")
	}
	if ctx.RESTConfig == nil {
		return errors.New("kubernetes rest config not initialized")
	}

	// Create Cluster-API clientset
	clusterapiClientset, err := clusterclientset.NewForConfig(ctx.RESTConfig)
	if err != nil {
		return err
	}
	clusterapiClient := clusterapiClientset.ClusterV1alpha1()

	// Apply MachineDeployments
	for _, workerset := range ctx.Cluster.Workers {
		deployment, err := createMachineDeployment(ctx.Cluster, workerset)
		if err != nil {
			return err
		}

		err = ensureMachineDeployment(clusterapiClient.MachineDeployments(deployment.Namespace), deployment)
		if err != nil {
			return err
		}
	}

	return nil
}

func createMachineDeployment(cluster *config.Cluster, workerset config.WorkerConfig) (*clusterv1alpha1.MachineDeployment, error) {
	provider := cluster.Provider.Name

	cloudProviderSpec, err := machineSpec(cluster, workerset, provider)
	if err != nil {
		return nil, err
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
		return nil, err
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

func ensureMachineDeployment(machineDeploymentClient clustertypes.MachineDeploymentInterface, required *clusterv1alpha1.MachineDeployment) error {
	existing, err := machineDeploymentClient.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = machineDeploymentClient.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = machineDeploymentClient.Update(existing)
	return err
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
			return nil, fmt.Errorf("could not parse tags for worker machines: %v", err)
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
