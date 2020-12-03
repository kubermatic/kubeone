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
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
)

var (
	containerRuntimeTemplates = map[string]string{
		"apt-docker-ce": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
			# Docker provides two different apt repos for ubuntu, bionic and focal. The focal repo currently
			# contains only Docker 19.03.14, which is not validated for all Kubernetes version.
			# Therefore, we use bionic repo which has all Docker versions.
			echo "deb https://download.docker.com/linux/ubuntu bionic stable" |
				sudo tee /etc/apt/sources.list.d/docker.list
			sudo apt-get update
			{{ end }}

			{{- if or .FORCE .UPGRADE }}
			sudo apt-mark unhold docker-ce docker-ce-cli
			{{- end }}

			{{ $DOCKER_VERSION_TO_INSTALL := "%s" }}
			{{ if semverCompare "< 1.17" .KUBERNETES_VERSION }}
			{{ $DOCKER_VERSION_TO_INSTALL = "%s" }}
			{{ end }}

			sudo DEBIAN_FRONTEND=noninteractive apt-get install \
				--option "Dpkg::Options::=--force-confold" \
				--no-install-recommends \
				-y \
				{{- if .FORCE }}
				--allow-downgrades \
				{{- end }}
				docker-ce=5:{{ $DOCKER_VERSION_TO_INSTALL }}* docker-ce-cli=5:{{ $DOCKER_VERSION_TO_INSTALL }}*
			sudo apt-mark hold docker-ce docker-ce-cli
			sudo systemctl daemon-reload
			sudo systemctl enable --now docker
			`,
			defaultDockerVersion,
			defaultLegacyDockerVersion,
		),

		"yum-docker-ce-amzn": heredoc.Docf(`
			{{- if or .FORCE .UPGRADE }}
			sudo yum versionlock delete docker || true
			{{- end }}

			{{ $DOCKER_VERSION_TO_INSTALL := "%s" }}
			{{ if semverCompare "< 1.17" .KUBERNETES_VERSION }}
			{{ $DOCKER_VERSION_TO_INSTALL = "%s" }}
			{{ end }}

			sudo yum install -y docker-{{ $DOCKER_VERSION_TO_INSTALL }}ce*
			sudo yum versionlock add docker

			sudo systemctl daemon-reload
			sudo systemctl enable --now docker
		`,
			defaultAmazonDockerVersion,
			defaultLegacyDockerVersion,
		),

		"yum-docker-ce": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			sudo yum install -y yum-utils
			sudo yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
			sudo yum-config-manager --save --setopt=docker-ce-stable.module_hotfixes=true >/dev/null
			# Docker provides two different apt repos for CentOS, 7 and 8. The 8 repo currently
			# contains only Docker 19.03.14, which is not validated for all Kubernetes version.
			# Therefore, we use 7 repo which has all Docker versions.
			sudo sed -i 's/\$releasever/7/g' /etc/yum.repos.d/docker-ce.repo
			{{ end }}

			{{ if or .FORCE .UPGRADE }}
			sudo yum versionlock delete docker-ce docker-ce-cli || true
			{{- end }}

			{{ $DOCKER_VERSION_TO_INSTALL := "%s" }}
			{{ if semverCompare "< 1.17" .KUBERNETES_VERSION }}
			{{ $DOCKER_VERSION_TO_INSTALL = "%s" }}
			{{ end }}

			sudo yum install -y docker-ce-{{ $DOCKER_VERSION_TO_INSTALL }}* docker-ce-cli-{{ $DOCKER_VERSION_TO_INSTALL }}*
			sudo yum versionlock add docker-ce docker-ce-cli
			sudo systemctl daemon-reload
			sudo systemctl enable --now docker
			`,
			defaultLegacyDockerVersion,
			defaultDockerVersion,
		),

		"containerd-github": heredoc.Docf(`
			{{ $CONTAINERD_VERSION := "%s" }}
			mkdir -p /tmp/containerd-{{ $CONTAINERD_VERSION }}
			pushd /tmp/containerd-{{ $CONTAINERD_VERSION }}
			curl --location --continue-at - \
				--output containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz.sha256sum \
				https://github.com/containerd/containerd/releases/download/v{{ $CONTAINERD_VERSION }}/containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz.sha256sum \
				--output containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz \
				https://github.com/containerd/containerd/releases/download/v{{ $CONTAINERD_VERSION }}/containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz
			sha256sum -c containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz.sha256sum
			tar xvf containerd-{{ $CONTAINERD_VERSION }}-linux-amd64.tar.gz
			sudo install --group=0 --owner=0 --preserve-timestamps ./bin/* --target-directory=/usr/local/bin/
			sudo mkdir -p /etc/containerd /etc/cni/net.d /opt/cni/bin

			cat <<EOF | sudo tee /etc/containerd/config.toml
			version = 2

			[metrics]
			  # metrics available at http://127.0.0.1:1338/v1/metrics
			  address = "127.0.0.1:1338"

			[plugins]
			[plugins."io.containerd.grpc.v1.cri"]
			[plugins."io.containerd.grpc.v1.cri".containerd]
			[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
			[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
			    runtime_type = "io.containerd.runc.v2"
			[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
			    SystemdCgroup = true
			EOF

			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix:///run/containerd/containerd.sock
			EOF

			cat <<EOF | sudo tee /etc/systemd/system/containerd.service
			[Unit]
			Description=containerd container runtime
			Documentation=https://containerd.io
			After=network.target local-fs.target

			[Service]
			ExecStartPre=-/sbin/modprobe overlay
			ExecStart=/usr/local/bin/containerd

			Type=notify
			Delegate=yes
			KillMode=process
			Restart=always
			RestartSec=5
			# Having non-zero Limit*s causes performance problems due to accounting overhead
			# in the kernel. We recommend using cgroups to do container-local accounting.
			LimitNPROC=infinity
			LimitCORE=infinity
			LimitNOFILE=1048576
			# Comment TasksMax if your systemd version does not supports it.
			# Only systemd 226 and above support this version.
			TasksMax=infinity
			OOMScoreAdjust=-999

			[Install]
			WantedBy=multi-user.target
			EOF

			sudo systemctl daemon-reload
			sudo systemctl enable --now containerd
			sudo systemctl restart containerd
			popd
			`,
			defaultContainerdVersion,
		),
	}
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(cmd string, variables map[string]interface{}) (string, error) {
	tpl := template.New("base").
		Funcs(sprig.TxtFuncMap()).
		Funcs(template.FuncMap{
			"dockerCfg": dockerCfg,
		})

	_, err := tpl.New("library").Parse(libraryTemplate)
	if err != nil {
		return "", err
	}

	for tplName, tplBody := range containerRuntimeTemplates {
		_, err = tpl.New(tplName).Parse(tplBody)
		if err != nil {
			return "", err
		}
	}

	_, err = tpl.Parse(cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse script template")
	}

	var buf strings.Builder
	buf.WriteString("set -xeu pipefail\n")
	buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	buf.WriteString("\n")

	if err := tpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to render script template")
	}

	return buf.String(), nil
}
