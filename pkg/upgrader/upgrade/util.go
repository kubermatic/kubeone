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

	"github.com/pkg/errors"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/scripts"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func determineHostname(s *state.State) error {
	s.Logger.Infoln("Determine hostname…")
	return s.RunTaskOnAllNodes(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		if node.Hostname != "" {
			return nil
		}

		hostnameCmd := scripts.GetHostname()

		// on azure the name of the Node should == name of the VM
		if s.Cluster.CloudProvider.Name == kubeoneapi.CloudProviderNameAzure {
			hostnameCmd = `hostname`
		}
		stdout, _, err := s.Runner.Run(hostnameCmd, nil)
		if err != nil {
			return err
		}

		node.SetHostname(stdout)
		return nil
	}, true)
}

func determineOS(s *state.State) error {
	s.Logger.Infoln("Determine operating system…")
	return s.RunTaskOnAllNodes(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		osID, _, err := s.Runner.Run("source /etc/os-release && echo -n $ID", nil)
		if err != nil {
			return err
		}

		node.SetOperatingSystem(osID)
		return nil
	}, true)
}

func labelNode(client dynclient.Client, host *kubeoneapi.HostConfig) error {
	retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: host.Hostname},
		}

		_, err := controllerutil.CreateOrUpdate(context.Background(), client, &node, func() error {
			if node.ObjectMeta.CreationTimestamp.IsZero() {
				return errors.New("node not found")
			}
			node.Labels[labelUpgradeLock] = ""
			return nil
		})
		return err
	})

	return errors.Wrapf(retErr, "failed to label node %q with label %q", host.Hostname, labelUpgradeLock)
}

func unlabelNode(client dynclient.Client, host *kubeoneapi.HostConfig) error {
	retErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: host.Hostname},
		}

		_, err := controllerutil.CreateOrUpdate(context.Background(), client, &node, func() error {
			if node.ObjectMeta.CreationTimestamp.IsZero() {
				return errors.New("node not found")
			}
			delete(node.ObjectMeta.Labels, labelUpgradeLock)
			return nil
		})
		return err
	})

	return errors.Wrapf(retErr, "failed to remove label %s from node %s", labelUpgradeLock, host.Hostname)
}
