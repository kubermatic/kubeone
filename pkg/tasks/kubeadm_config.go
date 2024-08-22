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
	"fmt"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/kubeadm"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func determinePauseImage(s *state.State) error {
	s.Logger.Infoln("Determining Kubernetes pause image...")

	return s.RunTaskOnLeaderWithMutator(determinePauseImageExecutor, func(original *state.State, tmp *state.State) {
		original.PauseImage = tmp.PauseImage
	})
}

func determinePauseImageExecutor(s *state.State, _ *kubeoneapi.HostConfig, _ executor.Interface) error {
	cmd, err := scripts.KubeadmPauseImageVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	out, _, err := s.Runner.RunRaw(cmd)
	if err != nil {
		return fail.SSH(err, "getting kubeadm PauseImage version")
	}

	s.PauseImage = s.Cluster.RegistryConfiguration.ImageRegistry("registry.k8s.io") + "/pause:" + out

	return nil
}

func generateKubeadm(s *state.State) error {
	s.Logger.Infoln("Generating kubeadm config file...")

	if err := determinePauseImage(s); err != nil {
		return err
	}

	kubeadmProvider, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	for idx := range s.Cluster.ControlPlane.Hosts {
		node := s.Cluster.ControlPlane.Hosts[idx]
		kubeadmConf, err := kubeadmProvider.Config(s, node)
		if err != nil {
			return err
		}

		// This needs further refactoring. The follower control plane nodes should not need the full configuration
		// any circumstances. However, the PR where this is being introduced is already huge, and I don't want
		// to make the reviewer's life miserable. :)
		if node.IsLeader {
			s.Configuration.AddFile(fmt.Sprintf("cfg/control_plane_full_%d.yaml", node.ID), kubeadmConf.ControlPlaneInitConfiguration)
			s.Configuration.AddFile(fmt.Sprintf("cfg/control_plane_%d.yaml", node.ID), kubeadmConf.ControlPlaneInitConfiguration)
		} else {
			s.Configuration.AddFile(fmt.Sprintf("cfg/control_plane_full_%d.yaml", node.ID), kubeadmConf.ControlPlaneInitConfiguration)
			s.Configuration.AddFile(fmt.Sprintf("cfg/control_plane_%d.yaml", node.ID), kubeadmConf.JoinConfiguration)
		}
	}

	for idx := range s.Cluster.StaticWorkers.Hosts {
		node := s.Cluster.StaticWorkers.Hosts[idx]
		kubeadmConf, err := kubeadmProvider.ConfigWorker(s, node)
		if err != nil {
			return err
		}

		s.Configuration.AddFile(fmt.Sprintf("cfg/worker_%d.yaml", node.ID), kubeadmConf.JoinConfiguration)
	}

	return s.RunTaskOnAllNodes(uploadKubeadmToNode, state.RunParallel)
}

func uploadKubeadmToNode(s *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
	return s.Configuration.UploadTo(conn, s.WorkDir)
}

func uploadKubeadmToConfigMaps(s *state.State) error {
	s.Logger.Info("Updating kubeadm ConfigMaps...")

	leader, err := s.Cluster.Leader()
	if err != nil {
		return err
	}

	kubeadmProvider, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	kubeadmConfig, err := kubeadmProvider.Config(s, leader)
	if err != nil {
		return err
	}

	s.Logger.Debug("Updating kube-system/kubeadm-config ConfigMap...")
	updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return updateConfigMap(s, "kubeadm-config", metav1.NamespaceSystem, "ClusterConfiguration", kubeadmConfig.ClusterConfiguration)
	})
	if updateErr != nil {
		return fail.Runtime(err, "updating kubeadm ConfigMaps")
	}

	s.Logger.Debug("Updating kube-system/kubelet-config ConfigMap...")
	updateErr = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return updateConfigMap(s, "kubelet-config", metav1.NamespaceSystem, "kubelet", kubeadmConfig.KubeletConfiguration)
	})
	if updateErr != nil {
		return fail.Runtime(err, "updating kubeadm ConfigMaps")
	}

	s.Logger.Debug("Updating kube-system/kube-proxy ConfigMap...")
	updateErr = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return updateConfigMap(s, "kube-proxy", metav1.NamespaceSystem, "config.conf", kubeadmConfig.KubeProxyConfiguration)
	})
	if updateErr != nil {
		return fail.Runtime(err, "updating kubeadm ConfigMaps")
	}

	return nil
}

func updateConfigMap(s *state.State, name, namespace, key, value string) error {
	configMap := corev1.ConfigMap{}
	objKey := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	err := s.DynamicClient.Get(context.Background(), objKey, &configMap)
	if err != nil {
		return fmt.Errorf("updating %s/%s ConfigMap: %w", objKey.Namespace, objKey.Name, err)
	}

	configMap.Data[key] = value

	return s.DynamicClient.Update(s.Context, &configMap)
}
