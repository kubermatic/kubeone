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
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"
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

	for i := range s.Cluster.ControlPlane.Hosts {
		s.LiveCluster.ControlPlane = append(s.LiveCluster.ControlPlane, state.Host{
			Hostname: s.Cluster.ControlPlane.Hosts[i].Hostname,
			Host:     &s.Cluster.ControlPlane.Hosts[i],
		})
	}

	return s.RunTaskOnControlPlane(investigateHost, state.RunParallel)
}

func investigateHost(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	var (
		idx int
		h   *state.Host
	)

	s.LiveCluster.Lock.Lock()
	for i := range s.LiveCluster.ControlPlane {
		host := s.LiveCluster.ControlPlane[i]
		if host.Hostname == node.Hostname {
			h = &host
			idx = i
			break
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

	fmt.Println("---------------")
	fmt.Printf("host: %q\n", h.Hostname)
	fmt.Printf("docker version: %q\n", h.ContainerRuntime.Version)
	fmt.Printf("docker is installed?: %t\n", h.ContainerRuntime.Status&state.ComponentInstalled != 0)
	fmt.Printf("docker is running?: %t\n", h.ContainerRuntime.Status&state.SystemDStatusRunning != 0)
	fmt.Printf("docker is active?: %t\n", h.ContainerRuntime.Status&state.SystemDStatusActive != 0)
	fmt.Printf("docker is restarting?: %t\n", h.ContainerRuntime.Status&state.SystemDStatusRestarting != 0)
	fmt.Println()

	fmt.Printf("kubelet version: %q\n", h.Kubelet.Version)
	fmt.Printf("kubelet is installed?: %t\n", h.Kubelet.Status&state.ComponentInstalled != 0)
	fmt.Printf("kubelet is running?: %t\n", h.Kubelet.Status&state.SystemDStatusRunning != 0)
	fmt.Printf("kubelet is active?: %t\n", h.Kubelet.Status&state.SystemDStatusActive != 0)
	fmt.Printf("kubelet is restarting?: %t\n", h.Kubelet.Status&state.SystemDStatusRestarting != 0)
	fmt.Printf("kubelet is initialized?: %t\n", h.Kubelet.Status&state.KubeletInitialized != 0)
	fmt.Println()

	s.LiveCluster.Lock.Lock()
	s.LiveCluster.ControlPlane[idx] = *h
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

	switch host.Host.OperatingSystem {
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

	switch host.Host.OperatingSystem {
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
