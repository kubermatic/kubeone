package installation

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

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

func saveKubeconfig(ctx *util.Context) error {
	kubeconfig, err := util.DownloadKubeconfig(ctx.Cluster)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s-kubeconfig", ctx.Cluster.Name)
	err = ioutil.WriteFile(fileName, kubeconfig, 0644)
	return errors.Wrap(err, "error saving kubeconfig file to the local machine")
}
