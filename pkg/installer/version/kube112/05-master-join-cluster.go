package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func joinControlplaneNode(ctx *util.Context) error {
	ctx.Logger.Infoln("Joining controlplane node…")
	return ctx.RunTaskOnFollowers(joinControlplaneNodeInternal, true)
}

func joinControlplaneNodeInternal(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	_, _, err := ctx.Runner.Run(`
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi

if [[ -f /etc/systemd/system/kubelet.service.d/10-kubeadm.conf.disabled ]]; then
  sudo mv /etc/systemd/system/kubelet.service.d/10-kubeadm.conf{.disabled,}
  sudo systemctl daemon-reload
fi

sudo systemctl stop kubelet
sudo {{ .JOIN_COMMAND }} \
     --experimental-control-plane \
     --node-name="{{ .NODE_NAME }}" \
     --ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests
`, util.TemplateVariables{
		"WORK_DIR":     ctx.WorkDir,
		"JOIN_COMMAND": ctx.JoinCommand,
		"NODE_NAME":    node.Hostname,
	})
	return err
}
