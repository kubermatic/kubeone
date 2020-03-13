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
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	errorsutil "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Ensure install/update machine-controller
func Ensure(s *state.State) error {
	if !s.Cluster.MachineController.Deploy {
		s.Logger.Info("Skipping machine-controller deployment because it was disabled in configuration.")
		return nil
	}

	s.Logger.Infoln("Installing machine-controller…")
	if err := Deploy(s); err != nil {
		return errors.Wrap(err, "failed to deploy machine-controller")
	}

	s.Logger.Infoln("Installing machine-controller webhooks…")
	if err := DeployWebhookConfiguration(s); err != nil {
		return errors.Wrap(err, "failed to deploy machine-controller webhook configuration")
	}

	return nil
}

// WaitReady waits for machine-controller and its webhook to became ready
func WaitReady(s *state.State) error {
	if !s.Cluster.MachineController.Deploy {
		return nil
	}

	s.Logger.Infoln("Waiting for machine-controller to come up…")

	// Wait a bit to let scheduler to react
	time.Sleep(10 * time.Second)

	if err := WaitForWebhook(s.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller-webhook did not come up")
	}

	if err := WaitForMachineController(s.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller did not come up")
	}

	if err := WaitForCRDs(s); err != nil {
		return errors.Wrap(err, "machine-controller CRDs did not come up")
	}
	return nil
}

// WaitForCRDs waits for machine-controller CRDs to be created and become established
func WaitForCRDs(s *state.State) error {
	var lastErr error
	_ = wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		lastErr = VerifyCRDs(s)
		return lastErr == nil, nil
	})
	return errors.Wrap(lastErr, "failed waiting for CRDs to become ready and established")
}

// VerifyCRDs verifies are Cluster-API CRDs deployed
func VerifyCRDs(s *state.State) error {
	bgCtx := context.Background()
	crd := &apiextensions.CustomResourceDefinition{}

	// Verify MachineDeployment CRD
	key := dynclient.ObjectKey{Name: "machinedeployments.cluster.k8s.io"}
	if err := s.DynamicClient.Get(bgCtx, key, crd); err != nil {
		return errors.Wrap(err, "MachineDeployments CRD is not deployed")
	}
	if err := verifyCRDEstablished(crd); err != nil {
		return errors.Wrap(err, "failed checking MachineDeployments CRD status")
	}

	// Verify MachineSet CRD
	key = dynclient.ObjectKey{Name: "machinesets.cluster.k8s.io"}
	if err := s.DynamicClient.Get(bgCtx, key, crd); err != nil {
		return errors.Wrap(err, "MachineSet CRD is not deployed")
	}
	if err := verifyCRDEstablished(crd); err != nil {
		return errors.Wrap(err, "failed checking MachineSets CRD status")
	}

	// Verify Machine CRD
	key = dynclient.ObjectKey{Name: "machines.cluster.k8s.io"}
	if err := s.DynamicClient.Get(bgCtx, key, crd); err != nil {
		return errors.Wrap(err, "MachineSet CRD is not deployed")
	}
	if err := verifyCRDEstablished(crd); err != nil {
		return errors.Wrap(err, "failed checking Machines CRD status")
	}

	return nil
}

func verifyCRDEstablished(crd *apiextensions.CustomResourceDefinition) error {
	for _, cond := range crd.Status.Conditions {
		if cond.Type == apiextensions.Established && cond.Status == apiextensions.ConditionTrue {
			return nil
		}
	}
	return errors.New("crd is not established")
}

// DestroyWorkers destroys all MachineDeployment, MachineSet and Machine objects
func DestroyWorkers(s *state.State) error {
	if !s.Cluster.MachineController.Deploy {
		s.Logger.Info("Skipping deleting workers because machine-controller is disabled in configuration.")
		return nil
	}
	if s.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	ctx := context.Background()

	// Annotate nodes with kubermatic.io/skip-eviction=true to skip eviction
	s.Logger.Info("Annotating nodes to skip eviction…")
	nodes := &corev1.NodeList{}
	if err := s.DynamicClient.List(ctx, nodes); err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}
	for _, node := range nodes.Items {
		nodeKey := dynclient.ObjectKey{Name: node.Name}

		retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			n := corev1.Node{}
			if err := s.DynamicClient.Get(ctx, nodeKey, &n); err != nil {
				return err
			}

			if n.Annotations == nil {
				n.Annotations = map[string]string{}
			}
			n.Annotations["kubermatic.io/skip-eviction"] = "true"
			return s.DynamicClient.Update(ctx, &n)
		})

		if retErr != nil {
			return errors.Wrapf(retErr, "unable to annotate node %s", node.Name)
		}
	}

	// Delete all MachineDeployment objects
	s.Logger.Info("Deleting MachineDeployment objects…")
	mdList := &clusterv1alpha1.MachineDeploymentList{}
	if err := s.DynamicClient.List(ctx, mdList, dynclient.InNamespace(MachineControllerNamespace)); err != nil {
		if !errorsutil.IsNotFound(err) {
			return errors.Wrap(err, "unable to list machinedeployment objects")
		}
	}
	for i := range mdList.Items {
		if err := s.DynamicClient.Delete(ctx, &mdList.Items[i]); err != nil {
			return errors.Wrapf(err, "unable to delete machinedeployment object %s", mdList.Items[i].Name)
		}
	}

	// Delete all MachineSet objects
	s.Logger.Info("Deleting MachineSet objects…")
	msList := &clusterv1alpha1.MachineSetList{}
	if err := s.DynamicClient.List(ctx, msList, dynclient.InNamespace(MachineControllerNamespace)); err != nil {
		if !errorsutil.IsNotFound(err) {
			return errors.Wrap(err, "unable to list machineset objects")
		}
	}
	for i := range msList.Items {
		if err := s.DynamicClient.Delete(ctx, &msList.Items[i]); err != nil {
			if !errorsutil.IsNotFound(err) {
				return errors.Wrapf(err, "unable to delete machineset object %s", msList.Items[i].Name)
			}
		}
	}

	// Delete all Machine objects
	s.Logger.Info("Deleting Machine objects…")
	mList := &clusterv1alpha1.MachineList{}
	if err := s.DynamicClient.List(ctx, mList, dynclient.InNamespace(MachineControllerNamespace)); err != nil {
		if !errorsutil.IsNotFound(err) {
			return errors.Wrap(err, "unable to list machine objects")
		}
	}
	for i := range mList.Items {
		if err := s.DynamicClient.Delete(ctx, &mList.Items[i]); err != nil {
			if !errorsutil.IsNotFound(err) {
				return errors.Wrapf(err, "unable to delete machine object %s", mList.Items[i].Name)
			}
		}
	}

	return nil
}

// WaitDestroy waits for all Machines to be deleted
func WaitDestroy(s *state.State) error {
	s.Logger.Info("Waiting for all machines to get deleted…")

	ctx := context.Background()
	return wait.Poll(5*time.Second, 5*time.Minute, func() (bool, error) {
		list := &clusterv1alpha1.MachineList{}
		if err := s.DynamicClient.List(ctx, list, dynclient.InNamespace(MachineControllerNamespace)); err != nil {
			return false, errors.Wrap(err, "unable to list machine objects")
		}
		if len(list.Items) != 0 {
			return false, nil
		}
		return true, nil
	})
}
