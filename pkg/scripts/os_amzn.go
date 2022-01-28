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
)

const (
	kubeadmAmazonLinuxTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo setenforce 0 || true
[ -f /etc/selinux/config ] && sudo sed -i 's/SELINUX=enforcing/SELINUX=permissive/g' /etc/selinux/config
sudo systemctl disable --now firewalld || true

source /etc/kubeone/proxy-env

{{ template "sysctl-k8s" }}
{{ template "journald-config" }}

yum_proxy=""
{{- if .PROXY }}
yum_proxy="proxy={{ .PROXY }} #kubeone"
{{ end }}
grep -v '#kubeone' /etc/yum.conf > /tmp/yum.conf || true
echo -n "${yum_proxy}" >> /tmp/yum.conf
sudo mv /tmp/yum.conf /etc/yum.conf

{{ if .CONFIGURE_REPOSITORIES }}
cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
{{ end }}

sudo yum install -y \
	yum-plugin-versionlock \
	device-mapper-persistent-data \
	lvm2 \
	conntrack-tools \
	ebtables \
	socat \
	iproute-tc \
	rsync

{{ if .INSTALL_DOCKER }}
{{ template "yum-docker-ce-amzn" . }}
{{ end }}

{{ if .INSTALL_CONTAINERD }}
{{ template "yum-containerd-amzn" . }}
{{ end }}

sudo mkdir -p /opt/bin /etc/kubernetes/pki /etc/kubernetes/manifests

rm -rf /tmp/k8s-binaries
mkdir -p /tmp/k8s-binaries
cd /tmp/k8s-binaries

{{- if .CNI_URL }}
sudo mkdir -p /opt/cni/bin
curl -L "{{ .CNI_URL }}" | sudo tar -C /opt/cni/bin -xz
{{- end }}

{{- if .NODE_BINARIES_URL }}
curl -L --output /tmp/k8s-binaries/node.tar.gz {{ .NODE_BINARIES_URL }}
tar xvf node.tar.gz
{{- end }}

{{- if and .KUBELET .NODE_BINARIES_URL }}
sudo install --owner=0 --group=0 --mode=0755 /tmp/k8s-binaries/kubernetes/node/bin/kubelet /opt/bin/kubelet
sudo ln -sf /opt/bin/kubelet /usr/bin/
rm /tmp/k8s-binaries/kubernetes/node/bin/kubelet

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
{{- end }}

{{- if and .KUBEADM .NODE_BINARIES_URL }}
sudo install --owner=0 --group=0 --mode=0755 /tmp/k8s-binaries/kubernetes/node/bin/kubeadm /opt/bin/kubeadm
sudo ln -sf /opt/bin/kubeadm /usr/bin/
rm /tmp/k8s-binaries/kubernetes/node/bin/kubeadm
sudo yum install -y cri-tools
{{- end }}

{{- if and .KUBECTL .KUBECTL_URL }}
curl -L --output /tmp/k8s-binaries/kubectl {{ .KUBECTL_URL }}
sudo install --owner=0 --group=0 --mode=0755 /tmp/k8s-binaries/kubectl /opt/bin/kubectl
sudo ln -sf /opt/bin/kubectl /usr/bin/
rm /tmp/k8s-binaries/kubectl
{{- end }}

{{ if .USE_KUBERNETES_REPO }}
{{- if or .FORCE .UPGRADE }}
sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni || true
{{- end }}

sudo yum install -y \
{{- if .KUBELET }}
	kubelet-{{ .KUBERNETES_VERSION }} \
{{- end }}
{{- if .KUBEADM }}
	kubeadm-{{ .KUBERNETES_VERSION }} \
{{- end }}
{{- if .KUBECTL }}
	kubectl-{{ .KUBERNETES_VERSION }} \
{{- end }}
	kubernetes-cni-{{ .KUBERNETES_CNI_VERSION }}
sudo yum versionlock add kubelet kubeadm kubectl kubernetes-cni
{{- end }}

sudo systemctl daemon-reload
sudo systemctl enable --now kubelet

{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{- end }}
`

	removeBinariesAmazonLinuxScriptTemplate = `
sudo systemctl stop kubelet || true

sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni
sudo yum remove -y \
	kubelet \
	kubeadm \
	kubectl \
	kubernetes-cni

# Stop kubelet
# Remove CNI and binaries
sudo rm -rf /opt/cni /opt/bin/kubeadm /opt/bin/kubectl /opt/bin/kubelet
# Remove symlinks
sudo rm -rf /usr/bin/kubeadm /usr/bin/kubectl /usr/bin/kubelet
# Remove systemd unit files
sudo rm -f /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

# Reload systemd
sudo systemctl daemon-reload
`
)

func KubeadmAmazonLinux(cluster *kubeoneapi.KubeOneCluster, force bool) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"KUBELET":                true,
		"KUBEADM":                true,
		"KUBECTL":                true,
		"NODE_BINARIES_URL":      cluster.AssetConfiguration.NodeBinaries.URL,
		"CNI_URL":                cluster.AssetConfiguration.CNI.URL,
		"KUBECTL_URL":            cluster.AssetConfiguration.Kubectl.URL,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"FORCE":                  force,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"USE_KUBERNETES_REPO":    cluster.AssetConfiguration.NodeBinaries.URL == "",
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	return Render(kubeadmAmazonLinuxTemplate, data)
}

func RemoveBinariesAmazonLinux() (string, error) {
	return Render(removeBinariesAmazonLinuxScriptTemplate, Data{})
}

func UpgradeKubeadmAndCNIAmazonLinux(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"UPGRADE":                true,
		"KUBEADM":                true,
		"NODE_BINARIES_URL":      cluster.AssetConfiguration.NodeBinaries.URL,
		"CNI_URL":                cluster.AssetConfiguration.CNI.URL,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"USE_KUBERNETES_REPO":    cluster.AssetConfiguration.NodeBinaries.URL == "",
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	return Render(kubeadmAmazonLinuxTemplate, data)
}

func UpgradeKubeletAndKubectlAmazonLinux(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"UPGRADE":                true,
		"KUBELET":                true,
		"KUBECTL":                true,
		"NODE_BINARIES_URL":      cluster.AssetConfiguration.NodeBinaries.URL,
		"KUBECTL_URL":            cluster.AssetConfiguration.Kubectl.URL,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"USE_KUBERNETES_REPO":    cluster.AssetConfiguration.NodeBinaries.URL == "",
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	return Render(kubeadmAmazonLinuxTemplate, data)
}
