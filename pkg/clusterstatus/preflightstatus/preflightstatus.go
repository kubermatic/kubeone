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

package preflightstatus

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	LabelControlPlaneNode = "node-role.kubernetes.io/control-plane"
	LabelUpgradeLock      = "kubeone.io/upgrade-in-progress"
)

// Run ensures that all prerequisites are satisfied
// TODO(xmudrii): Implement mechanism for skipping checks.
func Run(s *state.State, nodes corev1.NodeList) error {
	var errs []error

	// Verify that list of nodes match with the provided manifest
	s.Logger.Infoln("Verifying that nodes in the cluster match nodes defined in the manifest...")
	if err := verifyMatchNodes(s.Cluster.ControlPlane.Hosts, nodes, s.Logger, s.Verbose); err != nil {
		s.Logger.Errorln("Unable to match all control plane nodes in the cluster and all nodes defined in the manifest.")
		errs = append(errs, err...)
	}

	// Verify that all nodes in the cluster are ready
	s.Logger.Infoln("Verifying that all nodes in the cluster are ready...")
	if err := verifyNodesReady(nodes, s.Logger, s.Verbose); err != nil {
		s.Logger.Errorln("Unable to match all control plane nodes in the cluster and all nodes defined in the manifest.")
		errs = append(errs, err...)
	}

	// Verify that upgrade is not in progress
	s.Logger.Infoln("Verifying that there is no upgrade in progress...")
	if err := verifyNoUpgradeLabels(nodes, s.Logger, s.Verbose); err != nil {
		s.Logger.Errorf("Unable to verify that there is no upgrade in progress.")
		errs = append(errs, err...)
	}

	return utilerrors.NewAggregate(errs)
}

// verifyMatchNodes ensures match between nodes in the cluster and machines defined in the manifest
func verifyMatchNodes(hosts []kubeoneapi.HostConfig, nodes corev1.NodeList, logger logrus.FieldLogger, verbose bool) []error {
	if len(nodes.Items) != len(hosts) {
		logger.Errorf("Mismatch between nodes in the cluster (%d) and nodes defined in the manifest (%d).", len(nodes.Items), len(hosts))

		return []error{fail.RuntimeError{
			Op:  "checking match between Nodes and Hosts",
			Err: errors.Errorf("expected %d cluster nodes but got %d", len(nodes.Items), len(hosts)),
		}}
	}

	nodesFound := map[string]bool{}

	for _, node := range nodes.Items {
		nodesFound[node.Name] = false

		for _, host := range hosts {
			for _, addr := range node.Status.Addresses {
				switch addr.Type {
				case corev1.NodeInternalIP, corev1.NodeExternalIP:
					switch addr.Address {
					case host.PrivateAddress, host.PublicAddress:
						nodesFound[node.Name] = true
						if verbose {
							logger.Infof("Found endpoint %q (type %s) for the node %q.", addr.Address, addr.Type, node.ObjectMeta.Name)
						}
					}
				case corev1.NodeExternalDNS, corev1.NodeHostName, corev1.NodeInternalDNS:
				}
			}
		}
	}

	var errs []error

	for nodeName, found := range nodesFound {
		if !found {
			errs = append(errs, fail.RuntimeError{
				Op:  "matching found Nodes",
				Err: errors.Errorf("unable to match node %q to machines defined in the manifest", nodeName),
			})
		}
	}

	return errs
}

// verifyNodesReady ensures all nodes in the cluster are ready
func verifyNodesReady(nodes corev1.NodeList, logger logrus.FieldLogger, verbose bool) []error {
	var errs []error

	for _, n := range nodes.Items {
		found := false

		for _, c := range n.Status.Conditions {
			if c.Type == corev1.NodeReady {
				if verbose {
					logger.Infof("Node %q reporting %s=%s.", n.ObjectMeta.Name, c.Type, c.Status)
				}
				if c.Status == corev1.ConditionTrue {
					found = true
				}
			}
		}

		if !found {
			errs = append(errs, fail.RuntimeError{
				Op:  fmt.Sprintf("checking %q node readiness", n.ObjectMeta.Name),
				Err: errors.New("not ready"),
			})
		}
	}

	return errs
}

// verifyNoUpgradeLabels check labels on nodes to ensure there is no upgrade in progress
func verifyNoUpgradeLabels(nodes corev1.NodeList, logger logrus.FieldLogger, verbose bool) []error {
	var errs []error

	for _, n := range nodes.Items {
		_, ok := n.ObjectMeta.Labels[LabelUpgradeLock]
		if ok {
			logger.Errorf("Upgrade is in progress on the node %q (label %q is present).", n.ObjectMeta.Name, LabelUpgradeLock)
			errs = append(errs, fail.RuntimeError{
				Op:  fmt.Sprintf("checking presence of %q label on node %q", LabelUpgradeLock, n.ObjectMeta.Name),
				Err: errors.New("label is present"),
			})
		}

		if verbose && !ok {
			logger.Infof("Label %q isn't present on the node %q.", LabelUpgradeLock, n.ObjectMeta.Name)
		}
	}

	return errs
}
