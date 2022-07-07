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
	"bytes"
	"io"
	"path/filepath"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

func restartKubeAPIServer(s *state.State) error {
	s.Logger.Infoln("Restarting unhealthy API servers if needed...")

	return s.RunTaskOnControlPlane(func(s *state.State, node *kubeoneapi.HostConfig, _ executor.Interface) error {
		return restartKubeAPIServerOnOS(s, *node)
	}, state.RunSequentially)
}

func ensureRestartKubeAPIServer(s *state.State) error {
	s.Logger.Infoln("Restarting API servers...")

	return s.RunTaskOnControlPlane(func(s *state.State, node *kubeoneapi.HostConfig, _ executor.Interface) error {
		return ensureRestartKubeAPIServerOnOS(s, *node)
	}, state.RunSequentially)
}

func restartKubeAPIServerOnOS(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameAmazon:     restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameCentOS:     restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameDebian:     restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameFlatcar:    restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameRHEL:       restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameRockyLinux: restartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameUbuntu:     restartKubeAPIServerCrictl,
	})
}

func ensureRestartKubeAPIServerOnOS(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameAmazon:     ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameCentOS:     ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameDebian:     ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameFlatcar:    ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameRHEL:       ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameRockyLinux: ensureRestartKubeAPIServerCrictl,
		kubeoneapi.OperatingSystemNameUbuntu:     ensureRestartKubeAPIServerCrictl,
	})
}

func restartKubeAPIServerCrictl(s *state.State) error {
	cmd, err := scripts.RestartKubeAPIServerCrictl(false)
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "restarting kubeapi-server pod")
}

func ensureRestartKubeAPIServerCrictl(s *state.State) error {
	cmd, err := scripts.RestartKubeAPIServerCrictl(true)
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)

	return fail.SSH(err, "restarting kubeapi-server pod")
}

func labelNodes(s *state.State) error {
	candidateNodes := sets.NewString()
	nodeList := corev1.NodeList{}

	if err := s.DynamicClient.List(s.Context, &nodeList); err != nil {
		return fail.KubeClient(err, "getting %T", nodeList)
	}

	for _, node := range nodeList.Items {
		candidateNodes.Insert(node.Name)
		for _, addr := range node.Status.Addresses {
			candidateNodes.Insert(addr.Address)
		}
	}

	hostsSet := map[string]kubeoneapi.HostConfig{}

	for _, host := range append(s.Cluster.ControlPlane.Hosts, s.Cluster.StaticWorkers.Hosts...) {
		if candidateNodes.Has(host.Hostname) || candidateNodes.Has(host.PrivateAddress) || candidateNodes.Has(host.PublicAddress) {
			hostsSet[host.Hostname] = host
		}
	}

	for nodeName, host := range hostsSet {
		nodeName := nodeName
		host := host
		updateErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var node corev1.Node

			if err := s.DynamicClient.Get(s.Context, types.NamespacedName{Name: nodeName}, &node); err != nil {
				return err
			}

			node.Labels["v1.kubeone.io/operating-system"] = string(host.OperatingSystem)

			for labKey, labVal := range host.Labels {
				if strings.HasSuffix(labKey, "-") {
					// drop minus from the suffix
					labelToDelete := labKey[:len(labKey)-1]
					delete(node.Labels, labelToDelete)
				} else {
					node.Labels[labKey] = labVal
				}
			}

			return s.DynamicClient.Update(s.Context, &node)
		})

		if updateErr != nil {
			return fail.KubeClient(updateErr, "updating %s Node", nodeName)
		}
	}

	return nil
}

func patchStaticPods(s *state.State) error {
	return s.RunTaskOnControlPlane(func(ctx *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
		s.Logger.Infoln("Patching static pods...")

		sshfs := ctx.Runner.NewFS()
		f, err := sshfs.Open("/etc/kubernetes/manifests/kube-controller-manager.yaml")
		if err != nil {
			return err
		}
		defer f.Close()
		mgrPodManifest, _ := f.(executor.ExtendedFile)

		kubeManagerBuf, err := io.ReadAll(mgrPodManifest)
		if err != nil {
			return err
		}

		pod := corev1.Pod{}
		if err = yaml.Unmarshal(kubeManagerBuf, &pod); err != nil {
			return fail.Runtime(err, "unmarshalling kube-controller-manager.yaml")
		}

		cacertDir := cabundle.OriginalCertsDir
		if s.Cluster.CABundle != "" {
			cacertDir = cabundle.CustomCertsDir
		}

		for idx := range pod.Spec.Volumes {
			volume := pod.Spec.Volumes[idx]
			if volume.Name == "ca-certs" {
				volume.HostPath.Path = cacertDir
			}
		}

		foundEnvVar := false
		envVar := corev1.EnvVar{
			Name:  cabundle.SSLCertFileENV,
			Value: filepath.Join("/etc/ssl/certs", cabundle.FileName),
		}

		for idx := range pod.Spec.Containers[0].Env {
			env := pod.Spec.Containers[0].Env[idx]
			if env.Name == envVar.Name {
				env.Value = envVar.Value
				foundEnvVar = true
			}
		}

		if !foundEnvVar {
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, envVar)
		}

		buf, err := yaml.Marshal(&pod)
		if err != nil {
			return fail.Runtime(err, "marshalling kube-controller-manager.yaml")
		}

		if err = mgrPodManifest.Truncate(0); err != nil {
			return err
		}

		if _, err = mgrPodManifest.Seek(0, io.SeekStart); err != nil {
			return err
		}

		_, err = io.Copy(mgrPodManifest, bytes.NewBuffer(buf))

		return fail.Runtime(err, "writing kube-controller-manager.yaml")
	}, state.RunParallel)
}
