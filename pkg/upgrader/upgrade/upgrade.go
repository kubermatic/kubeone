package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/features"
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
		{fn: upgradeMachineDeployments, errMsg: "unable to upgrade MachineDeployments"},
	}

	for _, step := range commonSteps {
		if err := step.fn(ctx); err != nil {
			return errors.Wrap(err, step.errMsg)
		}
	}

	return nil
}
