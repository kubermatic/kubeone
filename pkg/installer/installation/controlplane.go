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

func joinControlplaneNode(ctx *util.Context) error {
	ctx.Logger.Infoln("Joining controlplane node…")
	return ctx.RunTaskOnFollowers(joinControlPlaneNodeInternal, false)
}

func joinControlPlaneNodeInternal(ctx *util.Context, node kubeoneapi.HostConfig, conn ssh.Connection) error {
	_, _, err := ctx.Runner.Run(`
if [[ -f /etc/kubernetes/kubelet.conf ]]; then exit 0; fi

sudo kubeadm join \
	--config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
	})
	return err
}
