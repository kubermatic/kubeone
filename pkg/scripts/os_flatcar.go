/*
Copyright 2021 The KubeOne Authors.

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
	"github.com/Masterminds/semver/v3"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/containerruntime"
	"k8c.io/kubeone/pkg/fail"
)

const (
	kubeadmFlatcarTemplate = `
source /etc/kubeone/proxy-env

{{ template "detect-host-cpu-architecture" }}
{{ template "sysctl-k8s" . }}
{{ template "journald-config" }}

{{- with .KUBERNETES_CNI_VERSION }}
sudo mkdir -p /opt/bin /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ . }}/cni-plugins-linux-${HOST_ARCH}-v{{ . }}.tgz" |
	sudo tar -C /opt/cni/bin -xz
sudo chown -R root:root /opt/cni/bin
{{- end }}
{{- with .CRITOOLS_VERSION }}
CRI_TOOLS_RELEASE="v{{ . }}"
curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRI_TOOLS_RELEASE}/crictl-${CRI_TOOLS_RELEASE}-linux-${HOST_ARCH}.tar.gz |
	sudo tar -C /opt/bin -xz
{{- end }}

{{- if .INSTALL_CONTAINERD }}
{{ template "flatcar-containerd" . }}
{{- end }}
binaries=()
{{- with .KUBELET }}
binaries+=('kubelet')
{{- end }}
{{- with .KUBECTL }}
binaries+=('kubectl')
{{- end }}
{{- with .KUBEADM }}
binaries+=('kubeadm')
{{- end }}

RELEASE="v{{ .KUBERNETES_VERSION }}"
for binary in "${binaries[@]}" ; do
	curl \
		--location \
		--output "/tmp/${binary}" \
		"https://dl.k8s.io/release/${RELEASE}/bin/linux/${HOST_ARCH}/${binary}"
	sudo install --owner=0 --group=0 --mode=0755 "/tmp/${binary}" "/opt/bin/${binary}"
	rm "/tmp/${binary}"
done

{{- with .KUBELET }}
cat <<EOF | sudo tee /etc/systemd/system/kubelet.service
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/home/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/opt/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo mkdir -p /etc/systemd/system/kubelet.service.d
cat <<EOF | sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/default/kubelet
ExecStart=
ExecStart=/opt/bin/kubelet \$KUBELET_KUBECONFIG_ARGS \$KUBELET_CONFIG_ARGS \$KUBELET_KUBEADM_ARGS \$KUBELET_EXTRA_ARGS
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now kubelet
{{- end }}

{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{- end }}
`

	removeBinariesFlatcarScriptTemplate = `
# Stop kubelet
sudo systemctl stop kubelet || true
# Remove CNI and binaries
sudo rm -rf /opt/cni /opt/bin/kubeadm /opt/bin/kubectl /opt/bin/kubelet
# Remove systemd unit files
sudo rm -f /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
# Reload systemd
sudo systemctl daemon-reload
`
)

func FlatcarScript(cluster *kubeoneapi.KubeOneCluster, params Params) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"UPGRADE":                params.Upgrade,
		"KUBELET":                params.Kubelet,
		"KUBECTL":                params.Kubectl,
		"KUBEADM":                params.Kubeadm,
		"FORCE":                  params.Force,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": flatcarCNIVersion(cluster.Versions.Kubernetes),
		"CRITOOLS_VERSION":       criToolsVersion(cluster.Versions.Kubernetes),
		"PROXY":                  proxy,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"IPV6_ENABLED":           cluster.ClusterNetwork.HasIPv6(),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmFlatcarTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmFlatcarTemplate script")
}

func RemoveBinariesFlatcar() (string, error) {
	result, err := Render(removeBinariesFlatcarScriptTemplate, nil)

	return result, fail.Runtime(err, "rendering removeBinariesFlatcarScriptTemplate script")
}

func flatcarCNIVersion(kubeVersion string) string {
	kubeSemVer := semver.MustParse(kubeVersion)

	switch kubeSemVer.Minor() {
	case 31:
		return "1.5.1"
	case 32:
		return "1.6.0"
	case 33:
		return "1.6.0"
	case 34:
		return "1.7.1"
	default:
		return "1.7.1"
	}
}

func criToolsVersion(kubeVersion string) string {
	// Validation passed at this point so we know that version is valid
	kubeSemVer := semver.MustParse(kubeVersion)

	switch kubeSemVer.Minor() {
	case 31:
		return "1.31.1"
	case 32:
		return "1.32.0"
	case 33:
		return "1.33.0"
	case 34:
		return "1.34.0"
	default:
		return "1.34.0"
	}
}
