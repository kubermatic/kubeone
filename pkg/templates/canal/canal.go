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

package canal

import (
	"bytes"
	"context"
	"text/template"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	kubeonecontext "github.com/kubermatic/kubeone/pkg/util/context"

	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	installCNIImage = "quay.io/calico/cni:v3.4.0"
	calicoImage     = "quay.io/calico/node:v3.4.0"
	flannelImage    = "quay.io/coreos/flannel:v0.9.1"

	// cniNetworkConfig configures installation on the each node. The special values in this config will be
	// automatically populated
	cniNetworkConfig = `
{
	"name": "k8s-pod-network",
	"cniVersion": "0.3.0",
	"plugins": [
		{
			"type": "calico",
			"log_level": "info",
			"datastore_type": "kubernetes",
			"nodename": "__KUBERNETES_NODE_NAME__",
			"ipam": {
				"type": "host-local",
				"subnet": "usePodCidr"
			},
			"policy": {
					"type": "k8s"
			},
			"kubernetes": {
					"kubeconfig": "__KUBECONFIG_FILEPATH__"
			}
		},
		{
			"type": "portmap",
			"snat": true,
			"capabilities": {"portMappings": true}
		}
	]
}
`
	// Flannel network configuration (mounted into the flannel container)
	flannelNetworkConfig = `
{
	"Network": "{{ .POD_SUBNET }}",
	"Backend": {
		"Type": "vxlan"
	}
}
`
)

// Deploy deploys Canal (Calico + Flannel) CNI on the cluster
func Deploy(ctx *kubeonecontext.Context) error {
	if ctx.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	// Populate Flannel network configuration
	tpl, err := template.New("base").Parse(flannelNetworkConfig)
	if err != nil {
		return errors.Wrap(err, "failed to parse canal config")
	}

	variables := map[string]interface{}{
		"POD_SUBNET": ctx.Cluster.ClusterNetwork.PodSubnet,
	}

	buf := bytes.Buffer{}
	if err = tpl.Execute(&buf, variables); err != nil {
		return errors.Wrap(err, "failed to render canal config")
	}

	bgCtx := context.Background()
	// ConfigMap
	cm := configMap()
	cm.Data["net-conf.json"] = buf.String()
	if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, cm); err != nil {
		return errors.Wrap(err, "failed to ensure canal ConfigMap")
	}

	// DaemonSet
	ds := daemonSet()
	if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, ds); err != nil {
		return errors.Wrap(err, "failed to ensure canal DaemonSet")
	}

	// ServiceAccount
	sa := serviceAccount()
	if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, sa); err != nil {
		return errors.Wrap(err, "failed to ensure canal ServiceAccount")
	}

	// CRDs
	crdGenerators := []func() *apiextensions.CustomResourceDefinition{
		felixConfigurationCRD,
		bgpConfigurationCRD,
		ipPoolsConfigurationCRD,
		hostEndpointsConfigurationCRD,
		clusterInformationsConfigurationCRD,
		globalNetworkPoliciesConfigurationCRD,
		globalNetworksetsConfigurationCRD,
		networkPoliciesConfigurationCRD,
	}

	for _, crdGen := range crdGenerators {
		if err = simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, crdGen()); err != nil {
			return errors.Wrap(err, "failed to ensure canal CustomResourceDefinition")
		}
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err = kubeconfig.HackIssue321InitDynamicClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to re-init dynamic client")
	}

	// ClusterRoles
	crGenerators := []func() *rbacv1.ClusterRole{
		calicoClusterRole,
		flannelClusterRole,
	}

	for _, crGen := range crGenerators {
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, crGen()); err != nil {
			return errors.Wrap(err, "failed to ensure canal ClusterRole")
		}
	}

	// ClusterRoleBindings
	crbGenerators := []func() *rbacv1.ClusterRoleBinding{
		calicoClusterRoleBinding,
		flannelClusterRoleBinding,
		canalClusterRoleBinding,
	}
	for _, crbGen := range crbGenerators {
		if err := simpleCreateOrUpdate(bgCtx, ctx.DynamicClient, crbGen()); err != nil {
			return errors.Wrap(err, "failed to ensure canal ClusterRoleBinding")
		}
	}

	return nil
}
