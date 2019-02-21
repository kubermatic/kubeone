package upgrade

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// runPreflightChecks runs all preflight checks
func runPreflightChecks(ctx *util.Context) error {
	// Verify clientset is initialized so we can reach API server
	if ctx.Clientset == nil {
		return errors.New("kubernetes clientset not initialized")
	}

	// Check are Docker, Kubelet and Kubeadm installed
	if err := checkPrerequisites(ctx); err != nil {
		return errors.Wrap(err, "unable to check are prerequisites installed")
	}

	// Get list of nodes and verify number of nodes
	nodes, err := ctx.Clientset.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelControlPlaneNode, ""),
	})
	if err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}
	if len(nodes.Items) != len(ctx.Cluster.Hosts) {
		return errors.Errorf("expected %d cluster nodes but got %d", len(ctx.Cluster.Hosts), len(nodes.Items))
	}

	// Run preflight checks on nodes
	ctx.Logger.Infoln("Running preflight checks…")

	ctx.Logger.Infoln("Verifying are all nodes running…")
	if err := verifyNodesRunning(nodes, ctx.Verbose); err != nil {
		return errors.Wrap(err, "unable to verify are nodes running")
	}

	ctx.Logger.Infoln("Verifying are correct labels set on nodes…")
	if err := verifyLabels(nodes, ctx.Verbose); err != nil {
		if ctx.ForceUpgrade {
			ctx.Logger.Warningf("unable to verify node labels: %v", err)
		} else {
			return errors.Wrap(err, "unable to verify node labels")
		}
	}

	ctx.Logger.Infoln("Verifying do all node IP addresses match with our state…")
	if err := verifyEndpoints(nodes, ctx.Cluster.Hosts, ctx.Verbose); err != nil {
		return errors.Wrap(err, "unable to verify node endpoints")
	}

	ctx.Logger.Infoln("Verifying is it possible to upgrade to the desired version…")
	if err := verifyVersion(ctx.Cluster.Versions.Kubernetes, nodes, ctx.Verbose); err != nil {
		return errors.Wrap(err, "unable to verify components version")
	}
	if err := verifyVersionSkew(ctx, nodes, ctx.Verbose); err != nil {
		if ctx.ForceUpgrade {
			ctx.Logger.Warningf("version skew check failed: %v", err)
		} else {
			return errors.Wrap(err, "version skew check failed")
		}
	}

	return nil
}

// checkPrerequisites checks are Docker, Kubelet, and Kubeadm installed on every machine in the cluster
func checkPrerequisites(ctx *util.Context) error {
	return ctx.RunTaskOnAllNodes(func(ctx *util.Context, _ *config.HostConfig, _ ssh.Connection) error {
		ctx.Logger.Infoln("Checking are all prerequisites installed…")
		_, _, err := ctx.Runner.Run(checkPrerequisitesCommand, util.TemplateVariables{})
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
func verifyEndpoints(nodes *corev1.NodeList, hosts []*config.HostConfig, verbose bool) error {
	for _, n := range nodes.Items {
		found := false
		for _, addr := range n.Status.Addresses {
			if verbose && addr.Type == corev1.NodeExternalIP {
				fmt.Printf("[%s] Endpoint: %s\n", n.ObjectMeta.Name, addr.Address)
			}
			for _, host := range hosts {
				if addr.Type == corev1.NodeExternalIP && host.PublicAddress == addr.Address {
					found = true
				}
			}
		}
		if !found {
			return errors.New("cannot match node by ip address")
		}
	}
	return nil
}

// verifyVersion verifies is it possible to upgrade to the requested version
func verifyVersion(version string, nodes *corev1.NodeList, verbose bool) error {
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

	if reqVer.Compare(kubelet) <= 0 {
		return errors.New("unable to upgrade to same or lower version")
	}

	return nil
}

// verifyVersionSkew ensures the requested version matches the version skew policy
func verifyVersionSkew(ctx *util.Context, nodes *corev1.NodeList, verbose bool) error {
	reqVer, err := semver.NewVersion(ctx.Cluster.Versions.Kubernetes)
	if err != nil {
		return errors.Wrap(err, "provided version is invalid")
	}

	// Check API server version
	var apiserverVersion *semver.Version
	apiserverPods, err := ctx.Clientset.CoreV1().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{
		LabelSelector: "component=kube-apiserver",
	})
	if err != nil {
		return errors.Wrap(err, "unable to list apiserver pods")
	}
	// This ensures all API server pods are running the same apiserver version
	for _, p := range apiserverPods.Items {
		ver, apiserverErr := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if apiserverErr != nil {
			return errors.Wrap(err, "unable to parse apiserver version")
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
