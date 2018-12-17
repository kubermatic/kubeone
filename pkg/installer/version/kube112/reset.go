package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
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
sudo rm -f /etc/systemd/system/kubelet.service.d/cfg/20-cloudconfig-kubelet.conf
if [[ -f /etc/systemd/system/kubelet.service.d/10-kubeadm.conf ]]; then
	sudo mv /etc/systemd/system/kubelet.service.d/10-kubeadm.conf{,.disabled}
fi
sudo systemctl daemon-reload
rm -rf "{{ .WORK_DIR }}"
`
