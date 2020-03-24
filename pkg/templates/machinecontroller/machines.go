/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machinecontroller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/state"

	clustercommon "github.com/kubermatic/machine-controller/pkg/apis/cluster/common"
	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeployMachineDeployments deploys MachineDeployments that create appropriate machines
func DeployMachineDeployments(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes dynamic client in not initialized")
	}

	ctx := context.Background()

	// Apply MachineDeployments
	for _, workerset := range s.Cluster.Workers {
		machinedeployment, err := createMachineDeployment(s.Cluster, workerset)
		if err != nil {
			return errors.Wrap(err, "failed to generate MachineDeployment")
		}

		err = clientutil.CreateOrUpdate(ctx, s.DynamicClient, machinedeployment)
		if err != nil {
			return errors.Wrap(err, "failed to ensure MachineDeployment")
		}
	}

	return nil
}

func createMachineDeployment(cluster *kubeoneapi.KubeOneCluster, workerset kubeoneapi.WorkerConfig) (*clusterv1alpha1.MachineDeployment, error) {
	provider := cluster.CloudProvider.Name

	cloudProviderSpec, err := machineSpec(cluster, workerset, provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate machineSpec")
	}

	cloudProviderSpecJSON, err := json.Marshal(cloudProviderSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal cloudProviderSpec to JSON")
	}

	workerset.Config.CloudProviderSpec = cloudProviderSpecJSON

	encoded, err := json.Marshal(struct {
		kubeoneapi.ProviderSpec
		CloudProvider kubeoneapi.CloudProviderName `json:"cloudProvider"`
	}{
		ProviderSpec:  workerset.Config,
		CloudProvider: provider,
	})
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

	if workerset.Config.Network != nil {
		// we have static network config
		maxSurge = intstr.FromInt(0)
		maxUnavailable = intstr.FromInt(1)
	}

	return &clusterv1alpha1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceSystem,
			Name:      workerset.Name,
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
					Taints: workerset.Config.Taints,
				},
			},
		},
	}, nil
}

func machineSpec(cluster *kubeoneapi.KubeOneCluster, workerset kubeoneapi.WorkerConfig, provider kubeoneapi.CloudProviderName) (map[string]interface{}, error) {
	var err error

	specRaw := workerset.Config.CloudProviderSpec
	if specRaw == nil {
		return nil, errors.New("could't find cloudProviderSpec")
	}

	if provider == kubeoneapi.CloudProviderNameAWS {
		var awsSpec AWSSpec

		err = json.Unmarshal(specRaw, &awsSpec)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse AWS Spec for worker machines")
		}

		tagName := fmt.Sprintf("kubernetes.io/cluster/%s", cluster.Name)
		tagValue := "shared"
		if awsSpec.Tags == nil {
			awsSpec.Tags = make(map[string]string)
		}
		awsSpec.Tags[tagName] = tagValue

		// effectively overwrite specRaw retrieved earlier
		specRaw, err = json.Marshal(awsSpec)
		if err != nil {
			return nil, errors.Wrap(err, "could not marshal AWSSpec")
		}
	}

	spec := make(map[string]interface{})
	err = json.Unmarshal(specRaw, &spec)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse the workerset spec")
	}

	return spec, nil
}
