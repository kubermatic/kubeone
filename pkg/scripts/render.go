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
			sudo yum versionlock delete docker cri-tools || true
			{{- end }}

			{{ $CRICTL_VERSION_TO_INSTALL := "%s" }}
			{{ $DOCKER_VERSION_TO_INSTALL := "%s" }}
			{{ if semverCompare "< 1.17" .KUBERNETES_VERSION }}
			{{ $DOCKER_VERSION_TO_INSTALL = "%s" }}
			{{ end }}

			sudo yum install -y docker-{{ $DOCKER_VERSION_TO_INSTALL }}ce* cri-tools-{{ $CRICTL_VERSION_TO_INSTALL }}*
			sudo yum versionlock add docker cri-tools

			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix:///var/run/dockershim.sock
			EOF

			sudo systemctl daemon-reload
			sudo systemctl enable --now docker
		`,
			defaultAmazonCrictlVersion,
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
			defaultDockerVersion,
			defaultLegacyDockerVersion,
		),

		"apt-containerd": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			sudo apt-get update
			sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common lsb-release
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg |
				sudo apt-key add -
			sudo add-apt-repository "deb https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
			{{ end }}

			{{ if or .FORCE .UPGRADE }}
			sudo apt-mark unhold containerd.io
			{{ end }}

			sudo apt-get install -y containerd.io=%s*
			sudo apt-mark hold containerd.io

			cat <<EOF | sudo tee /etc/containerd/config.toml
			{{ containerdCfg .INSECURE_REGISTRY -}}
			EOF

			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix:///run/containerd/containerd.sock
			EOF

			sudo mkdir -p /etc/systemd/system/containerd.service.d
			cat <<EOF | sudo tee /etc/systemd/system/containerd.service.d/environment.conf
			[Service]
			Restart=always
			EnvironmentFile=-/etc/environment
			EOF

			sudo systemctl daemon-reload
			sudo systemctl enable --now containerd
			sudo systemctl restart containerd
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

			sudo yum install -y yum-plugin-versionlock

			{{ if or .FORCE .UPGRADE }}
			sudo yum versionlock delete containerd.io || true
			{{- end }}

			sudo yum install -y containerd.io-%s
			sudo yum versionlock add containerd.io

			cat <<EOF | sudo tee /etc/containerd/config.toml
			{{ containerdCfg .INSECURE_REGISTRY -}}
			EOF

			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix:///run/containerd/containerd.sock
			EOF

			sudo mkdir -p /etc/systemd/system/containerd.service.d
			cat <<EOF | sudo tee /etc/systemd/system/containerd.service.d/environment.conf
			[Service]
			Restart=always
			EnvironmentFile=-/etc/environment
			EOF

			sudo systemctl daemon-reload
			sudo systemctl enable --now containerd
			sudo systemctl restart containerd
			`,
			defaultContainerdVersion,
		),

		"yum-containerd-amzn": heredoc.Docf(`
			{{- if or .FORCE .UPGRADE }}
			sudo yum versionlock delete containerd cri-tools || true
			{{- end }}

			sudo yum install -y containerd-%s* cri-tools-%s*
			sudo yum versionlock add containerd cri-tools

			cat <<EOF | sudo tee /etc/containerd/config.toml
			{{ containerdCfg .INSECURE_REGISTRY -}}
			EOF

			cat <<EOF | sudo tee /etc/crictl.yaml
			runtime-endpoint: unix:///run/containerd/containerd.sock
			EOF

			sudo mkdir -p /etc/systemd/system/containerd.service.d
			cat <<EOF | sudo tee /etc/systemd/system/containerd.service.d/environment.conf
			[Service]
			Restart=always
			EnvironmentFile=-/etc/environment
			EOF

			sudo systemctl daemon-reload
			sudo systemctl enable --now containerd
			sudo systemctl restart containerd
			`,
			defaultAmazonContainerdVersion,
			defaultAmazonCrictlVersion,
		),
	}
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(cmd string, variables map[string]interface{}) (string, error) {
	tpl := template.New("base").
		Funcs(sprig.TxtFuncMap()).
		Funcs(template.FuncMap{
			"dockerCfg":     dockerCfg,
			"containerdCfg": containerdCfg,
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
