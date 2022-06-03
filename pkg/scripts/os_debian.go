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
	kubeadmDebianTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo systemctl disable --now ufw || true

source /etc/kubeone/proxy-env

{{ template "sysctl-k8s" . }}
{{ template "journald-config" }}

sudo mkdir -p /etc/apt/apt.conf.d
cat <<EOF | sudo tee /etc/apt/apt.conf.d/proxy.conf
{{- if .HTTPS_PROXY }}
Acquire::https::Proxy "{{ .HTTPS_PROXY }}";
{{- end }}
{{- if .HTTP_PROXY }}
Acquire::http::Proxy "{{ .HTTP_PROXY }}";
{{- end }}
EOF

sudo apt-get update
sudo DEBIAN_FRONTEND=noninteractive apt-get install --option "Dpkg::Options::=--force-confold" -y --no-install-recommends \
	apt-transport-https \
	ca-certificates \
	curl \
	gnupg \
	lsb-release \
	{{- if .INSTALL_ISCSI_AND_NFS }}
	open-iscsi \
	nfs-common \
	{{- end }}
	rsync

{{- if .INSTALL_ISCSI_AND_NFS }}
sudo systemctl enable --now iscsid
{{- end }}

{{- if .CONFIGURE_REPOSITORIES }}
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

# You'd think that kubernetes-$(lsb_release -sc) belongs there instead, but the debian repo
# contains neither kubeadm nor kubelet, and the docs themselves suggest using xenial repo.
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
{{- end }}

kube_ver="{{ .KUBERNETES_VERSION }}*"
cni_ver="{{ .KUBERNETES_CNI_VERSION }}*"

{{- if or .FORCE .UPGRADE }}
sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
{{- end }}

{{ if .INSTALL_DOCKER }}
{{ template "apt-docker-ce" . }}
{{ end }}

{{ if .INSTALL_CONTAINERD }}
{{ template "apt-containerd" . }}
{{ end }}

sudo DEBIAN_FRONTEND=noninteractive apt-get install \
	--option "Dpkg::Options::=--force-confold" \
	--no-install-recommends \
	{{- if .FORCE }}
	--allow-downgrades \
	{{- end }}
	-y \
{{- if .KUBELET }}
	kubelet=${kube_ver} \
{{- end }}
{{- if .KUBEADM }}
	kubeadm=${kube_ver} \
{{- end }}
{{- if .KUBECTL }}
	kubectl=${kube_ver} \
{{- end }}
	kubernetes-cni=${cni_ver}

sudo apt-mark hold kubelet kubeadm kubectl kubernetes-cni

sudo systemctl daemon-reload
sudo systemctl enable --now kubelet

{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{- end }}
`

	removeBinariesDebianScriptTemplate = `
sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
sudo apt-get remove --purge -y \
	kubeadm \
	kubectl \
	kubelet
sudo apt-get remove --purge -y kubernetes-cni || true
sudo rm -rf /opt/cni
sudo rm -f /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo systemctl daemon-reload
`
)

func KubeadmDebian(cluster *kubeoneapi.KubeOneCluster, force bool) (string, error) {
	data := Data{
		"KUBELET":                true,
		"KUBEADM":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
		"FORCE":                  force,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmDebianTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmDebianTemplate script")
}

func RemoveBinariesDebian() (string, error) {
	result, err := Render(removeBinariesDebianScriptTemplate, Data{})

	return result, fail.Runtime(err, "rendering removeBinariesDebianScriptTemplate script")
}

func UpgradeKubeadmAndCNIDebian(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	data := Data{
		"UPGRADE":                true,
		"KUBEADM":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmDebianTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmDebianTemplate script")
}

func UpgradeKubeletAndKubectlDebian(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	data := Data{
		"UPGRADE":                true,
		"KUBELET":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmDebianTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmDebianTemplate script")
}
