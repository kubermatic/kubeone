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
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	kubeadmCertCommand = `
if [[ -d ./{{ .WORK_DIR }}/pki ]]; then
   sudo rsync -av ./{{ .WORK_DIR }}/pki/ /etc/kubernetes/pki/
   sudo chown -R root:root /etc/kubernetes
   rm -rf ./{{ .WORK_DIR }}/pki
fi
sudo kubeadm {{ .VERBOSE }} init phase certs all --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
	kubeadmInitCommand = `
if [[ -f /etc/kubernetes/admin.conf ]]; then
	sudo kubeadm {{ .VERBOSE }} token create {{ .TOKEN }} --ttl {{ .TOKEN_DURATION }}
	exit 0;
fi
sudo kubeadm {{ .VERBOSE }} init --config=./{{ .WORK_DIR }}/cfg/master_{{ .NODE_ID }}.yaml
`
)

func kubeadmCertsOnLeader(s *state.State) error {
	s.Logger.Infoln("Configuring certs and etcd on first controller…")
	return s.RunTaskOnLeader(kubeadmCertsExecutor)
}

func kubeadmCertsOnFollower(s *state.State) error {
	s.Logger.Infoln("Configuring certs and etcd on consecutive controller…")
	return s.RunTaskOnFollowers(kubeadmCertsExecutor, true)
}

func kubeadmCertsExecutor(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Ensuring Certificates…")
	_, _, err := s.Runner.Run(kubeadmCertCommand, runner.TemplateVariables{
		"WORK_DIR": s.WorkDir,
		"NODE_ID":  strconv.Itoa(node.ID),
		"VERBOSE":  s.KubeAdmVerboseFlag(),
	})
	return err
}

func initKubernetesLeader(s *state.State) error {
	s.Logger.Infoln("Initializing Kubernetes on leader…")
	return s.RunTaskOnLeader(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		s.Logger.Infoln("Running kubeadm…")

		_, _, err := s.Runner.Run(kubeadmInitCommand, runner.TemplateVariables{
			"WORK_DIR":       s.WorkDir,
			"NODE_ID":        strconv.Itoa(node.ID),
			"VERBOSE":        s.KubeAdmVerboseFlag(),
			"TOKEN":          s.JoinToken,
			"TOKEN_DURATION": "1h",
		})

		return err
	})
}
