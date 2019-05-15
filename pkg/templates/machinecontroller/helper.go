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

	"github.com/kubermatic/kubeone/pkg/util"

	corev1 "k8s.io/api/core/v1"
	errorsutil "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func simpleCreateOrUpdate(ctx context.Context, client dynclient.Client, obj runtime.Object) error {
	okFunc := func(runtime.Object) error { return nil }
	_, err := controllerutil.CreateOrUpdate(ctx, client, obj, okFunc)
	return err
}

// Ensure install/update machine-controller
func Ensure(ctx *util.Context) error {
	if !ctx.Cluster.MachineController.Deploy {
		ctx.Logger.Info("Skipping machine-controller deployment because it was disabled in configuration.")
		return nil
	}

	ctx.Logger.Infoln("Installing machine-controller…")
	if err := Deploy(ctx); err != nil {
		return errors.Wrap(err, "failed to deploy machine-controller")
	}

	ctx.Logger.Infoln("Installing machine-controller webhooks…")
	if err := DeployWebhookConfiguration(ctx); err != nil {
		return errors.Wrap(err, "failed to deploy machine-controller webhook configuration")
	}

	return nil
}

// WaitReady waits for machine-controller and its webhook to became ready
func WaitReady(ctx *util.Context) error {
	if !ctx.Cluster.MachineController.Deploy {
		return nil
	}

	ctx.Logger.Infoln("Waiting for machine-controller to come up…")

	// Wait a bit to let scheduler to react
	time.Sleep(10 * time.Second)

	if err := WaitForWebhook(ctx.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller-webhook did not come up")
	}

	if err := WaitForMachineController(ctx.DynamicClient); err != nil {
		return errors.Wrap(err, "machine-controller did not come up")
	}
	return nil
}

// DestroyWorkers destroys all MachineDeployment, MachineSet and Machine objects
func DestroyWorkers(ctx *util.Context) error {
	if !ctx.Cluster.MachineController.Deploy {
		ctx.Logger.Info("Skipping deleting workers because machine-controller is disabled in configuration.")
		return nil
	}
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgCtx := context.Background()

	// Annotate nodes with kubermatic.io/skip-eviction=true to skip eviction
	ctx.Logger.Info("Annotating nodes to skip eviction…")
	nodes := &corev1.NodeList{}
	if err := ctx.DynamicClient.List(bgCtx, &dynclient.ListOptions{}, nodes); err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}
	for _, node := range nodes.Items {
		nodeKey := dynclient.ObjectKey{Name: node.Name}

		retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			n := corev1.Node{}
			if err := ctx.DynamicClient.Get(bgCtx, nodeKey, &n); err != nil {
				return err
			}

			if n.Annotations == nil {
				n.Annotations = map[string]string{}
			}
			n.Annotations["kubermatic.io/skip-eviction"] = "true"
			return ctx.DynamicClient.Update(bgCtx, &n)
		})

		if retErr != nil {
			return errors.Wrapf(retErr, "unable to annotate node %s", node.Name)
		}
	}

	// Delete all MachineDeployment objects
	ctx.Logger.Info("Deleting MachineDeployment objects…")
	mdList := &clusterv1alpha1.MachineDeploymentList{}
	if err := ctx.DynamicClient.List(bgCtx, dynclient.InNamespace(MachineControllerNamespace), mdList); err != nil {
		if errorsutil.IsTimeout(err) || errorsutil.IsServerTimeout(err) {
			return errors.Wrap(err, "unable to list machinedeployment objects")
		}
		ctx.Logger.Info("Skipping deleting worker nodes because MachineDeployments CRD is not deployed")
		return nil
	}
	for i := range mdList.Items {
		if err := ctx.DynamicClient.Delete(bgCtx, &mdList.Items[i]); err != nil {
			return errors.Wrapf(err, "unable to delete machinedeployment object %s", mdList.Items[i].Name)
		}
	}

	// Delete all MachineSet objects
	ctx.Logger.Info("Deleting MachineSet objects…")
	msList := &clusterv1alpha1.MachineSetList{}
	if err := ctx.DynamicClient.List(bgCtx, dynclient.InNamespace(MachineControllerNamespace), msList); err != nil {
		return errors.Wrap(err, "unable to list machineset objects")
	}
	for i := range msList.Items {
		if err := ctx.DynamicClient.Delete(bgCtx, &msList.Items[i]); err != nil {
			return errors.Wrapf(err, "unable to delete machineset object %s", msList.Items[i].Name)
		}
	}

	// Delete all Machine objects
	ctx.Logger.Info("Deleting Machine objects…")
	mList := &clusterv1alpha1.MachineList{}
	if err := ctx.DynamicClient.List(bgCtx, dynclient.InNamespace(MachineControllerNamespace), mList); err != nil {
		return errors.Wrap(err, "unable to list machine objects")
	}
	for i := range mList.Items {
		if err := ctx.DynamicClient.Delete(bgCtx, &mList.Items[i]); err != nil {
			return errors.Wrapf(err, "unable to delete machine object %s", mList.Items[i].Name)
		}
	}

	// Wait for all Machines to be deleted
	ctx.Logger.Info("Waiting for all machines to get deleted…")
	return wait.Poll(5*time.Second, 3*time.Minute, func() (bool, error) {
		list := &clusterv1alpha1.MachineList{}
		if err := ctx.DynamicClient.List(bgCtx, dynclient.InNamespace(MachineControllerNamespace), list); err != nil {
			return false, errors.Wrap(err, "unable to list machine objects")
		}
		if len(list.Items) != 0 {
			return false, nil
		}
		return true, nil
	})
}
