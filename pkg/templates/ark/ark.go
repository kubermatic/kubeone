package ark

import (
	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/templates"
)

// Manifest returns the YAML-encoded manifest containing all
// resources for deployig Ark into a cluster.
func Manifest(cluster *config.Cluster) (string, error) {
	deploymentManifest, err := deployment(cluster)
	if err != nil {
		return "", err
	}

	items := []interface{}{
		// Ark CRDs
		backupsCRD(),
		schedulesCRD(),
		restoresCRD(),
		downloadRequestsCRD(),
		deleteBackupRequest(),
		podVolumeBackupsCRD(),
		podVolumeRestoresCRD(),
		resticRepositoriesCRD(),
		backupStorageLocationsCRD(),
		volumeSnapshotLocationsCRD(),

		// Ark Prerequisites
		namespace(),
		serviceAccount(),
		rbacRole(),

		// Credentials
		// TODO(xmudrii): Credentials needed for other provider stores.
		awsCredentials(cluster),

		// Backup location confiugration
		backupLocation(cluster),
		volumeSnapshotLocation(cluster),

		// Deployment
		deploymentManifest,

		// Restic
		resticDaemonset(),

		// Etcd automatic backup schedule
		etcdBackupSchedule(cluster),
	}

	return templates.KubernetesToYAML(items)
}
