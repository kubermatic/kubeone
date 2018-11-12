package kube112

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
	if err := generateKubeadm(ctx); err != nil {
		return fmt.Errorf("failed to generate kubeadm config files: %v", err)
	}
	if err := initKubernetesLeader(ctx); err != nil {
		return fmt.Errorf("failed to init kubernetes on leader: %v", err)
	}
	if err := downloadCA(ctx); err != nil {
		return fmt.Errorf("unable to download ca from leader: %v", err)
	}
	if err := deployCA(ctx); err != nil {
		return fmt.Errorf("unable to deploy ca on nodes: %v", err)
	}
	if err := joinMasterCluster(ctx); err != nil {
		return fmt.Errorf("unable to join other masters a cluster: %v", err)
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
