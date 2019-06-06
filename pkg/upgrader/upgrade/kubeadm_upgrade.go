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

	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
	"github.com/kubermatic/kubeone/pkg/util"
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

func upgradeLeaderControlPlane(ctx *util.Context) error {
	kadm, err := kubeadm.New(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubeadm")
	}

	_, _, err = ctx.Runner.Run(kubeadmUpgradeLeaderCommand, util.TemplateVariables{
		"KUBEADM_UPGRADE": kadm.UpgradeLeaderCMD(),
		"VERSION":         ctx.Cluster.Versions.Kubernetes,
		"WORK_DIR":        ctx.WorkDir,
	})

	return err
}

func upgradeFollowerControlPlane(ctx *util.Context) error {
	kadm, err := kubeadm.New(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubadm")
	}

	_, _, err = ctx.Runner.Run(kubeadmUpgradeFollowerCommand, util.TemplateVariables{
		"KUBEADM_UPGRADE": kadm.UpgradeFollowerCMD(),
	})
	return err
}
