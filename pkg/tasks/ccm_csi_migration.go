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
	"fmt"
	"strconv"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/nodeutils"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"

	"github.com/kubermatic/machine-controller/pkg/apis/cluster/common"
	clusterv1alpha1 "github.com/kubermatic/machine-controller/pkg/apis/cluster/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	provisionedByAnnotation            = "pv.kubernetes.io/provisioned-by"
	provisionedByOpenStackInTreeCinder = "kubernetes.io/cinder"
	provisionedByOpenStackCSICinder    = "cinder.csi.openstack.org"
)

func ccmMigrationValidateConfig(s *state.State) error {
	if !s.Cluster.CloudProvider.External {
		return fail.NewConfigError("validation", ".cloudProvider.external must be enabled to start the migration")
	}

	if !s.Cluster.CSIMigrationSupported() {
		return fail.NewConfigError("validation", "ccm/csi migration is not supported for the specified provider")
	}

	if !s.LiveCluster.CCMStatus.InTreeCloudProviderEnabled {
		return fail.NewConfigError("validation", "the cluster is already running external ccm")
	} else if s.LiveCluster.CCMStatus.ExternalCCMDeployed && !s.CCMMigrationComplete {
		return fail.NewConfigError("validation", "the ccm/csi migration is currently in progress, run command with --complete to finish it")
	}

	if s.Cluster.CloudProvider.Vsphere != nil && s.Cluster.CloudProvider.CSIConfig == "" {
		return fail.NewConfigError("validation", "the ccm/csi migration for vsphere requires providing csi configuration using .cloudProvider.csiConfig field")
	}

	return nil
}

func readyToCompleteCCMMigration(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	machines := clusterv1alpha1.MachineList{}
	if err := s.DynamicClient.List(s.Context, &machines); err != nil {
		return fail.KubeClient(err, "getting %T", machines)
	}

	migrated := true
	for i := range machines.Items {
		flag := common.GetKubeletFlags(machines.Items[i].Annotations)[common.ExternalCloudProviderKubeletFlag]
		if boolFlag, err := strconv.ParseBool(flag); !boolFlag || err != nil {
			migrated = false

			break
		}
	}

	if !migrated {
		return fail.NewRuntimeError("checking CCM migration readiness status", "not all machines are rolled-out or migration not started yet")
	}

	return nil
}

func ccmMigrationRegenerateControlPlaneManifestsAndKubeletConfig(s *state.State) error {
	return s.RunTaskOnControlPlane(ccmMigrationRegenerateControlPlaneManifestsAndKubeletConfigInternal, state.RunSequentially)
}

func ccmMigrationRegenerateControlPlaneManifestsAndKubeletConfigInternal(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	logger := s.Logger.WithField("node", node.PublicAddress)
	logger.Info("Starting CCM/CSI migration...")

	drainer := nodeutils.NewDrainer(s.RESTConfig, logger)

	logger.Infoln("Cordoning node...")
	if err := drainer.Cordon(s.Context, node.Hostname, true); err != nil {
		return err
	}

	logger.Infoln("Draining node...")
	if err := drainer.Drain(s.Context, node.Hostname); err != nil {
		return err
	}

	logger.Info("Regenerating API server and kube-controller-manager manifests, and Kubelet configuration...")

	cmd, err := scripts.CCMMigrationRegenerateControlPlaneConfigs(s.WorkDir, node.ID, s.KubeadmVerboseFlag())
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return fail.SSH(err, "regenerate control-plane manifests for CCM migration")
	}

	var (
		apiserverPodName         = fmt.Sprintf("kube-apiserver-%s", node.Hostname)
		controllerManagerPodName = fmt.Sprintf("kube-controller-manager-%s", node.Hostname)
		timeout                  = 2 * time.Minute
	)

	logger.Debugf("Waiting %s for control plane components to stabilize...", timeout)
	time.Sleep(timeout)

	logger.Debugf("Waiting up to %s for Kubelet to become running...", timeout)
	err = waitForKubeletReady(conn, timeout)
	if err != nil {
		return err
	}

	logger.Infof("Waiting up to %s for API server to become healthy...", timeout)
	err = waitForStaticPodReady(s, timeout, apiserverPodName, metav1.NamespaceSystem)
	if err != nil {
		return err
	}

	logger.Infof("Waiting up to %s for kube-controller-manager roll-out...", timeout)
	err = waitForStaticPodReady(s, timeout, controllerManagerPodName, metav1.NamespaceSystem)
	if err != nil {
		return err
	}

	logger.Infoln("Uncordoning node...")
	if err := drainer.Cordon(s.Context, node.Hostname, false); err != nil {
		return err
	}

	return nil
}

func ccmMigrationUpdateStaticWorkersKubeletConfig(s *state.State) error {
	return s.RunTaskOnStaticWorkers(ccmMigrationUpdateStaticWorkersKubeletConfigInternal, state.RunSequentially)
}

func ccmMigrationUpdateStaticWorkersKubeletConfigInternal(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
	logger := s.Logger.WithField("node", node.PublicAddress)
	logger.Info("Updating config and restarting Kubelet...")

	drainer := nodeutils.NewDrainer(s.RESTConfig, logger)

	logger.Infoln("Cordoning node...")
	if err := drainer.Cordon(s.Context, node.Hostname, true); err != nil {
		return err
	}

	logger.Infoln("Draining node...")
	if err := drainer.Drain(s.Context, node.Hostname); err != nil {
		return err
	}

	// Update kubelet config and flags
	logger.Info("Updating Kubelet config...")
	if err := ccmMigrationUpdateKubeletConfigFile(s); err != nil {
		return err
	}
	if err := ccmMigrationUpdateKubeletFlags(s); err != nil {
		return err
	}

	// Restart Kubelet
	logger.Info("Restarting Kubelet...")
	_, _, err := s.Runner.RunRaw(scripts.RestartKubelet())
	if err != nil {
		return fail.SSH(err, "restarting kubelet for CCM migration")
	}

	timeout := 2 * time.Minute
	logger.Debugf("Waiting up to %s for Kubelet to become running...", timeout)
	if err := waitForKubeletReady(conn, timeout); err != nil {
		return err
	}

	logger.Infoln("Uncordoning node...")
	if err := drainer.Cordon(s.Context, node.Hostname, false); err != nil {
		return err
	}

	return nil
}

func ccmMigrationUpdateKubeletConfigFile(s *state.State) error {
	return updateRemoteFile(s, kubeletConfigFile, func(content []byte) ([]byte, error) {
		// Unmarshal and update the config
		kubeletConfig, err := unmarshalKubeletConfig(content)
		if err != nil {
			return nil, err
		}

		if kubeletConfig.FeatureGates == nil {
			kubeletConfig.FeatureGates = map[string]bool{}
		}
		if s.ShouldEnableCSIMigration() {
			featureGates, _, fgErr := s.Cluster.CSIMigrationFeatureGates(s.ShouldUnregisterInTreeCloudProvider())
			if fgErr != nil {
				return nil, fgErr
			}
			for k, v := range featureGates {
				kubeletConfig.FeatureGates[k] = v
			}
		}

		return marshalKubeletConfig(kubeletConfig)
	})
}

func ccmMigrationUpdateKubeletFlags(s *state.State) error {
	return updateRemoteFile(s, kubeadmEnvFlagsFile, func(content []byte) ([]byte, error) {
		kubeletFlags, err := unmarshalKubeletFlags(content)
		if err != nil {
			return nil, err
		}

		kubeletFlags["--cloud-provider"] = "external"
		delete(kubeletFlags, "--cloud-config")
		buf := marshalKubeletFlags(kubeletFlags)

		return buf, nil
	})
}

func waitForStaticPodReady(s *state.State, timeout time.Duration, podName, podNamespace string) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	if podName == "" || podNamespace == "" {
		return fail.KubeClient(fmt.Errorf("static pod name and namespace are required"), "waiting for static pods")
	}

	return wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		if s.Verbose {
			s.Logger.Debugf("Waiting for pod %q to become healthy...", podName)
		}

		pod := corev1.Pod{}
		key := client.ObjectKey{
			Name:      podName,
			Namespace: podNamespace,
		}
		err := s.DynamicClient.Get(s.Context, key, &pod)
		if err != nil {
			// NB: We're intentionally ignoring error here to prevent failures while
			// Kubelet is rolling-out the static pod.
			if s.Verbose {
				s.Logger.Debugf("Failed to get pod %q: %v", podName, err)
			}

			return false, nil
		}

		// Ensure pod is running
		if pod.Status.Phase != corev1.PodRunning {
			if s.Verbose {
				s.Logger.Debugf("Pod %q is not yet running", podName)
			}

			return false, nil
		}

		// Ensure pod and all containers are ready
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status != corev1.ConditionTrue {
				if s.Verbose {
					s.Logger.Debugf("Pod %q is not yet ready", podName)
				}

				return false, nil
			} else if cond.Type == corev1.ContainersReady && cond.Status != corev1.ConditionTrue {
				if s.Verbose {
					s.Logger.Debugf("Containers for pod %q are not yet ready", podName)
				}

				return false, nil
			}
		}

		return true, nil
	})
}

func waitForKubeletReady(conn executor.Interface, timeout time.Duration) error {
	err := wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		kubeletStatus, sErr := systemdStatus(conn, "kubelet")
		if sErr != nil {
			return false, sErr
		}

		if kubeletStatus&state.SystemDStatusRunning != 0 && kubeletStatus&state.SystemDStatusRestarting == 0 {
			return true, nil
		}

		return false, nil
	})

	return fail.Runtime(err, "waiting for kubelet readiness")
}

func migrateOpenStackPVs(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	s.Logger.Infof("Patching OpenStack PersistentVolumes with annotation \"%s=%s\"...", provisionedByAnnotation, provisionedByOpenStackCSICinder)

	pvList := corev1.PersistentVolumeList{}
	if err := s.DynamicClient.List(s.Context, &pvList, &client.ListOptions{}); err != nil {
		return fail.KubeClient(err, "getting %T", pvList)
	}

	for i, pv := range pvList.Items {
		pv := pv
		if pv.Annotations[provisionedByAnnotation] == provisionedByOpenStackInTreeCinder {
			pvKey := client.ObjectKeyFromObject(&pv)
			if s.Verbose {
				s.Logger.Debugf("Patching PersistentVolume %q...", pvKey)
			}

			oldPv := pv.DeepCopy()
			pv.Annotations[provisionedByAnnotation] = provisionedByOpenStackCSICinder

			if err := s.DynamicClient.Patch(s.Context, &pvList.Items[i], client.MergeFrom(oldPv)); err != nil {
				return fail.KubeClient(err, "patching %T %s", pv, pvKey)
			}
		}
	}

	return nil
}
