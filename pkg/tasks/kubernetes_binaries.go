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
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/state"
)

func upgradeKubeletAndKubectlBinaries(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameUbuntu:  upgradeKubeletAndKubectlBinariesDebian,
		kubeoneapi.OperatingSystemNameCoreOS:  upgradeKubeletAndKubectlBinariesCoreOS,
		kubeoneapi.OperatingSystemNameFlatcar: upgradeKubeletAndKubectlBinariesCoreOS,
		kubeoneapi.OperatingSystemNameCentOS:  upgradeKubeletAndKubectlBinariesCentOS,
		kubeoneapi.OperatingSystemNameRHEL:    upgradeKubeletAndKubectlBinariesCentOS,
	})
}

func upgradeKubeadmAndCNIBinaries(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameUbuntu:  upgradeKubeadmAndCNIBinariesDebian,
		kubeoneapi.OperatingSystemNameCoreOS:  upgradeKubeadmAndCNIBinariesCoreOS,
		kubeoneapi.OperatingSystemNameFlatcar: upgradeKubeadmAndCNIBinariesCoreOS,
		kubeoneapi.OperatingSystemNameCentOS:  upgradeKubeadmAndCNIBinariesCentOS,
		kubeoneapi.OperatingSystemNameRHEL:    upgradeKubeadmAndCNIBinariesCentOS,
	})
}

func upgradeKubeletAndKubectlBinariesDebian(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlDebian(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeletAndKubectlBinariesCoreOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlCoreOS(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeletAndKubectlBinariesCentOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlCentOS(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesDebian(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNIDebian(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesCentOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNICentOS(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesCoreOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNICoreOS(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}
