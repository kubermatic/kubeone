/*
Copyright 2020 The KubeOne Authors.

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

package features

import (
	"context"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/templates/kubeadm/kubeadmargs"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podNodeSelectorAdmissionPlugin = "PodNodeSelector"
	nodeSelectorAnnotation         = "scheduler.alpha.kubernetes.io/node-selector"
)

func activateKubeadmPodNodeSelector(feature *kubeoneapi.PodNodeSelector, args *kubeadmargs.Args) {
	if feature == nil || !feature.Enable {
		return
	}

	args.APIServer.AppendMapStringStringExtraArg(apiServerAdmissionPluginsFlag, podNodeSelectorAdmissionPlugin)
	args.APIServer.ExtraArgs[apiServerAdmissionControlConfigFlag] = apiServerAdmissionControlConfigPath
}

// installPodNodeSelector annotates the kube-system namespace and deletes all
// pending pods in the kube-system namespace
func installPodNodeSelector(ctx context.Context, c client.Client, feature *kubeoneapi.PodNodeSelector) error {
	if feature == nil || !feature.Enable {
		return nil
	}

	if err := annotateKubeSystemNamespace(ctx, c); err != nil {
		return err
	}

	return deletePendingPods(ctx, c)
}

// annotateKubeSystemNamespace adds the scheduler.alpha.kubernetes.io/node-selector: ""
// annotation to the kube-system namespace. This ensures that critical pods, such
// as CNI and kube-proxy, can get scheduled on all nodes in the cluster.
func annotateKubeSystemNamespace(ctx context.Context, c client.Client) error {
	ns := corev1.Namespace{}
	key := client.ObjectKey{
		Name: metav1.NamespaceSystem,
	}

	if err := c.Get(ctx, key, &ns); err != nil {
		return fail.KubeClient(err, "getting %T %s", ns, key)
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Annotations[nodeSelectorAnnotation] = ""

	return clientutil.CreateOrUpdate(ctx, c, &ns)
}

// deletePendingPods polls for pending pods in the kube-system namespace
// and deletes them. The annotation has effect only for the newly-created pods,
// so pods created before annotating the namespace might have incorrect node
// selectors.
func deletePendingPods(ctx context.Context, c client.Client) error {
	podList := &corev1.PodList{}
	if err := c.List(ctx, podList, client.InNamespace(metav1.NamespaceSystem)); err != nil {
		return fail.KubeClient(err, "listing kube-system pods")
	}

	errs := []error{}
	for i, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodPending {
			if delErr := c.Delete(ctx, &podList.Items[i]); delErr != nil {
				errs = append(errs, delErr)
			}
		}
	}

	return fail.KubeClient(utilerrors.NewAggregate(errs), "deleting pending kube-system pods")
}
