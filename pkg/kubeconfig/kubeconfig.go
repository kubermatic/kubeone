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

package kubeconfig

import (
	"github.com/pkg/errors"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/state"
)

// Download downloads Kubeconfig over SSH
func Download(s *state.State) ([]byte, error) {
	// connect to host
	host, err := s.Cluster.Leader()
	if err != nil {
		return nil, err
	}

	conn, err := s.Connector.Connect(host)
	if err != nil {
		return nil, err
	}

	return CatKubernetesAdminConf(conn)
}

func CatKubernetesAdminConf(conn ssh.Connection) ([]byte, error) {
	konfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return nil, err
	}

	return []byte(konfig), nil
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(s *state.State) error {
	s.Logger.Infoln("Building Kubernetes clientset…")

	kubeconfig, err := Download(s)
	if err != nil {
		return errors.Wrap(err, "unable to download kubeconfig")
	}

	s.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to build config from kubeconfig bytes")
	}

	tunn, err := s.Connector.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return errors.Wrap(err, "failed to get SSH tunnel")
	}

	s.RESTConfig.Dial = tunn.TunnelTo

	return errors.WithStack(HackIssue321InitDynamicClient(s))
}

// HackIssue321InitDynamicClient initialize controller-runtime/client
// name comes from: https://github.com/kubernetes-sigs/controller-runtime/issues/321
func HackIssue321InitDynamicClient(s *state.State) error {
	var err error
	if s.RESTConfig == nil {
		return errors.New("rest config is not initialized")
	}

	s.DynamicClient, err = client.New(s.RESTConfig, client.Options{})
	return errors.Wrap(err, "unable to build dynamic client")
}
