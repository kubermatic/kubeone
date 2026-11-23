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
	"context"
	"errors"
	"time"

	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kubeProxyAppLabelValue   = "kube-proxy"
	kubeProxyConfigMapKey    = "config.conf"
	kubeProxyRestartTimeout  = 10 * time.Minute
	nftablesMinKernelVersion = ">= 5.13"
)

var nftablesKernelConstraint = semverutil.MustParseConstraint(nftablesMinKernelVersion)

// detectKubeProxyMode returns the mode configured in the live kube-proxy ConfigMap. kube-proxy treats an empty mode
// as iptables, so that's returned in that case. An empty string is returned if the ConfigMap doesn't exist (e.g.
// kube-proxy installation is skipped).
func detectKubeProxyMode(s *state.State) (string, error) {
	var cm corev1.ConfigMap
	if err := s.DynamicClient.Get(s.Context, KubeProxyObjectKey, &cm); err != nil {
		if k8serrors.IsNotFound(err) {
			return "", nil
		}

		return "", fail.KubeClient(err, "getting kube-proxy ConfigMap")
	}

	var kubeProxyConfig struct {
		Mode string `yaml:"mode"`
	}
	if err := yaml.Unmarshal([]byte(cm.Data[kubeProxyConfigMapKey]), &kubeProxyConfig); err != nil {
		return "", fail.Runtime(err, "parsing kube-proxy configuration")
	}

	if kubeProxyConfig.Mode == "" {
		return kubeoneapi.KubeProxyModeIPTables, nil
	}

	return kubeProxyConfig.Mode, nil
}

// kubeProxyModeChanged returns true if the cluster runs kube-proxy in a mode other than the configured one.
func kubeProxyModeChanged(s *state.State) bool {
	desired := s.Cluster.ClusterNetwork.KubeProxy.ProxyMode()
	live := s.LiveCluster.KubeProxyMode

	return desired != "" && live != "" && desired != live
}

// safeguardNFTablesKernel makes sure that all SSH-managed hosts can run the nftables kube-proxy mode.
func safeguardNFTablesKernel(s *state.State) error {
	if s.Cluster.ClusterNetwork.KubeProxy.ProxyMode() != kubeoneapi.KubeProxyModeNFTables {
		return nil
	}

	var hostErrors []error

	for _, host := range append(s.LiveCluster.ControlPlane, s.LiveCluster.StaticWorkers...) {
		kernel := host.KernelRelease.Version
		if kernel == nil || nftablesKernelConstraint.Check(kernel) {
			continue
		}

		hostErrors = append(hostErrors, fail.NewConfigError("checking kernel release version",
			"host %q has kernel version %q, but the nftables kube-proxy mode requires kernel %s; "+
				"upgrade the kernel or set .clusterNetwork.kubeProxy.iptables: {} to use the iptables mode",
			host.Config.Hostname,
			kernel,
			nftablesMinKernelVersion,
		))
	}

	return errors.Join(hostErrors...)
}

// restartKubeProxyPods deletes all kube-proxy pods at once, so all nodes switch to the new kube-proxy mode
// immediately, instead of running mixed modes for the duration of the DaemonSet rollout. On startup, kube-proxy
// removes the leftovers (rules, ipsets, kube-ipvs0 interface) of the modes it's not configured for.
func restartKubeProxyPods(s *state.State) error {
	s.Logger.Infof("Switching kube-proxy mode from %q to %q, restarting all kube-proxy pods...",
		s.LiveCluster.KubeProxyMode,
		s.Cluster.ClusterNetwork.KubeProxy.ProxyMode(),
	)

	selector := []dynclient.ListOption{
		dynclient.InNamespace(KubeProxyObjectKey.Namespace),
		dynclient.MatchingLabels{k8sAppLabel: kubeProxyAppLabelValue},
	}

	var oldPods corev1.PodList
	if err := s.DynamicClient.List(s.Context, &oldPods, selector...); err != nil {
		return fail.KubeClient(err, "listing kube-proxy pods")
	}

	oldPodUIDs := sets.New[types.UID]()
	for _, pod := range oldPods.Items {
		oldPodUIDs.Insert(pod.UID)
	}

	err := s.DynamicClient.DeleteAllOf(s.Context, &corev1.Pod{},
		dynclient.InNamespace(KubeProxyObjectKey.Namespace),
		dynclient.MatchingLabels{k8sAppLabel: kubeProxyAppLabelValue},
	)
	if err != nil {
		return fail.KubeClient(err, "deleting kube-proxy pods")
	}

	s.Logger.Info("Waiting for kube-proxy pods to become ready...")

	podsReady := func(ctx context.Context) (bool, error) {
		var ds appsv1.DaemonSet
		if err := s.DynamicClient.Get(ctx, KubeProxyObjectKey, &ds); err != nil {
			return false, nil //nolint:nilerr // retry on transient errors
		}

		var pods corev1.PodList
		if err := s.DynamicClient.List(ctx, &pods, selector...); err != nil {
			return false, nil //nolint:nilerr // retry on transient errors
		}

		var ready int32
		for _, pod := range pods.Items {
			if oldPodUIDs.Has(pod.UID) {
				// old pod is still terminating
				return false, nil
			}
			if pod.DeletionTimestamp == nil && podIsReady(&pod) {
				ready++
			}
		}

		return ready > 0 && ready >= ds.Status.DesiredNumberScheduled, nil
	}

	if err := wait.PollUntilContextTimeout(s.Context, 5*time.Second, kubeProxyRestartTimeout, true, podsReady); err != nil {
		return fail.KubeClient(err, "waiting for kube-proxy pods to become ready")
	}

	return nil
}

func podIsReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}

	return false
}
