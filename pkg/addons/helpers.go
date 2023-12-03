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
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	awsCSIDriverName            = "ebs.csi.aws.com"
	azureDiskCSIDriverName      = "disk.csi.azure.com"
	hetznerCSIDriverName        = "csi.hetzner.cloud"
	gceStandardStorageClassName = "standard"
	vSphereDeploymentName       = "vsphere-cloud-controller-manager"
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

	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, awsCSIDriver())
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

func awsCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: awsCSIDriverName,
		},
	}
}

func migrateOpenStackCCM(s *state.State) error {
	key := client.ObjectKey{
		Name:      "openstack-cloud-controller-manager",
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
		Name:      "openstack-cinder-csi-controllerplugin",
		Namespace: metav1.NamespaceSystem,
	}

	expectedOpenStackCSIPodSelectors["component"] = "controllerplugin"

	return migrateDeploymentIfPodSelectorDifferent(s, key, expectedOpenStackCSIPodSelectors)
}

func migrateOpenStackCSINode(s *state.State) error {
	key := client.ObjectKey{
		Name:      "openstack-cinder-csi-nodeplugin",
		Namespace: metav1.NamespaceSystem,
	}

	expectedOpenStackCSIPodSelectors["component"] = "nodeplugin"

	return migrateDaemonsetIfPodSelectorDifferent(s, key, expectedOpenStackCSIPodSelectors)
}

func migrateGCEStandardStorageClass(s *state.State) error {
	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, gceStandardStorageClass())
}

func gceStandardStorageClass() *storagev1.StorageClass {
	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gceStandardStorageClassName,
		},
	}
}

func migrateAzureDiskCSIDriver(s *state.State) error {
	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, azureDiskCSIDriver())
}

func azureDiskCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: azureDiskCSIDriverName,
		},
	}
}

func migrateHetznerCSIDriver(s *state.State) error {
	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, hetznerDiskCSIDriver())
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

func hetznerDiskCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: hetznerCSIDriverName,
		},
	}
}

func migrateVsphereAddon(s *state.State) error {
	return clientutil.DeleteIfExists(s.Context, s.DynamicClient, vSphereService())
}

func vSphereService() *corev1.Service {
	// We're intentionally keeping only Service metadata, as it's enough for
	// deleting the object
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vSphereDeploymentName,
			Namespace: metav1.NamespaceSystem,
		},
	}
}

func migratePacketToEquinixCCM(s *state.State) error {
	return DeleteAddonByName(s, resources.AddonCCMPacket)
}

func removeCSIVsphereFromKubeSystem(s *state.State) error {
	return DeleteAddonByName(s, resources.AddonCSIVsphereKubeSystem)
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
