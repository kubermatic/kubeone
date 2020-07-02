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

package state

import (
	"sync"

	"github.com/Masterminds/semver"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubermatic/kubeone/pkg/apis/kubeone"
)

type Cluster struct {
	ControlPlane []Host
	Workers      []Host
	Lock         sync.Mutex
}

type Host struct {
	ContainerRuntime ComponentStatus
	Hostname         string
	IsInCluster      bool
	Kubeconfig       []byte
	Kubernetes       ComponentStatus
	OS               kubeone.OperatingSystemName
	PrivateAddress   string
	PublicAddress    string
}

type ComponentStatus struct {
	Version *semver.Version
	Status  int64
}

const (
	SystemDStatusUnknown    = 1 << iota // systemd unit unknown
	SystemdDStatusDead                  // systemd unit dead
	SystemDStatusRestarting             // systemd unit restarting
	ComponentInstalled                  // installed (package, or direct download)
	SystemDStatusActive                 // systemd unit is activated
	SystemDStatusRunning                // systemd unit is running
)

func (c *Cluster) Healthy() bool {
	for i := range c.ControlPlane {
		if !c.ControlPlane[i].Healthy() {
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

func (h *Host) RestConfig() (*rest.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
}

func (h *Host) Ready() bool {
	return h.IsInCluster && h.Healthy()
}

func (h *Host) Healthy() bool {
	return h.ContainerRuntime.Healthy() && h.Kubernetes.Healthy()
}

func (cs *ComponentStatus) Healthy() bool {
	return cs.Status&SystemDStatusRunning != 0 && cs.Status&SystemDStatusRestarting == 0
}
