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
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/admissionconfig"
)

func installPrerequisites(s *state.State) error {
	s.Logger.Infoln("Installing prerequisites…")

	return s.RunTaskOnAllNodes(installPrerequisitesOnNode, state.RunParallel)
}

func generateConfigurationFiles(s *state.State) error {
	s.Configuration.AddFile("cfg/cloud-config", s.Cluster.CloudProvider.CloudConfig)

	if s.Cluster.Features.StaticAuditLog != nil && s.Cluster.Features.StaticAuditLog.Enable {
		if err := s.Configuration.AddFilePath("cfg/audit-policy.yaml", s.Cluster.Features.StaticAuditLog.Config.PolicyFilePath, s.ManifestFilePath); err != nil {
			return errors.Wrap(err, "unable to add policy file")
		}
	}
	if s.Cluster.Features.PodNodeSelector != nil && s.Cluster.Features.PodNodeSelector.Enable {
		admissionCfg, err := admissionconfig.NewAdmissionConfig(s.Cluster.Versions.Kubernetes, s.Cluster.Features.PodNodeSelector)
		if err != nil {
			return errors.Wrap(err, "failed to generate admissionconfiguration manifest")
		}
		s.Configuration.AddFile("cfg/admission-config.yaml", admissionCfg)

		if err := s.Configuration.AddFilePath("cfg/podnodeselector.yaml", s.Cluster.Features.PodNodeSelector.Config.ConfigFilePath, s.ManifestFilePath); err != nil {
			return errors.Wrap(err, "failed to add podnodeselector config file")
		}
	}

	return nil
}

func installPrerequisitesOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("os", node.OperatingSystem)

	logger.Infoln("Creating environment file…")
	if err := createEnvironmentFile(s); err != nil {
		return errors.Wrap(err, "failed to create environment file")
	}

	logger.Infoln("Configuring proxy…")
	if err := configureProxy(s); err != nil {
		return errors.Wrap(err, "failed to configure proxy for docker daemon")
	}

	logger.Infoln("Installing kubeadm…")
	return errors.Wrap(installKubeadm(s, *node), "failed to install kubeadm")
}

func createEnvironmentFile(s *state.State) error {
	cmd, err := scripts.EnvironmentFile(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return err
}

func installKubeadm(s *state.State, node kubeoneapi.HostConfig) error {
	return runOnOS(s, node.OperatingSystem, map[kubeoneapi.OperatingSystemName]runOnOSFn{
		kubeoneapi.OperatingSystemNameUbuntu:  installKubeadmDebian,
		kubeoneapi.OperatingSystemNameCoreOS:  installKubeadmCoreOS,
		kubeoneapi.OperatingSystemNameFlatcar: installKubeadmCoreOS,
		kubeoneapi.OperatingSystemNameCentOS:  installKubeadmCentOS,
		kubeoneapi.OperatingSystemNameRHEL:    installKubeadmCentOS,
	})
}

func installKubeadmDebian(s *state.State) error {
	cmd, err := scripts.KubeadmDebian(s.Cluster, s.ForceInstall)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func installKubeadmCentOS(s *state.State) error {
	cmd, err := scripts.KubeadmCentOS(s.Cluster, s.ForceInstall)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func installKubeadmCoreOS(s *state.State) error {
	cmd, err := scripts.KubeadmCoreOS(s.Cluster)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func uploadConfigurationFiles(s *state.State) error {
	return s.RunTaskOnAllNodes(uploadConfigurationFilesToNode, state.RunParallel)
}

func uploadConfigurationFilesToNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Uploading config files…")

	if err := s.Configuration.UploadTo(conn, s.WorkDir); err != nil {
		return errors.Wrap(err, "failed to upload")
	}

	cmd, err := scripts.SaveCloudConfig(s.WorkDir)
	if err != nil {
		return err
	}

	// move config files to their permanent locations
	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return err
	}

	cmd, err = scripts.SaveAuditPolicyConfig(s.WorkDir)
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return err
	}

	cmd, err = scripts.SavePodNodeSelectorConfig(s.WorkDir)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}

func configureProxy(s *state.State) error {
	if s.Cluster.Proxy.HTTP == "" && s.Cluster.Proxy.HTTPS == "" && s.Cluster.Proxy.NoProxy == "" {
		return nil
	}

	s.Logger.Infoln("Configuring docker/kubelet proxy…")
	cmd, err := scripts.DaemonsProxy()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}
