package kube111

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates"
)

func InstallPrerequisites(ctx *util.Context) error {
	ctx.Logger.Infoln("Installing prerequisites…")

	err := generateConfigurationFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %v", err)
	}

	return util.RunTaskOnNodes(ctx, installPrerequisitesOnNode)
}

func generateConfigurationFiles(ctx *util.Context) error {
	kubeadm, err := templates.KubeadmConfig(ctx.Manifest)
	if err != nil {
		return fmt.Errorf("failed to create kubeadm configuration: %v", err)
	}

	ctx.Configuration.AddFile("cfg/master.yaml", kubeadm)

	for idx := range ctx.Manifest.Hosts {
		etcd, err := templates.EtcdConfig(ctx.Manifest, idx)
		if err != nil {
			return fmt.Errorf("failed to create etcd configuration: %v", err)
		}

		ctx.Configuration.AddFile(fmt.Sprintf("etcd/etcd_%d.yaml", idx), etcd)
	}

	ctx.Configuration.AddFile("cfg/20-cloudconfig-kubelet.conf", fmt.Sprintf(`
[Service]
Environment="KUBELET_EXTRA_ARGS= --cloud-provider=%s --cloud-config=/etc/kubernetes/cloud-config"`,
		ctx.Manifest.Provider.Name))

	ctx.Configuration.AddFile("cfg/cloud-config", ctx.Manifest.Provider.CloudConfig)

	return nil
}

func installPrerequisitesOnNode(ctx *util.Context, node manifest.HostManifest, _ int, conn ssh.Connection) error {
	ctx.Logger.Infoln("Determine operating system…")
	os, err := determineOS(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to determine operating system: %v", err)
	}

	logger := ctx.Logger.WithField("os", os)

	logger.Infoln("Installing kubeadm…")
	err = installKubeadm(ctx, conn, os)
	if err != nil {
		return fmt.Errorf("failed to install kubeadm: %v", err)
	}

	logger.Infoln("Deploying configuration files…")
	err = deployConfigurationFiles(ctx, conn, os)
	if err != nil {
		return fmt.Errorf("failed to upload configuration files: %v", err)
	}

	return nil
}

func determineOS(ctx *util.Context, conn ssh.Connection) (string, error) {
	stdout, _, _, err := util.RunCommand(conn, "cat /etc/os-release | grep '^ID=' | sed s/^ID=//", ctx.Verbose)

	return stdout, err
}

func installKubeadm(ctx *util.Context, conn ssh.Connection, os string) error {
	var err error

	switch os {
	case "ubuntu":
		fallthrough
	case "debian":
		err = installKubeadmDebian(ctx, conn)

	case "coreos":
		err = installKubeadmCoreOS(ctx, conn)

	default:
		err = fmt.Errorf("'%s' is not a supported operating system", os)
	}

	return err
}

func installKubeadmDebian(ctx *util.Context, conn ssh.Connection) error {
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Manifest.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Manifest.Versions.Docker,
	})

	return err
}

const kubeadmDebianCommand = `
set -xeu pipefail
sudo swapoff -a

source /etc/os-release

sudo apt-get update
sudo apt-get install -y --no-install-recommends \
     apt-transport-https \
     ca-certificates \
     curl \
     htop \
     lsb-release \
     rsync \
     tree

curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
curl -fsSL https://download.docker.com/linux/${ID}/gpg | sudo apt-key add -

echo "deb [arch=amd64] https://download.docker.com/linux/${ID} $(lsb_release -sc) stable" | \
     sudo tee /etc/apt/sources.list.d/docker.list

# You'd think that kubernetes-$(lsb_release -sc) belongs there instead, but the debian repo
# contains neither kubeadm nor kubelet, and the docs themselves suggest using xenial repo.
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | \
     sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update

docker_ver=$(apt-cache madison docker-ce | grep "{{ .DOCKER_VERSION }}" | head -1 | awk '{print $3}')
kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold docker-ce kubelet kubeadm kubectl
sudo apt-get install -y --no-install-recommends \
     docker-ce=${docker_ver} \
     kubeadm=${kube_ver} \
     kubectl=${kube_ver} \
     kubelet=${kube_ver}
sudo apt-mark hold docker-ce kubelet kubeadm kubectl
sudo systemctl daemon-reload
`

func installKubeadmCoreOS(ctx *util.Context, conn ssh.Connection) error {
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmCoreOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Manifest.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Manifest.Versions.Docker,
		"CNI_VERSION":        "v0.7.1",
	})

	return err
}

const kubeadmCoreOSCommand = `
set -xeu pipefail

sudo mkdir -p /opt/cni/bin /etc/kubernetes/pki /etc/kubernetes/manifests
curl -L "https://github.com/containernetworking/plugins/releases/download/{{ .CNI_VERSION }}/cni-plugins-amd64-{{ .CNI_VERSION }}.tgz" | \
     sudo tar -C /opt/cni/bin -xz

RELEASE="v{{ .KUBERNETES_VERSION }}"

sudo mkdir -p /opt/bin
cd /opt/bin
sudo curl -L --remote-name-all \
     https://storage.googleapis.com/kubernetes-release/release/${RELEASE}/bin/linux/amd64/{kubeadm,kubelet,kubectl}
sudo chmod +x {kubeadm,kubelet,kubectl}

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

func deployConfigurationFiles(ctx *util.Context, conn ssh.Connection, operatingSystem string) error {
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	// move config files to their permanent locations
	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
set -xeu pipefail

sudo mkdir -p /etc/systemd/system/kubelet.service.d/ /etc/kubernetes
sudo mv ./{{ .WORK_DIR }}/cfg/20-cloudconfig-kubelet.conf /etc/systemd/system/kubelet.service.d/
sudo mv ./{{ .WORK_DIR }}/cfg/cloud-config /etc/kubernetes/cloud-config
sudo chown root:root /etc/kubernetes/cloud-config
sudo chmod 600 /etc/kubernetes/cloud-config
`, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}
