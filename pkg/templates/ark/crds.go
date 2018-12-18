package ark

// backupsCRD creates Backup CRD
func backupsCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: backups.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: backups
    kind: Backup
`
}

// schedulesCRD creates Schedule CRD
func schedulesCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: schedules.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: schedules
    kind: Schedule
`
}

// restoresCRD creates Restore CRD
func restoresCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: restores.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: restores
    kind: Restore
`
}

// downloadRequestsCRD creates DownloadRequest CRD
func downloadRequestsCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: downloadrequests.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: downloadrequests
    kind: DownloadRequest
`
}

// deleteBackupRequest creates BackupRequest CRD
func deleteBackupRequest() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: deletebackuprequests.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: deletebackuprequests
    kind: DeleteBackupRequest
`
}

// podVolumeBackupsCRD creates PodVolumeBackup CRD
func podVolumeBackupsCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: podvolumebackups.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: podvolumebackups
    kind: PodVolumeBackup
`
}

// podVolumeRestoresCRD creates PodVolumeRestore CRD
func podVolumeRestoresCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: podvolumerestores.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: podvolumerestores
    kind: PodVolumeRestore
`
}

// resticRepositoriesCRD creates ResticRepository CRD
func resticRepositoriesCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: resticrepositories.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: resticrepositories
    kind: ResticRepository
`
}

// backupStorageLocationsCRD creates BackupStorageLocation CRD
func backupStorageLocationsCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: backupstoragelocations.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: backupstoragelocations
    kind: BackupStorageLocation
`
}

// volumeSnapshotLocationsCRD creates VolumeSnapshot CRD
func volumeSnapshotLocationsCRD() string {
	return `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: volumesnapshotlocations.ark.heptio.com
  labels:
    component: ark
spec:
  group: ark.heptio.com
  version: v1
  scope: Namespaced
  names:
    plural: volumesnapshotlocations
    kind: VolumeSnapshotLocation
`
}
