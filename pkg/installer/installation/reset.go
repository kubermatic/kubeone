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
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/runner"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Reset undos all changes made by KubeOne to the configured machines.
func Reset(s *state.State) error {
	s.Logger.Infoln("Resetting cluster…")

	if s.DestroyWorkers {
		if err := destroyWorkers(s); err != nil {
			return err
		}
	}

	if err := s.RunTaskOnAllNodes(resetNode, true); err != nil {
		return err
	}

	if s.RemoveBinaries {
		if err := s.RunTaskOnAllNodes(removeBinaries, true); err != nil {
			return errors.Wrap(err, "unable to remove kubernetes binaries")
		}
	}

	return nil
}

func destroyWorkers(s *state.State) error {
	var lastErr error
	s.Logger.Infoln("Destroying worker nodes…")

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = kubeconfig.BuildKubernetesClientset(s)
		if lastErr != nil {
			s.Logger.Warn("Unable to connect to the control plane API. Retrying…")
			return false, nil
		}
		return true, nil

	})
	if lastErr != nil {
		s.Logger.Warn("Unable to connect to the control plane API and destroy worker nodes")
		s.Logger.Warn("You can skip destroying worker nodes and destroy them manually using `--destroy-workers=false`")
		return errors.Wrap(lastErr, "unable to build kubernetes clientset")
	}

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = machinecontroller.DestroyWorkers(s)
		if lastErr != nil {
			s.Logger.Warn("Unable to destroy worker nodes. Retrying…")
			return false, nil
		}
		return true, nil
	})
	if lastErr != nil {
		return errors.Wrap(lastErr, "unable to delete all worker nodes")
	}

	_ = wait.ExponentialBackoff(defaultRetryBackoff(3), func() (bool, error) {
		lastErr = machinecontroller.WaitDestroy(s)
		if lastErr != nil {
			s.Logger.Warn("Waiting for all machines to be deleted…")
			return false, nil
		}
		return true, nil
	})
	if lastErr != nil {
		return errors.Wrap(lastErr, "error waiting for machines to be deleted")
	}

	return nil
}

func resetNode(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Resetting node…")

	_, _, err := s.Runner.Run(resetScript, runner.TemplateVariables{
		"WORK_DIR": s.WorkDir,
		"VERBOSE":  s.KubeAdmVerboseFlag(),
	})

	return err
}

func removeBinaries(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Removing Kubernetes binaries")

	// Determine operating system
	os, err := determineOS(s)
	if err != nil {
		return errors.Wrap(err, "failed to determine operating system")
	}
	node.SetOperatingSystem(os)

	// Remove Kubernetes binaries
	switch node.OperatingSystem {
	case "ubuntu", "debian":
		err = removeBinariesDebian(s)
	case "centos":
		err = removeBinariesCentOS(s)
	case "coreos":
		err = removeBinariesCoreOS(s)
	default:
		err = errors.Errorf("'%s' is not a supported operating system", node.OperatingSystem)
	}

	return err
}

func removeBinariesDebian(s *state.State) error {
	_, _, err := s.Runner.Run(removeBinariesDebianCommand, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removeBinariesCentOS(s *state.State) error {
	_, _, err := s.Runner.Run(removeBinariesCentOSCommand, runner.TemplateVariables{
		"KUBERNETES_VERSION": s.Cluster.Versions.Kubernetes,
		"CNI_VERSION":        s.Cluster.Versions.KubernetesCNIVersion(),
	})

	return errors.WithStack(err)
}

func removeBinariesCoreOS(s *state.State) error {
	_, _, err := s.Runner.Run(removeBinariesCoreOSCommand, runner.TemplateVariables{})

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
sudo kubeadm {{ .VERBOSE }} reset --force || true
sudo rm -f /etc/kubernetes/cloud-config
sudo rm -rf /var/lib/etcd/
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
