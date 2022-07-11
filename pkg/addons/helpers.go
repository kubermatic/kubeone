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

	embeddedaddons "k8c.io/kubeone/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/clientutil"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	azureDiskCSIDriverName      = "disk.csi.azure.com"
	gceStandardStorageClassName = "standard"
	vSphereDeploymentName       = "vsphere-cloud-controller-manager"
)

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
