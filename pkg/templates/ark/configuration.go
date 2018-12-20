package ark

import (
	"fmt"

	"github.com/kubermatic/kubeone/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arkv1 "github.com/heptio/ark/pkg/apis/ark/v1"
)

// TODO(xmudrii): Other providers
func awsCredentials(cluster *config.Cluster) corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloud-credentials",
			Namespace: "heptio-ark",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"cloud": []byte(fmt.Sprintf("[default]\naws_access_key_id=%s\naws_secret_access_key=%s", cluster.Backup.S3AccessKey, cluster.Backup.S3SecretAccessKey)),
		},
	}
}

func backupLocation(cluster *config.Cluster) arkv1.BackupStorageLocation {
	return arkv1.BackupStorageLocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ark.heptio.com/v1",
			Kind:       "BackupStorageLocation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "heptio-ark",
		},
		Spec: arkv1.BackupStorageLocationSpec{
			Provider: cluster.Backup.Provider,
			StorageType: arkv1.StorageType{
				ObjectStorage: &arkv1.ObjectStorageLocation{
					Bucket: cluster.Backup.BucketName,
				},
			},
			Config: cluster.Backup.BackupStorageConfig,
		},
	}
}

func volumeSnapshotLocation(cluster *config.Cluster) arkv1.VolumeSnapshotLocation {
	return arkv1.VolumeSnapshotLocation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ark.heptio.com/v1",
			Kind:       "VolumeSnapshotLocation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-default",
			Namespace: "heptio-ark",
		},
		Spec: arkv1.VolumeSnapshotLocationSpec{
			Provider: cluster.Backup.Provider,
			Config:   cluster.Backup.VolumesSnapshotConfig,
		},
	}
}
