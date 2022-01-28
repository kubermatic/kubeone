/*
Copyright 2022 The KubeOne Authors.

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

package operatingsystemmanager

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const appLabelKey = "app"

// WaitReady waits for operating-system-manager and its webhook to become ready
func WaitReady(s *state.State) error {
	if !s.Cluster.OperatingSystemManagerEnabled() {
		return nil
	}

	s.Logger.Infoln("Waiting for operating-system-manager to come up...")

	if err := waitForWebhook(s.Context, s.DynamicClient); err != nil {
		return errors.Wrap(err, "operating-system-manager-webhook did not come up")
	}

	if err := waitForController(s.Context, s.DynamicClient); err != nil {
		return errors.Wrap(err, "operating-system-manager did not come up")
	}

	if err := waitForCRDs(s); err != nil {
		return errors.Wrap(err, "operating-system-manager CRDs did not come up")
	}

	return nil
}

// waitForCRDs waits for operating-system-manager CRDs to be created and become established
func waitForCRDs(s *state.State) error {
	condFn := clientutil.CRDsReadyCondition(s.Context, s.DynamicClient, CRDNames())
	err := wait.Poll(5*time.Second, 3*time.Minute, condFn)

	return errors.Wrap(err, "failed waiting for CRDs to become ready and established")
}

// waitForController waits for operating-system-manager controller to become running
func waitForController(ctx context.Context, client dynclient.Client) error {
	condFn := clientutil.PodsReadyCondition(ctx, client, dynclient.ListOptions{
		Namespace: resources.OperatingSystemManagerNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			appLabelKey: resources.OperatingSystemManagerName,
		}),
	})

	return wait.Poll(5*time.Second, 3*time.Minute, condFn)
}

// waitForWebhook waits for operating-system-manager-webhook to become running
func waitForWebhook(ctx context.Context, client dynclient.Client) error {
	condFn := clientutil.PodsReadyCondition(ctx, client, dynclient.ListOptions{
		Namespace: resources.OperatingSystemManagerNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			appLabelKey: resources.OperatingSystemManagerWebhookName,
		}),
	})

	return wait.Poll(5*time.Second, 3*time.Minute, condFn)
}

func CRDNames() []string {
	return []string{
		"operatingsystemprofiles.operatingsystemmanager.k8c.io",
		"operatingsystemconfigs.operatingsystemmanager.k8c.io",
	}
}
