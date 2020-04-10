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
	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

const (
	centosDockerVersion = "18.09.9-3.el7"
)

const (
	kubeadmDebianTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab

. /etc/os-release
. /etc/kubeone/proxy-env

HOST_ARCH=""
case $(uname -m) in
x86_64)
    HOST_ARCH="amd64"
    ;;
aarch64)
    HOST_ARCH="arm64"
    ;;
*)
    echo "unsupported CPU architecture, exiting"
    exit 1
    ;;
esac

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2",
	"log-driver": "json-file",
	"log-opts": {
		"max-size": "100m"
	}
}
EOF

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
	htop \
	lsb-release \
	rsync

{{ if .CONFIGURE_REPOSITORIES }}
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
curl -fsSL https://download.docker.com/linux/${ID}/gpg | sudo apt-key add -

echo "deb https://download.docker.com/linux/${ID} $(lsb_release -sc) stable" |
	sudo tee /etc/apt/sources.list.d/docker.list

# You'd think that kubernetes-$(lsb_release -sc) belongs there instead, but the debian repo
# contains neither kubeadm nor kubelet, and the docs themselves suggest using xenial repo.
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
{{ end }}

docker_ver=$(apt-cache madison docker-ce | grep "{{ .DOCKER_VERSION }}" | head -1 | awk '{print $3}')
kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold docker-ce kubelet kubeadm kubectl kubernetes-cni
sudo DEBIAN_FRONTEND=noninteractive apt-get install --option "Dpkg::Options::=--force-confold" -y --no-install-recommends \
	docker-ce=${docker_ver} \
	kubeadm=${kube_ver} \
	kubectl=${kube_ver} \
	kubelet=${kube_ver} \
	kubernetes-cni=${cni_ver}
sudo apt-mark hold docker-ce kubelet kubeadm kubectl kubernetes-cni
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	kubeadmCentOSTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo setenforce 0 || true
sudo sed -i 's/SELINUX=enforcing/SELINUX=permissive/g' /etc/sysconfig/selinux

. /etc/kubeone/proxy-env

sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2",
	"log-driver": "json-file",
	"log-opts": {
		"max-size": "100m"
	}
}
EOF

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system

yum_proxy=""
{{ if .PROXY }}
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
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kube*
EOF
{{ end }}

sudo yum install -y yum-utils
sudo yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo

sudo yum install -y --disableexcludes=kubernetes \
	docker-ce-{{ .DOCKER_VERSION }} \
	kubelet-{{ .KUBERNETES_VERSION }}-0 \
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0

sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	kubeadmCoreOSTemplate = `
. /etc/kubeone/proxy-env

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2",
	"log-driver": "json-file",
	"log-opts": {
		"max-size": "100m"
	}
}
EOF
sudo systemctl restart docker

sudo mkdir -p /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .CNI_VERSION }}/cni-plugins-${HOST_ARCH}-v{{ .CNI_VERSION }}.tgz" |
	sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /opt/bin
cd /opt/bin
k8s_rel_baseurl=https://storage.googleapis.com/kubernetes-release/release
for binary in kubeadm kubelet kubectl; do
	curl -L --output /tmp/$binary \
		$k8s_rel_baseurl/${RELEASE}/bin/linux/${HOST_ARCH}/$binary
	sudo install --owner=0 --group=0 --mode=0755 /tmp/$binary /opt/bin/$binary
	rm /tmp/$binary
done

curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/kubelet.service" |
	sed "s:/usr/bin:/opt/bin:g" |
	sudo tee /etc/systemd/system/kubelet.service

sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/10-kubeadm.conf" |
	sed "s:/usr/bin:/opt/bin:g" |
	sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

sudo systemctl daemon-reload
sudo systemctl enable docker.service kubelet.service
sudo systemctl start docker.service kubelet.service
`

	removeBinariesDebianScriptTemplate = `
kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
sudo apt-get remove --purge -y \
	kubeadm=${kube_ver} \
	kubectl=${kube_ver} \
	kubelet=${kube_ver} \
	kubernetes-cni=${cni_ver}
`

	removeBinariesCentOSScriptTemplate = `
sudo yum remove -y \
	kubelet-{{ .KUBERNETES_VERSION }}-0\
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0
`

	removeBinariesCoreOSScriptTemplate = `
# Remove CNI and binaries
sudo rm -rf /opt/cni /opt/bin/kubeadm /opt/bin/kubectl /opt/bin/kubelet
# Remove systemd unit files
sudo rm /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
`

	upgradeKubeadmAndCNIDebianScriptTemplate = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubeadm | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubeadm kubernetes-cni
sudo DEBIAN_FRONTEND=noninteractive apt-get install --option "Dpkg::Options::=--force-confold" -y --no-install-recommends \
	kubeadm=${kube_ver} \
	kubernetes-cni=${cni_ver}
sudo apt-mark hold kubeadm kubernetes-cni
`

	upgradeKubeadmAndCNICentOSScriptTemplate = `
source /etc/kubeone/proxy-env

sudo yum install -y --disableexcludes=kubernetes \
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0
`
	upgradeKubeadmAndCNICoreOSScriptTemplate = `
source /etc/kubeone/proxy-env

sudo mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .CNI_VERSION }}/cni-plugins-${HOST_ARCH}-v{{ .CNI_VERSION }}.tgz" |
	sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /var/tmp/kube-binaries
cd /var/tmp/kube-binaries
sudo curl -L --remote-name-all \
	https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/${HOST_ARCH}/kubeadm

sudo mkdir -p /opt/bin
cd /opt/bin
sudo systemctl stop kubelet
sudo mv /var/tmp/kube-binaries/kubeadm .
sudo chmod +x kubeadm
`

	upgradeKubeletAndKubectlDebianScriptTemplate = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet kubectl
sudo DEBIAN_FRONTEND=noninteractive apt-get install --option "Dpkg::Options::=--force-confold" -y --no-install-recommends \
	kubelet=${kube_ver} \
	kubectl=${kube_ver}
sudo apt-mark hold kubelet kubectl
`

	upgradeKubeletAndKubectlCentOSScriptTemplate = `
source /etc/kubeone/proxy-env

sudo yum install -y --disableexcludes=kubernetes \
	kubelet-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0
`

	upgradeKubeletAndKubectlCoreOSScriptTemplate = `
source /etc/kubeone/proxy-env

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

curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/kubelet.service" |
	sed "s:/usr/bin:/opt/bin:g" |
	sudo tee /etc/systemd/system/kubelet.service

sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/10-kubeadm.conf" |
	sed "s:/usr/bin:/opt/bin:g" |
	sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

	 
sudo systemctl daemon-reload
sudo systemctl start kubelet
`
)

func KubeadmDebian(cluster *kubeone.KubeOneCluster, dockerVersion string) (string, error) {
	return Render(kubeadmDebianTemplate, Data{
		"DOCKER_VERSION":         dockerVersion,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"CNI_VERSION":            cluster.Versions.KubernetesCNIVersion(),
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
	})
}

func KubeadmCentOS(cluster *kubeone.KubeOneCluster, proxy string) (string, error) {
	return Render(kubeadmCentOSTemplate, Data{
		"DOCKER_VERSION":         centosDockerVersion,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"CNI_VERSION":            cluster.Versions.KubernetesCNIVersion(),
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
	})
}

func KubeadmCoreOS(cluster *kubeone.KubeOneCluster) (string, error) {
	return Render(kubeadmCoreOSTemplate, Data{
		"KUBERNETES_VERSION": cluster.Versions.Kubernetes,
		"CNI_VERSION":        cluster.Versions.KubernetesCNIVersion(),
	})
}

func RemoveBinariesDebian(k8sVersion, cniVersion string) (string, error) {
	return Render(removeBinariesDebianScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
		"CNI_VERSION":        cniVersion,
	})
}

func RemoveBinariesCentOS(k8sVersion, cniVersion string) (string, error) {
	return Render(removeBinariesCentOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
		"CNI_VERSION":        cniVersion,
	})
}

func RemoveBinariesCoreOS() (string, error) {
	return Render(removeBinariesCoreOSScriptTemplate, nil)
}

func UpgradeKubeadmAndCNIDebian(k8sVersion, cniVersion string) (string, error) {
	return Render(upgradeKubeadmAndCNIDebianScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
		"CNI_VERSION":        cniVersion,
	})
}

func UpgradeKubeadmAndCNICentOS(k8sVersion, cniVersion string) (string, error) {
	return Render(upgradeKubeadmAndCNICentOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
		"CNI_VERSION":        cniVersion,
	})
}

func UpgradeKubeadmAndCNICoreOS(k8sVersion, cniVersion string) (string, error) {
	return Render(upgradeKubeadmAndCNICoreOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
		"CNI_VERSION":        cniVersion,
	})
}

func UpgradeKubeletAndKubectlDebian(k8sVersion string) (string, error) {
	return Render(upgradeKubeletAndKubectlDebianScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
	})
}

func UpgradeKubeletAndKubectlCentOS(k8sVersion string) (string, error) {
	return Render(upgradeKubeletAndKubectlCentOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
	})
}

func UpgradeKubeletAndKubectlCoreOS(k8sVersion string) (string, error) {
	return Render(upgradeKubeletAndKubectlCoreOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
	})
}
