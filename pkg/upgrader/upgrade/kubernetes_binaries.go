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

package upgrade

import (
	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/state"
)

const (
	upgradeKubeBinariesDebianCommand = `
source /etc/os-release
source /etc/kubeone/proxy-env

sudo apt-get update

kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubeadm kubelet kubectl kubernetes-cni
sudo DEBIAN_FRONTEND=noninteractive apt-get install --option "Dpkg::Options::=--force-confold" -y --no-install-recommends \
     kubeadm=${kube_ver} \
     kubectl=${kube_ver} \
     kubelet=${kube_ver} \
     kubernetes-cni=${cni_ver}
sudo apt-mark hold kubeadm kubelet kubectl kubernetes-cni
`
	upgradeKubeBinariesCentOSCommand = `
source /etc/kubeone/proxy-env

sudo yum install -y --disableexcludes=kubernetes \
			kubelet-{{ .KUBERNETES_VERSION }}-0 \
			kubeadm-{{ .KUBERNETES_VERSION }}-0 \
			kubectl-{{ .KUBERNETES_VERSION }}-0 \
			kubernetes-cni-{{ .CNI_VERSION }}-0
`
	upgradeKubeBinariesCoreOSCommand = `
source /etc/kubeone/proxy-env

sudo mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/v{{ .CNI_VERSION }}/cni-plugins-amd64-v{{ .CNI_VERSION }}.tgz" | \
     sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"



sudo mkdir -p /var/tmp/kube-binaries
cd /var/tmp/kube-binaries
sudo curl -L --remote-name-all \
     https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/amd64/{kubeadm,kubelet,kubectl}


sudo mkdir -p /opt/bin
cd /opt/bin
sudo systemctl stop kubelet
sudo mv /var/tmp/kube-binaries/{kubeadm,kubelet,kubectl} .
sudo chmod +x {kubeadm,kubelet,kubectl}

curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/kubelet.service" | \
     sed "s:/usr/bin:/opt/bin:g" | \
	  sudo tee /etc/systemd/system/kubelet.service

sudo mkdir -p /etc/systemd/system/kubelet.service.d
curl -sSL "https://raw.githubusercontent.com/kubernetes/kubernetes/${RELEASE}/build/debs/10-kubeadm.conf" | \
     sed "s:/usr/bin:/opt/bin:g" | \
	 sudo tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
	 
sudo systemctl daemon-reload
sudo systemctl start kubelet
`
)

func upgradeKubernetesBinaries(s *state.State, node kubeoneapi.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = upgradeKubernetesBinariesDebian(s)

	case "coreos":
		err = upgradeKubernetesBinariesCoreOS(s)

	case "centos":
		err = upgradeKubernetesBinariesCentOS(s)

	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func upgradeKubernetesBinariesDebian(s *state.State) error {
	_, _, err := s.Runner.Run(upgradeKubeBinariesDebianCommand, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func upgradeKubernetesBinariesCentOS(s *state.State) error {
	_, _, err := s.Runner.Run(upgradeKubeBinariesCentOSCommand, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func upgradeKubernetesBinariesCoreOS(s *state.State) error {
	_, _, err := s.Runner.Run(upgradeKubeBinariesCoreOSCommand, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}
