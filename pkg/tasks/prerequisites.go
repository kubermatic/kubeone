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
	"fmt"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates"
	"k8c.io/kubeone/pkg/templates/admissionconfig"
	encryptionproviders "k8c.io/kubeone/pkg/templates/encryptionproviders"

	"k8s.io/apimachinery/pkg/runtime"
)

func installPrerequisites(s *state.State) error {
	s.Logger.Infoln("Installing prerequisites...")

	return s.RunTaskOnAllNodes(installPrerequisitesOnNode, state.RunParallel)
}

func generateConfigurationFiles(s *state.State) error {
	s.Configuration.AddFile("cfg/cloud-config", s.Cluster.CloudProvider.CloudConfig)
	s.Configuration.AddFile("ca-certs/"+cabundle.FileName, s.Cluster.CABundle)

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

	if s.ShouldEnableEncryption() || s.EncryptionEnabled() {
		configFileName := s.GetEncryptionProviderConfigName()
		var config string
		// User provided custom config
		if s.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration != "" {
			config = s.Cluster.Features.EncryptionProviders.CustomEncryptionConfiguration
			s.Configuration.AddFile(fmt.Sprintf("cfg/%s", configFileName), config)
		} else if s.ShouldEnableEncryption() { // automatically generate config
			encryptionProvidersConfig, err := encryptionproviders.NewEncyrptionProvidersConfig(s)
			if err != nil {
				return err
			}
			config, err = templates.KubernetesToYAML([]runtime.Object{encryptionProvidersConfig})
			if err != nil {
				return err
			}
			s.Configuration.AddFile(fmt.Sprintf("cfg/%s", configFileName), config)
		}
	}

	return nil
}

func installPrerequisitesOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	logger := s.Logger.WithField("os", node.OperatingSystem)

	logger.Infoln("Creating environment file...")
	if err := createEnvironmentFile(s); err != nil {
		return errors.Wrap(err, "failed to create environment file")
	}

	logger.Infoln("Configuring proxy...")
	if err := configureProxy(s); err != nil {
		return errors.Wrap(err, "failed to configure proxy for docker daemon")
	}

	logger.Infoln("Installing kubeadm...")
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
		kubeoneapi.OperatingSystemNameAmazon:  installKubeadmAmazonLinux,
		kubeoneapi.OperatingSystemNameCentOS:  installKubeadmCentOS,
		kubeoneapi.OperatingSystemNameDebian:  installKubeadmDebian,
		kubeoneapi.OperatingSystemNameFlatcar: installKubeadmFlatcar,
		kubeoneapi.OperatingSystemNameRHEL:    installKubeadmCentOS,
		kubeoneapi.OperatingSystemNameUbuntu:  installKubeadmDebian,
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

func installKubeadmAmazonLinux(s *state.State) error {
	cmd, err := scripts.KubeadmAmazonLinux(s.Cluster, s.ForceInstall)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func installKubeadmFlatcar(s *state.State) error {
	cmd, err := scripts.KubeadmFlatcar(s.Cluster)
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
	s.Logger.Infoln("Uploading config files...")

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
	if err != nil {
		return err
	}

	cmd, err = scripts.SaveEncryptionProvidersConfig(s.WorkDir, s.GetEncryptionProviderConfigName())
	if err != nil {
		return err
	}
	_, _, err = s.Runner.RunRaw(cmd)
	if err != nil {
		return err
	}

	return nil
}

func configureProxy(s *state.State) error {
	if s.Cluster.Proxy.HTTP == "" && s.Cluster.Proxy.HTTPS == "" && s.Cluster.Proxy.NoProxy == "" {
		return nil
	}

	s.Logger.Infoln("Configuring docker/kubelet proxy...")
	cmd, err := scripts.DaemonsProxy()
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)
	return err
}
