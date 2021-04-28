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
	"text/template"
	"time"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/kubeconfig"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/images"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	canalComponentLabel = "canal"

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
      "log_file_path": "/var/log/calico/cni/cni.log",
      "datastore_type": "kubernetes",
      "nodename": "__KUBERNETES_NODE_NAME__",
      "mtu": __CNI_MTU__,
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
    },
    {
      "type": "bandwidth",
      "capabilities": {
        "bandwidth": true
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

func canalCRDs() []client.Object {
	return []client.Object{
		felixConfigurationCRD(),
		ipamBlockCRD(),
		blockAffinityCRD(),
		ipamHandleCRD(),
		ipamConfigCRD(),
		bgpPeerCRD(),
		bgpConfigurationCRD(),
		ipPoolCRD(),
		kubeControllersConfigurationCRD(),
		hostEndpointCRD(),
		clusterInformationCRD(),
		globalNetworkPolicyCRD(),
		globalNetworksetCRD(),
		networkPolicyCRD(),
		networkSetCRD(),
	}
}

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

	ctx := s.Context

	calicoCNIImage := s.Images.Get(images.CalicoCNI)
	calicoNodeImage := s.Images.Get(images.CalicoNode)
	calicoControllerImage := s.Images.Get(images.CalicoController)
	flannelImage := s.Images.Get(images.Flannel)

	crds := canalCRDs()
	k8sobjects := append(crds,
		// RBAC
		calicoKubeControllersClusterRole(),
		calicoNodeClusterRole(),
		flannelClusterRole(),
		calicoKubeControllersClusterRoleBinding(),
		flannelClusterRoleBinding(),
		canalClusterRoleBinding(),

		// workloads
		configMap(buf, s.Cluster.ClusterNetwork.CNI.Canal.MTU),
		daemonsetServiceAccount(),
		deploymentServiceAccount(),
		daemonSet(s.PatchCNI, s.Cluster.ClusterNetwork.PodSubnet, calicoCNIImage, calicoNodeImage, flannelImage),
		controllerDeployment(calicoControllerImage),
	)

	withLabel := clientutil.WithComponentLabel(canalComponentLabel)
	for _, obj := range k8sobjects {
		if err = clientutil.CreateOrUpdate(ctx, s.DynamicClient, obj, withLabel); err != nil {
			return errors.WithStack(err)
		}
	}

	gkResources := []string{}
	for _, crd := range crds {
		gkResources = append(gkResources, crd.(metav1.Object).GetName())
	}

	condFn := clientutil.CRDsReadyCondition(ctx, s.DynamicClient, gkResources)

	err = wait.Poll(5*time.Second, 1*time.Minute, condFn)
	if err != nil {
		return errors.Wrap(err, "failed to establish calico CRDs")
	}

	// HACK: re-init dynamic client in order to re-init RestMapper, to drop caches
	err = kubeconfig.HackIssue321InitDynamicClient(s)
	return errors.Wrap(err, "failed to re-init dynamic client")
}
