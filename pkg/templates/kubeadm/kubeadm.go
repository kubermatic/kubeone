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

package kubeadm

import (
	"github.com/Masterminds/semver/v3"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/semverutil"
	"k8c.io/kubeone/pkg/state"
)

const (
	kubeadmUpgradeNodeCommand = "kubeadm upgrade node"
)

var preV131Constraint = semverutil.MustParseConstraint("< 1.31")

type Config struct {
	ControlPlaneInitConfiguration string
	ClusterConfiguration          string
	JoinConfiguration             string
	KubeletConfiguration          string
	KubeProxyConfiguration        string
}

// Kubedm interface abstract differences between different kubeadm versions
type Kubedm interface {
	Config(s *state.State, instance kubeoneapi.HostConfig) (*Config, error)
	ConfigWorker(s *state.State, instance kubeoneapi.HostConfig) (*Config, error)
	UpgradeLeaderCommand() string
	UpgradeFollowerCommand() string
	UpgradeStaticWorkerCommand() string
}

// New is a constructor for the kubeadm configuration
func New(ver string) (Kubedm, error) {
	sver, err := semver.NewVersion(ver)
	if err != nil {
		return nil, err
	}

	if preV131Constraint.Check(sver) {
		return &kubeadmv1beta3{version: ver}, nil
	}

	return &kubeadmv1beta4{version: ver}, nil
}
