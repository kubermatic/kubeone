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
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/util"
)

func downloadCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Generating PKI…")

	return ctx.RunTaskOnLeader(func(ctx *util.Context, _ *config.HostConfig, conn ssh.Connection) error {
		ctx.Logger.Infoln("Running kubeadm…")

		_, _, err := ctx.Runner.Run(`
mkdir -p ./{{ .WORK_DIR }}/pki/etcd
sudo cp /etc/kubernetes/pki/ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.pub ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/etcd/ca.{crt,key} ./{{ .WORK_DIR }}/pki/etcd/

sudo chown -R "$USER:$USER" ./{{ .WORK_DIR }}
`, util.TemplateVariables{
			"WORK_DIR": ctx.WorkDir,
		})
		if err != nil {
			return err
		}

		ctx.Logger.Infoln("Downloading PKI files…")

		err = ctx.Configuration.Download(conn, ctx.WorkDir+"/pki", "pki")
		if err != nil {
			return errors.Wrap(err, "failed to download PKI files")
		}

		if ctx.BackupFile != "" {
			ctx.Logger.Infoln("Creating local backup…")

			err = ctx.Configuration.Backup(ctx.BackupFile)
			if err != nil {
				// do not stop in case of failed backups, the user can
				// always create the backup themselves if needed
				ctx.Logger.Warnf("Failed to create backup: %v", err)
			}
		}

		return nil
	})
}

func deployCA(ctx *util.Context) error {
	ctx.Logger.Infoln("Deploying PKI…")
	return ctx.RunTaskOnFollowers(deployCAOnNode, true)
}

func deployCAOnNode(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Uploading files…")
	return ctx.Configuration.UploadTo(conn, ctx.WorkDir)
}
