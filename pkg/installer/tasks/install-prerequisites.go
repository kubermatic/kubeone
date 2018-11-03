package tasks

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/template"
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/manifest"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates"
)

type InstallPrerequisitesTask struct{}

func (t *InstallPrerequisitesTask) Execute(ctx *Context) error {
	ctx.Logger.Infoln("Installing prerequisites…")

	err := t.generateConfigurationFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %v", err)
	}

	for _, node := range ctx.Manifest.Hosts {
		logger := ctx.Logger.WithFields(logrus.Fields{
			"node": node.Address,
		})

		err = t.executeNode(ctx, node, logger)
		if err != nil {
			break
		}
	}

	return err
}

func (t *InstallPrerequisitesTask) generateConfigurationFiles(ctx *Context) error {
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

func (t *InstallPrerequisitesTask) executeNode(ctx *Context, node manifest.HostManifest, logger logrus.FieldLogger) error {
	logger.Infoln("Connecting…")
	conn, err := ctx.Connector.Connect(node)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	logger.Infoln("Determine operating system…")
	os, err := t.determineOS(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to determine operating system: %v", err)
	}

	logger = logger.WithField("os", os)

	logger.Infoln("Installing kubeadm…")
	err = t.installKubeadm(ctx, conn, os)
	if err != nil {
		return fmt.Errorf("failed to install kubeadm: %v", err)
	}

	logger.Infoln("Deploying configuration files…")
	err = t.deployConfigurationFiles(ctx, conn, os)
	if err != nil {
		return fmt.Errorf("failed to upload configuration files: %v", err)
	}

	return nil
}

func (t *InstallPrerequisitesTask) determineOS(ctx *Context, conn ssh.Connection) (string, error) {
	stdout, _, _, err := conn.Exec("cat /etc/os-release | grep '^ID=' | sed s/^ID=//")

	return stdout, err
}

func (t *InstallPrerequisitesTask) installKubeadm(ctx *Context, conn ssh.Connection, os string) error {
	var err error

	switch os {
	case "ubuntu":
		fallthrough
	case "debian":
		err = t.installKubeadmDebian(ctx, conn)

	case "coreos":
		err = t.installKubeadmCoreOS(ctx, conn)

	default:
		err = fmt.Errorf("'%s' is not a supported operating system", os)
	}

	return err
}

func makeShellCommand(cmd string, variables map[string]string) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to parse shell script: %v", err)
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render shell script: %v", err)
	}

	return buf.String(), nil
}

func (t *InstallPrerequisitesTask) installKubeadmDebian(ctx *Context, conn ssh.Connection) error {
	command, err := makeShellCommand(kubeadmDebianCommand, map[string]string{
		"KUBERNETES_VERSION": ctx.Manifest.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Manifest.Versions.Docker,
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := conn.Exec(command)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

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

func (t *InstallPrerequisitesTask) installKubeadmCoreOS(ctx *Context, conn ssh.Connection) error {
	command, err := makeShellCommand(kubeadmCoreOSCommand, map[string]string{
		"KUBERNETES_VERSION": ctx.Manifest.Versions.Kubernetes,
		"DOCKER_VERSION":     ctx.Manifest.Versions.Docker,
		"CNI_VERSION":        "v0.7.1",
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := conn.Exec(command)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

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

func (t *InstallPrerequisitesTask) deployConfigurationFiles(ctx *Context, conn ssh.Connection, operatingSystem string) error {
	err := ctx.Configuration.UploadTo(conn, ctx.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	// move config files to their permanent locations
	command, err := makeShellCommand(`
set -xeu pipefail

sudo mkdir -p /etc/systemd/system/kubelet.service.d/ /etc/kubernetes
sudo mv ./{{ .WORK_DIR }}/cfg/20-cloudconfig-kubelet.conf /etc/systemd/system/kubelet.service.d/
sudo mv ./{{ .WORK_DIR }}/cfg/cloud-config /etc/kubernetes/cloud-config
sudo chown root:root /etc/kubernetes/cloud-config
sudo chmod 600 /etc/kubernetes/cloud-config
`, map[string]string{
		"WORK_DIR": ctx.WorkDir,
	})
	if err != nil {
		return fmt.Errorf("failed to construct shell script: %v", err)
	}

	_, stderr, _, err := conn.Exec(command)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
	}

	return err
}
