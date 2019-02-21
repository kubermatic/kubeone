package upgrade

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

const (
	labelUpgradeLock      = "kubeone.io/upgrading-in-process"
	labelControlPlaneNode = "node-role.kubernetes.io/master"
)

// Upgrade performs all the steps required to upgrade Kubernetes on
// cluster provisioned using KubeOne
func Upgrade(ctx *util.Context) error {
	if err := util.BuildKubernetesClientset(ctx); err != nil {
		return fmt.Errorf("unable to build kubernetes clientset: %v", err)
	}
	if err := runPreflightChecks(ctx); err != nil {
		return fmt.Errorf("preflight checks failed: %v", err)
	}

	return nil
}
