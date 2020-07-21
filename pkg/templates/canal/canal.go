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
	"time"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/clientutil"
	"github.com/kubermatic/kubeone/pkg/kubeconfig"
	"github.com/kubermatic/kubeone/pkg/state"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	installCNIImage = "calico/cni:v3.10.0"
	calicoImage     = "calico/node:v3.10.0"
	flannelImage    = "quay.io/kubermatic/coreos_flannel:v0.11.0@sha256:3de983d62621898fe58ffd9537a4845c7112961a775efb205cab56e089e163b6"

	// cniNetworkConfig configures installation on the each node. The special values in this config will be
	// automatically populated
	cniNetworkConfig = `
{
  "name": "k8s-pod-network",
  "cniVersion": "0.3.1",
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
      "capabilities": {
        "portMappings": true
      }
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
func Deploy(s *state.State) error {
	if s.DynamicClient == nil {
		return errors.New("kubernetes dynamic client is not initialized")
	}

	// Populate Flannel network configuration
	tpl, err := template.New("base").Parse(flannelNetworkConfig)
	if err != nil {
		return errors.Wrap(err, "failed to parse canal config")
	}

	variables := map[string]interface{}{
		"POD_SUBNET": s.Cluster.ClusterNetwork.PodSubnet,
	}

	buf := bytes.Buffer{}
	if err = tpl.Execute(&buf, variables); err != nil {
		return errors.Wrap(err, "failed to render canal config")
	}

	ctx := context.Background()

	k8sobjects := []runtime.Object{
		// CRDs
		felixConfigurationCRD(),
		ipamBlockCRD(),
		blockAffinityCRD(),
		ipamHandleCRD(),
		ipamConfigCRD(),
		bgpPeerCRD(),
		bgpConfigurationCRD(),
		ipPoolCRD(),
		hostEndpointCRD(),
		clusterInformationCRD(),
		globalNetworkPolicyCRD(),
		globalNetworksetCRD(),
		networkPolicyCRD(),
		networkSetCRD(),

		// RBAC
		calicoClusterRole(),
		flannelClusterRole(),
		calicoClusterRoleBinding(),
		flannelClusterRoleBinding(),
		canalClusterRoleBinding(),

		// workloads
		configMap(buf),
		daemonSet(s.PatchCNI),
		serviceAccount(),
	}

	for _, obj := range k8sobjects {
		if err = clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj); err != nil {
			return errors.WithStack(err)
		}
	}

	gkResources := []string{}
	for _, obj := range k8sobjects {
		if gvk := obj.GetObjectKind().GroupVersionKind(); gvk.Group == "crd.projectcalico.org" {
			gkResources = append(gkResources, gvk.GroupKind().String())
		}
	}

	var waitErr error

	for _, res := range gkResources {
		waitErr = wait.Poll(5*time.Second, 1*time.Minute, func() (bool, error) {
			ok, crdErr := clientutil.VerifyCRD(ctx, s.DynamicClient, res)
			if crdErr != nil {
				return false, nil
			}
			if !ok {
				return ok, nil
			}
			return true, nil
		})
	}
	if waitErr != nil {
		return errors.Wrap(waitErr, "failed to establish calico CRDs")
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err = kubeconfig.HackIssue321InitDynamicClient(s)

	return errors.Wrap(err, "failed to re-init dynamic client")
}
