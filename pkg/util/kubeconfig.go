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

package util

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"

	"k8s.io/client-go/tools/clientcmd"
)

// DownloadKubeconfig downloads Kubeconfig over SSH
func DownloadKubeconfig(cluster *config.Cluster) ([]byte, error) {
	// connect to leader
	leader, err := cluster.Leader()
	if err != nil {
		return nil, err
	}
	connector := ssh.NewConnector()

	conn, err := connector.Connect(*leader)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// get the kubeconfig
	kubeconfig, _, _, err := conn.Exec("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return nil, err
	}

	return []byte(kubeconfig), nil
}

// BuildKubernetesClientset builds core kubernetes and apiextensions clientsets
func BuildKubernetesClientset(ctx *Context) error {
	ctx.Logger.Infoln("Building Kubernetes clientset…")

	kubeconfig, err := DownloadKubeconfig(ctx.Cluster)
	if err != nil {
		return errors.Wrap(err, "unable to download kubeconfig")
	}

	ctx.RESTConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to build config from kubeconfig bytes")
	}

	err = HackIssue321InitDynamicClient(ctx)
	return errors.Wrap(err, "unable to build dynamic client")
}

// HackIssue321InitDynamicClient initialize controller-runtime/client
// name comes from: https://github.com/kubernetes-sigs/controller-runtime/issues/321
func HackIssue321InitDynamicClient(ctx *Context) error {
	if ctx.RESTConfig == nil {
		return errors.New("rest config is not initialized")
	}

	var err error
	ctx.DynamicClient, err = client.New(ctx.RESTConfig, client.Options{})
	return errors.Wrap(err, "unable to build dynamic client")
}
