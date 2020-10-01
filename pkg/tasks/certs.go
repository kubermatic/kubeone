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
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

func deployPKIToFollowers(s *state.State) error {
	s.Logger.Infoln("Deploying PKI...")
	return s.RunTaskOnFollowers(deployCAOnNode, state.RunParallel)
}

func deployCAOnNode(s *state.State, node *kubeoneapi.HostConfig, conn ssh.Connection) error {
	s.Logger.Infoln("Uploading PKI files...")
	return s.Configuration.UploadTo(conn, s.WorkDir)
}
