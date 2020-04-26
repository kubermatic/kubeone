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
	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

const (
	daemonsProxyScript = `
sudo mkdir -p /etc/systemd/system/docker.service.d
cat <<EOF | sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf
[Service]
EnvironmentFile=/etc/kubeone/proxy-env
EOF

sudo mkdir -p /etc/systemd/system/kubelet.service.d
cat <<EOF | sudo tee /etc/systemd/system/kubelet.service.d/http-proxy.conf
[Service]
EnvironmentFile=/etc/kubeone/proxy-env
EOF

sudo systemctl daemon-reload
if sudo systemctl status docker &>/dev/null; then sudo systemctl restart docker; fi
if sudo systemctl status kubelet &>/dev/null; then sudo systemctl restart kubelet; fi
`

	environmentFileCmd = `
sudo mkdir -p /etc/kubeone
cat <<EOF | sudo tee /etc/kubeone/proxy-env
{{ with .HTTP_PROXY -}}
HTTP_PROXY="{{ . }}"
http_proxy="{{ . }}"
export HTTP_PROXY http_proxy
{{ end }}

{{- with .HTTPS_PROXY -}}
HTTPS_PROXY="{{ . }}"
https_proxy="{{ . }}"
export HTTPS_PROXY https_proxy
{{ end }}

{{- with .NO_PROXY -}}
NO_PROXY="{{ . }}"
no_proxy="{{ . }}"
export NO_PROXY no_proxy
{{ end }}
EOF

envtmp=/tmp/k1-etc-environment
grep -v '#kubeone$' /etc/environment > $envtmp || true
set +o pipefail # grep exits non-zero without match
grep = /etc/kubeone/proxy-env | sed 's/$/#kubeone/' >> $envtmp
sudo tee /etc/environment < $envtmp
`
)

func EnvironmentFile(cluster *kubeone.KubeOneCluster) (string, error) {
	return Render(environmentFileCmd, Data{
		"HTTP_PROXY":  cluster.Proxy.HTTP,
		"HTTPS_PROXY": cluster.Proxy.HTTPS,
		"NO_PROXY":    cluster.Proxy.NoProxy,
	})
}

func DaemonsProxy() (string, error) {
	return Render(daemonsProxyScript, nil)
}
