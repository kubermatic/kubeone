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

package installation

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchCoreDNS(ctx *util.Context) error {
	if !ctx.Cluster.Provider.External {
		return nil
	}

	ctx.Logger.Infoln("Patching coreDNS with uninitialized toleration…")

	if ctx.DynamicClient == nil {
		return errors.New("kubernetes client not initialized")
	}

	bgCtx := context.Background()
	dep := &appsv1.Deployment{}
	key := client.ObjectKey{
		Name:      "coredns",
		Namespace: metav1.NamespaceSystem,
	}

	err := ctx.DynamicClient.Get(bgCtx, key, dep)
	if err != nil {
		return errors.Wrap(err, "failed to get coredns deployment")
	}

	dep.Spec.Template.Spec.Tolerations = append(dep.Spec.Template.Spec.Tolerations,
		corev1.Toleration{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: corev1.TaintEffectNoSchedule,
		},
	)

	return errors.Wrap(ctx.DynamicClient.Update(bgCtx, dep), "failed to update coredns deployment")
}
