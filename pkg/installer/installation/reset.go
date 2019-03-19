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
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Reset undos all changes made by KubeOne to the configured machines.
func Reset(ctx *util.Context) error {
	ctx.Logger.Infoln("Resetting kubeadm…")

	if ctx.DestroyWorkers {
		if err := ctx.RunTaskOnLeader(destroyWorkers); err != nil {
			return err
		}
	}

	return ctx.RunTaskOnAllNodes(resetNode, true)
}

func destroyWorkers(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Destroying worker nodes…")

	_, _, err := ctx.Runner.Run(destroyScript, util.TemplateVariables{
		"WORK_DIR":   ctx.WorkDir,
		"MACHINE_NS": machinecontroller.MachineControllerNamespace,
	})

	return err
}

func resetNode(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Resetting node…")

	_, _, err := ctx.Runner.Run(resetScript, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

const destroyScript = `
if kubectl cluster-info > /dev/null; then
  kubectl annotate --all --overwrite node kubermatic.io/skip-eviction=true
  kubectl delete machinedeployment -n "{{ .MACHINE_NS }}" --all
  kubectl delete machineset -n "{{ .MACHINE_NS }}" --all
  kubectl delete machine -n "{{ .MACHINE_NS }}" --all

  for try in {1..30}; do
    if kubectl get machine -n "{{ .MACHINE_NS }}" 2>&1 | grep -q  'No resources found.'; then
      exit 0
    fi
    sleep 10s
  done

  echo "Error: Couldn't delete all machines!"
  exit 1
fi
`

const resetScript = `
sudo kubeadm reset --force
sudo rm /etc/kubernetes/cloud-config
rm -rf "{{ .WORK_DIR }}"
`
