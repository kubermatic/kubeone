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
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/clusterstatus/preflightstatus"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// runPreflightChecks runs all preflight checks
func runPreflightChecks(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{labelControlPlaneNode: ""}),
	}

	err := s.DynamicClient.List(context.Background(), &nodes, &nodeListOpts)
	if err != nil {
		return errors.Wrap(err, "unable to list nodes")
	}

	// Run preflight checks on nodes
	s.Logger.Infoln("Running preflight checks...")
	if err := preflightstatus.Run(s, nodes); err != nil {
		return errors.Wrap(err, "unable to verify prerequisites")
	}

	s.Logger.Infoln("Verifying is it possible to upgrade to the desired version...")
	if err := verifyVersion(s.Logger, s.Cluster.Versions.Kubernetes, &nodes, s.Verbose, s.ForceUpgrade); err != nil {
		return errors.Wrap(err, "unable to verify components version")
	}

	if canPass, err := verifyVersionSkew(s, &nodes, s.Verbose); err != nil {
		if s.ForceUpgrade && canPass {
			s.Logger.Warningf("version skew check failed: %v", err)
		} else {
			return errors.Wrap(err, "version skew check failed")
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
func verifyVersionSkew(s *state.State, nodes *corev1.NodeList, verbose bool) (bool, error) {
	reqVer, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return false, errors.Wrap(err, "provided version is invalid")
	}

	// Check API server version
	var apiserverVersion *semver.Version

	apiserverPods := &corev1.PodList{}
	apiserverListOpts := &dynclient.ListOptions{
		Namespace:     metav1.NamespaceSystem,
		LabelSelector: labels.SelectorFromSet(map[string]string{"component": "kube-apiserver"}),
	}

	err = s.DynamicClient.List(context.Background(), apiserverPods, apiserverListOpts)
	if err != nil {
		return false, errors.Wrap(err, "unable to list apiserver pods")
	}

	// This ensures all API server pods are running the same apiserver version
	for _, p := range apiserverPods.Items {
		ver, apiserverErr := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if apiserverErr != nil {
			return false, errors.Wrap(apiserverErr, "unable to parse apiserver version")
		}
		if verbose {
			fmt.Printf("Pod %s is running apiserver version %s\n", p.ObjectMeta.Name, ver.String())
		}
		if apiserverVersion == nil {
			apiserverVersion = ver
		}
		if apiserverVersion.Compare(ver) != 0 {
			return true, errors.New("all apiserver pods must be running same version before upgrade")
		}
	}

	err = checkVersionSkew(reqVer, apiserverVersion, 1)
	if err != nil {
		return true, errors.Wrap(err, "apiserver version check failed")
	}

	// Check Kubelet version
	for _, n := range nodes.Items {
		kubeletVer, kubeletErr := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if kubeletErr != nil {
			return false, errors.Wrap(err, "unable to parse kubelet version")
		}
		if verbose {
			fmt.Printf("Node %s is running kubelet version %s\n", n.ObjectMeta.Name, kubeletVer.String())
		}
		// Check is requested version different than current and ensure version skew policy
		err = checkVersionSkew(reqVer, kubeletVer, 2)
		if err != nil {
			return true, errors.Wrap(err, "kubelet version check failed")
		}
		if kubeletVer.Minor() > apiserverVersion.Minor() {
			return true, errors.New("kubelet cannot be newer than apiserver")
		}
	}

	return false, nil
}

func parseContainerImageVersion(image string) (*semver.Version, error) {
	ver := strings.Split(image, ":")
	if len(ver) != 2 {
		return nil, errors.Errorf("invalid container image format: %s", image)
	}
	return semver.NewVersion(ver[1])
}

func checkVersionSkew(reqVer, currVer *semver.Version, diff uint64) error {
	// Check is requested version different than current and ensure version skew policy
	if currVer.Equal(reqVer) {
		return errors.New("requested version is same as current")
	}

	// Check are we upgrading to newer minor or patch release
	if int64(reqVer.Minor())-int64(currVer.Minor()) < 0 ||
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
