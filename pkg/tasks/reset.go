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
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/machinecontroller"

	"k8s.io/apimachinery/pkg/util/wait"
)

func destroyWorkers(s *state.State) error {
	if !s.DestroyWorkers {
		return nil
	}

	var lastErr error
	s.Logger.Infoln("Destroying worker nodes...")

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = kubeconfig.BuildKubernetesClientset(s)
		if lastErr != nil {
			s.Logger.Warn("Unable to connect to the control plane API. Retrying...")
			return false, nil
		}
		return true, nil
	})
	if lastErr != nil {
		s.Logger.Warn("Unable to connect to the control plane API and destroy worker nodes")
		s.Logger.Warn("You can skip destroying worker nodes and destroy them manually using `--destroy-workers=false`")
		return errors.Wrap(lastErr, "unable to build kubernetes clientset")
	}

	mcCRDs := []string{}
	for _, crd := range machinecontroller.CRDs() {
		mcCRDs = append(mcCRDs, crd.GetName())
	}

	condFn := clientutil.CRDsReadyCondition(s.Context, s.DynamicClient, mcCRDs)
	lastErr = wait.ExponentialBackoff(defaultRetryBackoff(3), condFn)
	if lastErr != nil {
		s.Logger.Info("Skipping deleting worker nodes because machine-controller CRDs are not deployed")
		return nil
	}

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = machinecontroller.DestroyWorkers(s)
		if lastErr != nil {
			s.Logger.Warn("Unable to destroy worker nodes. Retrying...")
			return false, nil
		}
		return true, nil
	})
	if lastErr != nil {
		return errors.Wrap(lastErr, "unable to delete all worker nodes")
	}

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = machinecontroller.WaitDestroy(s)
		if lastErr != nil {
			s.Logger.Warn("Waiting for all machines to be deleted...")
			return false, nil
		}
		return true, nil
	})
	if lastErr != nil {
		return errors.Wrap(lastErr, "error waiting for machines to be deleted")
	}

	return nil
}

func resetAllNodes(s *state.State) error {
	s.Logger.Infoln("Resettings all the nodes...")

	return s.RunTaskOnAllNodes(resetNode, state.RunSequentially)
}

func resetNode(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Resetting node...")

	cmd, err := scripts.KubeadmReset(s.KubeadmVerboseFlag(), s.WorkDir)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}

func removeBinariesAllNodes(s *state.State) error {
	if !s.RemoveBinaries {
		return nil
	}

	s.Logger.Infoln("Removing binaries from nodes...")
	return s.RunTaskOnAllNodes(removeBinaries, state.RunParallel)
}

func removeBinaries(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Removing Kubernetes binaries")
	var err error

	// Determine operating system
	if err = determineOS(s); err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}

	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameAmazon:  removeBinariesAmazonLinux,
		kubeoneapi.OperatingSystemNameCentOS:  removeBinariesCentOS,
		kubeoneapi.OperatingSystemNameDebian:  removeBinariesDebian,
		kubeoneapi.OperatingSystemNameFlatcar: removeBinariesFlatcar,
		kubeoneapi.OperatingSystemNameRHEL:    removeBinariesCentOS,
		kubeoneapi.OperatingSystemNameUbuntu:  removeBinariesDebian,
	})
}

func removeBinariesDebian(s *state.State) error {
	cmd, err := scripts.RemoveBinariesDebian()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return errors.WithStack(err)
}

func removeBinariesCentOS(s *state.State) error {
	cmd, err := scripts.RemoveBinariesCentOS()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return errors.WithStack(err)
}

func removeBinariesAmazonLinux(s *state.State) error {
	cmd, err := scripts.RemoveBinariesAmazonLinux()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return errors.WithStack(err)
}

func removeBinariesFlatcar(s *state.State) error {
	cmd, err := scripts.RemoveBinariesFlatcar()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return errors.WithStack(err)
}
