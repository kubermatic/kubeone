/*
Copyright 2026 The KubeOne Authors.

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
	"fmt"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonescheme "k8c.io/kubeone/pkg/apis/kubeone/scheme"
	kubeonev1beta3 "k8c.io/kubeone/pkg/apis/kubeone/v1beta3"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	// register cloud providers for control plane provisioning
	_ "k8c.io/kubeone/pkg/cloudprovider/gce"
	_ "k8c.io/kubeone/pkg/cloudprovider/hetzner"
	_ "k8c.io/kubeone/pkg/cloudprovider/kubevirt"
	_ "k8c.io/kubeone/pkg/cloudprovider/openstack"
)

func WithFindControlPlane(t Tasks) Tasks {
	for _, p := range cloudprovider.ControlPlaneProviders() {
		if lb, ok := p.(cloudprovider.LoadBalancerProvider); ok {
			t = t.append(Task{
				Description: fmt.Sprintf("Find %s load balancer", p.Name()),
				Predicate:   lb.HasLoadBalancer,
				Fn:          lb.LookupLoadBalancer,
			})
		}
		t = t.append(Task{
			Description: fmt.Sprintf("Find %s VMs", p.Name()),
			Predicate:   p.Enabled,
			Fn:          p.LookupVMs,
		})
	}

	return t.append(
		Task{
			Operation: "defaulting cluster hosts",
			Predicate: func(s *state.State) bool { return len(s.Cluster.ControlPlane.NodeSets) != 0 },
			Fn:        defaultCluster,
		},
	).append(
		WithHostnameOS(nil)...,
	)
}

func WithEnsureControlPlane(steps Tasks, cluster *kubeoneapi.KubeOneCluster) (Tasks, error) {
	for _, p := range cloudprovider.ControlPlaneProviders() {
		if !p.MatchesConfig(cluster) {
			continue
		}

		machines, err := p.GenerateMachines(cluster.Name, cluster.ControlPlane.NodeSets, cluster.Versions.Kubernetes)
		if err != nil {
			return nil, err
		}

		if lb, ok := p.(cloudprovider.LoadBalancerProvider); ok {
			steps = steps.append(Task{
				Description: fmt.Sprintf("Ensure %s load balancer", p.Name()),
				Predicate:   lb.HasLoadBalancer,
				Fn:          lb.EnsureLoadBalancer,
			})
		}

		for _, machine := range machines {
			m := machine
			steps = steps.append(Task{
				Description: fmt.Sprintf("Ensure %s control-plane %q VM", p.Name(), m.Name),
				Predicate:   p.Enabled,
				Fn: func(s *state.State) error {
					return p.EnsureVM(s, m)
				},
			})
		}

		break
	}

	return steps.append(Task{
		Operation: "defaulting cluster hosts",
		Predicate: func(s *state.State) bool { return len(s.Cluster.ControlPlane.NodeSets) != 0 },
		Fn:        defaultCluster,
	}), nil
}

func defaultCluster(st *state.State) error {
	v1beta3Cluster := kubeonev1beta3.NewKubeOneCluster()
	if err := kubeonescheme.Scheme.Convert(st.Cluster, v1beta3Cluster, nil); err != nil {
		return fail.Config(err, fmt.Sprintf("converting internal to %s object", v1beta3Cluster.GroupVersionKind()))
	}

	kubeonescheme.Scheme.Default(v1beta3Cluster)

	if err := kubeonescheme.Scheme.Convert(v1beta3Cluster, st.Cluster, nil); err != nil {
		return fail.Config(err, fmt.Sprintf("converting %s to internal object", v1beta3Cluster.GroupVersionKind()))
	}

	return nil
}
