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
		"/lib/systemd/system/kubelet.service",
		"/lib/systemd/system/kubelet.service.d/*",
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
	if err := wait.PollUntilContextTimeout(s.Context, 5*time.Second, 5*time.Minute, true, readyNodes); err != nil {
		return fail.KubeClient(err, "waiting for all Nodes to be Ready")
	}

	return s.RunTaskOnAllNodes(func(ctx *state.State, _ *kubeoneapi.HostConfig, conn executor.Interface) error {
		for _, pathList := range systemFiles {
			for _, path := range pathList {
				nodeFS := executorfs.New(conn)
				matches, err := fs.Glob(nodeFS, path)
				if err != nil {
					return fail.SSH(err, "expanding glob pattern")
				}

				for _, match := range matches {
					match = strings.TrimSpace(match)
					file, err := nodeFS.Open(match)
					if err != nil {
						return err
					}

					fw, ok := file.(executor.ExtendedFile)
					if !ok {
						return fail.RuntimeError{
							Op:  "checking if file satisfy sshiofs.ExtendedFile interface",
							Err: errors.New("can not change file permissions"),
						}
					}

					ctx.Logger.Debugf("chmod 0600 %q", match)
					if err = fw.Chmod(0o600); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}, state.RunParallel)
}
