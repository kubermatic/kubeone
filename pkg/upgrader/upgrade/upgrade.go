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
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

const (
	labelUpgradeLock      = "kubeone.io/upgrade-in-progress"
	labelControlPlaneNode = "node-role.kubernetes.io/master"
)

// Upgrade performs all the steps required to upgrade Kubernetes on
// cluster provisioned using KubeOne
func Upgrade(ctx *util.Context) error {
	// commonSteps are same for all worker nodes and they are safe to be run in parallel
	commonSteps := []struct {
		fn     func(ctx *util.Context) error
		errMsg string
	}{
		{fn: util.BuildKubernetesClientset, errMsg: "unable to build kubernetes clientset"},
		{fn: determineHostname, errMsg: "unable to determine hostname"},
		{fn: determineOS, errMsg: "unable to determine operating system"},
		{fn: runPreflightChecks, errMsg: "preflight checks failed"},
		{fn: upgradeLeader, errMsg: "unable to upgrade leader control plane"},
		{fn: upgradeFollower, errMsg: "unable to upgrade follower control plane"},
		{fn: features.Activate, errMsg: "unable to activate features"},
		{fn: certificate.DownloadCA, errMsg: "unable to download ca from leader"},
		{fn: machinecontroller.EnsureMachineController, errMsg: "failed to update machine-controller"},
		{fn: machinecontroller.WaitReady, errMsg: "failed to wait for machine-controller"},
		{fn: upgradeMachineDeployments, errMsg: "unable to upgrade MachineDeployments"},
	}

	for _, step := range commonSteps {
		if err := step.fn(ctx); err != nil {
			return errors.Wrap(err, step.errMsg)
		}
	}

	return nil
}
