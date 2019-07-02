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

package upgrade

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/kubeadm"
	"github.com/kubermatic/kubeone/pkg/util/context"
)

func generateKubeadmConfig(ctx *context.Context, node kubeoneapi.HostConfig) error {
	kadm, err := kubeadm.New(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "failed to init kubeadm")
	}

	kubeadmConf, err := kadm.Config(ctx, node)
	if err != nil {
		return errors.Wrap(err, "failed to create kubeadm configuration")
	}

	ctx.Configuration.AddFile("cfg/master_0.yaml", kubeadmConf)
	return nil
}

func uploadKubeadmConfig(ctx *context.Context, sshConn ssh.Connection) error {
	return errors.Wrap(ctx.Configuration.UploadTo(sshConn, ctx.WorkDir), "failed to upload")
}
