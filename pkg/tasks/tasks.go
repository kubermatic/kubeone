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

	"github.com/kubermatic/kubeone/pkg/addons"
	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/externalccm"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/templates/nodelocaldns"
)

type Tasks []Task

func (t Tasks) Run(s *state.State) error {
	for _, step := range t {
		if err := step.Run(s); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	return nil
}

func (t Tasks) append(newtasks ...Task) Tasks {
	return append(t, newtasks...)
}

func (t Tasks) prepend(newtasks ...Task) Tasks {
	return append(newtasks, t...)
}

// WithBinariesOnly will prepend passed tasks with tasks WithHostnameOS() and
// append install prerequisite binaries (docker, kubeadm, kubelet, etc...) on
// all hosts
func WithBinariesOnly(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(Task{Fn: installPrerequisites, ErrMsg: "failed to install prerequisites"})
}

// WithHostnameOS will prepend passed tasks with 2 basic tasks:
//  * detect OS on all cluster hosts
//  * detect hostnames  on all cluster hosts
func WithHostnameOS(t Tasks) Tasks {
	return t.prepend(
		Task{Fn: determineHostname, ErrMsg: "failed to detect hostname"},
		Task{Fn: determineOS, ErrMsg: "failed to detect OS"},
	)
}

// WithFullInstall with install binaries (using WithBinariesOnly) and
// orchestrate complete cluster init
func WithFullInstall(t Tasks) Tasks {
	return WithBinariesOnly(t).
		append(kubernetesConfigFiles()...).
		append(Tasks{
			{Fn: kubeadmCertsOnLeader, ErrMsg: "failed to provision certs and etcd on leader"},
			{Fn: certificate.DownloadCA, ErrMsg: "failed to download ca from leader"},
			{Fn: deployPKIToFollowers, ErrMsg: "failed to upload PKI"},
			{Fn: kubeadmCertsOnFollower, ErrMsg: "failed to provision certs and etcd on followers"},
			{Fn: initKubernetesLeader, ErrMsg: "failed to init kubernetes on leader"},
			{Fn: joinControlplaneNode, ErrMsg: "failed to join other masters a cluster"},
			{Fn: copyKubeconfig, ErrMsg: "failed to copy kubeconfig to home directory"},
			{Fn: saveKubeconfig, ErrMsg: "failed to save kubeconfig to the local machine"},
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
		}...).
		append(kubernetesResources()...).
		append(
			Task{Fn: createMachineDeployments, ErrMsg: "failed to create worker machines"},
		)
}

func WithUpgrade(t Tasks) Tasks {
	return WithHostnameOS(t).
		append(kubernetesConfigFiles()...).
		append(Tasks{
			{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "failed to build kubernetes clientset"},
			{Fn: runPreflightChecks, ErrMsg: "preflight checks failed", Retries: 1},
			{Fn: upgradeLeader, ErrMsg: "failed to upgrade leader control plane"},
			{Fn: upgradeFollower, ErrMsg: "failed to upgrade follower control plane"},
			{Fn: certificate.DownloadCA, ErrMsg: "failed to download ca from leader"},
		}...).
		append(kubernetesResources()...).
		append(
			Task{Fn: upgradeStaticWorkers, ErrMsg: "unable to upgrade static worker nodes"},
			Task{Fn: upgradeMachineDeployments, ErrMsg: "failed to upgrade MachineDeployments"},
		)
}

func WithReset(t Tasks) Tasks {
	return t.append(Tasks{
		{Fn: destroyWorkers, ErrMsg: "failed to destroy workers"},
		{Fn: resetAllNodes, ErrMsg: "failed to reset nodes"},
		{Fn: removeBinariesAllNodes, ErrMsg: "failed to remove binaries from nodes"},
	}...)
}

func kubernetesConfigFiles() Tasks {
	return Tasks{
		{Fn: generateKubeadm, ErrMsg: "failed to generate kubeadm config files"},
		{Fn: generateConfigurationFiles, ErrMsg: "failed to generate config files"},
		{Fn: uploadConfigurationFiles, ErrMsg: "failed to upload config files"},
	}
}

func kubernetesResources() Tasks {
	return Tasks{
		{Fn: nodelocaldns.Deploy, ErrMsg: "failed to deploy nodelocaldns"},
		{Fn: features.Activate, ErrMsg: "failed to activate features"},
		{Fn: ensureCNI, ErrMsg: "failed to install cni plugin"},
		{Fn: addons.Ensure, ErrMsg: "failed to apply addons"},
		{Fn: patchCoreDNS, ErrMsg: "failed to patch CoreDNS"},
		{Fn: credentials.Ensure, ErrMsg: "failed to ensure credentials secret"},
		{Fn: externalccm.Ensure, ErrMsg: "failed to ensure external CCM"},
		{Fn: patchCNI, ErrMsg: "failed to patch CNI"},
		{Fn: joinStaticWorkerNodes, ErrMsg: "failed to join worker nodes to the cluster"},
		{Fn: machinecontroller.Ensure, ErrMsg: "failed to install machine-controller"},
		{Fn: machinecontroller.WaitReady, ErrMsg: "failed to wait for machine-controller"},
	}
}
