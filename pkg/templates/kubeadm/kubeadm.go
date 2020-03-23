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
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	kubeadmUpgradeNodeCommand = "kubeadm upgrade node"
)

var (
	v13x = mustParseConstraint("1.13.x")
	v14x = mustParseConstraint("1.14.x")
	v15x = mustParseConstraint("1.15.x")
	v16x = mustParseConstraint("1.16.x")
)

// Kubedm interface abstract differences between different kubeadm versions
type Kubedm interface {
	Config(s *state.State, instance kubeoneapi.HostConfig, isWorker bool) (string, error)
	UpgradeLeaderCommand() string
	UpgradeFollowerCommand() string
	UpgradeWorkerHostCommand() string
}

// New constructor
func New(ver string) (Kubedm, error) {
	sver, err := semver.NewVersion(ver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}

	switch {
	case v13x.Check(sver):
		return &kubeadmv1beta1{version: ver}, nil
	case v14x.Check(sver):
		return &kubeadmv1beta1{version: ver}, nil
	case v15x.Check(sver):
		return &kubeadmv1beta2{version: ver}, nil
	case v16x.Check(sver):
		return &kubeadmv1beta2{version: ver}, nil
	}

	// By default use latest known kubeadm API version
	return &kubeadmv1beta2{version: ver}, nil
}

func mustParseConstraint(constraint string) *semver.Constraints {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}

	return c
}
