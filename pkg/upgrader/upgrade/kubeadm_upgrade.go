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

package upgrade

import (
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
)

const (
	kubeadmUpgradeLeaderCommand = `
sudo {{ .KUBEADM_UPGRADE }} \
	--config=./{{ .WORK_DIR }}/cfg/master_0.yaml \
	-y {{ .VERSION }}
`
	kubeadmUpgradeFollowerCommand = `
sudo {{ .KUBEADM_UPGRADE }}
`
)

func upgradeLeaderControlPlane(s *state.State) error {
	kadm, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubeadm")
	}

	_, _, err = s.Runner.Run(kubeadmUpgradeLeaderCommand, runner.TemplateVariables{
		"KUBEADM_UPGRADE": kadm.UpgradeLeaderCommand(),
		"VERSION":         s.Cluster.Versions.Kubernetes,
		"WORK_DIR":        s.WorkDir,
	})

	return err
}

func upgradeFollowerControlPlane(s *state.State) error {
	kadm, err := kubeadm.New(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubadm")
	}

	_, _, err = s.Runner.Run(kubeadmUpgradeFollowerCommand, runner.TemplateVariables{
		"KUBEADM_UPGRADE": kadm.UpgradeFollowerCommand(),
	})
	return err
}
