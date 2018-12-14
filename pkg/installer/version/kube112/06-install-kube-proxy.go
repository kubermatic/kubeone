package kube112

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

func installKubeProxy(ctx *util.Context) error {
	return util.RunTaskOnLeader(ctx, func(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Installing kube-proxy…")

		_, _, err := ctx.Runner.Run(`
mkdir -p ~/.kube
sudo cp /etc/kubernetes/admin.conf ~/.kube/config
sudo chown -R $(id -u):$(id -g) ~/.kube

cd "{{ .WORK_DIR }}"

sudo kubectl -n kube-system get configmap kube-proxy -o yaml \
	| sed -i -e 's#server:.*#server: https://{{ .IP_ADDRESS }}:6443#g' \
	| sudo kubectl replace -f -
sudo kubectl -n kube-system delete pod -l k8s-app=kube-proxy
`, util.TemplateVariables{
			"WORK_DIR":   ctx.WorkDir,
			"IP_ADDRESS": node.PublicAddress,
		})

		return err
	})
}
