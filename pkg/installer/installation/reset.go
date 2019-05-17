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
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"
)

// Reset undos all changes made by KubeOne to the configured machines.
func Reset(ctx *util.Context) error {
	ctx.Logger.Infoln("Resetting cluster…")

	if ctx.DestroyWorkers {
		if err := destroyWorkers(ctx); err != nil {
			return err
		}
	}

	if err := ctx.RunTaskOnAllNodes(resetNode, true); err != nil {
		return err
	}

	if ctx.RemovePackages {
		if err := ctx.RunTaskOnAllNodes(removePackages, true); err != nil {
			return errors.Wrap(err, "unable to remove kubernetes packages")
		}
	}

	return nil
}

func destroyWorkers(ctx *util.Context) error {
	ctx.Logger.Infoln("Destroying worker nodes…")

	if err := util.BuildKubernetesClientset(ctx); err != nil {
		return errors.Wrap(err, "unable to build kubernetes clientset")
	}
	if err := machinecontroller.DestroyWorkers(ctx); err != nil {
		return errors.Wrap(err, "unable to delete all worker nodes")
	}

	return nil
}

func resetNode(ctx *util.Context, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Resetting node…")

	_, _, err := ctx.Runner.Run(resetScript, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

func removePackages(ctx *util.Context, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Removing Kubernetes packages…")

	// Determine operating system
	os, err := determineOS(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}
	node.SetOperatingSystem(os)

	// Remove Kubernetes packages
	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = removePackagesDebian(ctx)
	case "centos":
		err = removePackagesCentOS(ctx)
	case "coreos":
		err = removePackagesCoreOS(ctx)
	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func removePackagesDebian(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removePackagesDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        ctx.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removePackagesCentOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removePackagesCentOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        ctx.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removePackagesCoreOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removePackagesCoreOSCommand, util.TemplateVariables{})

	return errors.WithStack(err)
}

const (
	removePackagesDebianCommand = `
kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
sudo apt-get remove --purge -y \
     kubeadm=${kube_ver} \
     kubectl=${kube_ver} \
     kubelet=${kube_ver} \
     kubernetes-cni=${cni_ver}
`
	removePackagesCentOSCommand = `
sudo yum remove -y \
	kubelet-{{ .KUBERNETES_VERSION }}-0\
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0
`
	removePackagesCoreOSCommand = `
# Remove CNI and binaries
sudo rm -rf /opt/cni /opt/bin/kubeadm /opt/bin/kubectl /opt/bin/kubelet
# Remove systemd unit files
sudo rm /etc/systemd/system/kubelet.service /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
`
	resetScript = `
sudo kubeadm reset --force
sudo rm /etc/kubernetes/cloud-config
rm -rf "{{ .WORK_DIR }}"
`
)
