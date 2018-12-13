package kube112

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/flannel"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
)

func installPrerequisites(ctx *util.Context) error {
	ctx.Logger.Infoln("Installing prerequisites…")

	if err := generateConfigurationFiles(ctx); err != nil {
		return fmt.Errorf("failed to create configuration: %v", err)
	}

	return util.RunTaskOnAllNodes(ctx, installPrerequisitesOnNode)
}

func generateConfigurationFiles(ctx *util.Context) error {
	ctx.Configuration.AddFile("cfg/20-cloudconfig-kubelet.conf", fmt.Sprintf(`
[Service]
Environment="KUBELET_EXTRA_ARGS= --cloud-provider=%s --cloud-config=/etc/kubernetes/cloud-config"`,
		ctx.Cluster.Provider.Name))

	ctx.Configuration.AddFile("cfg/cloud-config", ctx.Cluster.Provider.CloudConfig)

	mc, err := machinecontroller.Deployment(ctx.Cluster)
	if err != nil {
		return fmt.Errorf("failed to create machine-controller configuration: %v", err)
	}
	ctx.Configuration.AddFile("machine-controller.yaml", mc)

	if len(ctx.Cluster.Workers) > 0 {
		machines, deployErr := machinecontroller.MachineDeployments(ctx.Cluster)
		if err != nil {
			return fmt.Errorf("failed to create worker machine configuration: %v", deployErr)
		}
		ctx.Configuration.AddFile("workers.yaml", machines)
	}

	flannel, err := flannel.Configuration(ctx.Cluster)
	if err != nil {
		return fmt.Errorf("failed to create flannel configuration: %v", err)
	}
	ctx.Configuration.AddFile("kube-flannel.yaml", flannel)

	return nil
}

func installPrerequisitesOnNode(ctx *util.Context, node *config.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Determine operating system…")
	os, err := determineOS(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to determine operating system: %v", err)
	}

	node.OperatingSystem = os

	ctx.Logger.Infoln("Determine hostname…")
	hostname, err := determineHostname(ctx, conn, node)
	if err != nil {
		return fmt.Errorf("failed to determine hostname: %v", err)
	}

	node.Hostname = hostname

	logger := ctx.Logger.WithField("os", os)

	logger.Infoln("Installing kubeadm…")
	err = installKubeadm(ctx, conn, node)
	if err != nil {
		return fmt.Errorf("failed to install kubeadm: %v", err)
	}

	logger.Infoln("Deploying configuration files…")
	err = deployConfigurationFiles(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to upload configuration files: %v", err)
	}

	return nil
}

func determineOS(ctx *util.Context, conn ssh.Connection) (string, error) {
	stdout, _, _, err := util.RunCommand(conn, "cat /etc/os-release | grep '^ID=' | sed s/^ID=//", ctx.Verbose)

	return stdout, err
}

func determineHostname(ctx *util.Context, conn ssh.Connection, _ *config.HostConfig) (string, error) {
	stdout, _, _, err := util.RunCommand(conn, "hostname -f", ctx.Verbose)

	return stdout, err
}

func installKubeadm(ctx *util.Context, conn ssh.Connection, node *config.HostConfig) error {
	var err error

	switch node.OperatingSystem {
	case "ubuntu":
		fallthrough
	case "debian":
		err = installKubeadmDebian(ctx, conn)

	case "coreos":
		err = installKubeadmCoreOS(ctx, conn)

	default:
		err = fmt.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func installKubeadmDebian(ctx *util.Context, conn ssh.Connection) error {
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Cluster.Versions.Docker,
	})

	return err
}

const kubeadmDebianCommand = `
sudo swapoff -a
sudo sed -i '/.*swap.*/d' /etc/fstab

source /etc/os-release


# Short-Circuit the installation if it was arleady executed
if type docker && type kubelet; then exit 0; fi

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
sudo mv /etc/systemd/system/kubelet.service.d/10-kubeadm.conf /etc/systemd/system/kubelet.service.d/10-kubeadm.conf.disabled
sudo systemctl daemon-reload
`

func installKubeadmCoreOS(ctx *util.Context, conn ssh.Connection) error {
	_, _, _, err := util.RunShellCommand(conn, ctx.Verbose, kubeadmCoreOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Cluster.Versions.Docker,
		"CNI_VERSION":        "v0.7.1",
	})

	return err
}

const kubeadmCoreOSCommand = `
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

func deployConfigurationFiles(ctx *util.Context, conn ssh.Connection) error {
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	// move config files to their permanent locations
	_, _, _, err = util.RunShellCommand(conn, ctx.Verbose, `
sudo cp /lib/systemd/system/kubelet.service /etc/systemd/system/kubelet.service
sudo sed -i 's#ExecStart=/usr/bin/kubelet.*#ExecStart=/usr/bin/kubelet --pod-manifest-path=/etc/kubernetes/manifests#g' /etc/systemd/system/kubelet.service
sudo mkdir -p /etc/kubernetes/manifests
sudo systemctl daemon-reload
sudo systemctl restart kubelet
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
