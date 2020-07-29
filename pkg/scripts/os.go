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
	defaultKubernetesCNIVersion = "0.8.6"
)

const (
	kubeadmDebianTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo systemctl disable --now ufw || true

source /etc/kubeone/proxy-env

{{ template "docker-daemon-config" }}
{{ template "sysctl-k8s" }}
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
	lsb-release \
	rsync

{{- if .CONFIGURE_REPOSITORIES }}
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

{{- /* TODO(kron4eg): replace bionic with focal someday */}}
echo "deb https://download.docker.com/linux/ubuntu bionic stable" |
	sudo tee /etc/apt/sources.list.d/docker.list

# You'd think that kubernetes-$(lsb_release -sc) belongs there instead, but the debian repo
# contains neither kubeadm nor kubelet, and the docs themselves suggest using xenial repo.
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
{{- end }}

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .KUBERNETES_CNI_VERSION }}" | head -1 | awk '{print $3}')

{{- if or .FORCE .UPGRADE }}
sudo apt-mark unhold docker-ce docker-ce-cli kubelet kubeadm kubectl kubernetes-cni
{{- end }}

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
	kubernetes-cni=${cni_ver} \
	{{ aptDocker .KUBERNETES_VERSION }}

sudo apt-mark hold docker-ce docker-ce-cli kubelet kubeadm kubectl kubernetes-cni

sudo systemctl daemon-reload
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	kubeadmCentOSTemplate = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo setenforce 0 || true
sudo sed -i 's/SELINUX=enforcing/SELINUX=permissive/g' /etc/sysconfig/selinux
sudo sed -i 's/SELINUX=enforcing/SELINUX=permissive/g' /etc/selinux/config
sudo systemctl disable --now firewalld || true

source /etc/kubeone/proxy-env

{{ template "docker-daemon-config" }}
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
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

sudo yum install -y yum-utils
sudo yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
sudo yum-config-manager --save --setopt=docker-ce-stable.module_hotfixes=true >/dev/null
{{ end }}

sudo yum install -y \
	yum-plugin-versionlock \
	device-mapper-persistent-data \
	lvm2

{{- if or .FORCE .UPGRADE }}
sudo yum versionlock delete docker-ce docker-ce-cli kubelet kubeadm kubectl kubernetes-cni || true
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
	kubernetes-cni-{{ .KUBERNETES_CNI_VERSION }} \
	{{ yumDocker .KUBERNETES_VERSION }}
sudo yum versionlock add docker-ce docker-ce-cli kubelet kubeadm kubectl kubernetes-cni

sudo systemctl daemon-reload
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet

{{- if or .FORCE .KUBELET }}
sudo systemctl restart kubelet
{{- end }}
`

	kubeadmCoreOSTemplate = `
source /etc/kubeone/proxy-env

{{ template "detect-host-cpu-architecture" }}
{{ template "docker-daemon-config" }}
{{ template "sysctl-k8s" }}
{{ template "journald-config" }}

sudo systemctl restart docker

sudo mkdir -p /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .KUBERNETES_CNI_VERSION }}/cni-plugins-linux-${HOST_ARCH}-v{{ .KUBERNETES_CNI_VERSION }}.tgz" |
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
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	removeBinariesDebianScriptTemplate = `
sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
sudo apt-get remove --purge -y \
	kubeadm \
	kubectl \
	kubelet
sudo apt-get remove --purge -y kubernetes-cni || true
`

	removeBinariesCentOSScriptTemplate = `
sudo yum versionlock delete kubelet kubeadm kubectl kubernetes-cni || true
sudo yum remove -y \
	kubelet \
	kubeadm \
	kubectl
sudo yum remove -y kubernetes-cni || true
`

	removeBinariesCoreOSScriptTemplate = `
# Remove CNI and binaries
sudo rm -rf /opt/cni /opt/bin/kubeadm /opt/bin/kubectl /opt/bin/kubelet
# Remove systemd unit files
sudo rm /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
`

	upgradeKubeadmAndCNICoreOSScriptTemplate = `
{{ template "detect-host-cpu-architecture" }}

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
sudo systemctl stop kubelet
sudo mv /var/tmp/kube-binaries/kubeadm .
sudo chmod +x kubeadm
`

	upgradeKubeletAndKubectlCoreOSScriptTemplate = `
source /etc/kubeone/proxy-env

{{ template "detect-host-cpu-architecture" }}

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

func KubeadmDebian(cluster *kubeone.KubeOneCluster, force bool) (string, error) {
	return Render(kubeadmDebianTemplate, Data{
		"KUBELET":                true,
		"KUBEADM":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
		"FORCE":                  force,
	})
}

func KubeadmCentOS(cluster *kubeone.KubeOneCluster, force bool) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	return Render(kubeadmCentOSTemplate, Data{
		"KUBELET":                true,
		"KUBEADM":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
		"FORCE":                  force,
	})
}

func KubeadmCoreOS(cluster *kubeone.KubeOneCluster) (string, error) {
	return Render(kubeadmCoreOSTemplate, Data{
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
	})
}

func RemoveBinariesDebian() (string, error) {
	return Render(removeBinariesDebianScriptTemplate, Data{})
}

func RemoveBinariesCentOS() (string, error) {
	return Render(removeBinariesCentOSScriptTemplate, Data{})
}

func RemoveBinariesCoreOS() (string, error) {
	return Render(removeBinariesCoreOSScriptTemplate, nil)
}

func UpgradeKubeadmAndCNIDebian(cluster *kubeone.KubeOneCluster) (string, error) {
	return Render(kubeadmDebianTemplate, Data{
		"UPGRADE":                true,
		"KUBEADM":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
	})
}

func UpgradeKubeadmAndCNICentOS(cluster *kubeone.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	return Render(kubeadmCentOSTemplate, Data{
		"UPGRADE":                true,
		"KUBEADM":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
	})
}

func UpgradeKubeadmAndCNICoreOS(k8sVersion string) (string, error) {
	return Render(upgradeKubeadmAndCNICoreOSScriptTemplate, Data{
		"KUBERNETES_VERSION":     k8sVersion,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
	})
}

func UpgradeKubeletAndKubectlDebian(cluster *kubeone.KubeOneCluster) (string, error) {
	return Render(kubeadmDebianTemplate, Data{
		"UPGRADE":                true,
		"KUBELET":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"HTTP_PROXY":             cluster.Proxy.HTTP,
		"HTTPS_PROXY":            cluster.Proxy.HTTPS,
	})
}

func UpgradeKubeletAndKubectlCentOS(cluster *kubeone.KubeOneCluster) (string, error) {
	proxy := cluster.Proxy.HTTPS
	if proxy == "" {
		proxy = cluster.Proxy.HTTP
	}

	return Render(kubeadmCentOSTemplate, Data{
		"UPGRADE":                true,
		"KUBELET":                true,
		"KUBECTL":                true,
		"KUBERNETES_VERSION":     cluster.Versions.Kubernetes,
		"KUBERNETES_CNI_VERSION": defaultKubernetesCNIVersion,
		"CONFIGURE_REPOSITORIES": cluster.SystemPackages.ConfigureRepositories,
		"PROXY":                  proxy,
	})
}

func UpgradeKubeletAndKubectlCoreOS(k8sVersion string) (string, error) {
	return Render(upgradeKubeletAndKubectlCoreOSScriptTemplate, Data{
		"KUBERNETES_VERSION": k8sVersion,
	})
}
