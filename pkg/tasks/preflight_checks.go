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
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// runPreflightChecks runs all preflight checks
func runPreflightChecks(s *state.State) error {
	if s.DynamicClient == nil {
		return fail.NoKubeClient()
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{labelControlPlaneNode: ""}),
	}

	err := s.DynamicClient.List(context.Background(), &nodes, &nodeListOpts)
	if err != nil {
		return fail.KubeClient(err, "getting %T", nodes)
	}

	// Run preflight checks on nodes
	s.Logger.Infoln("Running preflight checks...")
	if err := preflightstatus.Run(s, nodes); err != nil {
		return err
	}

	s.Logger.Infoln("Verifying is it possible to upgrade to the desired version...")
	if err := verifyVersion(s.Logger, s.Cluster.Versions.Kubernetes, &nodes, s.Verbose, s.ForceUpgrade); err != nil {
		return err
	}

	if canPass, err := verifyVersionSkew(s, &nodes, s.Verbose); err != nil {
		if s.ForceUpgrade && canPass {
			s.Logger.Warningf("version skew check failed: %v", err)
		} else {
			return err
		}
	}

	return nil
}

// verifyVersion verifies is it possible to upgrade to the requested version
func verifyVersion(logger logrus.FieldLogger, version string, nodes *corev1.NodeList, verbose, force bool) error {
	reqVer, err := semver.NewVersion(version)
	if err != nil {
		return fail.ConfigValidation(err)
	}

	kubelet, err := semver.NewVersion(nodes.Items[0].Status.NodeInfo.KubeletVersion)
	if err != nil {
		return fail.Runtime(err, "parsing node kubelet version")
	}

	if verbose {
		fmt.Printf("Kubelet version on the control plane node: %s\n", kubelet.String())
		fmt.Printf("Requested version: %s\n", reqVer.String())
	}

	if reqVer.Compare(kubelet) < 0 {
		return fail.Runtime(fmt.Errorf("unable to upgrade to lower version"), "checking version skew")
	}

	if reqVer.Compare(kubelet) == 0 {
		if force {
			logger.Warningf("upgrading to the same kubernetes version")

			return nil
		}

		return fail.Runtime(fmt.Errorf("unable to upgrade to the same version"), "checking version skew")
	}

	return nil
}

// verifyVersionSkew ensures the requested version matches the version skew policy
func verifyVersionSkew(s *state.State, nodes *corev1.NodeList, verbose bool) (bool, error) {
	reqVer, err := semver.NewVersion(s.Cluster.Versions.Kubernetes)
	if err != nil {
		return false, fail.Runtime(err, "checking skew version")
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
		return false, fail.KubeClient(err, "getting kube-apiserver pods")
	}

	// This ensures all API server pods are running the same apiserver version
	for _, p := range apiserverPods.Items {
		ver, apiserverErr := parseContainerImageVersion(p.Spec.Containers[0])
		if apiserverErr != nil {
			return false, fail.Runtime(apiserverErr, "parsing kube-apiserver container image")
		}
		if verbose {
			fmt.Printf("Pod %s is running apiserver version %s\n", p.ObjectMeta.Name, ver.String())
		}
		if apiserverVersion == nil {
			apiserverVersion = ver
		}
		if apiserverVersion.Compare(ver) != 0 {
			return true, fail.RuntimeError{
				Op:  "checking kube-apiserver pods versions",
				Err: errors.New("must be running same version before upgrade"),
			}
		}
	}

	err = checkVersionSkew(reqVer, apiserverVersion, 1)
	if err != nil {
		return true, err
	}

	// Check Kubelet version
	for _, n := range nodes.Items {
		kubeletVer, kubeletErr := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if kubeletErr != nil {
			return false, fail.Runtime(kubeletErr, "parsing %q node kubelet version", n.Name)
		}
		if verbose {
			fmt.Printf("Node %s is running kubelet version %s\n", n.ObjectMeta.Name, kubeletVer.String())
		}
		// Check is requested version different than current and ensure version skew policy
		err = checkVersionSkew(reqVer, kubeletVer, 2)
		if err != nil {
			return true, err
		}
		if kubeletVer.Minor() > apiserverVersion.Minor() {
			return true, fail.RuntimeError{
				Op:  fmt.Sprintf("comparing kubelet on %q Node and kube-apiserver versions", n.Name),
				Err: errors.New("kubelet cannot be newer than apiserver"),
			}
		}
	}

	return false, nil
}

func parseContainerImageVersion(container corev1.Container) (*semver.Version, error) {
	ver := strings.Split(container.Image, ":")
	if len(ver) != 2 {
		return nil, fmt.Errorf("invalid container image format: %s", container.Image)
	}

	return semver.NewVersion(ver[1])
}

func checkVersionSkew(reqVer, currVer *semver.Version, diff uint64) error {
	// Check is requested version different than current and ensure version skew policy
	if currVer.Equal(reqVer) {
		return fail.Runtime(fmt.Errorf("requested version is same as current"), "checking version skew policy")
	}

	// Check are we upgrading to newer minor or patch release
	if int64(reqVer.Minor())-int64(currVer.Minor()) < 0 ||
		(reqVer.Minor() == currVer.Minor() && reqVer.Patch() < currVer.Patch()) {
		return fail.Runtime(fmt.Errorf("requested version can't be lower than current"), "checking version skew policy")
	}

	// Ensure the version skew policy
	// https://kubernetes.io/docs/setup/version-skew-policy/#supported-version-skew
	if reqVer.Minor()-currVer.Minor() > diff {
		return fail.RuntimeError{
			Op:  "checking version skew policy",
			Err: errors.Errorf("component can be only %d minor version older than requested version", diff),
		}
	}

	return nil
}
