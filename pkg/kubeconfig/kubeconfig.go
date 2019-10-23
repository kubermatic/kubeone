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

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/ssh"
	"github.com/kubermatic/kubeone/pkg/state"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Download downloads Kubeconfig over SSH
func Download(cluster *kubeoneapi.KubeOneCluster) ([]byte, error) {
	// connect to leader
	leader, err := cluster.Leader()
	if err != nil {
		return nil, err
	}
	connector := ssh.NewConnector()

	conn, err := connector.Connect(leader)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// get the kubeconfig
	konfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return nil, err
	}

	return []byte(konfig), nil
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(s *state.State) error {
	s.Logger.Infoln("Building Kubernetes clientset…")

	kubeconfig, err := Download(s.Cluster)
	if err != nil {
		return errors.Wrap(err, "unable to download kubeconfig")
	}

	s.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to build config from kubeconfig bytes")
	}

	s.DynamicClient, err = client.New(s.RESTConfig, client.Options{})
	return errors.Wrap(err, "unable to build dynamic client")
}
