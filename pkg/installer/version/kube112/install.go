package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	steps := []struct {
		fn     func(ctx *util.Context) error
		errFmt string
	}{
		{fn: installPrerequisites, errFmt: "failed to install prerequisites: %v"},
		{fn: deployCA, errFmt: "unable to deploy ca on nodes: %v"},
		{
			fn: func(ctx *util.Context) error {
				return ctx.RunTaskOnAllNodes(setupEtcd, true)
			},
			errFmt: "failed to setup etcd: %v",
		},
		{fn: generateKubeadm, errFmt: "failed to generate kubeadm config files: %v"},
		{fn: backup, errFmt: ""},
		{fn: kubeadmCertsAndEtcdOnLeader, errFmt: "failed to provision certs and etcd on leader: %v"},
		{fn: kubeadmCertsAndEtcdOnFollower, errFmt: "failed to provision certs and etcd on followers: %v"},
		{fn: initKubernetesLeader, errFmt: "failed to init kubernetes on leader: %v"},
		{fn: createJoinToken, errFmt: "failed to create join token: %v"},
		{fn: joinControlplaneNode, errFmt: "unable to join other masters a cluster: %v"},
		{fn: installKubeProxy, errFmt: "failed to install kube proxy: %v"},
		{
			fn: func(ctx *util.Context) error {
				return applyCNI(ctx, "canal")
			},
			errFmt: "failed to install cni plugin canal: %v",
		},
		{fn: installMachineController, errFmt: "failed to install machine-controller: %v"},
		{fn: createWorkerMachines, errFmt: "failed to create worker machines: %v"},
		{fn: deployArk, errFmt: "failed to deploy ark: %v"},
	}

	for _, step := range steps {
		if err := step.fn(ctx); err != nil {
			return fmt.Errorf("v1.12 install: "+step.errFmt, err)
		}
	}

	return nil
}
