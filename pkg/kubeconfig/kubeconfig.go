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
	"context"
	"io/fs"
	"net"
	"os"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func File(st *state.State) (*os.File, func(), error) {
	konfigBuf, err := Download(st)
	if err != nil {
		return nil, nil, err
	}

	tmpKubeConf, err := os.CreateTemp("", "kubeone-kubeconfig-*")
	if err != nil {
		return nil, nil, fail.Runtime(err, "creating temp file for helm kubeconfig")
	}

	cleanupFn := func() {
		name := tmpKubeConf.Name()
		tmpKubeConf.Close()
		os.Remove(name)
	}

	n, err := tmpKubeConf.Write(konfigBuf)
	if err != nil {
		cleanupFn()
		return nil, nil, fail.Runtime(err, "wring temp file for helm kubeconfig")
	}
	if n != len(konfigBuf) {
		cleanupFn()
		return nil, nil, fail.NewRuntimeError("incorrect number of bytes written to temp kubeconfig", "")
	}

	return tmpKubeConf, cleanupFn, nil
}

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

	err = TunnelRestConfig(s, s.RESTConfig)
	if err != nil {
		return err
	}

	dynamicClient, err := client.New(s.RESTConfig, client.Options{})
	if err != nil {
		return fail.KubeClient(err, "building dynamic kubernetes client")
	}

	s.DynamicClient = dynamicClient

	return nil
}

func TunnelRestConfig(s *state.State, rc *rest.Config) error {
	rc.WarningHandler = rest.NewWarningWriter(os.Stderr, rest.WarningWriterOptions{
		Deduplicate: true,
	})

	rc.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		dial := TunnelDialerFactory(s.Executor, s.Cluster.RandomHost())

		return dial(ctx, network, address)
	}

	return nil
}

func TunnelDialerFactory(adapter executor.Adapter, host kubeoneapi.HostConfig) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		tunn, err := adapter.Tunnel(host)
		if err != nil {
			return nil, fail.KubeClient(err, "getting SSH tunnel")
		}

		netConn, err := tunn.TunnelTo(ctx, network, address)
		if err != nil {
			tunn.Close()

			return nil, err
		}

		return netConn, err
	}
}
