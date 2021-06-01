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

package scripts

import (
	"github.com/MakeNowJust/heredoc/v2"
)

const (
	defaultKubernetesCNIVersion = "0.8.7"
	defaultCriToolsVersion      = "1.21.0"
)

var migrateToContainerdScriptTemplate = heredoc.Doc(`
	sudo systemctl stop kubelet
	sudo docker ps -q | xargs sudo docker stop
	sudo docker ps -qa | xargs sudo docker rm
	sudo systemctl disable --now docker

	{{ template "containerd-config" . }}

	source /var/lib/kubelet/kubeadm-flags.env
	KUBELET_EXTRA_ARGS="$KUBELET_EXTRA_ARGS --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock"

	cat <<EOF | sudo tee /var/lib/kubelet/kubeadm-flags.env
	KUBELET_EXTRA_ARGS="$KUBELET_EXTRA_ARGS"
	EOF

	sudo systemctl restart kubelet
`)

func MigrateToContainerd(insecureRegistry string) (string, error) {
	return Render(migrateToContainerdScriptTemplate, Data{
		"INSECURE_REGISTRY": insecureRegistry,
	})
}
