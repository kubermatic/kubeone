package v1_11

import (
	"fmt"
	"time"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	if err := InstallPrerequisites(ctx); err != nil {
		return fmt.Errorf("failed to install prerequisites: %v", err)
	}

	if err := GenerateCA(ctx); err != nil {
		return fmt.Errorf("failed to generate CA: %v", err)
	}

	if err := DeployCA(ctx); err != nil {
		return fmt.Errorf("failed to deploy CA: %v", err)
	}

	if err := WaitForEtcd(ctx); err != nil {
		return fmt.Errorf("etcd bootstrapping failed: %v", err)
	}

	if err := InitKubernetes(ctx); err != nil {
		return fmt.Errorf("failed to init Kubernetes: %v", err)
	}

	if err := Wait(ctx, 30*time.Second); err != nil {
		return err
	}

	if err := InstallKubeProxy(ctx); err != nil {
		return fmt.Errorf("failed to install kube proxy: %v", err)
	}

	if err := Wait(ctx, 30*time.Second); err != nil {
		return err
	}

	if err := CreateJoinToken(ctx); err != nil {
		return fmt.Errorf("failed to create join token: %v", err)
	}

	return nil
}
