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

package certificate

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

// DownloadCA grabs CA certs/keys from leader host
func DownloadCA(s *state.State) error {
	s.Logger.Info("Downloading PKI…")

	return s.RunTaskOnLeader(func(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
		_, _, err := s.Runner.Run(`
mkdir -p ./{{ .WORK_DIR }}/pki/etcd
sudo cp /etc/kubernetes/pki/ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/sa.pub ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.crt ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/front-proxy-ca.key ./{{ .WORK_DIR }}/pki/
sudo cp /etc/kubernetes/pki/etcd/ca.{crt,key} ./{{ .WORK_DIR }}/pki/etcd/

sudo chown -R "$(id -u):$(id -g)" ./{{ .WORK_DIR }}
`, runner.TemplateVariables{
			"WORK_DIR": s.WorkDir,
		})
		if err != nil {
			return err
		}

		s.Logger.Infoln("Downloading PKI files…")

		err = s.Configuration.Download(conn, s.WorkDir+"/pki", "pki")
		if err != nil {
			return errors.Wrap(err, "failed to download PKI files")
		}

		if s.BackupFile != "" {
			s.Logger.Infoln("Creating local backup…")

			err = s.Configuration.Backup(s.BackupFile)
			if err != nil {
				// do not stop in case of failed backups, the user can
				// always create the backup themselves if needed
				s.Logger.Warnf("Failed to create backup: %v", err)
			}
		}

		return nil
	})
}
