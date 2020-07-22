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

package tasks

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/clusterstatus/apiserverstatus"
	"github.com/kubermatic/kubeone/pkg/clusterstatus/etcdstatus"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemdShowStatusCMD = `systemctl show %s -p LoadState,ActiveState,SubState`

	dockerVersionDPKG = `dpkg-query --show --showformat='${Version}' docker-ce | cut -d: -f2 | cut -d~ -f1`
	dockerVersionRPM  = `rpm -qa --queryformat '%{RPMTAG_VERSION}' docker-ce`

	kubeletVersionDPKG = `dpkg-query --show --showformat='${Version}' kubelet | cut -d- -f1`
	kubeletVersionRPM  = `rpm -qa --queryformat '%{RPMTAG_VERSION}' kubelet`
	kubeletVersionCLI  = `kubelet --version | cut -d' ' -f2`

	kubeletInitializedCMD = `test -f /etc/kubernetes/kubelet.conf`
)

func runProbes(s *state.State) error {
	expectedVersion, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	s.LiveCluster = &state.Cluster{
		ExpectedVersion: expectedVersion,
	}

	s.Logger.Info("Running host probes…")
	for i := range s.Cluster.ControlPlane.Hosts {
		s.LiveCluster.ControlPlane = append(s.LiveCluster.ControlPlane, state.Host{
			Config: &s.Cluster.ControlPlane.Hosts[i],
		})
	}
	for i := range s.Cluster.StaticWorkers.Hosts {
		s.LiveCluster.StaticWorkers = append(s.LiveCluster.StaticWorkers, state.Host{
			Config: &s.Cluster.StaticWorkers.Hosts[i],
		})
	}

	if err := s.RunTaskOnAllNodes(investigateHost, state.RunParallel); err != nil {
		return err
	}

	if s.LiveCluster.IsProvisioned() {
		return investigateCluster(s)
	}

	return nil
}

func investigateHost(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	var (
		idx          int
		h            *state.Host
		controlPlane bool
	)

	s.LiveCluster.Lock.Lock()
	for i := range s.LiveCluster.ControlPlane {
		host := s.LiveCluster.ControlPlane[i]
		if host.Config.Hostname == node.Hostname {
			h = &host
			idx = i
			controlPlane = true
			break
		}
	}
	if h == nil {
		for i := range s.LiveCluster.StaticWorkers {
			host := s.LiveCluster.StaticWorkers[i]
			if host.Config.Hostname == node.Hostname {
				h = &host
				idx = i
				break
			}
		}
	}
	s.LiveCluster.Lock.Unlock()

	if h == nil {
		return errors.New("didn't matched live cluster against provided")
	}

	if err := detectDockerStatusVersion(h, conn); err != nil {
		return err
	}

	if err := detectKubeletStatusVersion(h, conn); err != nil {
		return err
	}

	if err := detectKubeletInitialized(h, conn); err != nil {
		return err
	}

	s.LiveCluster.Lock.Lock()
	if controlPlane {
		s.LiveCluster.ControlPlane[idx] = *h
	} else {
		s.LiveCluster.StaticWorkers[idx] = *h
	}
	s.LiveCluster.Lock.Unlock()
	return nil
}

func investigateCluster(s *state.State) error {
	if !s.LiveCluster.IsProvisioned() {
		return errors.New("unable to investigate non-provisioned cluster")
	}

	s.Logger.Info("Electing cluster leader…")
	s.LiveCluster.Lock.Lock()
	for i := range s.LiveCluster.ControlPlane {
		s.LiveCluster.ControlPlane[i].Config.IsLeader = false
	}

	leaderElected := false
	for i := range s.LiveCluster.ControlPlane {
		apiserverStatus, _ := apiserverstatus.Get(s, *s.LiveCluster.ControlPlane[i].Config)
		if apiserverStatus != nil && apiserverStatus.Health {
			s.LiveCluster.ControlPlane[i].APIServer.Status |= state.PodRunning
			if !leaderElected {
				s.LiveCluster.ControlPlane[i].Config.IsLeader = true
				leaderElected = true
				s.Logger.Infof("Elected leader %q…\n", s.LiveCluster.ControlPlane[i].Config.Hostname)
			}
		}
	}
	if !leaderElected {
		s.Logger.Errorln("Failed to elect leader.")
		s.Logger.Errorln("Quorum is mostly like lost, manual cluster repair might be needed.")
		s.Logger.Errorln("Consider the KubeOne documentation for further steps.")
		return errors.New("leader not elected, quorum mostly like lost")
	}

	etcdMembers, err := etcdstatus.MemberList(s)
	if err != nil {
		return err
	}
	for i := range s.LiveCluster.ControlPlane {
		etcdStatus, _ := etcdstatus.Get(s, *s.LiveCluster.ControlPlane[i].Config, etcdMembers)
		if etcdStatus != nil {
			if etcdStatus.Member && etcdStatus.Health {
				s.LiveCluster.ControlPlane[i].Etcd.Status |= state.PodRunning
			}
		}
	}
	s.LiveCluster.Lock.Unlock()

	if s.DynamicClient == nil {
		if err := kubeconfig.BuildKubernetesClientset(s); err != nil {
			return err
		}
	}

	s.Logger.Info("Running cluster probes…")

	// Get the node list
	nodes := corev1.NodeList{}
	if err := s.DynamicClient.List(s.Context, &nodes, &dynclient.ListOptions{}); err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}

	// Parse the node list
	knownHostsIdentities := sets.NewString()
	knownNodesIdentities := sets.NewString()

	for _, host := range s.LiveCluster.ControlPlane {
		knownHostsIdentities.Insert(host.Config.Hostname)
	}
	for _, host := range s.LiveCluster.StaticWorkers {
		knownHostsIdentities.Insert(host.Config.Hostname)
	}

	s.LiveCluster.Lock.Lock()
	for _, node := range nodes.Items {
		knownNodesIdentities.Insert(node.Name)
		if knownHostsIdentities.Has(node.Name) {
			found := false
			for i := range s.LiveCluster.ControlPlane {
				if node.Name == s.LiveCluster.ControlPlane[i].Config.Hostname {
					s.LiveCluster.ControlPlane[i].IsInCluster = true
					found = true
					break
				}
			}
			if found {
				continue
			}
			for i := range s.LiveCluster.StaticWorkers {
				if node.Name == s.LiveCluster.StaticWorkers[i].Config.Hostname {
					s.LiveCluster.StaticWorkers[i].IsInCluster = true
					break
				}
			}
		}
	}
	s.LiveCluster.Lock.Unlock()

	return nil
}

func detectDockerStatusVersion(host *state.Host, conn ssh.Connection) error {
	var err error
	host.ContainerRuntime.Status, err = systemdStatus(conn, "docker")
	if err != nil {
		return err
	}

	if host.ContainerRuntime.Status&state.ComponentInstalled == 0 {
		// docker is not installed
		return nil
	}

	var dockerVersionCmd string

	switch host.Config.OperatingSystem {
	case kubeoneapi.OperatingSystemNameCentOS, kubeoneapi.OperatingSystemNameRHEL:
		dockerVersionCmd = dockerVersionRPM
	case kubeoneapi.OperatingSystemNameUbuntu:
		dockerVersionCmd = dockerVersionDPKG
	case kubeoneapi.OperatingSystemNameFlatcar, kubeoneapi.OperatingSystemNameCoreOS:
		// we don't care about version because on container linux we don't manage docker
		host.ContainerRuntime.Version = &semver.Version{}
		return nil
	default:
		return nil
	}

	out, _, _, err := conn.Exec(dockerVersionCmd)
	if err != nil {
		return err
	}

	ver, err := semver.NewVersion(strings.TrimSpace(out))
	if err != nil {
		return errors.Wrapf(err, "version was: %q", out)
	}
	host.ContainerRuntime.Version = ver

	return nil
}

func detectKubeletStatusVersion(host *state.Host, conn ssh.Connection) error {
	var err error
	host.Kubelet.Status, err = systemdStatus(conn, "kubelet")
	if err != nil {
		return err
	}

	if host.Kubelet.Status&state.ComponentInstalled == 0 {
		// kubelet is not installed
		return nil
	}

	var kubeletVersionCmd string

	switch host.Config.OperatingSystem {
	case kubeoneapi.OperatingSystemNameCentOS, kubeoneapi.OperatingSystemNameRHEL:
		kubeletVersionCmd = kubeletVersionRPM
	case kubeoneapi.OperatingSystemNameUbuntu:
		kubeletVersionCmd = kubeletVersionDPKG
	case kubeoneapi.OperatingSystemNameFlatcar, kubeoneapi.OperatingSystemNameCoreOS:
		kubeletVersionCmd = kubeletVersionCLI
	default:
		return nil
	}

	out, _, _, err := conn.Exec(kubeletVersionCmd)
	if err != nil {
		return err
	}

	ver, err := semver.NewVersion(strings.TrimSpace(out))
	if err != nil {
		return err
	}
	host.Kubelet.Version = ver

	return nil
}

func detectKubeletInitialized(host *state.Host, conn ssh.Connection) error {
	_, _, exitcode, err := conn.Exec(kubeletInitializedCMD)
	if err != nil && exitcode <= 0 {
		// If there's an error and exit code is 0, there's mostly like a connection
		// error. If exit code is -1, there might be a session problem.
		return err
	}

	if exitcode == 0 {
		host.Kubelet.Status |= state.KubeletInitialized
	}

	return nil
}

func systemdStatus(conn ssh.Connection, service string) (uint64, error) {
	out, _, _, err := conn.Exec(fmt.Sprintf(systemdShowStatusCMD, service))
	if err != nil {
		return 0, err
	}

	out = strings.ReplaceAll(out, "=", ": ")
	m := map[string]string{}
	if err = yaml.Unmarshal([]byte(out), &m); err != nil {
		return 0, err
	}

	var status uint64

	if m["LoadState"] == "loaded" {
		status |= state.ComponentInstalled
	}

	switch m["ActiveState"] {
	case "active", "activating":
		status |= state.SystemDStatusActive
	}

	switch m["SubState"] {
	case "running":
		status |= state.SystemDStatusRunning
	case "auto-restart":
		status |= state.SystemDStatusRestarting
	case "dead":
		status |= state.SystemdDStatusDead
	default:
		status |= state.SystemDStatusUnknown
	}

	return status, nil
}
