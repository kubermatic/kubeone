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

package installation

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
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
		{fn: certificate.DownloadCA, errMsg: "unable to download ca from leader"},
		{fn: deployCA, errMsg: "unable to deploy ca on nodes"},
		{fn: kubeadmCertsOnFollower, errMsg: "failed to provision certs and etcd on followers"},
		{fn: initKubernetesLeader, errMsg: "failed to init kubernetes on leader"},
		{fn: joinControlplaneNode, errMsg: "unable to join other masters a cluster"},
		{fn: copyKubeconfig, errMsg: "unable to copy kubeconfig to home directory"},
		{fn: saveKubeconfig, errMsg: "unable to save kubeconfig to the local machine"},
		{fn: util.BuildKubernetesClientset, errMsg: "unable to build kubernetes clientset"},
		{fn: features.Activate, errMsg: "unable to activate features"},
		{fn: applyCanalCNI, errMsg: "failed to install cni plugin canal"},
		{fn: machinecontroller.EnsureMachineController, errMsg: "failed to install machine-controller"},
		{fn: machinecontroller.WaitReady, errMsg: "failed to wait for machine-controller"},
		{fn: createWorkerMachines, errMsg: "failed to create worker machines"},
	}

	for _, step := range installSteps {
		if err := step.fn(ctx); err != nil {
			return errors.Wrap(err, step.errMsg)
		}
	}

	return nil
}
