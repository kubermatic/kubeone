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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/test/e2e/provisioner"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// testRunIdentifier aka. the build number, a unique identifier for the test run.
var (
	testRunIdentifier  string
	testInitialVersion string
	testTargetVersion  string
	testProvider       string
	testOSControlPlane string
	testOSWorkers      string
)

func init() {
	flag.StringVar(&testRunIdentifier, "identifier", "", "The unique identifier for this test run")
	flag.StringVar(&testProvider, "provider", "", "Provider to run tests on")
	flag.StringVar(&testInitialVersion, "initial-version", "", "Cluster version to provision for tests")
	flag.StringVar(&testTargetVersion, "target-version", "", "Cluster version to provision for tests")
	flag.StringVar(&testOSControlPlane, "os-control-plane", "", "Operating system to use for control plane nodes")
	flag.StringVar(&testOSWorkers, "os-workers", "", "Operating system to use for worker nodes")
	flag.Parse()
}

func setupTearDown(p provisioner.Provisioner, k Kubeone) func(t *testing.T) {
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

func waitForNodesReady(client dynclient.Client, expectedNumberOfNodes int) error {
	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}
		nodeListOpts := dynclient.ListOptions{}

		err := client.List(context.Background(), &nodeListOpts, &nodes)
		if err != nil {
			return false, errors.Wrap(err, "unable to list nodes")
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
	nodeListOpts := dynclient.ListOptions{}
	_ = nodeListOpts.SetLabelSelector(fmt.Sprintf("%s=%s", labelControlPlaneNode, ""))
	err = client.List(context.Background(), &nodeListOpts, &nodes)
	if err != nil {
		return errors.Wrap(err, "failed to list nodes")
	}

	// Kubelet version check
	for _, n := range nodes.Items {
		kubeletVer, err := semver.NewVersion(n.Status.NodeInfo.KubeletVersion)
		if err != nil {
			return err
		}
		if reqVer.Compare(kubeletVer) != 0 {
			return errors.Errorf("kubelet version mismatch: expected %v, got %v", reqVer.String(), kubeletVer.String())
		}
	}

	apiserverPods := corev1.PodList{}
	podsListOpts := dynclient.ListOptions{Namespace: namespace}
	_ = podsListOpts.SetLabelSelector("component=kube-apiserver")
	err = client.List(context.Background(), &podsListOpts, &apiserverPods)
	if err != nil {
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
