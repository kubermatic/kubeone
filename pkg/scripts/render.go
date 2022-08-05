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
	"github.com/Masterminds/sprig/v3"

	"k8c.io/kubeone/pkg/fail"
)

var (
	containerRuntimeTemplates = map[string]string{
		"container-runtime-daemon-config": heredoc.Doc(`
			{{- if .CONTAINER_RUNTIME_CONFIG_PATH }}
			sudo mkdir -p $(dirname {{ .CONTAINER_RUNTIME_CONFIG_PATH }})
			sudo touch {{ .CONTAINER_RUNTIME_CONFIG_PATH }}
			sudo chmod 600 {{ .CONTAINER_RUNTIME_CONFIG_PATH }}
			cat <<EOF | sudo tee {{ .CONTAINER_RUNTIME_CONFIG_PATH }}
			{{ .CONTAINER_RUNTIME_CONFIG }}
			EOF
			{{- end }}

			{{- if .CONTAINER_RUNTIME_SOCKET }}
			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix://{{ .CONTAINER_RUNTIME_SOCKET }}
			EOF
			{{- end }}
		`),

		"containerd-systemd-setup": heredoc.Doc(`
			sudo systemctl daemon-reload
			sudo systemctl enable containerd
			sudo systemctl restart containerd
		`),

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

			sudo apt-mark unhold docker-ce docker-ce-cli containerd.io || true
			{{- $DOCKER_VERSION_TO_INSTALL := "%s" }}

			sudo DEBIAN_FRONTEND=noninteractive apt-get install \
				--option "Dpkg::Options::=--force-confold" \
				--no-install-recommends \
				-y \
				{{- if .FORCE }}
				--allow-downgrades \
				{{- end }}
				docker-ce=5:{{ $DOCKER_VERSION_TO_INSTALL }} \
				docker-ce-cli=5:{{ $DOCKER_VERSION_TO_INSTALL }} \
				containerd.io=%s
			sudo apt-mark hold docker-ce docker-ce-cli containerd.io
			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			sudo systemctl enable --now docker
			if systemctl status kubelet 2>&1 > /dev/null; then
				sudo systemctl restart kubelet
				sleep 10
			fi
			`,
			latestDockerVersion,
			defaultContainerdVersion,
		),

		"yum-docker-ce-amzn": heredoc.Docf(`
			sudo yum versionlock delete docker containerd || true

			{{- $DOCKER_VERSION_TO_INSTALL := "%s" }}

			sudo yum install -y \
				docker-{{ $DOCKER_VERSION_TO_INSTALL }} \
				containerd.io-%s
			sudo yum versionlock add docker containerd
			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			sudo systemctl enable --now docker
			if systemctl status kubelet 2>&1 > /dev/null; then
				sudo systemctl restart kubelet
				sleep 10
			fi
		`,
			latestDockerVersion,
			defaultAmazonContainerdVersion,
		),

		"yum-docker-ce": heredoc.Docf(`
			{{- if .CONFIGURE_REPOSITORIES }}
			sudo yum install -y yum-utils
			sudo yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
			sudo yum-config-manager --save --setopt=docker-ce-stable.module_hotfixes=true >/dev/null
			{{- end }}

			sudo yum versionlock delete docker-ce docker-ce-cli containerd.io || true

			{{- $DOCKER_VERSION_TO_INSTALL := "%s" }}

			sudo yum install -y \
				docker-ce-{{ $DOCKER_VERSION_TO_INSTALL }} \
				docker-ce-cli-{{ $DOCKER_VERSION_TO_INSTALL }} \
				containerd.io-%s
			sudo yum versionlock add docker-ce docker-ce-cli containerd.io
			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			sudo systemctl enable --now docker
			if systemctl status kubelet 2>&1 > /dev/null; then
				sudo systemctl restart kubelet
				sleep 10
			fi
			`,
			latestDockerVersion,
			defaultContainerdVersion,
		),

		"apt-containerd": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			sudo apt-get update
			sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common lsb-release
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg |
				sudo apt-key add -
			sudo add-apt-repository "deb https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
			{{ end }}

			sudo apt-mark unhold containerd.io || true
			sudo DEBIAN_FRONTEND=noninteractive apt-get install \
				--option "Dpkg::Options::=--force-confold" \
				--no-install-recommends \
				{{- if .FORCE }}
				--allow-downgrades \
				{{- end }}
				-y \
				containerd.io=%s
			sudo apt-mark hold containerd.io

			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			`,
			defaultContainerdVersion,
		),

		"yum-containerd": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			sudo yum install -y yum-utils
			sudo yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
			{{- /*
			Due to DNF modules we have to do this on docker-ce repo
			More info at: https://bugzilla.redhat.com/show_bug.cgi?id=1756473
			*/}}
			sudo yum-config-manager --save --setopt=docker-ce-stable.module_hotfixes=true
			{{ end }}

			sudo yum versionlock delete containerd.io || true
			sudo yum install -y containerd.io-%s
			sudo yum versionlock add containerd.io

			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			`,
			defaultContainerdVersion,
		),

		"yum-containerd-amzn": heredoc.Docf(`
			sudo yum versionlock delete containerd || true
			sudo yum install -y containerd-%s
			sudo yum versionlock add containerd

			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			`,
			defaultAmazonContainerdVersion,
		),

		"flatcar-containerd": heredoc.Doc(`
			{{ template "container-runtime-daemon-config" . }}
			{{ template "flatcar-systemd-drop-in" . }}
			{{ template "containerd-systemd-setup" . }}
			`,
		),

		"flatcar-docker": heredoc.Doc(`
			{{ template "container-runtime-daemon-config" . }}
			sudo systemctl daemon-reload
			sudo systemctl enable --now docker
			sudo systemctl restart docker
			if systemctl status kubelet 2>&1 > /dev/null; then
				sudo systemctl restart kubelet
				sleep 10
			fi			
			`,
		),

		"flatcar-systemd-drop-in": heredoc.Doc(`
			sudo mkdir -p /etc/systemd/system/containerd.service.d
			cat <<EOF | sudo tee /etc/systemd/system/containerd.service.d/10-kubeone.conf
			[Service]
			Restart=always
			Environment=CONTAINERD_CONFIG=/etc/containerd/config.toml
			ExecStart=
			ExecStart=/usr/bin/env PATH=\${TORCX_BINDIR}:\${PATH} \${TORCX_BINDIR}/containerd --config \${CONTAINERD_CONFIG}
			EOF
		`),
	}
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(cmd string, variables map[string]interface{}) (string, error) {
	tpl := template.New("base").
		Funcs(sprig.TxtFuncMap())

	_, err := tpl.New("library").Parse(libraryTemplate)
	if err != nil {
		return "", fail.Runtime(err, "parsing library template")
	}

	for tplName, tplBody := range containerRuntimeTemplates {
		_, err = tpl.New(tplName).Parse(tplBody)
		if err != nil {
			return "", fail.Runtime(err, "parsing %s template", tplName)
		}
	}

	_, err = tpl.Parse(cmd)
	if err != nil {
		return "", fail.Runtime(err, "parsing command template")
	}

	var buf strings.Builder
	buf.WriteString("set -xeuo pipefail\n")
	buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	buf.WriteString("\n")

	if err := tpl.Execute(&buf, variables); err != nil {
		return "", fail.Runtime(err, "rendering template")
	}

	return buf.String(), nil
}
