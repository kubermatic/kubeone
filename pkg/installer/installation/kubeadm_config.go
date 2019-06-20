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

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
	"github.com/kubermatic/kubeone/pkg/util"
)

func generateKubeadm(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating kubeadm config file…")

	kadm, err := kubeadm.New(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubeadm")
	}

	for idx := range ctx.Cluster.Hosts {
		kubeadm, err := kadm.Config(ctx, ctx.Cluster.Hosts[idx])
		if err != nil {
			return errors.Wrap(err, "failed to create kubeadm configuration")
		}

		ctx.Configuration.AddFile(fmt.Sprintf("cfg/master_%d.yaml", idx), kubeadm)
	}

	return ctx.RunTaskOnAllNodes(generateKubeadmOnNode, true)
}

func generateKubeadmOnNode(ctx *util.Context, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	return errors.Wrap(ctx.Configuration.UploadTo(conn, ctx.WorkDir), "failed to upload")
}
