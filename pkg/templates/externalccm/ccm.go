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

package externalccm

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	uninitializedTaint = "node.cloudprovider.kubernetes.io/uninitialized"
)

// Ensure external CCM deployen if Provider.External
func Ensure(s *state.State) error {
	if !s.Cluster.CloudProvider.External {
		return nil
	}

	return errors.Wrap(waitForInitializedNodes(s), "failed waiting for nodes to be initialized by CCM")
}

func waitForInitializedNodes(s *state.State) error {
	ctx := context.Background()

	s.Logger.Info("Waiting for nodes to initialize by CCM...")

	return wait.Poll(5*time.Second, 10*time.Minute, func() (bool, error) {
		nodes := corev1.NodeList{}

		if err := s.DynamicClient.List(ctx, &nodes); err != nil {
			return false, err
		}

		for _, node := range nodes.Items {
			for _, taint := range node.Spec.Taints {
				if taint.Key == uninitializedTaint && taint.Value == "true" {
					return false, nil
				}
			}
		}

		return true, nil
	})
}
