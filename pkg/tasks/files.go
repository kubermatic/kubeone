/*
Copyright 2024 The KubeOne Authors.

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
	"io/fs"
	"strings"
	"time"

	"github.com/pkg/errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var systemFiles = map[string][]string{
	"kubelet": {
		"/lib/systemd/system/kubelet.service*",
		"/lib/systemd/system/kubelet.service.d/*",
		"/etc/systemd/system/kubelet.service*",
		"/etc/systemd/system/kubelet.service.d/*",
		"/var/lib/kubelet/config.yaml",
	},
	"cni": {
		"/etc/cni/net.d/*",
	},
}

func fixFilePermissions(s *state.State) error {
	s.Logger.Info("Fixing permissions of the kubernetes system files...")

	readyNodes := func(ctx context.Context) (bool, error) {
		s.Logger.Debug(".")

		nodeList := corev1.NodeList{}
		if err := s.DynamicClient.List(ctx, &nodeList); err != nil {
			return false, fail.KubeClient(err, "getting %T", nodeList)
		}

		for _, node := range nodeList.Items {
			for _, cond := range node.Status.Conditions {
				if cond.Type == corev1.NodeReady {
					if cond.Status != corev1.ConditionTrue {
						// Some Nodes are not yet Ready. Let's continue to poll.
						return false, nil
					}
				}
			}
		}

		return true, nil
	}

	s.Logger.Debug("Waiting for all nodes to be Ready.")

	// In order to satisfy CIS benchmark 1.8 we have to chmod 0600 files in /etc/cni/net.d/, however this directory is
	// created only by the CNI driver and we have to wait until it's available (which happens when the Node becomes``
	// Ready).
	if err := wait.PollUntilContextTimeout(s.Context, 5*time.Second, 5*time.Minute, true, readyNodes); err != nil {
		return fail.KubeClient(err, "waiting for all Nodes to be Ready")
	}

	// CIS benchmark tests mode of files only on control-plane nodes
	return s.RunTaskOnControlPlane(func(ctx *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
		for _, pathList := range systemFiles {
			for _, path := range pathList {
				nodeFS := executorfs.New(conn)
				matches, err := fs.Glob(nodeFS, path)
				if err != nil {
					return fail.SSH(err, "expanding glob pattern")
				}

				for _, match := range matches {
					match = strings.TrimSpace(match)
					if match == path {
						// glob returns the pattern itself in case when there was no match
						continue
					}

					file, err := nodeFS.Open(match)
					if err != nil {
						return err
					}

					fw, ok := file.(executor.ExtendedFile)
					if !ok {
						return fail.RuntimeError{
							Op:  "checking if file satisfy sshiofs.ExtendedFile interface",
							Err: errors.New("file is not executor.ExtendedFile"),
						}
					}

					fi, err := fw.Stat()
					if err != nil {
						return err
					}

					var mode fs.FileMode = 0o600
					if fi.IsDir() {
						mode = 0o700
					}

					ctx.Logger.Debugf("chmod %o %q", mode, match)
					if err = fw.Chmod(mode); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}, state.RunParallel)
}
