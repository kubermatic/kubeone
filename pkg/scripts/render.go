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

var containerRuntimeTemplates = map[string]string{
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

	"apt-containerd": heredoc.Docf(`
			{{ if .CONFIGURE_REPOSITORIES }}
			sudo apt-get update
			sudo apt-get install -y apt-transport-https ca-certificates curl lsb-release
			sudo install -m 0755 -d /etc/apt/keyrings
			sudo rm -f /etc/apt/keyrings/docker.gpg
			curl -fsSL https://download.docker.com/linux/$(lsb_release -si | tr '[:upper:]' '[:lower:]')/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
			sudo chmod a+r /etc/apt/keyrings/docker.gpg

			echo "deb [signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(lsb_release -si | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list
			sudo apt-get update
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
			sudo yum-config-manager --add-repo=https://download.docker.com/linux/rhel/docker-ce.repo
			{{ end }}

			sudo yum versionlock delete containerd.io || true
			sudo yum install -y containerd.io-%s
			sudo yum versionlock add containerd.io

			{{ template "container-runtime-daemon-config" . }}
			{{ template "containerd-systemd-setup" . -}}
			`,
		defaultContainerdVersion,
	),

	"flatcar-containerd": heredoc.Doc(`
			{{ template "container-runtime-daemon-config" . }}
			{{ template "flatcar-systemd-drop-in" . }}
			{{ template "containerd-systemd-setup" . }}
			`,
	),

	"flatcar-systemd-drop-in": heredoc.Doc(`
			sudo mkdir -p /etc/systemd/system/containerd.service.d
			cat <<EOF | sudo tee /etc/systemd/system/containerd.service.d/10-kubeone.conf
			[Service]
			Restart=always
			Environment=CONTAINERD_CONFIG=/etc/containerd/config.toml
			ExecStart=
			ExecStart=/usr/bin/env PATH=\${TORCX_BINDIR}:\${PATH} containerd --config \${CONTAINERD_CONFIG}
			EOF
		`),
}

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

type Params struct {
	Force   bool
	Upgrade bool
	Kubelet bool
	Kubectl bool
	Kubeadm bool
}

func (params Params) String() string {
	var str strings.Builder
	var components []string

	if params.Upgrade {
		str.WriteString("upgrading ")
	} else {
		str.WriteString("installing ")
	}

	if params.Kubeadm {
		components = append(components, "kubeadm")
	}
	if params.Kubectl {
		components = append(components, "kubectl")
	}
	if params.Kubelet {
		components = append(components, "kubelet")
	}

	if len(components) == 0 {
		str.WriteString("nothing ")
	} else {
		str.WriteString(strings.Join(components, " and "))
		str.WriteString(" ")
	}

	if params.Force {
		str.WriteString("using force")
	}

	return strings.TrimSpace(str.String())
}
