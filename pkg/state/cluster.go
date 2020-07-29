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
	"errors"
	"sync"

	"github.com/Masterminds/semver"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8c.io/kubeone/pkg/apis/kubeone"
)

type Cluster struct {
	ControlPlane    []Host
	StaticWorkers   []Host
	ExpectedVersion *semver.Version
	Lock            sync.Mutex
}

type Host struct {
	Config *kubeone.HostConfig

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
	Status uint64
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

/*
	Cluster level checks
*/

// IsProvisioned returns is the target cluster provisioned.
// The cluster is consider provisioned if there is at least one initialized host
func (c *Cluster) IsProvisioned() bool {
	for i := range c.ControlPlane {
		if c.ControlPlane[i].Initialized() {
			return true
		}
	}

	return false
}

// Healthy checks the cluster overall healthiness
func (c *Cluster) Healthy() bool {
	for i := range c.ControlPlane {
		if !c.ControlPlane[i].ControlPlaneHealthy() {
			return false
		}
	}

	for i := range c.StaticWorkers {
		if !c.StaticWorkers[i].WorkerHealthy() {
			return false
		}
	}

	return true
}

// BrokenHosts returns a list of broken hosts that needs to be removed manually
func (c *Cluster) BrokenHosts() []string {
	brokenNodes := []string{}
	for i := range c.ControlPlane {
		if c.ControlPlane[i].IsInCluster && !c.ControlPlane[i].ControlPlaneHealthy() {
			brokenNodes = append(brokenNodes, c.ControlPlane[i].Config.Hostname)
		}
	}
	for i := range c.StaticWorkers {
		if c.StaticWorkers[i].IsInCluster && !c.StaticWorkers[i].WorkerHealthy() {
			brokenNodes = append(brokenNodes, c.StaticWorkers[i].Config.Hostname)
		}
	}
	return brokenNodes
}

func (c *Cluster) SafeToDeleteHosts() []string {
	safeToDelete := []string{}
	deleteCandidate := []string{}
	tolerance := c.EtcdToleranceRemain()

	for i := range c.ControlPlane {
		if !c.ControlPlane[i].IsInCluster {
			continue
		}
		if !c.ControlPlane[i].Etcd.Healthy() {
			safeToDelete = append(safeToDelete, c.ControlPlane[i].Config.Hostname)
		} else if !c.ControlPlane[i].APIServer.Healthy() {
			deleteCandidate = append(deleteCandidate, c.ControlPlane[i].Config.Hostname)
		}
	}
	tolerance -= len(safeToDelete)
	if tolerance > 0 && len(deleteCandidate) > 0 {
		if tolerance >= len(deleteCandidate) {
			safeToDelete = append(safeToDelete, deleteCandidate...)
		} else {
			safeToDelete = append(safeToDelete, deleteCandidate[:tolerance]...)
		}
	}

	// Worker nodes are always safe to delete as quorum is not affected
	for i := range c.StaticWorkers {
		if c.StaticWorkers[i].IsInCluster && !c.StaticWorkers[i].WorkerHealthy() {
			safeToDelete = append(safeToDelete, c.StaticWorkers[i].Config.Hostname)
		}
	}

	return safeToDelete
}

// EtcdToleranceRemain returns how many non-working nodes can be removed at the same time.
func (c *Cluster) EtcdToleranceRemain() int {
	var healthyEtcd int
	for i := range c.ControlPlane {
		if c.ControlPlane[i].IsInCluster && c.ControlPlane[i].Etcd.Healthy() {
			healthyEtcd++
		}
	}

	quorum := int(float64((healthyEtcd / 2) + 1))
	tolerance := healthyEtcd - quorum

	return tolerance
}

// UpgradeNeeded compares actual and expected Kubernetes versions for control plane and static worker nodes
func (c *Cluster) UpgradeNeeded() (bool, error) {
	for i := range c.ControlPlane {
		verDiff := c.ExpectedVersion.Compare(c.ControlPlane[i].Kubelet.Version)
		if verDiff > 0 {
			return true, nil
		} else if verDiff < 0 {
			return false, errors.New("cluster downgrades are disallowed")
		}
	}

	for i := range c.StaticWorkers {
		verDiff := c.ExpectedVersion.Compare(c.StaticWorkers[i].Kubelet.Version)
		if verDiff > 0 {
			return true, nil
		} else if verDiff < 0 {
			return false, errors.New("cluster downgrades are disallowed")
		}
	}

	return false, nil
}

func (c *Cluster) SafeToRepair(targetVersion string) (bool, string) {
	targetVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return false, ""
	}

	var highestVer *semver.Version
	for _, host := range c.ControlPlane {
		if !host.IsInCluster {
			continue
		}
		if highestVer == nil || host.Kubelet.Version.GreaterThan(highestVer) {
			highestVer = host.Kubelet.Version
		}
	}

	if highestVer != nil && targetVer.GreaterThan(highestVer) {
		return false, highestVer.String()
	}

	return true, targetVer.String()
}

/*
	Host level checks
*/

// RestConfig grabs Kubeconfig from a node
func (h *Host) RestConfig() (*rest.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
}

// Initialized checks is a host provisioned and is kubelet initialized
func (h *Host) Initialized() bool {
	return h.IsProvisioned() && h.Kubelet.Status&KubeletInitialized != 0
}

// IsProvisioned checks are CRI and Kubelet provisioned on a host
func (h *Host) IsProvisioned() bool {
	return h.ContainerRuntime.IsProvisioned() && h.Kubelet.IsProvisioned()
}

// ControlPlaneHealthy checks is a control-plane host part of the cluster and are CRI, Kubelet, and API server healthy
func (h *Host) ControlPlaneHealthy() bool {
	return h.healthy() && h.APIServer.Healthy()
}

// WorkerHealthy checks is a worker host part of the cluster and are CRI and Kubelet healthy
func (h *Host) WorkerHealthy() bool {
	return h.healthy()
}

func (h *Host) healthy() bool {
	return h.IsInCluster && h.ContainerRuntime.Healthy() && h.Kubelet.Healthy()
}

/*
	Component status level checks
*/

// IsProvisioned checks is a component running, installed and active
func (cs *ComponentStatus) IsProvisioned() bool {
	return cs.Status&(SystemDStatusRunning|ComponentInstalled|SystemDStatusActive) != 0
}

// Healthy checks is a component running and not restarting
func (cs *ComponentStatus) Healthy() bool {
	return cs.Status&SystemDStatusRunning != 0 && cs.Status&SystemDStatusRestarting == 0
}

/*
	Container status level checks
*/

// Healthy checks is a pod running
func (cs *ContainerStatus) Healthy() bool {
	return cs.Status&PodRunning != 0
}
