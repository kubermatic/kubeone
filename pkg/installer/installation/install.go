package installation

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	installSteps := []struct {
		fn     func(*util.Context) error
		errMsg string
	}{
		{fn: installPrerequisites, errMsg: "failed to install prerequisites"},
		{fn: generateKubeadm, errMsg: "failed to generate kubeadm config files"},
		{fn: kubeadmCertsOnLeader, errMsg: "failed to provision certs and etcd on leader"},
		{fn: downloadCA, errMsg: "unable to download ca from leader"},
		{fn: deployCA, errMsg: "unable to deploy ca on nodes"},
		{fn: kubeadmCertsOnFollower, errMsg: "failed to provision certs and etcd on followers"},
		{fn: initKubernetesLeader, errMsg: "failed to init kubernetes on leader"},
		{fn: joinControlplaneNode, errMsg: "unable to join other masters a cluster"},
		{fn: copyKubeconfig, errMsg: "unable to copy kubeconfig to home directory"},
		{fn: util.BuildKubernetesClientset, errMsg: "unable to build kubernetes clientset"},
		{fn: applyCanalCNI, errMsg: "failed to install cni plugin canal"},
		{fn: installMachineController, errMsg: "failed to install machine-controller"},
		{fn: createWorkerMachines, errMsg: "failed to create worker machines"},
		{fn: deployArk, errMsg: "failed to deploy ark"},
	}

	for _, step := range installSteps {
		if err := step.fn(ctx); err != nil {
			return errors.Wrap(err, step.errMsg)
		}
	}

	return nil
}
