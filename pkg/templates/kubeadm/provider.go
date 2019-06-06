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
	"github.com/kubermatic/kubeone/pkg/util"
)

var (
	v13x = mustParseConstraint("1.13.x")
	v14x = mustParseConstraint("1.14.x")
	v15x = mustParseConstraint("1.15.x")
)

// KubeADM interface abstract differences between different kubeadm versions
type KubeADM interface {
	Config(ctx *util.Context, instance kubeoneapi.HostConfig) (string, error)
	UpgradeLeaderCMD() string
	UpgradeFollowerCMD() string
}

// New constructor
func New(ver string) (KubeADM, error) {
	sver, err := semver.NewVersion(ver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse version")
	}
	switch {
	case v13x.Check(sver):
		return &kubeadmv1beta1{}, nil
	case v14x.Check(sver):
		return &kubeadmv1beta1{}, nil
	case v15x.Check(sver):
		return &kubeadmv1beta2{}, nil
	}

	// By default use latest known kubeadm API version
	return &kubeadmv1beta2{}, nil
}

func mustParseConstraint(constraint string) *semver.Constraints {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}

	return c
}
