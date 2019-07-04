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

package upgrader

import (
	"github.com/sirupsen/logrus"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/configupload"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/upgrader/upgrade"
)

// Options groups the various possible options for running KubeOne upgrade
type Options struct {
	ForceUpgrade              bool
	UpgradeMachineDeployments bool
	Verbose                   bool
}

// Upgrader is entrypoint for the upgrade process
type Upgrader struct {
	cluster *kubeoneapi.KubeOneCluster
	logger  *logrus.Logger
}

// NewUpgrader returns a new upgrader, responsible for running the upgrade process
func NewUpgrader(cluster *kubeoneapi.KubeOneCluster, logger *logrus.Logger) *Upgrader {
	return &Upgrader{
		cluster: cluster,
		logger:  logger,
	}
}

// Upgrade run the upgrade process
func (u *Upgrader) Upgrade(options *Options) error {
	return upgrade.Upgrade(u.createState(options))
}

// createState creates a basic, non-host bound context with all relevant information, but no Runner yet.
// The various task helper functions will take care of setting up Runner structs for each task individually
func (u *Upgrader) createState(options *Options) *state.State {
	return &state.State{
		Cluster:                   u.cluster,
		Connector:                 ssh.NewConnector(),
		Configuration:             configupload.NewConfiguration(),
		WorkDir:                   "kubeone",
		Logger:                    u.logger,
		Verbose:                   options.Verbose,
		ForceUpgrade:              options.ForceUpgrade,
		UpgradeMachineDeployments: options.UpgradeMachineDeployments,
	}
}
