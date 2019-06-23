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
	"time"

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"
	"github.com/kubermatic/kubeone/pkg/util"

	"k8s.io/apimachinery/pkg/util/wait"
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

	if ctx.RemoveBinaries {
		if err := ctx.RunTaskOnAllNodes(removeBinaries, true); err != nil {
			return errors.Wrap(err, "unable to remove kubernetes binaries")
		}
	}

	return nil
}

func destroyWorkers(ctx *util.Context) error {
	ctx.Logger.Infoln("Destroying worker nodes…")

	waitErr := wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		err := util.BuildKubernetesClientset(ctx)
		return err == nil, errors.Wrap(err, "unable to build kubernetes clientset")
	})
	if waitErr != nil {
		ctx.Logger.Warn("Unable to connect to the control plane API and destroy worker nodes")
		ctx.Logger.Warn("You can skip destorying worker nodes and destroy them manually using `--destroy-workers=false`")
		return waitErr
	}

	waitErr = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		err := machinecontroller.DestroyWorkers(ctx)
		return err == nil, errors.Wrap(err, "unable to delete all worker nodes")
	})

	return waitErr
}

func resetNode(ctx *util.Context, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Resetting node…")

	_, _, err := ctx.Runner.Run(resetScript, util.TemplateVariables{
		"WORK_DIR": ctx.WorkDir,
	})

	return err
}

func removeBinaries(ctx *util.Context, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	ctx.Logger.Infoln("Removing Kubernetes binaries")

	// Determine operating system
	os, err := determineOS(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}
	node.SetOperatingSystem(os)

	// Remove Kubernetes binaries
	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = removeBinariesDebian(ctx)
	case "centos":
		err = removeBinariesCentOS(ctx)
	case "coreos":
		err = removeBinariesCoreOS(ctx)
	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func removeBinariesDebian(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removeBinariesDebianCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        ctx.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removeBinariesCentOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removeBinariesCentOSCommand, util.TemplateVariables{
		"KUBERNETES_VERSION": ctx.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        ctx.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removeBinariesCoreOS(ctx *util.Context) error {
	_, _, err := ctx.Runner.Run(removeBinariesCoreOSCommand, util.TemplateVariables{})

	return errors.WithStack(err)
}

const (
	removeBinariesDebianCommand = `
kube_ver=$(apt-cache madison kubelet | grep "{{ .KUBERNETES_VERSION }}" | head -1 | awk '{print $3}')
cni_ver=$(apt-cache madison kubernetes-cni | grep "{{ .CNI_VERSION }}" | head -1 | awk '{print $3}')

sudo apt-mark unhold kubelet kubeadm kubectl kubernetes-cni
sudo apt-get remove --purge -y \
     kubeadm=${kube_ver} \
     kubectl=${kube_ver} \
     kubelet=${kube_ver} \
     kubernetes-cni=${cni_ver}
`
	removeBinariesCentOSCommand = `
sudo yum remove -y \
	kubelet-{{ .KUBERNETES_VERSION }}-0\
	kubeadm-{{ .KUBERNETES_VERSION }}-0 \
	kubectl-{{ .KUBERNETES_VERSION }}-0 \
	kubernetes-cni-{{ .CNI_VERSION }}-0
`
	removeBinariesCoreOSCommand = `
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

func defaultRetryBackoff(retries int) wait.Backoff {
	if retries == 0 {
		retries = 1
	}

	return wait.Backoff{
		Steps:    retries,
		Duration: 5 * time.Second,
		Factor:   2.0,
	}
}
