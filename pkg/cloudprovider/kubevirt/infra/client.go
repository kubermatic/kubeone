/*
Copyright 2026 The KubeOne Authors.

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

// Package infra manages KubeOne-owned objects in the KubeVirt infra cluster:
// identities (ServiceAccounts, Roles, kubeconfigs) of components that run in
// the InfraNamespace, and their cleanup.
package infra

import (
	"encoding/base64"
	"fmt"

	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"k8s.io/client-go/tools/clientcmd"
	ctrlrt "sigs.k8s.io/controller-runtime/pkg/client"
)

// UserKubeconfig returns the user-provided infra cluster kubeconfig
// (KUBEVIRT_KUBECONFIG), base64 decoded if needed.
func UserKubeconfig(s *state.State) ([]byte, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	raw := providerCreds[credentials.KubevirtKubeconfigKey]
	if raw == "" {
		return nil, fail.Config(fmt.Errorf("kubevirt kubeconfig is empty"), "reading kubevirt kubeconfig")
	}

	if decoded, decErr := base64.StdEncoding.DecodeString(raw); decErr == nil {
		return decoded, nil
	}

	return []byte(raw), nil
}

// Client returns a client for the infra cluster, built from the user-provided
// kubeconfig, and the InfraNamespace.
func Client(s *state.State) (ctrlrt.Client, string, error) {
	kubeconfig, err := UserKubeconfig(s)
	if err != nil {
		return nil, "", err
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, "", fail.Config(err, "parsing kubevirt kubeconfig")
	}

	infraClient, err := ctrlrt.New(restConfig, ctrlrt.Options{})

	return infraClient, s.Cluster.CloudProvider.Kubevirt.InfraNamespace, fail.Runtime(err, "creating kubevirt infra cluster client")
}
