package ark

import (
	"errors"

	"github.com/kubermatic/kubeone/pkg/installer/util"
	"github.com/kubermatic/kubeone/pkg/templates"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arkv1 "github.com/heptio/ark/pkg/apis/ark/v1"
	arkclientset "github.com/heptio/ark/pkg/generated/clientset/versioned/typed/ark/v1"
)

const (
	arkContainerImage = "gcr.io/heptio-images/ark:v0.10.0"
	arkNamespace      = "heptio-ark"
	arkServiceAccount = "ark"
)

// Deploy deploys Ark on the target cluster
func Deploy(ctx *util.Context) error {
	if ctx.Clientset == nil {
		return errors.New("kubernetes clientset not initialized")
	}
	if ctx.APIExtensionClientset == nil {
		return errors.New("kubernetes apiextension clientset not initialized")
	}

	// Kubernetes clientsets
	coreClient := ctx.Clientset.CoreV1()
	rbacClient := ctx.Clientset.RbacV1()
	appsClient := ctx.Clientset.AppsV1()

	// Ark clientset
	arkClient, err := arkclientset.NewForConfig(ctx.RESTConfig)
	if err != nil {
		return err
	}

	// CRDs
	crdGenerators := []func() *apiextensions.CustomResourceDefinition{
		backupsCRD,
		schedulesCRD,
		restoresCRD,
		downloadRequestsCRD,
		deleteBackupRequestCRD,
		podVolumeBackupsCRD,
		podVolumeRestoresCRD,
		resticRepositoriesCRD,
		backupStorageLocationsCRD,
		volumeSnapshotLocationsCRD,
	}
	crdClient := ctx.APIExtensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions()

	for _, crdGen := range crdGenerators {
		crdErr := templates.EnsureCRD(crdClient, crdGen())
		if crdErr != nil {
			return crdErr
		}
	}

	// Namespace
	err = templates.EnsureNamespace(coreClient.Namespaces(), namespace())
	if err != nil {
		return err
	}

	// ServiceAccount
	sa := serviceAccount()
	err = templates.EnsureServiceAccount(coreClient.ServiceAccounts(sa.Namespace), sa)
	if err != nil {
		return err
	}

	// RBAC Role
	err = templates.EnsureClusterRoleBinding(rbacClient.ClusterRoleBindings(), clusterRoleBinding())
	if err != nil {
		return err
	}

	// Credentials
	credentials := awsCredentials(ctx.Cluster)
	err = templates.EnsureSecret(coreClient.Secrets(credentials.Namespace), credentials)
	if err != nil {
		return err
	}

	// Backup and Volume Locations
	bsl := backupStorageLocation(ctx.Cluster)
	err = ensureBackupStorageLocation(arkClient.BackupStorageLocations(bsl.Namespace), bsl)
	if err != nil {
		return err
	}
	vsl := volumeSnapshotLocation(ctx.Cluster)
	err = ensureVolumeSnapshotLocation(arkClient.VolumeSnapshotLocations(vsl.Namespace), vsl)
	if err != nil {
		return err
	}

	// Ark Deployment
	dep := deployment(ctx.Cluster)
	err = templates.EnsureDeployment(appsClient.Deployments(dep.Namespace), dep)
	if err != nil {
		return err
	}

	// Restic DaemonSet
	ds := resticDaemonset()
	err = templates.EnsureDaemonSet(appsClient.DaemonSets(ds.Namespace), ds)
	if err != nil {
		return err
	}

	return nil
}

func ensureBackupStorageLocation(backupLocationInterface arkclientset.BackupStorageLocationInterface, required *arkv1.BackupStorageLocation) error {
	existing, err := backupLocationInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = backupLocationInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = backupLocationInterface.Update(existing)
	return err
}

func ensureVolumeSnapshotLocation(snapshotLocationInterface arkclientset.VolumeSnapshotLocationInterface, required *arkv1.VolumeSnapshotLocation) error {
	existing, err := snapshotLocationInterface.Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = snapshotLocationInterface.Create(required)
		return err
	}
	if err != nil {
		return err
	}

	modified := false
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Annotations, required.ObjectMeta.Annotations)
	templates.MergeStringMap(&modified, &existing.ObjectMeta.Labels, required.ObjectMeta.Labels)
	if equality.Semantic.DeepEqual(required.Spec, existing.Spec) && !modified {
		return nil
	}

	_, err = snapshotLocationInterface.Update(existing)
	return err
}
