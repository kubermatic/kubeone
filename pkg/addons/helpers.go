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

package addons

import (
	"io/fs"
	"reflect"

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	awsCSIDriverName                       = "ebs.csi.aws.com"
	azureDiskCSIDriverName                 = "disk.csi.azure.com"
	azurediskCSINodeSecretBindingName      = "csi-azuredisk-node-secret-binding" //nolint:gosec
	azurediskCSINodeSecretRoleName         = "csi-azuredisk-node-secret-role"    //nolint:gosec
	gceStandardStorageClassName            = "standard"
	hetznerCSIControllerDeploymentName     = "hcloud-csi-controller"
	hetznerCSIDriverName                   = "csi.hetzner.cloud"
	hetznerCSINodeDaemonSetName            = "hcloud-csi-node"
	openstackCCMName                       = "openstack-cloud-controller-manager"
	openstackCinderCSIControllerPluginName = "openstack-cinder-csi-controllerplugin"
	openstackCinderCSINodePluginName       = "openstack-cinder-csi-nodeplugin"
	vSphereDeploymentName                  = "vsphere-cloud-controller-manager"
)

var (
	expectedAWSCSIPodSelectors = map[string]string{
		"app.kubernetes.io/name":     "aws-ebs-csi-driver",
		"app.kubernetes.io/instance": "aws-ebs-csi-driver",
	}
	expectedOpenStackCCMPodSelectors = map[string]string{
		"component": "controllermanager",
		"app":       "openstack-cloud-controller-manager",
		"release":   "openstack-ccm",
	}
	expectedOpenStackCSIPodSelectors = map[string]string{
		"app":     "openstack-cinder-csi",
		"release": "cinder-csi",
	}
)

func migrateAWSCSIDriver(s *state.State) error {
	if err := migrateAWSCSIController(s); err != nil {
		return err
	}

	if err := migrateAWSCSINode(s); err != nil {
		return err
	}

	return clientutil.DeleteIfExists(
		s.Context,
		s.DynamicClient,
		genNamedObject[storagev1.CSIDriver](awsCSIDriverName),
	)
}

func migrateAWSCSIController(s *state.State) error {
	key := client.ObjectKey{
		Name:      "ebs-csi-controller",
		Namespace: metav1.NamespaceSystem,
	}

	expectedAWSCSIPodSelectors["app"] = "ebs-csi-controller"

	return migrateDeploymentIfPodSelectorDifferent(s, key, expectedAWSCSIPodSelectors)
}

func migrateAWSCSINode(s *state.State) error {
	key := client.ObjectKey{
		Name:      "ebs-csi-node",
		Namespace: metav1.NamespaceSystem,
	}

	expectedAWSCSIPodSelectors["app"] = "ebs-csi-node"

	return migrateDaemonsetIfPodSelectorDifferent(s, key, expectedAWSCSIPodSelectors)
}

func migrateOpenStackCCM(s *state.State) error {
	key := client.ObjectKey{
		Name:      openstackCCMName,
		Namespace: metav1.NamespaceSystem,
	}

	return migrateDaemonsetIfPodSelectorDifferent(s, key, expectedOpenStackCCMPodSelectors)
}

func migrateOpenStackCSIDriver(s *state.State) error {
	if err := migrateOpenStackCSIController(s); err != nil {
		return err
	}

	return migrateOpenStackCSINode(s)
}

func migrateOpenStackCSIController(s *state.State) error {
	key := client.ObjectKey{
		Name:      openstackCinderCSIControllerPluginName,
		Namespace: metav1.NamespaceSystem,
	}

	expectedOpenStackCSIPodSelectors["component"] = "controllerplugin"

	return migrateDeploymentIfPodSelectorDifferent(s, key, expectedOpenStackCSIPodSelectors)
}

func migrateOpenStackCSINode(s *state.State) error {
	key := client.ObjectKey{
		Name:      openstackCinderCSINodePluginName,
		Namespace: metav1.NamespaceSystem,
	}

	expectedOpenStackCSIPodSelectors["component"] = "nodeplugin"

	return migrateDaemonsetIfPodSelectorDifferent(s, key, expectedOpenStackCSIPodSelectors)
}

func migrateGCEStandardStorageClass(s *state.State) error {
	return clientutil.DeleteIfExists(
		s.Context,
		s.DynamicClient,
		genNamedObject[storagev1.StorageClass](gceStandardStorageClassName),
	)
}

func migrateAzureFileCSI(s *state.State) error {
	dsKey := client.ObjectKey{
		Name:      "csi-azurefile-nodemanager",
		Namespace: metav1.NamespaceSystem,
	}
	dsLabels := map[string]string{
		"app":                        "csi-azurefile-nodemanager",
		"app.kubernetes.io/name":     "azurefile-csi-driver",
		"app.kubernetes.io/instance": "azurefile-csi-driver",
	}
	if err := migrateDaemonsetIfPodSelectorDifferent(s, dsKey, dsLabels); err != nil {
		return nil
	}

	deployKey := client.ObjectKey{
		Name:      "csi-azurefile-controllermanager",
		Namespace: metav1.NamespaceSystem,
	}
	deployLabels := map[string]string{
		"app.kubernetes.io/name":     "azurefile-csi-driver",
		"app.kubernetes.io/instance": "azurefile-csi-driver",
		"app":                        "csi-azurefile-controllermanager",
	}
	if err := migrateDeploymentIfPodSelectorDifferent(s, deployKey, deployLabels); err != nil {
		return nil
	}

	return nil
}

func migrateAzureDiskCSI(s *state.State) error {
	if err := migrateAzureDiskNodeCRBIfLegacy(s); err != nil {
		return err
	}

	return clientutil.DeleteIfExists(
		s.Context,
		s.DynamicClient,
		genNamedObject[storagev1.CSIDriver](azureDiskCSIDriverName),
	)
}

func migrateAzureDiskNodeCRBIfLegacy(s *state.State) error {
	crb := &rbacv1.ClusterRoleBinding{}
	key := client.ObjectKey{
		Name: azurediskCSINodeSecretBindingName,
	}

	if err := s.DynamicClient.Get(s.Context, key, crb); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if crb.RoleRef.Name == azurediskCSINodeSecretRoleName {
		return clientutil.DeleteIfExists(s.Context, s.DynamicClient, crb)
	}

	return nil
}

func migrateHetznerCSI(s *state.State) error {
	hzDeploymentKey := client.ObjectKey{
		Name:      hetznerCSIControllerDeploymentName,
		Namespace: metav1.NamespaceSystem,
	}
	hzDeploymentLabels := map[string]string{
		"app.kubernetes.io/name":      "hcloud-csi",
		"app.kubernetes.io/instance":  "hcloud-csi",
		"app.kubernetes.io/component": "controller",
	}
	if err := migrateDeploymentIfPodSelectorDifferent(s, hzDeploymentKey, hzDeploymentLabels); err != nil {
		return err
	}

	hzDeamonSetKey := client.ObjectKey{
		Name:      hetznerCSINodeDaemonSetName,
		Namespace: metav1.NamespaceSystem,
	}
	hzDaemonSetLabels := map[string]string{
		"app.kubernetes.io/name":      "hcloud-csi",
		"app.kubernetes.io/instance":  "hcloud-csi",
		"app.kubernetes.io/component": "node",
	}
	if err := migrateDaemonsetIfPodSelectorDifferent(s, hzDeamonSetKey, hzDaemonSetLabels); err != nil {
		return err
	}

	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, genNamedObject[storagev1.CSIDriver](hetznerCSIDriverName))
}

func migrateHetznerCCM(s *state.State) error {
	key := client.ObjectKey{
		Name:      "hcloud-cloud-controller-manager",
		Namespace: metav1.NamespaceSystem,
	}

	return migrateDeploymentIfPodSelectorDifferent(s, key, map[string]string{
		"app.kubernetes.io/instance": "hccm",
		"app.kubernetes.io/name":     "hcloud-cloud-controller-manager",
	})
}

func migrateVsphereAddon(s *state.State) error {
	return clientutil.DeleteIfExists(
		s.Context,
		s.DynamicClient,
		genNamedObject[corev1.Service](vSphereDeploymentName, metav1.NamespaceSystem),
	)
}

func migratePacketToEquinixCCM(s *state.State) error {
	return DeleteAddonByName(s, resources.AddonCCMPacket)
}

func removeCSIVsphereFromKubeSystem(s *state.State) error {
	return DeleteAddonByName(s, resources.AddonCSIVsphereKubeSystem)
}

func migrateMetricsServer(state *state.State) error {
	return migrateDeploymentIfPodSelectorDifferent(state,
		client.ObjectKey{
			Name:      "metrics-server",
			Namespace: metav1.NamespaceSystem,
		},
		map[string]string{
			"app.kubernetes.io/instance": "metrics-server",
			"app.kubernetes.io/name":     "metrics-server",
		})
}

func migrateDeploymentIfPodSelectorDifferent(s *state.State, key client.ObjectKey, expectedPodSelectors map[string]string) error {
	deploy := &appsv1.Deployment{}
	if err := s.DynamicClient.Get(s.Context, key, deploy); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if deploy.Spec.Selector != nil && reflect.DeepEqual(deploy.Spec.Selector.MatchLabels, expectedPodSelectors) {
		return nil
	}

	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, deploy)
}

func migrateDaemonsetIfPodSelectorDifferent(s *state.State, key client.ObjectKey, expectedPodSelectors map[string]string) error {
	ds := &appsv1.DaemonSet{}
	if err := s.DynamicClient.Get(s.Context, key, ds); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if ds.Spec.Selector != nil && reflect.DeepEqual(ds.Spec.Selector.MatchLabels, expectedPodSelectors) {
		return nil
	}

	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, ds)
}

// EmbeddedAddonsOnly checks if all specified addons are embedded addons
func EmbeddedAddonsOnly(addons []kubeoneapi.Addon) (bool, error) {
	// Read the directory entries for embedded addons
	embeddedAddons, err := fs.ReadDir(embeddedaddons.FS, ".")
	if err != nil {
		return false, fail.Runtime(err, "reading embedded addons directory")
	}

	// Iterate over addons specified in the KubeOneCluster object
	for _, addon := range addons {
		embedded := false
		// Iterate over embedded addons directory to check if the addon exists
		for _, embeddedAddon := range embeddedAddons {
			if embeddedAddon.Name() == addon.Name {
				embedded = true

				break
			}
		}
		// At each iteration of the outer loop check if the "addon" was a customer/non-embedded addon
		if !embedded {
			return false, nil
		}
	}

	return true, nil
}

// genNamedObject generic function to generate named metav1.Object
//
// Usage:
//
//	genNamedObject[appsv1.Deployment]("my-name", "my-namespace"), will return an equivalent
//	*appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "my-name",
//			Namespace: "my-namespace",
//		}
//	}
func genNamedObject[T any, PT interface {
	*T
	metav1.Object
}](names ...string,
) PT {
	t := PT(new(T))

	name := names[0]
	t.SetName(name)

	if len(names) > 1 {
		namespace := names[1]
		t.SetNamespace(namespace)
	}

	return t
}
