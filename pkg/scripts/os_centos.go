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
	kubeadmCentOSTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo setenforce 0 || true
[ -f /etc/selinux/config ] && sudo sed -i 's/SELINUX=enforcing/SELINUX=permissive/g' /etc/selinux/config
sudo systemctl disable --now firewalld || true

source /etc/kubeone/proxy-env

{{ template "sysctl-k8s" . }}
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

source /etc/os-release
if [ "$ID" == "centos" ] && [ "$VERSION_ID" == "8" ]; then
	sudo sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
	sudo sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
fi
{{ end }}

sudo yum install -y \
	yum-plugin-versionlock \
	device-mapper-persistent-data \
	lvm2 \
	conntrack-tools \
	ebtables \
	socat \
	iproute-tc \
	{{- if .INSTALL_ISCSI_AND_NFS }}
	iscsi-initiator-utils \
	nfs-utils \
	{{- end }}
	rsync

{{- if .INSTALL_ISCSI_AND_NFS }}
sudo systemctl enable --now iscsid
{{- end }}

{{ if .INSTALL_DOCKER }}
{{ template "yum-docker-ce" . }}
{{ end }}

{{ if .INSTALL_CONTAINERD }}
{{ template "yum-containerd" . }}
{{ end }}

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

sudo systemctl daemon-reload
sudo systemctl enable --now kubelet
{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{ end }}
`
	removeBinariesCentOSScriptTemplate = `
sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni || true
sudo yum remove -y \
	kubelet \
	kubeadm \
	kubectl
sudo yum remove -y kubernetes-cni || true
sudo rm -rf /opt/cni
sudo rm -f /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
sudo systemctl daemon-reload
`
	disableNMCloudSetup = `
if systemctl status 'nm-cloud-setup.timer' 2> /dev/null | grep -Fq "Active: active"; then
sudo systemctl stop nm-cloud-setup.timer
sudo systemctl disable nm-cloud-setup.service
sudo systemctl disable nm-cloud-setup.timer
sudo reboot
fi
`
)

func KubeadmCentOS(cluster *kubeoneapi.KubeOneCluster, force bool) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"KUBELET":                true,
		"KUBEADM":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"FORCE":                  force,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmCentOSTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmCentOSTemplate script")
}

func RemoveBinariesCentOS() (string, error) {
	result, err := Render(removeBinariesCentOSScriptTemplate, Data{})

	return result, fail.Runtime(err, "rendering removeBinariesCentOSScriptTemplate script")
}

func UpgradeKubeadmAndCNICentOS(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"UPGRADE":                true,
		"KUBEADM":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmCentOSTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmCentOSTemplate script")
}

func UpgradeKubeletAndKubectlCentOS(cluster *kubeoneapi.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	data := Data{
		"UPGRADE":                true,
		"KUBELET":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"INSTALL_DOCKER":         cluster.ContainerRuntime.Docker,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"CILIUM":                 ciliumCNI(cluster),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmCentOSTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmCentOSTemplate script")
}

func DisableNMCloudSetup() (string, error) {
	result, err := Render(disableNMCloudSetup, nil)

	return result, fail.Runtime(err, "rendering disableNMCloudSetup script")
}
