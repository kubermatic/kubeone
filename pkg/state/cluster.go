/*
Copyright 2020 The KubeOne Authors.

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

package state

import (
	"sync"

	"github.com/Masterminds/semver"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

type Cluster struct {
	ControlPlane    []Host
	Workers         []Host
	ExpectedVersion *semver.Version
	Lock            sync.Mutex
}

type Host struct {
	Hostname string
	Host     *kubeone.HostConfig

	ContainerRuntime ComponentStatus
	Kubelet          ComponentStatus

	// Applicable only for CP nodes
	APIServer ContainerStatus
	Etcd      ContainerStatus

	IsInCluster bool
	Kubeconfig  []byte
}

type ComponentStatus struct {
	Version *semver.Version
	Status  uint64
}

type ContainerStatus struct {
	Version *semver.Version
	Status  uint64
}

const (
	SystemDStatusUnknown    = 1 << iota // systemd unit unknown
	SystemdDStatusDead                  // systemd unit dead
	SystemDStatusRestarting             // systemd unit restarting
	ComponentInstalled                  // installed (package, or direct download)
	SystemDStatusActive                 // systemd unit is activated
	SystemDStatusRunning                // systemd unit is running
	KubeletInitialized                  // kubelet config found (means node is initialized)
	PodRunning                          // pod is running
)

func (c *Cluster) IsProvisioned() bool {
	for i := range c.ControlPlane {
		if c.ControlPlane[i].Initialized() {
			return true
		}
	}

	return false
}

func (c *Cluster) IsDegraded() bool {
	for i := range c.ControlPlane {
		if c.ControlPlane[i].IsDegraded() {
			return true
		}
	}

	return false
}

func (c *Cluster) Healthy() bool {
	if !c.QuorumSatisfied() {
		return false
	}

	for i := range c.ControlPlane {
		if !c.ControlPlane[i].ControlPlaneHealthy() {
			return false
		}
	}

	for i := range c.Workers {
		if !c.Workers[i].Healthy() {
			return false
		}
	}

	return true
}

func (c *Cluster) QuorumSatisfied() bool {
	var healthyNodes int
	quorum := int(float64(((len(c.ControlPlane) / 2) + 1)))
	tolerance := len(c.ControlPlane) - quorum

	for i := range c.ControlPlane {
		if c.ControlPlane[i].Healthy() {
			healthyNodes++
		}
	}

	return healthyNodes >= tolerance
}

func (c *Cluster) UpgradeNeeded() bool {
	for i := range c.ControlPlane {
		// TODO: We should eventually error if expected version is lower than
		// current, since downgrades aren't allowed
		if c.ExpectedVersion.GreaterThan(c.ControlPlane[i].Kubelet.Version) {
			return true
		}
	}

	for i := range c.Workers {
		if c.ExpectedVersion.GreaterThan(c.Workers[i].Kubelet.Version) {
			return true
		}
	}

	return false
}

func (c *Cluster) UpgradeMachinesNeeded() bool {
	// TODO: implement
	return false
}

func (h *Host) RestConfig() (*rest.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
}

func (h *Host) Initialized() bool {
	return h.IsProvisioned() && h.Kubelet.Status&KubeletInitialized != 0
}

func (h *Host) Ready() bool {
	return h.IsInCluster && h.Healthy()
}

func (h *Host) IsProvisioned() bool {
	return h.ContainerRuntime.IsProvisioned() && h.Kubelet.IsProvisioned()
}

func (h *Host) Healthy() bool {
	return h.ContainerRuntime.Healthy() && h.Kubelet.Healthy()
}

func (h *Host) ControlPlaneHealthy() bool {
	return h.Healthy() && h.Etcd.Healthy() && h.APIServer.Healthy()
}

func (h *Host) IsDegraded() bool {
	return !h.IsInCluster || h.APIServer.Status&PodRunning == 0
}

func (cs *ComponentStatus) IsProvisioned() bool {
	return cs.Status&(SystemDStatusRunning|ComponentInstalled|SystemDStatusActive) != 0
}

func (cs *ComponentStatus) Healthy() bool {
	return cs.Status&SystemDStatusRunning != 0 && cs.Status&SystemDStatusRestarting == 0
}

func (cs *ContainerStatus) Healthy() bool {
	return cs.Status&PodRunning != 0
}
