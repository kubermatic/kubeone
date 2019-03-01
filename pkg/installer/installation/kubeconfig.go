package installation

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func copyKubeconfig(ctx *util.Context) error {
	return ctx.RunTaskOnAllNodes(func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Copying Kubeconfig to home directory…")

		_, _, err := ctx.Runner.Run(`
mkdir -p $HOME/.kube/
sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -u) $HOME/.kube/config
`, util.TemplateVariables{})
		if err != nil {
			return err
		}

		return nil
	}, true)
}
