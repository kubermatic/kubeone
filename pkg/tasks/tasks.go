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

package tasks

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/addons"
	"github.com/kubermatic/kubeone/pkg/certificate"
	"github.com/kubermatic/kubeone/pkg/credentials"
	"github.com/kubermatic/kubeone/pkg/features"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/externalccm"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/templates/nodelocaldns"
)

type Opts func(Tasks) Tasks

type Tasks []Task

func (t Tasks) Run(s *state.State) error {
	for _, step := range t {
		if err := step.Run(s); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	return nil
}

func WithBinariesOnly(t Tasks) Tasks {
	return append(t,
		Task{Fn: installPrerequisites, ErrMsg: "failed to install prerequisites"},
	)
}

func WithFullInstall(t Tasks) Tasks {
	return append(t, Tasks{
		{Fn: installPrerequisites, ErrMsg: "failed to install prerequisites"},
		{Fn: generateKubeadm, ErrMsg: "failed to generate kubeadm config files"},
		{Fn: kubeadmCertsOnLeader, ErrMsg: "failed to provision certs and etcd on leader"},
		{Fn: certificate.DownloadCA, ErrMsg: "unable to download ca from leader"},
		{Fn: deployCA, ErrMsg: "unable to deploy ca on nodes"},
		{Fn: kubeadmCertsOnFollower, ErrMsg: "failed to provision certs and etcd on followers"},
		{Fn: initKubernetesLeader, ErrMsg: "failed to init kubernetes on leader"},
		{Fn: joinControlplaneNode, ErrMsg: "unable to join other masters a cluster"},
		{Fn: copyKubeconfig, ErrMsg: "unable to copy kubeconfig to home directory"},
		{Fn: saveKubeconfig, ErrMsg: "unable to save kubeconfig to the local machine"},
		{Fn: kubeconfig.BuildKubernetesClientset, ErrMsg: "unable to build kubernetes clientset"},
		{Fn: nodelocaldns.Deploy, ErrMsg: "unable to deploy nodelocaldns"},
		{Fn: features.Activate, ErrMsg: "unable to activate features"},
		{Fn: ensureCNI, ErrMsg: "failed to install cni plugin"},
		{Fn: patchCoreDNS, ErrMsg: "failed to patch CoreDNS"},
		{Fn: credentials.Ensure, ErrMsg: "unable to ensure credentials secret"},
		{Fn: externalccm.Ensure, ErrMsg: "failed to install external CCM"},
		{Fn: patchCNI, ErrMsg: "failed to patch CNI"},
		{Fn: joinStaticWorkerNodes, ErrMsg: "unable to join worker nodes to the cluster"},
		{Fn: machinecontroller.Ensure, ErrMsg: "failed to install machine-controller"},
		{Fn: machinecontroller.WaitReady, ErrMsg: "failed to wait for machine-controller"},
		{Fn: createWorkerMachines, ErrMsg: "failed to create worker machines"},
		{Fn: addons.Ensure, ErrMsg: "failed to apply addons"},
	}...)
}

func WithUpgrade(t Tasks) Tasks {
	return t
}

func WithReset(t Tasks) Tasks {
	return t
}

func New(opts ...Opts) Tasks {
	return nil
}
