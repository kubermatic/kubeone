package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	if err := installPrerequisites(ctx); err != nil {
		return fmt.Errorf("failed to install prerequisites: %v", err)
	}
	if err := ctx.RunTaskOnAllNodes(setupEtcd, true); err != nil {
		return fmt.Errorf("failed to setup etcd: %v", err)
	}
	if err := generateKubeadm(ctx); err != nil {
		return fmt.Errorf("failed to generate kubeadm config files: %v", err)
	}
	if err := kubeadmCertsAndEtcdOnLeader(ctx); err != nil {
		return fmt.Errorf("failed to provision certs and etcd on leader: %v", err)
	}
	if err := downloadCA(ctx); err != nil {
		return fmt.Errorf("unable to download ca from leader: %v", err)
	}
	if err := deployCA(ctx); err != nil {
		return fmt.Errorf("unable to deploy ca on nodes: %v", err)
	}
	if err := kubeadmCertsAndEtcdOnFollower(ctx); err != nil {
		return fmt.Errorf("failed to provision certs and etcd on followers: %v", err)
	}
	if err := initKubernetesLeader(ctx); err != nil {
		return fmt.Errorf("failed to init kubernetes on leader: %v", err)
	}
	if err := createJoinToken(ctx); err != nil {
		return fmt.Errorf("failed to create join token: %v", err)
	}
	if err := joinControlplaneNode(ctx); err != nil {
		return fmt.Errorf("unable to join other masters a cluster: %v", err)
	}
	if err := installKubeProxy(ctx); err != nil {
		return fmt.Errorf("failed to install kube proxy: %v", err)
	}
	if err := applyCNI(ctx, "canal"); err != nil {
		return fmt.Errorf("failed to install cni plugin flannel: %v", err)
	}
	if err := installMachineController(ctx); err != nil {
		return fmt.Errorf("failed to install machine-controller: %v", err)
	}
	if err := createWorkerMachines(ctx); err != nil {
		return fmt.Errorf("failed to create worker machines: %v", err)
	}
	if err := deployArk(ctx); err != nil {
		return fmt.Errorf("failed to deploy ark: %v", err)
	}

	return nil
}
