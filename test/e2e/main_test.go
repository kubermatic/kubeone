// +build e2e

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

package e2e

import (
	"context"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"k8c.io/kubeone/test/e2e/provisioner"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testCredentialsFile string
	testRunIdentifier   string
	testInitialVersion  string
	testTargetVersion   string
	testProvider        string
	testOSControlPlane  string
	testOSWorkers       string
)

func init() {
	flag.StringVar(&testCredentialsFile, "credentials", "", "path to the credentials file")
	flag.StringVar(&testRunIdentifier, "identifier", "", "The unique identifier for this test run")
	flag.StringVar(&testProvider, "provider", "", "Provider to run tests on")
	flag.StringVar(&testInitialVersion, "initial-version", "", "Cluster version to provision for tests")
	flag.StringVar(&testTargetVersion, "target-version", "", "Cluster version to provision for tests")
	flag.StringVar(&testOSControlPlane, "os-control-plane", "", "Operating system to use for control plane nodes")
	flag.StringVar(&testOSWorkers, "os-workers", "", "Operating system to use for worker nodes")
	flag.Parse()
}

// This is a workaround for a change in the testing framework
// affecting Go 1.13 and newer.
// More details: https://github.com/golang/go/issues/31859#issuecomment-489889428
var _ = func() bool {
	testing.Init()
	return true
}()

func setupTearDown(p provisioner.Provisioner, k *Kubeone) func(t *testing.T) {
	return func(t *testing.T) {
		t.Log("cleanup ....")

		errKubeone := k.Reset()
		errProvisioner := p.Cleanup()

		if errKubeone != nil {
			t.Errorf("%v", errKubeone)
		}

		if errProvisioner != nil {
			t.Errorf("%v", errProvisioner)
		}
	}
}

func waitForNodesReady(t *testing.T, client dynclient.Client, expectedNumberOfNodes int) error {
	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}

		if err := client.List(context.Background(), &nodes); err != nil {
			t.Logf("error: %v", err)
			return false, nil
		}

		if len(nodes.Items) != expectedNumberOfNodes {
			return false, nil
		}

		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}
		return true, nil
	})
}

func verifyVersion(client dynclient.Client, namespace string, targetVersion string) error {
	reqVer, err := semver.NewVersion(targetVersion)
	if err != nil {
		return errors.Wrap(err, "desired version is invalid")
	}

	nodes := corev1.NodeList{}
	nodeListOpts := dynclient.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			labelControlPlaneNode: "",
		}),
	}

	if err = client.List(context.Background(), &nodes, &nodeListOpts); err != nil {
		return errors.Wrap(err, "failed to list nodes")
	}

	// Kubelet version check
	for _, n := range nodes.Items {
		kubeletVer, errSemver := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if errSemver != nil {
			return errSemver
		}
		if reqVer.Compare(kubeletVer) != 0 {
			return errors.Errorf("kubelet version mismatch: expected %v, got %v", reqVer.String(), kubeletVer.String())
		}
	}

	apiserverPods := corev1.PodList{}
	podsListOpts := dynclient.ListOptions{
		Namespace: namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"component": "kube-apiserver",
		}),
	}

	if err = client.List(context.Background(), &apiserverPods, &podsListOpts); err != nil {
		return errors.Wrap(err, "unable to list apiserver pods")
	}

	for _, p := range apiserverPods.Items {
		apiserverVer, err := parseContainerImageVersion(p.Spec.Containers[0].Image)
		if err != nil {
			return errors.Wrap(err, "unable to parse apiserver version")
		}
		if reqVer.Compare(apiserverVer) != 0 {
			return errors.Errorf("apiserver version mismatch: expected %v, got %v", reqVer.String(), apiserverVer.String())
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
