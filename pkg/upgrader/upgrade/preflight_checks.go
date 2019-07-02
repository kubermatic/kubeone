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

package upgrade

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	kubeonecontext "github.com/kubermatic/kubeone/pkg/util/context"
	"github.com/kubermatic/kubeone/pkg/util/runner"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// runPreflightChecks runs all preflight checks
func runPreflightChecks(ctx *kubeonecontext.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	// Check are Docker, Kubelet and Kubeadm installed
	if err := checkPrerequisites(ctx); err != nil {
		return errors.Wrap(err, "unable to check are prerequisites installed")
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{}
	err := nodeListOpts.SetLabelSelector(fmt.Sprintf("%s=%s", labelControlPlaneNode, ""))
	if err != nil {
		return errors.Wrap(err, "failed to set node selector labels")
	}

	err = ctx.DynamicClient.List(context.Background(), &nodeListOpts, &nodes)
	if err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}

	if len(nodes.Items) != len(ctx.Cluster.Hosts) {
		return errors.Errorf("expected %d cluster nodes but got %d", len(ctx.Cluster.Hosts), len(nodes.Items))
	}

	// Run preflight checks on nodes
	ctx.Logger.Infoln("Running preflight checks…")

	ctx.Logger.Infoln("Verifying are all nodes running…")
	if err := verifyNodesRunning(&nodes, ctx.Verbose); err != nil {
		return errors.Wrap(err, "unable to verify are nodes running")
	}

	ctx.Logger.Infoln("Verifying are correct labels set on nodes…")
	if err := verifyLabels(&nodes, ctx.Verbose); err != nil {
		if ctx.ForceUpgrade {
			ctx.Logger.Warningf("unable to verify node labels: %v", err)
		} else {
			return errors.Wrap(err, "unable to verify node labels")
		}
	}

	ctx.Logger.Infoln("Verifying do all node IP addresses match with our state…")
	if err := verifyEndpoints(&nodes, ctx.Cluster.Hosts, ctx.Verbose); err != nil {
		return errors.Wrap(err, "unable to verify node endpoints")
	}

	ctx.Logger.Infoln("Verifying is it possible to upgrade to the desired version…")
	if err := verifyVersion(ctx.Logger, ctx.Cluster.Versions.Kubernetes, &nodes, ctx.Verbose, ctx.ForceUpgrade); err != nil {
		return errors.Wrap(err, "unable to verify components version")
	}

	if err := verifyVersionSkew(ctx, &nodes, ctx.Verbose); err != nil {
		if ctx.ForceUpgrade {
			ctx.Logger.Warningf("version skew check failed: %v", err)
		} else {
			return errors.Wrap(err, "version skew check failed")
		}
	}

	return nil
}

// checkPrerequisites checks are Docker, Kubelet, and Kubeadm installed on every machine in the cluster
func checkPrerequisites(ctx *kubeonecontext.Context) error {
	return ctx.RunTaskOnAllNodes(func(ctx *kubeonecontext.Context, _ *kubeoneapi.HostConfig, _ ssh.Connection) error {
		ctx.Logger.Infoln("Checking are all prerequisites installed…")
		_, _, err := ctx.Runner.Run(checkPrerequisitesCommand, runner.TemplateVariables{})
		return err
	}, true)
}

const checkPrerequisitesCommand = `
# Check is Docker installed
if ! type docker &>/dev/null; then exit 1; fi
# Check is Kubelet installed
if ! type kubelet &>/dev/null; then exit 1; fi
# Check is Kubeadm installed
if ! type kubeadm &>/dev/null; then exit 1; fi
# Check do Kubernetes directories and files exist
if [[ ! -d "/etc/kubernetes/manifests" ]]; then exit 1; fi
if [[ ! -d "/etc/kubernetes/pki" ]]; then exit 1; fi
if [[ ! -f "/etc/kubernetes/kubelet.conf" ]]; then exit 1; fi
# Check are kubelet running
if ! sudo systemctl is-active --quiet kubelet &>/dev/null; then exit 1; fi
`

// verifyControlPlaneRunning ensures all control plane nodes are running
func verifyNodesRunning(nodes *corev1.NodeList, verbose bool) error {
	for _, n := range nodes.Items {
		found := false
		for _, c := range n.Status.Conditions {
			if c.Type == corev1.NodeReady {
				if verbose {
					fmt.Printf("[%s] %s (%v)\n", n.ObjectMeta.Name, c.Type, c.Status)
				}
				if c.Status == corev1.ConditionTrue {
					found = true
				}
			}
		}
		if !found {
			return errors.Errorf("node %s is not running", n.ObjectMeta.Name)
		}
	}
	return nil
}

// verifyLabels ensures all control plane nodes don't have the lock label or upgrade is run with the force flag
func verifyLabels(nodes *corev1.NodeList, verbose bool) error {
	for _, n := range nodes.Items {
		_, ok := n.ObjectMeta.Labels[labelUpgradeLock]
		if ok {
			return errors.Errorf("label %s is present on node %s", labelUpgradeLock, n.ObjectMeta.Name)
		}
		if verbose {
			fmt.Printf("[%s] Label %s isn't present\n", n.ObjectMeta.Name, labelUpgradeLock)
		}
	}
	return nil
}

// verifyEndpoints verifies are IP addresses defined in the KubeOne manifest same as IP addresses of nodes
func verifyEndpoints(nodes *corev1.NodeList, hosts []kubeoneapi.HostConfig, verbose bool) error {
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeInternalIP, corev1.NodeExternalIP:
				if verbose {
					fmt.Printf("[%s] %s Endpoint: %s\n", node.ObjectMeta.Name, addr.Type, addr.Address)
				}
			default:
				// we don't care about other types of NodeAddress
				continue
			}

			found := false
			for _, host := range hosts {
				switch addr.Type {
				case corev1.NodeExternalIP:
					if addr.Address == host.PublicAddress {
						found = true
					}
				case corev1.NodeInternalIP:
					if addr.Address == host.PrivateAddress {
						found = true
					}
				}
			}

			if !found {
				return errors.New("cannot match node by ip address")
			}
		}
	}

	return nil
}

// verifyVersion verifies is it possible to upgrade to the requested version
func verifyVersion(logger logrus.FieldLogger, version string, nodes *corev1.NodeList, verbose, force bool) error {
	reqVer, err := semver.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "provided version is invalid")
	}

	kubelet, err := semver.NewVersion(nodes.Items[0].Status.NodeInfo.KubeletVersion)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("Kubelet version on the control plane node: %s", kubelet.String())
		fmt.Printf("Requested version: %s", reqVer.String())
	}

	if reqVer.Compare(kubelet) < 0 {
		return errors.New("unable to upgrade to lower version")
	}
	if reqVer.Compare(kubelet) == 0 {
		if force {
			logger.Warningf("upgrading to the same kubernetes version")
			return nil
		}
		return errors.New("unable to upgrade to the same version")
	}

	return nil
}

// verifyVersionSkew ensures the requested version matches the version skew policy
func verifyVersionSkew(ctx *kubeonecontext.Context, nodes *corev1.NodeList, verbose bool) error {
	reqVer, err := semver.NewVersion(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "provided version is invalid")
	}

	// Check API server version
	var apiserverVersion *semver.Version

	apiserverPods := &corev1.PodList{}
	apiserverListOpts := &dynclient.ListOptions{Namespace: metav1.NamespaceSystem}

	err = apiserverListOpts.SetLabelSelector("component=kube-apiserver")
	if err != nil {
		return errors.Wrap(err, "failed to set labels selector for kube-apiserver")
	}

	err = ctx.DynamicClient.List(context.Background(), apiserverListOpts, apiserverPods)
	if err != nil {
		return errors.Wrap(err, "unable to list apiserver pods")
	}

	// This ensures all API server pods are running the same apiserver version
	for _, p := range apiserverPods.Items {
		ver, apiserverErr := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if apiserverErr != nil {
			return errors.Wrap(apiserverErr, "unable to parse apiserver version")
		}
		if verbose {
			fmt.Printf("Pod %s is running apiserver version %s\n", p.ObjectMeta.Name, ver.String())
		}
		if apiserverVersion == nil {
			apiserverVersion = ver
		}
		if apiserverVersion.Compare(ver) != 0 {
			return errors.New("all apiserver pods must be running same version before upgrade")
		}
	}

	err = checkVersionSkew(reqVer, apiserverVersion, 1)
	if err != nil {
		return errors.Wrap(err, "apiserver version check failed")
	}

	// Check Kubelet version
	for _, n := range nodes.Items {
		kubeletVer, kubeletErr := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if kubeletErr != nil {
			return errors.Wrap(err, "unable to parse kubelet version")
		}
		if verbose {
			fmt.Printf("Node %s is running kubelet version %s\n", n.ObjectMeta.Name, kubeletVer.String())
		}
		// Check is requested version different than current and ensure version skew policy
		err = checkVersionSkew(reqVer, kubeletVer, 2)
		if err != nil {
			return errors.Wrap(err, "kubelet version check failed")
		}
		if kubeletVer.Minor() > apiserverVersion.Minor() {
			return errors.New("kubelet cannot be newer than apiserver")
		}
	}

	return nil
}

func parseContainerImageVersion(image string) (*semver.Version, error) {
	ver := strings.Split(image, ":")
	if len(ver) != 2 {
		return nil, errors.Errorf("invalid container image format: %s", image)
	}
	return semver.NewVersion(ver[1])
}

func checkVersionSkew(reqVer, currVer *semver.Version, diff int64) error {
	// Check is requested version different than current and ensure version skew policy
	if currVer.Equal(reqVer) {
		return errors.New("requested version is same as current")
	}
	// Check are we upgrading to newer minor or patch release
	if reqVer.Minor()-currVer.Minor() < 0 ||
		(reqVer.Minor() == currVer.Minor() && reqVer.Patch() < currVer.Patch()) {
		return errors.New("requested version can't be lower than current")
	}
	// Ensure the version skew policy
	// https://kubernetes.io/docs/setup/version-skew-policy/#supported-version-skew
	if reqVer.Minor()-currVer.Minor() > diff {
		return errors.Errorf("version skew check failed: component can be only %d minor version older than requested version", diff)
	}
	return nil
}
