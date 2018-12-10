package ark

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates"
)

// Manifest returns the YAML-encoded manifest containing all
// resources for deployig Ark into a cluster.
func Manifest(cluster *config.Cluster) (string, error) {
	items := []interface{}{
		// Ark CRDs
		arkBackupsCRD(),
		arkSchedulesCRD(),
		arkRestoresCRD(),
		arkDownloadRequestsCRD(),
		arkDeleteBackupRequest(),
		arkPodVolumeBackupsCRD(),
		arkPodVolumeRestoresCRD(),
		arkResticRepositoriesCRD(),
		arkBackupStorageLocationsCRD(),
		arkVolumeSnapshotLocationsCRD(),

		// Ark Prerequisites
		arkNamespace(),
		arkServiceAccount(),
		arkRBACRole(),

		// Configuration
		createArkAWSCredentials(cluster),
		createArkBackupLocation(cluster),
		createArkVolumeSnapshotLocation(cluster),

		// Deployment
		// TODO(xmudrii): Restic
		arkDeployment(),
	}

	return templates.KubernetesToYAML(items)
}
