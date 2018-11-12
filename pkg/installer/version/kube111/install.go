package kube111

import (
	"fmt"
	"time"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	if err := installPrerequisites(ctx); err != nil {
		return fmt.Errorf("failed to install prerequisites: %v", err)
	}

	if err := generateCA(ctx); err != nil {
		return fmt.Errorf("failed to generate CA: %v", err)
	}

	if err := deployCA(ctx); err != nil {
		return fmt.Errorf("failed to deploy CA: %v", err)
	}

	if err := waitForEtcd(ctx); err != nil {
		return fmt.Errorf("etcd bootstrapping failed: %v", err)
	}

	if err := initKubernetes(ctx); err != nil {
		return fmt.Errorf("failed to init Kubernetes: %v", err)
	}

	if err := wait(ctx, 30*time.Second); err != nil {
		return err
	}

	if err := installKubeProxy(ctx); err != nil {
		return fmt.Errorf("failed to install kube proxy: %v", err)
	}

	if err := wait(ctx, 30*time.Second); err != nil {
		return err
	}

	if err := createJoinToken(ctx); err != nil {
		return fmt.Errorf("failed to create join token: %v", err)
	}

	return nil
}
