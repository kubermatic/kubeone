package ark

// arkBackupsCRD creates Backup CRD
func arkBackupsCRD() string {
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

// arkSchedulesCRD creates Schedule CRD
func arkSchedulesCRD() string {
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

// arkRestorescRD creates Restore CRD
func arkRestoresCRD() string {
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

// arkDownloadRequestsCRD creates DownloadRequest CRD
func arkDownloadRequestsCRD() string {
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

// arkDeleteBackupRequest creates BackupRequest CRD
func arkDeleteBackupRequest() string {
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

// arkPodVolumeBackupsCRD creates PodVolumeBackup CRD
func arkPodVolumeBackupsCRD() string {
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

// arkPodVolumeRestoresCRD creates PodVolumeRestore CRD
func arkPodVolumeRestoresCRD() string {
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

// arkRestingRepositoriesCRD creates ResticRepository CRD
func arkResticRepositoriesCRD() string {
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

// arkBackupStorageLocationsCRD creates BackupStorageLocation CRD
func arkBackupStorageLocationsCRD() string {
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

// arkVolumeSnapshotLocationCRD creates VolumeSnapshot CRD
func arkVolumeSnapshotLocationsCRD() string {
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
