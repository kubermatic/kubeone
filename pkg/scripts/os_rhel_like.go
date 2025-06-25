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
	kubeadmRHELLikeTemplate = `
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
LATEST_STABLE=$(curl -sL https://dl.k8s.io/release/stable.txt | sed 's/\.[0-9]*$//')
cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://pkgs.k8s.io/core:/stable:/{{ .KUBERNETES_MAJOR_MINOR }}/rpm/
enabled=1
gpgcheck=1
gpgkey=https://pkgs.k8s.io/core:/stable:/${LATEST_STABLE}/rpm/repodata/repomd.xml.key
exclude=kubelet kubeadm kubectl cri-tools kubernetes-cni
EOF

source /etc/os-release
if [ "$ID" == "centos" ] && [ "$VERSION_ID" == "8" ]; then
	sudo sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
	sudo sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
fi

# We must clean 'yum' cache upon changing the package repository
# because older 'yum' versions (e.g. CentOS and Amazon Linux 2)
# don't detect the change otherwise.
sudo yum clean all
sudo yum makecache
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

{{ if .INSTALL_CONTAINERD }}
{{ template "yum-containerd" . }}
{{ end }}

{{- if or .FORCE .UPGRADE }}
sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni cri-tools || true
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
	kubernetes-cni \
	cri-tools
sudo yum versionlock add kubelet kubeadm kubectl kubernetes-cni cri-tools

sudo systemctl daemon-reload
sudo systemctl enable --now kubelet
{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{ end }}
`

	removeBinariesRHELLikeScriptTemplate = `
sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni cri-tools || true
sudo yum remove -y \
	kubelet \
	kubeadm \
	kubectl
sudo yum remove -y kubernetes-cni cri-tools || true
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

func RHELLikeScript(cluster *kubeoneapi.KubeOneCluster, params Params) (string, error) {
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
		"KUBERNETES_MAJOR_MINOR": cluster.Versions.KubernetesMajorMinorVersion(),
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"INSTALL_CONTAINERD":     cluster.ContainerRuntime.Containerd,
		"INSTALL_ISCSI_AND_NFS":  installISCSIAndNFS(cluster),
		"IPV6_ENABLED":           cluster.ClusterNetwork.HasIPv6(),
	}

	if err := containerruntime.UpdateDataMap(cluster, data); err != nil {
		return "", err
	}

	result, err := Render(kubeadmRHELLikeTemplate, data)

	return result, fail.Runtime(err, "rendering kubeadmRHELLikeTemplate script")
}

func DisableNMCloudSetup() (string, error) {
	result, err := Render(disableNMCloudSetup, nil)

	return result, fail.Runtime(err, "rendering disableNMCloudSetup script")
}

func RemoveBinariesRHELLike() (string, error) {
	result, err := Render(removeBinariesRHELLikeScriptTemplate, Data{})

	return result, fail.Runtime(err, "rendering removeBinariesCentOSScriptTemplate script")
}
