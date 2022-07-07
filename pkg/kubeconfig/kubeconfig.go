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
	"io/fs"
	"os"

	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Download downloads Kubeconfig over SSH
func Download(s *state.State) ([]byte, error) {
	// connect to host
	host, err := s.Cluster.Leader()
	if err != nil {
		return nil, err
	}

	conn, err := s.Executor.Open(host)
	if err != nil {
		return nil, err
	}

	return catKubernetesAdminConf(conn)
}

func catKubernetesAdminConf(conn executor.Interface) ([]byte, error) {
	return fs.ReadFile(executorfs.New(conn), "/etc/kubernetes/admin.conf")
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(s *state.State) error {
	s.Logger.Infoln("Building Kubernetes clientset...")

	kubeconfig, err := Download(s)
	if err != nil {
		return err
	}

	s.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return fail.KubeClient(err, "building config from kubeconfig")
	}

	s.RESTConfig.WarningHandler = rest.NewWarningWriter(os.Stderr, rest.WarningWriterOptions{
		Deduplicate: true,
	})

	tunn, err := s.Executor.Tunnel(s.Cluster.RandomHost())
	if err != nil {
		return fail.KubeClient(err, "getting SSH tunnel")
	}

	s.RESTConfig.Dial = tunn.TunnelTo

	s.DynamicClient, err = client.New(s.RESTConfig, client.Options{})
	if err != nil {
		return fail.KubeClient(err, "building dynamic kubernetes client")
	}

	return nil
}
