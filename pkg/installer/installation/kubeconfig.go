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
