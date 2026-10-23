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

package infra

import (
	"fmt"
	"net"
	"strconv"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/fail"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

// inClusterServer is the infra API server address as seen from pods running
// in the infra cluster.
const inClusterServer = "https://kubernetes.default.svc"

// InClusterKubeconfig builds a kubeconfig for the infra cluster from an infra
// ServiceAccount token Secret. It only works from inside the infra cluster.
func InClusterKubeconfig(clusterName, namespace string, token *corev1.Secret) ([]byte, error) {
	return tokenKubeconfig("infra", inClusterServer, namespace, clusterName, token)
}

// TenantKubeconfig builds a kubeconfig for the tenant cluster from a tenant
// ServiceAccount token Secret. The token's ca.crt is the tenant cluster CA.
func TenantKubeconfig(clusterName, server string, token *corev1.Secret) ([]byte, error) {
	return tokenKubeconfig(clusterName, server, "", clusterName, token)
}

// TenantServer returns the tenant API server URL for the given endpoint.
func TenantServer(endpoint kubeoneapi.APIEndpoint) string {
	return "https://" + net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port))
}

func tokenKubeconfig(contextCluster, server, namespace, clusterName string, token *corev1.Secret) ([]byte, error) {
	if !tokenPopulated(token) {
		return nil, fail.NewRuntimeError("building kubeconfig", "ServiceAccount token Secret %s/%s is not populated", token.Namespace, token.Name)
	}

	user := fmt.Sprintf("%s-%s", clusterName, token.Annotations[corev1.ServiceAccountNameKey])

	kubeconfig := clientapi.NewConfig()
	kubeconfig.Clusters[contextCluster] = &clientapi.Cluster{
		Server:                   server,
		CertificateAuthorityData: token.Data[corev1.ServiceAccountRootCAKey],
	}
	kubeconfig.AuthInfos[user] = &clientapi.AuthInfo{
		Token: string(token.Data[corev1.ServiceAccountTokenKey]),
	}
	kubeconfig.Contexts[user] = &clientapi.Context{
		Cluster:   contextCluster,
		AuthInfo:  user,
		Namespace: namespace,
	}
	kubeconfig.CurrentContext = user

	out, err := clientcmd.Write(*kubeconfig)

	return out, fail.Runtime(err, "marshalling kubeconfig")
}
