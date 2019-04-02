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
	"github.com/kubermatic/kubeone/pkg/task"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Install performs all the steps required to install Kubernetes on
// an empty, pristine machine.
func Install(ctx *util.Context) error {
	installSteps := []task.Task{
		{Fn: installPrerequisites, ErrMsg: "failed to install prerequisites"},
		{Fn: generateKubeadm, ErrMsg: "failed to generate kubeadm config files"},
		{Fn: kubeadmCertsOnLeader, ErrMsg: "failed to provision certs and etcd on leader"},
		{Fn: certificate.DownloadCA, ErrMsg: "unable to download ca from leader", Retries: 3},
		{Fn: deployCA, ErrMsg: "unable to deploy ca on nodes", Retries: 3},
		{Fn: kubeadmCertsOnFollower, ErrMsg: "failed to provision certs and etcd on followers"},
		{Fn: initKubernetesLeader, ErrMsg: "failed to init kubernetes on leader"},
		{Fn: joinControlplaneNode, ErrMsg: "unable to join other masters a cluster"},
		{Fn: copyKubeconfig, ErrMsg: "unable to copy kubeconfig to home directory", Retries: 3},
		{Fn: saveKubeconfig, ErrMsg: "unable to save kubeconfig to the local machine", Retries: 3},
		{Fn: util.BuildKubernetesClientset, ErrMsg: "unable to build kubernetes clientset", Retries: 3},
		{Fn: features.Activate, ErrMsg: "unable to activate features"},
		{Fn: applyCanalCNI, ErrMsg: "failed to install cni plugin canal", Retries: 3},
		{Fn: machinecontroller.EnsureMachineController, ErrMsg: "failed to install machine-controller", Retries: 3},
		{Fn: machinecontroller.WaitReady, ErrMsg: "failed to wait for machine-controller", Retries: 3},
		{Fn: createWorkerMachines, ErrMsg: "failed to create worker machines", Retries: 3},
	}

	for _, step := range installSteps {
		if err := step.Run(ctx); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	return nil
}
