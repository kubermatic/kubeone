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

package installation

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	dockerVersion = "18.09.7"
)

func installPrerequisites(s *state.State) error {
	s.Logger.Infoln("Installing prerequisites…")

	err := generateConfigurationFiles(s)
	if err != nil {
		return errors.Wrap(err, "unable to generate configuration files")
	}

	return s.RunTaskOnAllNodes(installPrerequisitesOnNode, true)
}

func generateConfigurationFiles(s *state.State) error {
	s.Configuration.AddFile("cfg/cloud-config", s.Cluster.CloudProvider.CloudConfig)

	if s.Cluster.Features.StaticAuditLog != nil && s.Cluster.Features.StaticAuditLog.Enable {
		err := s.Configuration.AddFilePath("cfg/audit-policy.yaml", s.Cluster.Features.StaticAuditLog.Config.PolicyFilePath)
		if err != nil {
			return errors.Wrap(err, "unable to add policy file")
		}
	}

	return nil
}

func installPrerequisitesOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Determine operating system…")
	os, err := determineOS(s)
	if err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}

	node.SetOperatingSystem(os)

	if node.Hostname == "" {
		s.Logger.Infoln("Determine hostname…")
		hostname, hostnameErr := determineHostname(s, *node)
		if hostnameErr != nil {
			return errors.Wrap(hostnameErr, "failed to determine hostname")
		}
		node.SetHostname(hostname)
	}

	s.Logger.Infoln("Creating environment file…")
	err = createEnvironmentFile(s)
	if err != nil {
		return errors.Wrap(err, "failed to create environment file")
	}

	logger := s.Logger.WithField("os", os)

	logger.Infoln("Installing kubeadm…")
	err = installKubeadm(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to install kubeadm")
	}

	err = configureProxy(s)
	if err != nil {
		return errors.Wrap(err, "failed to configure proxy for docker daemon")
	}

	logger.Infoln("Deploying configuration files…")
	err = deployConfigurationFiles(s)
	if err != nil {
		return errors.Wrap(err, "failed to upload configuration files")
	}

	return nil
}

func determineOS(s *state.State) (string, error) {
	osID, _, err := s.Runner.Run(scripts.OSID(), nil)
	return osID, err
}

func determineHostname(s *state.State, _ kubeoneapi.HostConfig) (string, error) {
	hostnameCmd := scripts.Hostname()

	// on azure the name of the Node should == name of the VM
	if s.Cluster.CloudProvider.Name == kubeoneapi.CloudProviderNameAzure {
		hostnameCmd = `hostname`
	}
	stdout, _, err := s.Runner.Run(hostnameCmd, nil)

	return stdout, err
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
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = installKubeadmDebian(s)
	case "coreos":
		err = installKubeadmCoreOS(s)
	case "centos":
		err = installKubeadmCentOS(s)
	default:
		err = errors.Errorf("%q is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func installKubeadmDebian(s *state.State) error {
	cmd, err := scripts.KubeadmDebian(s.Cluster, dockerVersion)
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}

func installKubeadmCentOS(s *state.State) error {
	proxy := s.Cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = s.Cluster.Proxy.HTTP
	}

	cmd, err := scripts.KubeadmCentOS(s.Cluster, proxy)
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

func deployConfigurationFiles(s *state.State) error {
	err := s.Configuration.UploadTo(s.Runner.Conn, s.WorkDir)
	if err != nil {
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
