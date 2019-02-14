package canal

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"

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
func Deploy(ctx *util.Context) error {
	if ctx.Clientset == nil {
		return errors.New("kubernetes clientset not initialized")
	}
	if ctx.APIExtensionClientset == nil {
		return errors.New("kubernetes apiextension clientset not initialized")
	}

	// Populate Flannel network configuration
	tpl, err := template.New("base").Parse(flannelNetworkConfig)
	if err != nil {
		return fmt.Errorf("failed to parse canal config: %v", err)
	}

	variables := map[string]interface{}{
		"POD_SUBNET": ctx.Cluster.Network.PodSubnet(),
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, variables); err != nil {
		return fmt.Errorf("failed to render canal config: %v", err)
	}

	// Kubernetes clientsets
	coreClient := ctx.Clientset.CoreV1()
	rbacClient := ctx.Clientset.RbacV1()
	appsClient := ctx.Clientset.AppsV1()

	// ConfigMap
	cm := configMap()
	cm.Data["net-conf.json"] = buf.String()
	if err := templates.EnsureConfigMap(coreClient.ConfigMaps(cm.Namespace), cm); err != nil {
		return err
	}

	// DaemonSet
	ds := daemonSet()
	if err := templates.EnsureDaemonSet(appsClient.DaemonSets(ds.Namespace), ds); err != nil {
		return err
	}

	// ServiceAccount
	sa := serviceAccount()
	if err := templates.EnsureServiceAccount(coreClient.ServiceAccounts(sa.Namespace), sa); err != nil {
		return err
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
	crdClient := ctx.APIExtensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crdGen := range crdGenerators {
		if err := templates.EnsureCRD(crdClient, crdGen()); err != nil {
			return err
		}
	}

	// ClusterRoles
	crGenerators := []func() *rbacv1.ClusterRole{
		calicoClusterRole,
		flannelClusterRole,
	}
	for _, crGen := range crGenerators {
		if err := templates.EnsureClusterRole(rbacClient.ClusterRoles(), crGen()); err != nil {
			return err
		}
	}

	// ClusterRoleBindings
	crbGenerators := []func() *rbacv1.ClusterRoleBinding{
		calicoClusterRoleBinding,
		flannelClusterRoleBinding,
		canalClusterRoleBinding,
	}
	for _, crbGen := range crbGenerators {
		if err := templates.EnsureClusterRoleBinding(rbacClient.ClusterRoleBindings(), crbGen()); err != nil {
			return err
		}
	}

	return nil
}
