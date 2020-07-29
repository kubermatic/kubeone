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
	"io/ioutil"
	"os"
	"time"

	osrelease "github.com/dominodatalab/os-release"
	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	labelUpgradeLock      = "kubeone.io/upgrade-in-progress"
	labelControlPlaneNode = "node-role.kubernetes.io/master"
	// timeoutNodeUpgrade is time for how long kubeone will wait after finishing the upgrade
	// process on the node
	timeoutNodeUpgrade = 30 * time.Second
)

func determineHostname(s *state.State) error {
	s.Logger.Infoln("Determine hostname…")
	return s.RunTaskOnAllNodes(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		if node.Hostname != "" {
			return nil
		}

		hostnameCmd := scripts.Hostname()

		// on azure the name of the Node should == name of the VM
		if s.Cluster.CloudProvider.Azure != nil {
			hostnameCmd = `hostname`
		}
		stdout, _, err := s.Runner.Run(hostnameCmd, nil)
		if err != nil {
			return err
		}

		node.SetHostname(stdout)
		return nil
	}, state.RunParallel)
}

func determineOS(s *state.State) error {
	s.Logger.Infoln("Determine operating system…")
	return s.RunTaskOnAllNodes(func(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
		f, err := conn.File("/etc/os-release", os.O_RDONLY)
		if err != nil {
			return err
		}
		defer f.Close()

		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		osrData := osrelease.Parse(string(buf))
		node.SetOperatingSystem(kubeoneapi.OperatingSystemName(osrData.ID))
		return nil
	}, state.RunParallel)
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

type runOnOSFn func(*state.State) error

func runOnOS(s *state.State, osname kubeoneapi.OperatingSystemName, fnMap map[kubeoneapi.OperatingSystemName]runOnOSFn) error {
	fn, ok := fnMap[osname]
	if !ok {
		return errors.Errorf("%q is not a supported operating system", osname)
	}

	return errors.WithStack(fn(s))
}
