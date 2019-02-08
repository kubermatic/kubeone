package ark

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// backupsCRD creates Backup CRD
func backupsCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backups.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "backups",
				Kind:   "Backup",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// schedulesCRD creates Schedule CRD
func schedulesCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "schedules.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "schedules",
				Kind:   "Schedule",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// restoresCRD creates Restore CRD
func restoresCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "restores.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "restores",
				Kind:   "Restore",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// downloadRequestsCRD creates DownloadRequest CRD
func downloadRequestsCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "downloadrequests.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "downloadrequests",
				Kind:   "DownloadRequest",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// deleteBackupRequestCRD creates BackupRequest CRD
func deleteBackupRequestCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "deletebackuprequests.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "deletebackuprequests",
				Kind:   "DeleteBackupRequest",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// podVolumeBackupsCRD creates PodVolumeBackup CRD
func podVolumeBackupsCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "podvolumebackups.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "podvolumebackups",
				Kind:   "PodVolumeBackup",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// podVolumeRestoresCRD creates PodVolumeRestore CRD
func podVolumeRestoresCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "podvolumerestores.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "podvolumerestores",
				Kind:   "PodVolumeRestore",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// resticRepositoriesCRD creates ResticRepository CRD
func resticRepositoriesCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "resticrepositories.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "resticrepositories",
				Kind:   "ResticRepository",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// backupStorageLocationsCRD creates BackupStorageLocation CRD
func backupStorageLocationsCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backupstoragelocations.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "backupstoragelocations",
				Kind:   "BackupStorageLocation",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}

// volumeSnapshotLocationsCRD creates VolumeSnapshot CRD
func volumeSnapshotLocationsCRD() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1beta1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "volumesnapshotlocations.ark.heptio.com",
			Labels: map[string]string{
				"component": "ark",
			},
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
				},
			},
			Group: "ark.heptio.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural: "volumesnapshotlocations",
				Kind:   "VolumeSnapshotLocation",
			},
			Scope: apiextensions.NamespaceScoped,
		},
	}
}
