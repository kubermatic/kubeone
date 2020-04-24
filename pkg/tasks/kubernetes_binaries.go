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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	osNameDebian = "debian"
	osNameUbuntu = "ubuntu"
	osNameCoreos = "coreos"
	osNameCentos = "centos"
)

func upgradeKubeletAndKubectlBinaries(s *state.State, node kubeoneapi.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case osNameDebian, osNameUbuntu:
		err = upgradeKubeletAndKubectlBinariesDebian(s)
	case osNameCoreos:
		err = upgradeKubeletAndKubectlBinariesCoreOS(s)
	case osNameCentos:
		err = upgradeKubeletAndKubectlBinariesCentOS(s)
	default:
		err = errors.Errorf("%q is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubeadmAndCNIBinaries(s *state.State, node kubeoneapi.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case osNameDebian, osNameUbuntu:
		err = upgradeKubeadmAndCNIBinariesDebian(s)
	case osNameCoreos:
		err = upgradeKubeadmAndCNIBinariesCoreOS(s)
	case osNameCentos:
		err = upgradeKubeadmAndCNIBinariesCentOS(s)
	default:
		err = errors.Errorf("%q is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubeletAndKubectlBinariesDebian(s *state.State) error {
	cmd, err := scripts.UpgradeKubeletAndKubectlDebian(s.Cluster.Versions.Kubernetes)
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
	cmd, err := scripts.UpgradeKubeletAndKubectlCentOS(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesDebian(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNIDebian(s.Cluster.Versions.Kubernetes, s.Cluster.Versions.KubernetesCNIVersion())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesCentOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNICentOS(s.Cluster.Versions.Kubernetes, s.Cluster.Versions.KubernetesCNIVersion())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func upgradeKubeadmAndCNIBinariesCoreOS(s *state.State) error {
	cmd, err := scripts.UpgradeKubeadmAndCNICoreOS(s.Cluster.Versions.Kubernetes, s.Cluster.Versions.KubernetesCNIVersion())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}
