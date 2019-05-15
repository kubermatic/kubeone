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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Reset undos all changes made by KubeOne to the configured machines.
func Reset(ctx *util.Context) error {
	ctx.Logger.Infoln("Resetting kubeadm…")

	if ctx.DestroyWorkers {
		if err := destroyWorkers(ctx); err != nil {
			return errors.Wrap(err, "unable to destroy worker nodes")
		}
	}

	return ctx.RunTaskOnAllNodes(resetNode, true)
}

func destroyWorkers(ctx *util.Context) error {
	ctx.Logger.Infoln("Destroying worker nodes…")

	if err := util.BuildKubernetesClientset(ctx); err != nil {
		return errors.Wrap(err, "unable to build kubernetes clientset")
	}
	if err := machinecontroller.DeleteAllMachines(ctx); err != nil {
		return errors.Wrap(err, "unable to delete all worker nodes")
	}

	return nil
}

func resetNode(ctx *util.Context, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Resetting node…")

	_, _, err := ctx.Runner.Run(resetScript, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

const resetScript = `
sudo kubeadm reset --force
sudo rm /etc/kubernetes/cloud-config
rm -rf "{{ .WORK_DIR }}"
`
