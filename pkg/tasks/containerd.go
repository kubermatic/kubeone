/*
Copyright 2021 The KubeOne Authors.

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
	"errors"
	"time"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	kubeadmCRISocket = "kubeadm.alpha.kubernetes.io/cri-socket"
)

var (
	containerdKubeletFlags = map[string]string{
		"--container-runtime":          "remote",
		"--container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
	}
)

func validateContainerdInConfig(s *state.State) error {
	if s.Cluster.ContainerRuntime.Containerd == nil {
		return errors.New("containerd must be enabled in config")
	}

	return nil
}

func patchCRISocketAnnotation(s *state.State) error {
	var nodes corev1.NodeList

	if err := s.DynamicClient.List(s.Context, &nodes); err != nil {
		return err
	}

	for _, node := range nodes.Items {
		node := node
		if socketPath, found := node.Annotations[kubeadmCRISocket]; found {
			if socketPath != "/var/run/dockershim.sock" {
				continue
			}

			if node.Annotations == nil {
				node.Annotations = map[string]string{}
			}
			node.Annotations[kubeadmCRISocket] = "unix:///run/containerd/containerd.sock"

			if err := s.DynamicClient.Update(s.Context, &node); err != nil {
				return err
			}
		}
	}

	return nil
}

func migrateToContainerd(s *state.State) error {
	return s.RunTaskOnAllNodes(migrateToContainerdTask, state.RunSequentially)
}

func migrateToContainerdTask(s *state.State, node *kubeone.HostConfig, conn ssh.Connection) error {
	s.Logger.Info("Migrating container runtime to containerd")

	err := updateRemoteFile(s, kubeadmEnvFlagsFile, func(content []byte) ([]byte, error) {
		kubeletFlags, err := unmarshalKubeletFlags(content)
		if err != nil {
			return nil, err
		}

		for k, v := range containerdKubeletFlags {
			kubeletFlags[k] = v
		}

		buf := marshalKubeletFlags(kubeletFlags)
		return buf, nil
	})
	if err != nil {
		return err
	}

	generateContainerdConfig := node.OperatingSystem != kubeone.OperatingSystemNameFlatcar
	migrateScript, err := scripts.MigrateToContainerd(s.Cluster.RegistryConfiguration.InsecureRegistryAddress(), generateContainerdConfig)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(migrateScript)
	if err != nil {
		return err
	}

	s.Logger.Infof("Waiting all pods on %q to became Ready...", node.Hostname)
	err = wait.Poll(10*time.Second, 10*time.Minute, func() (bool, error) {
		var podsList corev1.PodList

		if perr := s.DynamicClient.List(s.Context, &podsList); perr != nil {
			return false, err
		}

		for _, pod := range podsList.Items {
			if pod.Spec.NodeName != node.Hostname {
				continue
			}

			if pod.Status.Phase != corev1.PodRunning {
				s.Logger.Debugf("Pod %s/%s is not running", pod.Namespace, pod.Name)
				return false, nil
			}

			for _, podcond := range pod.Status.Conditions {
				if podcond.Type == corev1.PodReady && podcond.Status != corev1.ConditionTrue {
					s.Logger.Debugf("Pod %s/%s is not ready", pod.Namespace, pod.Name)
					return false, nil
				}
			}

			for _, condstatus := range pod.Status.ContainerStatuses {
				if !condstatus.Ready {
					s.Logger.Debugf("Container %s in pod %s/%s is not ready", condstatus.Name, pod.Namespace, pod.Name)
					return false, nil
				}
			}
		}

		s.Logger.Debugf("All pods on %s Node are ready", node.Hostname)
		return true, nil
	})

	return err
}
