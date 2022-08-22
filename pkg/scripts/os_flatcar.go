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

sudo mkdir -p /opt/bin /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .KUBERNETES_CNI_VERSION }}/cni-plugins-linux-${HOST_ARCH}-v{{ .KUBERNETES_CNI_VERSION }}.tgz" |
	sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"
CRI_TOOLS_RELEASE="v{{ .CRITOOLS_VERSION }}"

curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRI_TOOLS_RELEASE}/crictl-${CRI_TOOLS_RELEASE}-linux-${HOST_ARCH}.tar.gz |
	sudo tar -C /opt/bin -xz

{{ if .INSTALL_DOCKER }}
{{ template "flatcar-docker" . }}
{{ end }}

{{ if .INSTALL_CONTAINERD }}
{{ template "flatcar-containerd" . }}
{{ end }}

cd /opt/bin
k8s_rel_baseurl=https://storage.googleapis.com/kubernetes-release/release
for binary in kubeadm kubelet kubectl; do
	curl -L --output /tmp/$binary \
		$k8s_rel_baseurl/${RELEASE}/bin/linux/${HOST_ARCH}/$binary
	sudo install --owner=0 --group=0 --mode=0755 /tmp/$binary /opt/bin/$binary
	rm /tmp/$binary
done

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

	upgradeKubeadmAndCNIFlatcarScriptTemplate = `
{{ template "detect-host-cpu-architecture" }}

{{- if .INSTALL_DOCKER -}}
{{ template "flatcar-docker" . }}
{{ end }}

{{- if .INSTALL_CONTAINERD -}}
{{ template "flatcar-containerd" . }}
{{ end }}

source /etc/kubeone/proxy-env

sudo mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .KUBERNETES_CNI_VERSION }}/cni-plugins-linux-${HOST_ARCH}-v{{ .KUBERNETES_CNI_VERSION }}.tgz" |
	sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /var/tmp/kube-binaries
cd /var/tmp/kube-binaries
sudo curl -L --remote-name-all \
	https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/${HOST_ARCH}/kubeadm

sudo mkdir -p /opt/bin
cd /opt/bin
sudo mv /var/tmp/kube-binaries/kubeadm .
sudo chmod +x kubeadm
`

	upgradeKubeletAndKubectlFlatcarScriptTemplate = `
source /etc/kubeone/proxy-env

{{ template "detect-host-cpu-architecture" }}

{{- if .INSTALL_DOCKER -}}
{{ template "flatcar-docker" . }}
{{ end }}

{{- if .INSTALL_CONTAINERD -}}
{{ template "flatcar-containerd" . }}
{{ end }}

RELEASE="v{{ .KUBERNETES_VERSION }}"
sudo mkdir -p /var/tmp/kube-binaries
cd /var/tmp/kube-binaries
sudo curl -L --remote-name-all \
	https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/${HOST_ARCH}/{kubelet,kubectl}
sudo mkdir -p /opt/bin
cd /opt/bin
sudo systemctl stop kubelet
sudo mv /var/tmp/kube-binaries/{kubelet,kubectl} .
sudo chmod +x {kubelet,kubectl}

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
sudo systemctl start kubelet
`
)

func KubeadmFlatcar(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	data := Data{
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CRITOOLS_VERSION":       defaultCriToolsVersion,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"CILIUM":                 ciliumCNI(cluster),
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

func UpgradeKubeadmAndCNIFlatcar(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	data := Data{
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(upgradeKubeadmAndCNIFlatcarScriptTemplate, data)

	return result, fail.Runtime(err, "rendering upgradeKubeadmAndCNIFlatcarScriptTemplate script")
}

func UpgradeKubeletAndKubectlFlatcar(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	data := Data{
		"KUBERNETES_VERSION": cluster.Versions.Kubernetes,
		"INSTALL_DOCKER":     cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD": cluster.ContainerRuntime.Containerd,
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(upgradeKubeletAndKubectlFlatcarScriptTemplate, data)

	return result, fail.Runtime(err, "rendering upgradeKubeletAndKubectlFlatcarScriptTemplate script")
}
