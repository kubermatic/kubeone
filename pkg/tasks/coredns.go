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

package tasks

import (
	"context"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchCoreDNS(s *state.State) error {
	if !s.Cluster.CloudProvider.External {
		return nil
	}

	s.Logger.Infoln("Patching coreDNS with uninitialized toleration...")

	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	ctx := context.Background()
	dep := appsv1.Deployment{}
	key := client.ObjectKey{
		Name:      "coredns",
		Namespace: metav1.NamespaceSystem,
	}

	err := s.DynamicClient.Get(ctx, key, &dep)
	if err != nil {
		return fail.KubeClient(err, "getting %T %s", dep, key)
	}

	dep.Spec.Template.Spec.Tolerations = append(dep.Spec.Template.Spec.Tolerations,
		corev1.Toleration{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: corev1.TaintEffectNoSchedule,
		},
	)

	err = s.DynamicClient.Update(ctx, &dep)

	return fail.KubeClient(err, "updating %T %s", dep, key)
}
