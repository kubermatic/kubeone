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
	"strconv"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

const (
	kubeadmCertCommand = `
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
       sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
       rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo kubeadm init phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
	kubeadmInitCommand = `
if [[ -f /etc/kubernetes/admin.conf ]]; then exit 0; fi
sudo kubeadm init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
)

func kubeadmCertsOnLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on first controller…")
	return ctx.RunTaskOnLeader(kubeadmCertsExecutor)
}

func kubeadmCertsOnFollower(ctx *util.Context) error {
	ctx.Logger.Infoln("Configuring certs and etcd on consecutive controller…")
	return ctx.RunTaskOnFollowers(kubeadmCertsExecutor, true)
}

func kubeadmCertsExecutor(ctx *util.Context, node kubeoneapi.HostConfig, conn ssh.Connection) error {

	ctx.Logger.Infoln("Ensuring Certificates…")
	_, _, err := ctx.Runner.Run(kubeadmCertCommand, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}

func initKubernetesLeader(ctx *util.Context) error {
	ctx.Logger.Infoln("Initializing Kubernetes on leader…")

	return ctx.RunTaskOnLeader(func(ctx *util.Context, node kubeoneapi.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, err := ctx.Runner.Run(kubeadmInitCommand, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
			"NODE_ID":  strconv.Itoa(node.ID),
		})

		return err
	})
}
