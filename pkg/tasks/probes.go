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
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clusterstatus/apiserverstatus"
	"k8c.io/kubeone/pkg/clusterstatus/etcdstatus"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	systemdShowStatusCMD    = `systemctl show %s -p LoadState,ActiveState,SubState`
	systemdShowExecStartCMD = `systemctl show %s -p ExecStart`

	kubeletInitializedCMD = `test -f /etc/kubernetes/kubelet.conf`

	k8sAppLabel               = "k8s-app"
	openstackCCMAppLabelValue = "openstack-cloud-controller-manager"
)

var KubeProxyObjectKey = dynclient.ObjectKey{
	Namespace: "kube-system",
	Name:      "kube-proxy",
}

func safeguard(s *state.State) error {
	if !s.LiveCluster.IsProvisioned() {
		return nil
	}

	if s.Cluster.ClusterNetwork.KubeProxy != nil && s.Cluster.ClusterNetwork.KubeProxy.SkipInstallation {
		var kubeProxyDs appsv1.DaemonSet
		if err := s.DynamicClient.Get(s.Context, KubeProxyObjectKey, &kubeProxyDs); err != nil {
			if !k8serrors.IsNotFound(err) {
				return fail.KubeClient(err, "getting kube-proxy daemonset")
			}
		} else {
			return fail.RuntimeError{
				Err: errors.New("is enabled but kube-proxy was already installed and requires manual deletion"),
				Op:  ".clusterNetwork.kubeProxy.skipInstallation",
			}
		}
	}

	var nodes corev1.NodeList
	if err := s.DynamicClient.List(s.Context, &nodes); err != nil {
		return fail.KubeClient(err, "getting %T", nodes)
	}

	cr := s.Cluster.ContainerRuntime
	configuredClusterContainerRuntime := cr.String()

	for _, node := range nodes.Items {
		if !s.Cluster.IsManagedNode(node.Name) {
			// skip nodes unknown to the current configuration (most likely, machine-controller nodes)
			continue
		}

		nodesContainerRuntime := strings.Split(node.Status.NodeInfo.ContainerRuntimeVersion, ":")[0]

		if nodesContainerRuntime != configuredClusterContainerRuntime {
			errMsg := "Migration is not supported yet"
			if nodesContainerRuntime == "docker" {
				errMsg = "Support for docker will be removed with Kubernetes 1.24 release. It is recommended to switch to containerd as container runtime using `kubeone migrate to-containerd`. To continue using Docker please specify ContainerRuntime explicitly in KubeOneCluster manifest"
			} else if cr.Containerd != nil {
				errMsg = "Use `kubeone migrate to-containerd`"
			}

			return fail.RuntimeError{
				Err: errors.Errorf("on node %q is %q, but %q is configured. %s",
					node.Name,
					nodesContainerRuntime,
					configuredClusterContainerRuntime,
					errMsg,
				),
				Op: "container runtime",
			}
		}
	}

	// Block kubeone apply if .cloudProvider.external is enabled on cluster with
	// in-tree cloud provider, but with no external CCM
	st := s.LiveCluster.CCMStatus
	if st != nil {
		if s.Cluster.CloudProvider.External {
			if st.InTreeCloudProviderEnabled && !st.ExternalCCMDeployed {
				return fail.RuntimeError{
					Err: errors.New("cluster is using in-tree provider. run ccm/csi migration by running 'kubeone migrate to-ccm-csi'"),
					Op:  ".cloudProvider.external is enabled",
				}
			}
		} else {
			if st.ExternalCCMDeployed {
				// Block disabling .cloudProvider.external
				return fail.RuntimeError{
					Err: errors.New("external ccm is deployed"),
					Op:  ".cloudProvider.external is disabled",
				}
			}
		}
	}

	return nil
}

func runProbes(s *state.State) error {
	expectedVersion, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return fail.ConfigValidation(err)
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

	clusterName, cnErr := detectClusterName(s)
	if cnErr != nil {
		return errors.Wrap(cnErr, "failed to detect the ccm --cluster-name flag value")
	}
	s.LiveCluster.Lock.Lock()
	s.LiveCluster.CCMClusterName = clusterName
	s.LiveCluster.Lock.Unlock()

	switch {
	case s.Cluster.ContainerRuntime.Containerd != nil:
		return nil
	case s.Cluster.ContainerRuntime.Docker != nil:
		return nil
	}

	gteKube124Condition, _ := semver.NewConstraint(">= 1.24")

	switch {
	case gteKube124Condition.Check(s.LiveCluster.ExpectedVersion):
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

func investigateHost(s *state.State, node *kubeoneapi.HostConfig, conn executor.Interface) error {
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

		return fail.RuntimeError{
			Err: errors.New("quorum mostly like lost"),
			Op:  "leader electing",
		}
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
		return fail.KubeClient(err, "getting %T", nodes)
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
		return err
	}
	if encryptionEnabled.Enabled {
		s.LiveCluster.Lock.Lock()
		s.LiveCluster.EncryptionConfiguration = &state.EncryptionConfiguration{Enable: true, Custom: encryptionEnabled.Custom}
		s.LiveCluster.Lock.Unlock()
		// no need to lock around FetchEncryptionProvidersFile because it handles locking internally.
		if fErr := fetchEncryptionProvidersFile(s); fErr != nil {
			return err
		}
	}

	ccmStatus, err := detectCCMMigrationStatus(s)
	if err != nil {
		return err
	}
	if ccmStatus != nil {
		s.LiveCluster.Lock.Lock()
		s.LiveCluster.CCMStatus = ccmStatus
		s.LiveCluster.Lock.Unlock()
	}

	return nil
}

type systemdUnitInfoOpt func(component *state.ComponentStatus, conn executor.Interface) error

func systemdUnitInfo(name string, conn executor.Interface, opts ...systemdUnitInfoOpt) (state.ComponentStatus, error) {
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

func withFlatcarContainerRuntimeVersion(component *state.ComponentStatus, conn executor.Interface) error {
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
	return func(component *state.ComponentStatus, conn executor.Interface) error {
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

func detectKubeletInitialized(host *state.Host, conn executor.Interface) error {
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

func systemdUnitExecStartPath(conn executor.Interface, unitName string) (string, error) {
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

func systemdStatus(conn executor.Interface, service string) (uint64, error) {
	out, _, _, err := conn.Exec(fmt.Sprintf(systemdShowStatusCMD, service))
	if err != nil {
		return 0, fail.Runtime(err, "ckecking %q systemd service status", service)
	}

	out = strings.ReplaceAll(out, "=", ": ")
	m := map[string]string{}
	if err = yaml.Unmarshal([]byte(out), &m); err != nil {
		return 0, fail.Runtime(err, "unmarshalling systemd status %q", service)
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
		return ees, fail.NoKubeClient()
	}

	pods := corev1.PodList{}
	err = s.DynamicClient.List(s.Context, &pods, &dynclient.ListOptions{
		Namespace: "kube-system",
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "kube-apiserver"})})
	if err != nil {
		return ees, fail.KubeClient(err, "listing kube-apiserver pods")
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

func detectCCMMigrationStatus(s *state.State) (*state.CCMStatus, error) {
	if s.DynamicClient == nil {
		return nil, fail.NoKubeClient()
	}

	pods := corev1.PodList{}
	err := s.DynamicClient.List(s.Context, &pods, &dynclient.ListOptions{
		Namespace: metav1.NamespaceSystem,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "kube-controller-manager",
		}),
	})
	if err != nil {
		return nil, fail.KubeClient(err, "listing kube-controller-manager pods")
	}

	// This uses regex so we can easily match any CSIMigration feature gate
	// and confirm it's enabled.
	csiFlagRegex := regexp.MustCompile(`CSIMigration[a-zA-Z]+=true`)
	status := &state.CCMStatus{}
	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers[0].Command {
			switch {
			case strings.HasPrefix(c, "--cloud-provider") && !strings.Contains(c, "external"):
				status.InTreeCloudProviderEnabled = true
			case strings.HasPrefix(c, "--feature-gates"):
				if csiFlagRegex.MatchString(c) {
					status.CSIMigrationEnabled = true
				}
				unregister := s.Cluster.InTreePluginUnregisterFeatureGate()

				foundUnregister := 0
				for _, u := range unregister {
					if strings.Contains(c, fmt.Sprintf("%s=true", u)) {
						foundUnregister++
					}
				}
				if len(unregister) > 0 && foundUnregister == len(unregister) {
					status.InTreeCloudProviderUnregistered = true
				}
			}
		}
	}

	ccmLabel := k8sAppLabel
	var ccmLabelValue string

	switch {
	case s.Cluster.CloudProvider.Azure != nil:
		ccmLabelValue = "azure-cloud-controller-manager"
	case s.Cluster.CloudProvider.Openstack != nil:
		ccmLabelValue = openstackCCMAppLabelValue
	case s.Cluster.CloudProvider.Vsphere != nil:
		ccmLabelValue = "vsphere-cloud-controller-manager"
	default:
		status.ExternalCCMDeployed = false

		return status, nil
	}

	pods = corev1.PodList{}
	err = s.DynamicClient.List(s.Context, &pods, &dynclient.ListOptions{
		Namespace: metav1.NamespaceSystem,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			ccmLabel: ccmLabelValue,
		}),
	})
	if err != nil {
		return nil, fail.KubeClient(err, "listing CCM pods")
	}
	if len(pods.Items) > 0 {
		status.ExternalCCMDeployed = true
	}

	return status, nil
}

// detectClusterName is used to detect the value that should be passed to the
// external CCM via the --cluster-name flag.
//
// This function is currently used for OpenStack clusters, because we initially
// didn't set this flag, in which case it defaults to `kubernetes`.
//
// Not setting the flag can cause issues if there are multiple clusters in the
// same tenant. For example, Load Balancers with the same name in different
// clusters will share the same Octavia LB.
//
// Changing the --cluster-name causes the CCM to lose all references to the
// Load Balancers on OpenStack, because the cluster name is used as part of
// the reference to the LB. Therefore, we need this function to ensure the
// backwards compatibility.
//
// The function works in the following way:
//   * if the cluster is not provisioned, or if the cluster is not an OpenStack
//     cluster, return the KubeOne cluster name
//   * if it's an existing OpenStack cluster:
//      * if cluster is running in-tree cloud provider: return the KubeOne
//        cluster name because the in-tree provider already has the
//        --cluster-name flag set
//      * if cluster is running external cloud provider: check if there is
//        `--cluster-name` flag on the OpenStack CCM. If there is, read the
//        value and return it, otherwise don't set the OpenStack cluster name,
//        in which case it defaults to `kubernetes`
//   * if cluster is migrated to external CCM, return the KubeOne cluster name
//
// If an operator wants to change the --cluster-name flag on OpenStack external
// CCM, they need to edit the CCM DaemonSet manually. KubeOne will
// automatically pick up the provided value when reconciling the cluster.
func detectClusterName(s *state.State) (string, error) {
	if !s.LiveCluster.IsProvisioned() ||
		s.LiveCluster.CCMStatus == nil ||
		s.Cluster.CloudProvider.Openstack == nil {
		return s.Cluster.Name, nil
	}

	if s.LiveCluster.CCMStatus.InTreeCloudProviderEnabled && !s.LiveCluster.CCMStatus.ExternalCCMDeployed {
		return s.Cluster.Name, nil
	}

	pods := corev1.PodList{}
	err := s.DynamicClient.List(s.Context, &pods, &dynclient.ListOptions{
		Namespace: metav1.NamespaceSystem,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			k8sAppLabel: openstackCCMAppLabelValue,
		}),
	})
	if err != nil {
		return "", fail.KubeClient(err, "openstack CCM pod listing")
	}

	if len(pods.Items) == 0 || len(pods.Items[0].Spec.Containers) == 0 {
		return "", fail.RuntimeError{
			Op:  "checking containers of openstack CCM pods",
			Err: errors.New("no containers found"),
		}
	}

	for _, container := range pods.Items[0].Spec.Containers {
		if container.Name != openstackCCMAppLabelValue {
			continue
		}
		for _, flag := range container.Command {
			if strings.HasPrefix(flag, "--cluster-name") {
				return strings.Split(flag, "=")[1], nil
			}
		}
	}

	// If we got here, the cluster is running external CCM, but we didn't
	// find the --cluster-name flag, therefore assume default value.
	return "", nil
}
