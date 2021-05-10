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

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clusterstatus/apiserverstatus"
	"k8c.io/kubeone/pkg/clusterstatus/etcdstatus"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemdShowStatusCMD    = `systemctl show %s -p LoadState,ActiveState,SubState`
	systemdShowExecStartCMD = `systemctl show %s -p ExecStart`

	kubeletInitializedCMD = `test -f /etc/kubernetes/kubelet.conf`
)

func safeguard(s *state.State) error {
	if !s.LiveCluster.IsProvisioned() {
		return nil
	}

	var nodes corev1.NodeList
	if err := s.DynamicClient.List(s.Context, &nodes); err != nil {
		return err
	}

	configuredClusterContainerRuntime := s.Cluster.ContainerRuntime.String()

	for _, node := range nodes.Items {
		if !s.Cluster.IsManagedNode(node.Name) {
			// skip nodes unknown to the current configuration (most likely, machine-controller nodes)
			continue
		}

		nodesContainerRuntime := strings.Split(node.Status.NodeInfo.ContainerRuntimeVersion, ":")[0]

		if nodesContainerRuntime != configuredClusterContainerRuntime {
			return errors.Errorf(
				"Container runtime on node %q is %q, but %q is configured. Migration is not supported yet.",
				node.Name,
				nodesContainerRuntime,
				configuredClusterContainerRuntime,
			)
		}
	}

	return nil
}

func runProbes(s *state.State) error {
	expectedVersion, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	s.LiveCluster = &state.Cluster{
		ExpectedVersion: expectedVersion,
		EncryptionConfiguration: &state.EncryptionConfiguration{
			Enable: false,
		},
	}

	s.Logger.Info("Running host probes...")
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
		if err := investigateCluster(s); err != nil {
			return err
		}
	}

	switch {
	case s.Cluster.ContainerRuntime.Containerd != nil:
		return nil
	case s.Cluster.ContainerRuntime.Docker != nil:
		return nil
	}

	gteKube121Condition, _ := semver.NewConstraint(">= 1.21")

	switch {
	case gteKube121Condition.Check(s.LiveCluster.ExpectedVersion):
		s.Cluster.ContainerRuntime.Containerd = &kubeoneapi.ContainerRuntimeContainerd{}

		if s.LiveCluster.IsProvisioned() {
			for _, host := range s.LiveCluster.ControlPlane {
				if host.ContainerRuntimeDocker.IsProvisioned() {
					s.Cluster.ContainerRuntime.Docker = &kubeoneapi.ContainerRuntimeDocker{}
					s.Cluster.ContainerRuntime.Containerd = nil
				}
			}
		}
	default:
		s.Cluster.ContainerRuntime.Docker = &kubeoneapi.ContainerRuntimeDocker{}
	}

	return nil
}

func versionCmdGenerator(execPath string) string {
	return fmt.Sprintf("%s --version | awk '{print $3}' | awk -F - '{print $1}'  | awk -F , '{print $1}'", execPath)
}

func kubeletVersionCmdGenerator(execPath string) string {
	return fmt.Sprintf("%s --version | awk '{print $2}'", execPath)
}

func investigateHost(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	var (
		idx          int
		foundHost    *state.Host
		controlPlane bool
	)

	s.LiveCluster.Lock.Lock()
	for i := range s.LiveCluster.ControlPlane {
		host := s.LiveCluster.ControlPlane[i]
		if host.Config.Hostname == node.Hostname {
			foundHost = &host
			idx = i
			controlPlane = true
			break
		}
	}
	if foundHost == nil {
		for i := range s.LiveCluster.StaticWorkers {
			host := s.LiveCluster.StaticWorkers[i]
			if host.Config.Hostname == node.Hostname {
				foundHost = &host
				idx = i
				break
			}
		}
	}
	s.LiveCluster.Lock.Unlock()

	if foundHost == nil {
		return errors.New("didn't matched live cluster against provided")
	}

	var err error

	containerRuntimeOpts := []systemdUnitInfoOpt{withComponentVersion(versionCmdGenerator)}

	if foundHost.Config.OperatingSystem == kubeoneapi.OperatingSystemNameFlatcar {
		// Flatcar is special
		containerRuntimeOpts = []systemdUnitInfoOpt{withFlatcarContainerRuntimeVersion}
	}

	foundHost.ContainerRuntimeContainerd, err = systemdUnitInfo("containerd", conn, containerRuntimeOpts...)
	if err != nil {
		return err
	}

	foundHost.ContainerRuntimeDocker, err = systemdUnitInfo("docker", conn, containerRuntimeOpts...)
	if err != nil {
		return err
	}

	foundHost.Kubelet, err = systemdUnitInfo("kubelet", conn, withComponentVersion(kubeletVersionCmdGenerator))
	if err != nil {
		return err
	}

	if err = detectKubeletInitialized(foundHost, conn); err != nil {
		return err
	}

	if foundHost.Initialized() && controlPlane {
		foundHost.EarliestCertExpiry, err = earliestCertExpiry(conn)
		if err != nil {
			return err
		}
	}

	s.LiveCluster.Lock.Lock()
	if controlPlane {
		s.LiveCluster.ControlPlane[idx] = *foundHost
	} else {
		s.LiveCluster.StaticWorkers[idx] = *foundHost
	}
	s.LiveCluster.Lock.Unlock()
	return nil
}

func investigateCluster(s *state.State) error {
	if !s.LiveCluster.IsProvisioned() {
		return errors.New("unable to investigate non-provisioned cluster")
	}

	s.Logger.Info("Electing cluster leader...")
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
				s.Logger.Infof("Elected leader %q...", s.LiveCluster.ControlPlane[i].Config.Hostname)
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
		if err = kubeconfig.BuildKubernetesClientset(s); err != nil {
			return err
		}
	}

	s.Logger.Info("Running cluster probes...")

	// Get the node list
	nodes := corev1.NodeList{}
	if err = s.DynamicClient.List(s.Context, &nodes, &dynclient.ListOptions{}); err != nil {
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
	encryptionEnabled, err := detectEncryptionProvidersEnabled(s)
	if err != nil {
		return errors.Wrap(err, "failed to check for EncryptionProviders")
	}
	if encryptionEnabled.Enabled {
		s.LiveCluster.Lock.Lock()
		s.LiveCluster.EncryptionConfiguration = &state.EncryptionConfiguration{Enable: true, Custom: encryptionEnabled.Custom}
		s.LiveCluster.Lock.Unlock()
		// no need to lock around FetchEncryptionProvidersFile because it handles locking internally.
		if err := fetchEncryptionProvidersFile(s); err != nil {
			return errors.Wrap(err, "failed to fetch EncryptionProviders configuration")
		}
	}
	return nil
}

type systemdUnitInfoOpt func(component *state.ComponentStatus, conn ssh.Connection) error

func systemdUnitInfo(name string, conn ssh.Connection, opts ...systemdUnitInfoOpt) (state.ComponentStatus, error) {
	var (
		compStatus = state.ComponentStatus{Name: name}
		err        error
	)

	compStatus.Status, err = systemdStatus(conn, name)
	if err != nil {
		return compStatus, err
	}

	if compStatus.Status&state.ComponentInstalled == 0 {
		// provided containerRuntime is not known to systemd, we consider this as not installed
		return compStatus, nil
	}

	for _, fn := range opts {
		if err := fn(&compStatus, conn); err != nil {
			return compStatus, err
		}
	}

	return compStatus, nil
}

func withFlatcarContainerRuntimeVersion(component *state.ComponentStatus, conn ssh.Connection) error {
	cmd := versionCmdGenerator(fmt.Sprintf("/run/torcx/bin/%s", component.Name))

	out, _, _, err := conn.Exec(cmd)
	if err != nil {
		return err
	}

	ver, err := semver.NewVersion(strings.TrimSpace(out))
	if err != nil {
		return errors.Wrapf(err, "%s version was: %q", component.Name, out)
	}

	component.Version = ver

	return nil
}

func withComponentVersion(versionCmdGenerator func(string) string) systemdUnitInfoOpt {
	return func(component *state.ComponentStatus, conn ssh.Connection) error {
		execPath, err := systemdUnitExecStartPath(conn, component.Name)
		if err != nil {
			return err
		}

		out, _, _, err := conn.Exec(versionCmdGenerator(execPath))
		if err != nil {
			return err
		}

		ver, err := semver.NewVersion(strings.TrimSpace(out))
		if err != nil {
			return errors.Wrapf(err, "%s version was: %q", component.Name, out)
		}

		component.Version = ver

		return nil
	}
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

func systemdUnitExecStartPath(conn ssh.Connection, unitName string) (string, error) {
	out, _, _, err := conn.Exec(fmt.Sprintf(systemdShowExecStartCMD, unitName))
	if err != nil {
		return "", err
	}

	lines := strings.Split(out, " ")
	for _, line := range lines {
		if strings.HasPrefix(line, "path=") {
			pathSplit := strings.Split(line, "=")
			if len(pathSplit) == 2 {
				return pathSplit[1], nil
			}
		}
	}

	return "", errors.Errorf("ExecStart not found in %q systemd unit", unitName)
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

type encryptionEnabledStatus struct {
	Enabled bool
	Custom  bool
}

func detectEncryptionProvidersEnabled(s *state.State) (ees encryptionEnabledStatus, err error) {
	if s.DynamicClient == nil {
		return ees, errors.New("kubernetes dynamic client is not initialized")
	}
	pods := corev1.PodList{}
	err = s.DynamicClient.List(s.Context, &pods, &dynclient.ListOptions{
		Namespace: "kube-system",
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "kube-apiserver"})})
	if err != nil {
		return ees, errors.Wrap(err, "unable to list pods")
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers[0].Command {
			if strings.HasPrefix(c, "--encryption-provider") {
				ees.Enabled = true
			}
			if strings.Contains(c, "encryption-providers/custom-encryption-providers.yaml") {
				ees.Custom = true
			}
		}
	}
	return ees, nil
}
