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

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/containerruntime"
)

const (
	defaultKubernetesCNIVersion = "0.8.7"
	defaultCriToolsVersion      = "1.21.0"
)

var migrateToContainerdScriptTemplate = heredoc.Doc(`
	sudo systemctl stop kubelet
	sudo docker ps -q | xargs sudo docker stop || true
	sudo docker ps -qa | xargs sudo docker rm || true

	{{ template "flatcar-containerd" . }}

	{{- /*
		/var/lib/kubelet/kubeadm-flags.env should be modified by the caller of
		this script, following flags should be added:
		* --container-runtime=remote
		* --container-runtime-endpoint=unix:///run/containerd/containerd.sock
	*/ -}}

	sudo systemctl restart kubelet
`)

func MigrateToContainerd(cluster *kubeone.KubeOneCluster) (string, error) {
	data := Data{}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	return Render(migrateToContainerdScriptTemplate, data)
}
