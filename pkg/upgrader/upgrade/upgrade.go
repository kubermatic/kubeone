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

package upgrade

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/task"
	"github.com/kubermatic/kubeone/pkg/templates/externalccm"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util/context"
	"github.com/kubermatic/kubeone/pkg/util/credentials"
	"github.com/kubermatic/kubeone/pkg/util/kubeconfig"
)

const (
	labelUpgradeLock      = "kubeone.io/upgrade-in-progress"
	labelControlPlaneNode = "node-role.kubernetes.io/master"
	// timeoutKubeletUpgrade is time for how long kubeone will wait after upgrading kubelet
	// and running the upgrade process on the node
	timeoutKubeletUpgrade = 1 * time.Minute
	// timeoutNodeUpgrade is time for how long kubeone will wait after finishing the upgrade
	// process on the node
	timeoutNodeUpgrade = 15 * time.Second
)

// Upgrade performs all the steps required to upgrade Kubernetes on
// cluster provisioned using KubeOne
func Upgrade(ctx *context.Context) error {
	// commonSteps are same for all worker nodes and they are safe to be run in parallel
	commonSteps := []task.Task{
		{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "unable to build kubernetes clientset"},
		{Fn: determineHostname, ErrMsg: "unable to determine hostname"},
		{Fn: determineOS, ErrMsg: "unable to determine operating system"},
		{Fn: runPreflightChecks, ErrMsg: "preflight checks failed"},
		{Fn: upgradeLeader, ErrMsg: "unable to upgrade leader control plane", Retries: 3},
		{Fn: upgradeFollower, ErrMsg: "unable to upgrade follower control plane", Retries: 3},
		{Fn: features.Activate, ErrMsg: "unable to activate features"},
		{Fn: certificate.DownloadCA, ErrMsg: "unable to download ca from leader", Retries: 3},
		{Fn: credentials.Ensure, ErrMsg: "unable to ensure credentials secret"},
		{Fn: externalccm.Ensure, ErrMsg: "failed to install external CCM"},
		{Fn: machinecontroller.Ensure, ErrMsg: "failed to update machine-controller", Retries: 3},
		{Fn: machinecontroller.WaitReady, ErrMsg: "failed to wait for machine-controller", Retries: 3},
		{Fn: upgradeMachineDeployments, ErrMsg: "unable to upgrade MachineDeployments", Retries: 3},
	}

	for _, step := range commonSteps {
		if err := step.Run(ctx); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	return nil
}
