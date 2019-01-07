package installation

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func joinControlplaneNode(ctx *util.Context) error {
	ctx.Logger.Infoln("Joining controlplane node…")
	return ctx.RunTaskOnFollowers(joinControlPlaneNodeInternal, false)
}

func joinControlPlaneNodeInternal(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	_, _, err := ctx.Runner.Run(`
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi

sudo {{ .JOIN_COMMAND }} \
     --experimental-control-plane \
     --node-name="{{ .NODE_NAME }}"
`, util.TemplateVariables{
		"WORK_DIR":     ctx.WorkDir,
		"JOIN_COMMAND": ctx.JoinCommand,
		"NODE_NAME":    node.Hostname,
	})
	return err
}
