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

package installation

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	dockerVersion = "18.09.7"
)

const (
	daemonsProxyScript = `
sudo mkdir -p /etc/systemd/system/docker.service.d
cat <<EOF | sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf
[Service]
EnvironmentFile=/etc/kubeone/proxy-env
EOF

sudo mkdir -p /etc/systemd/system/kubelet.service.d
cat <<EOF | sudo tee /etc/systemd/system/kubelet.service.d/http-proxy.conf
[Service]
EnvironmentFile=/etc/kubeone/proxy-env
EOF

sudo systemctl daemon-reload
if sudo systemctl status docker &>/dev/null; then sudo systemctl restart docker; fi
if sudo systemctl status kubelet &>/dev/null; then sudo systemctl restart kubelet; fi
`

	hostnameScript = `
fqdn=$(hostname -f)
[ "$fqdn" = localhost ] && fqdn=$(hostname)
echo "$fqdn"
`

	environmentFileScript = `
sudo mkdir -p /etc/kubeone
cat <<EOF | sudo tee /etc/kubeone/proxy-env
{{ if .HTTP_PROXY -}}
HTTP_PROXY="{{ .HTTP_PROXY }}"
http_proxy="{{ .HTTP_PROXY }}"
export HTTP_PROXY http_proxy

{{ end }}
{{- if .HTTPS_PROXY -}}
HTTPS_PROXY="{{ .HTTPS_PROXY }}"
https_proxy="{{ .HTTPS_PROXY }}"
export HTTPS_PROXY https_proxy

{{ end }}
{{- if .NO_PROXY -}}
NO_PROXY="{{ .NO_PROXY }}"
no_proxy="{{ .NO_PROXY }}"
export NO_PROXY no_proxy

{{ end }}
EOF
	`

	kubeadmDebianScript = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab

. /etc/os-release
. /etc/kubeone/proxy-env

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2"
}
EOF

sudo apt-get update
sudo apt-get install -y --no-install-recommends \
	apt-transport-https \
	ca-certificates \
	curl \
	htop \
	lsb-release \
	rsync \
	tree

curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | \
	sudo apt-key add -
curl -fsSL https://download.docker.com/linux/${ID}/gpg | \
	sudo apt-key add -

echo "deb [arch=amd64] https://download.docker.com/linux/${ID} $(lsb_release -sc) stable" | \
	sudo tee /etc/apt/sources.list.d/docker.list

# You'd think that kubernetes-$(lsb_release -sc) belongs there instead, but the debian repo
# contains neither kubeadm nor kubelet, and the docs themselves suggest using xenial repo.
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | \
	sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update

docker_ver=$(apt-cache madison docker-ce | \
	grep "{{ .DOCKER_VERSION }}" | head -1 | awk '{print $3}')
kube_ver=$(apt-cache madison kubelet | \
	grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | \
	grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold docker-ce kubelet kubeadm kubectl kubernetes-cni
sudo apt-get install -y --no-install-recommends \
	docker-ce=${docker_ver} \
	kubeadm=${kube_ver} \
	kubectl=${kube_ver} \
	kubelet=${kube_ver} \
	kubernetes-cni=${cni_ver}
sudo apt-mark hold docker-ce kubelet kubeadm kubectl kubernetes-cni
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	kubeadmCentOSScript = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab
sudo setenforce 0 || true
sudo sed -i s/SELINUX=enforcing/SELINUX=permissive/g /etc/sysconfig/selinux

. /etc/kubeone/proxy-env

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system

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

sudo yum install -y --disableexcludes=kubernetes \
	docker kubelet-{{ .KUBERNETES_VERSION }}-0\
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0
sudo systemctl enable --now docker
sudo systemctl enable --now kubelet
`

	kubeadmCoreOSScript = `
. /etc/kubeone/proxy-env

# Short-Circuit the installation if it was already executed
if type docker &>/dev/null && type kubelet &>/dev/null; then exit 0; fi

sudo mkdir -p /etc/docker
cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2"
}
EOF
sudo systemctl restart docker

sudo mkdir -p /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .CNI_VERSION }}/cni-plugins-amd64-v{{ .CNI_VERSION }}.tgz" | \
	sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /opt/bin
cd /opt/bin
k8s_rel_baseurl=https://storage.googleapis.com/kubernetes-release/release
for binary in kubeadm kubelet kubectl; do
	curl -L --output /tmp/$binary \
		$k8s_rel_baseurl/${RELEASE}/bin/linux/amd64/$binary
	sudo install --owner=0 --group=0 --mode=0755 /tmp/$binary /opt/bin/$binary
	rm /tmp/$binary
done

curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/kubelet.service" | \
	sed "s:/usr/bin:/opt/bin:g" | \
	sudo tee /etc/systemd/system/kubelet.service

sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/10-kubeadm.conf" | \
	sed "s:/usr/bin:/opt/bin:g" | \
	sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

sudo systemctl daemon-reload
sudo systemctl enable docker.service kubelet.service
sudo systemctl start docker.service kubelet.service
`
)

func installPrerequisites(s *state.State) error {
	s.Logger.Infoln("Installing prerequisites…")

	err := generateConfigurationFiles(s)
	if err != nil {
		return errors.Wrap(err, "unable to generate configuration files")
	}

	return s.RunTaskOnAllNodes(installPrerequisitesOnNode, true)
}

func generateConfigurationFiles(s *state.State) error {
	s.Configuration.AddFile("cfg/cloud-config", s.Cluster.CloudProvider.CloudConfig)

	if s.Cluster.Features.StaticAuditLog != nil && s.Cluster.Features.StaticAuditLog.Enable {
		err := s.Configuration.AddFilePath("cfg/audit-policy.yaml", s.Cluster.Features.StaticAuditLog.Config.PolicyFilePath)
		if err != nil {
			return errors.Wrap(err, "unable to add policy file")
		}
	}

	return nil
}

func installPrerequisitesOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Determine operating system…")
	os, err := determineOS(s)
	if err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}

	node.SetOperatingSystem(os)

	s.Logger.Infoln("Determine hostname…")
	if node.Hostname == "" {
		hostname, hostnameErr := determineHostname(s, *node)
		if hostnameErr != nil {
			return errors.Wrap(hostnameErr, "failed to determine hostname")
		}
		node.SetHostname(hostname)
	}

	s.Logger.Infoln("Creating environment file…")
	err = createEnvironmentFile(s)
	if err != nil {
		return errors.Wrap(err, "failed to create environment file")
	}

	logger := s.Logger.WithField("os", os)

	logger.Infoln("Installing kubeadm…")
	err = installKubeadm(s, *node)
	if err != nil {
		return errors.Wrap(err, "failed to install kubeadm")
	}

	err = configureProxy(s)
	if err != nil {
		return errors.Wrap(err, "failed to configure proxy for docker daemon")
	}

	logger.Infoln("Deploying configuration files…")
	err = deployConfigurationFiles(s)
	if err != nil {
		return errors.Wrap(err, "failed to upload configuration files")
	}

	return nil
}

func determineOS(s *state.State) (string, error) {
	osID, _, err := s.Runner.Run(". /etc/os-release && echo -n $ID", nil)
	return osID, err
}

func determineHostname(s *state.State, _ kubeoneapi.HostConfig) (string, error) {
	hostcmd := hostnameScript

	// on azure the name of the Node should == name of the VM
	if s.Cluster.CloudProvider.Name == kubeoneapi.CloudProviderNameAzure {
		hostcmd = `hostname`
	}
	stdout, _, err := s.Runner.Run(hostcmd, nil)

	return stdout, err
}

func createEnvironmentFile(s *state.State) error {
	_, _, err := s.Runner.Run(environmentFileScript, runner.TemplateVariables{
		"HTTP_PROXY":  s.Cluster.Proxy.HTTP,
		"HTTPS_PROXY": s.Cluster.Proxy.HTTPS,
		"NO_PROXY":    s.Cluster.Proxy.NoProxy,
	})

	return err
}

func installKubeadm(s *state.State, node kubeoneapi.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = installKubeadmDebian(s)

	case "coreos":
		err = installKubeadmCoreOS(s)

	case "centos":
		err = installKubeadmCentOS(s)

	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func installKubeadmDebian(s *state.State) error {
	_, _, err := s.Runner.Run(kubeadmDebianScript, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"DOCKER_VERSION":     dockerVersion,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func installKubeadmCentOS(s *state.State) error {
	_, _, err := s.Runner.Run(kubeadmCentOSScript, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})
	return err
}

func installKubeadmCoreOS(s *state.State) error {
	_, _, err := s.Runner.Run(kubeadmCoreOSScript, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return err
}

func deployConfigurationFiles(s *state.State) error {
	err := s.Configuration.UploadTo(s.Runner.Conn, s.WorkDir)
	if err != nil {
		return errors.Wrap(err, "failed to upload")
	}

	// move config files to their permanent locations
	_, _, err = s.Runner.Run(`
sudo mkdir -p /etc/systemd/system/kubelet.service.d/ /etc/kubernetes
sudo mv ./{{ .WORK_DIR }}/cfg/cloud-config /etc/kubernetes/cloud-config
sudo chown root:root /etc/kubernetes/cloud-config
sudo chmod 600 /etc/kubernetes/cloud-config

if [[ -f "./{{ .WORK_DIR }}/cfg/audit-policy.yaml" ]]; then
	sudo mkdir -p /etc/kubernetes/audit
	sudo mv ./{{ .WORK_DIR }}/cfg/audit-policy.yaml /etc/kubernetes/audit/policy.yaml
	sudo chown root:root /etc/kubernetes/audit/policy.yaml
fi
`, runner.TemplateVariables{
		"WORK_DIR": s.WorkDir,
	})

	return err
}

func configureProxy(s *state.State) error {
	if s.Cluster.Proxy.HTTP == "" && s.Cluster.Proxy.HTTPS == "" && s.Cluster.Proxy.NoProxy == "" {
		return nil
	}

	s.Logger.Infoln("Configuring docker proxy…")
	_, _, err := s.Runner.Run(daemonsProxyScript, runner.TemplateVariables{})

	return err
}
