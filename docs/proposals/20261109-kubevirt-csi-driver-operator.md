# KubeVirt: deploy the CSI driver through kubevirt-csi-driver-operator

**Status:** **Draft**
**Created:** 2026-11-09
**Last updated:** 2026-11-09
**Author:** Artiom Diomin ([@kron4eg](https://github.com/kron4eg))
**Issue:** [kubermatic/kubeone#3978](https://github.com/kubermatic/kubeone/issues/3978)

## Abstract

The `csi-kubevirt` addon deploys the KubeVirt CSI node plugin DaemonSet and the CSI controller Deployment as static
manifests. It creates no StorageClasses, so users have to write their own `csi.kubevirt.io` StorageClasses.

This proposal adds [kubermatic/kubevirt-csi-driver-operator](https://github.com/kubermatic/kubevirt-csi-driver-operator)
to the `csi-kubevirt` addon. The operator replaces the node plugin DaemonSet, and it creates the `CSIDriver`, the
StorageClasses and the VolumeSnapshotClasses from a `Tenant` object. KubeOne renders the `Tenant` from new fields in
the KubeOneCluster manifest. The CSI controller stays in the addon, because the operator does not deploy it.

## Goals

- Deploy `kubevirt-csi-driver-operator` as part of the `csi-kubevirt` addon.
- Configure StorageClasses and VolumeSnapshotClasses in the KubeOneCluster manifest, and keep them in sync with it on
  every `kubeone apply`, including changes to fields that are immutable on a StorageClass.
- Support volume expansion and snapshots in the CSI controller.
- Migrate existing KubeVirt clusters on `kubeone apply`. Existing PVs keep working.

## Non-goals

- Moving the CSI controller, the CCM or machine-controller into the infra cluster. The tenant cluster keeps the infra
  kubeconfig in `kubeone-ccm-credentials`, as today.
- Changing `--infra-cluster-labels` (`csi-driver/cluster=tenant`). Two clusters that share an `InfraNamespace` still
  cannot tell their DataVolumes apart. Changing the label on a live cluster affects DataVolumes that already exist, so
  this needs its own proposal.
- Changes to the operator. KubeOne works around the operator's StorageClass reconcile (see
  [StorageClass reconciliation](#storageclass-and-volumesnapshotclass-reconciliation)).

## Background

### Current `csi-kubevirt` addon

- `kube-system/kubevirt-csi-node` DaemonSet, ServiceAccount `kubevirt-csi-node-sa`, ClusterRole `kubevirt-csi-node-cr`
  and ClusterRoleBinding `kubevirt-csi-node-binding`.
- `kube-system/driver-config` ConfigMap with `infraClusterNamespace` and `infraClusterLabels`.
- `kube-system/kubevirt-csi-controller` Deployment with `csi-driver`, `csi-provisioner`, `csi-attacher` and
  `csi-liveness-probe` containers. It reads the infra kubeconfig from the `kubeone-ccm-credentials` Secret. It has no
  snapshotter and no resizer sidecar.
- No `CSIDriver`, no StorageClass and no VolumeSnapshotClass.

The addon is deployed by `ensureCSIAddons` (`pkg/addons/ensure.go`) when `cloudProvider.kubevirt` is set and
`cloudProvider.disableBundledCSIDrivers` is false.

### kubevirt-csi-driver-operator (v0.5.3)

The operator runs in the tenant cluster and only talks to the tenant API server. It reconciles a cluster-scoped
`Tenant` (`csiprovisioner.kubevirt.io/v1alpha1`) named `tenant` and creates, all with a controller ownerReference to
the `Tenant`:

- the `CSIDriver` `csi.kubevirt.io`;
- ServiceAccount `kubevirt-csi-node` and its RBAC;
- DaemonSet `kubevirt-csi-node` in the namespace `kubevirt-csi-driver`. The namespace is hard-coded. The pods use
  `hostNetwork: true`, tolerate all taints, and use the same `app: kubevirt-csi-driver` selector and the same images as
  KubeOne today (`kubermatic/kubevirt-csi-driver:v0.4.5`, `openshift/origin-csi-node-driver-registrar:4.20.0`,
  `openshift/origin-csi-livenessprobe:4.20.0`). The images are hard-coded and rewritten by `--overwrite-registry`;
- one StorageClass `kubevirt-<infraStorageClassName>` per `spec.storageClasses[]` entry;
- one VolumeSnapshotClass `kubevirt-<infraVolumeSnapshotClass>` per `spec.volumeSnapshotClasses[]` entry.

It does not deploy the CSI controller. On every reconcile it deletes the Deployment
`kubevirt-csi-driver/kubevirt-csi-controller`, which older operator versions created.

The operator ignores `Tenant.spec.imageRepository` and `imageTag`.

**StorageClass reconcile.** On an existing StorageClass the operator only updates the annotations, ownerReferences,
`parameters` and `volumeBindingMode`:

| Change to a `storageClasses[]` entry                  | Result                                                                                                                |
|-------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|
| New entry                                             | Created.                                                                                                              |
| `isDefaultClass`                                      | Updated.                                                                                                              |
| `bus`, `volumeBindingMode`                            | The update is rejected because the fields are immutable. The reconcile fails at this StorageClass every time, so the StorageClasses after it and all VolumeSnapshotClasses are not reconciled either. |
| `reclaimPolicy`                                       | Ignored (the field is also immutable).                                                                                |
| `allowVolumeExpansion`, `labels`, `zones`, `regions`  | Ignored. They are set only when the StorageClass is created.                                                          |
| Entry removed                                         | The StorageClass is kept.                                                                                             |

The operator watches the StorageClasses it owns, so it recreates a deleted StorageClass right away from the current
`Tenant` spec.

**VolumeSnapshotClass reconcile.** The operator updates all fields of an existing VolumeSnapshotClass. It does not
delete VolumeSnapshotClasses of removed entries, and it does not watch VolumeSnapshotClasses.

### KKP reference

KKP deploys the operator as a user cluster addon, and runs the CSI controller in the seed
([`pkg/resources/csi/kubevirt/deployment.go`](https://github.com/kubermatic/kubermatic/blob/main/pkg/resources/csi/kubevirt/deployment.go)).
The KKP controller has `csi-snapshotter` and `csi-resizer` sidecars (the resizer with
`--handle-volume-inuse-error=false`), passes the infra namespace and labels as arguments, and pins the driver version
to the one the operator deploys. Its ClusterRole matches the one of the KubeOne controller.

## Design

### Addon layout

The addon keeps the name `csi-kubevirt`. Its manifests are replaced:

| File                                 | Content |
|--------------------------------------|---------|
| `1-crd-tenant.yaml`                  | The `tenants.csiprovisioner.kubevirt.io` CRD, copied from the operator's `v0.5.3` kustomize output. |
| `2-kubevirt-csi-driver-operator.yaml`| Namespace `kubevirt-csi-driver`, the operator ServiceAccount, ClusterRole, ClusterRoleBinding and Deployment. |
| `3-kubevirt-csi-controller.yaml`     | The CSI controller in `kube-system` (see [CSI controller](#csi-controller)). |
| `4-tenant.yaml`                      | The `Tenant` named `tenant`, rendered from the KubeOneCluster manifest. |

`nativeclient.Applier` applies CRDs first, waits until they are established, and retries objects whose type the
RESTMapper does not know yet. The `Tenant` can therefore be part of the addon.

For KubeVirt, `csi-kubevirt` is applied after `csi-external-snapshotter`. The operator creates the
VolumeSnapshotClasses, and its reconcile fails until their CRD exists. Its retries back off exponentially, so on a
fresh install the VolumeSnapshotClasses could otherwise take many minutes to appear. The order for other providers
does not change.

Because the addon name does not change, the next apply prunes objects of the old addon through the addon inventory
ConfigMap `kube-system/kubeone-addon-csi-kubevirt` and the addon label. The node plugin still needs explicit handling,
see [Replacing the node plugin](#replacing-the-node-plugin).

### Operator

Namespace, ServiceAccount, ClusterRole, ClusterRoleBinding and Deployment follow the KKP addon:

- The ClusterRole is upstream's plus `get`, `list`, `watch`, `create`, `update` and `patch` on
  `snapshot.storage.k8s.io/volumesnapshotclasses`. Upstream's ClusterRole lacks it, and the upstream kustomize output
  does not create the namespace either.
- The Deployment runs `/manager --leader-elect`, plus `--overwrite-registry=<registry>` when
  `registryConfiguration.overwriteRegistry` is set. It runs as non-root with a `RuntimeDefault` seccomp profile, all
  capabilities dropped, `priorityClassName: system-cluster-critical`, and tolerates `CriticalAddonsOnly` and the
  control-plane taint.

### CSI controller

`kube-system/kubevirt-csi-controller` stays in the addon. It is outside the `kubevirt-csi-driver` namespace, so the
operator does not delete it. Changes:

- `--infra-cluster-namespace={{ .Config.CloudProvider.Kubevirt.InfraNamespace }}` and
  `--infra-cluster-labels=csi-driver/cluster=tenant` are passed as arguments. The `driver-config` ConfigMap is removed.
- New `csi-snapshotter` container (`--csi-address=/csi/csi.sock`, `--timeout=3m`).
- New `csi-resizer` container (`--csi-address=/csi/csi.sock`, `--timeout=3m`, `--handle-volume-inuse-error=false`).
- The ClusterRole `kubevirt-csi-controller-cr` gets `patch` on `persistentvolumeclaims`, and `update` and `patch` on
  `snapshot.storage.k8s.io/volumesnapshotcontents/status`.

The infra kubeconfig in `kubeone-ccm-credentials` needs `create`, `get` and `delete` on `volumesnapshots` in
`InfraNamespace` for snapshots to work. This only applies when `volumeSnapshotClasses` are configured, and is
documented.

The snapshot CRDs and the snapshot-controller are already deployed for KubeVirt by `csi-external-snapshotter`.

### Replacing the node plugin

The operator's DaemonSet `kubevirt-csi-driver/kubevirt-csi-node` and the old `kube-system/kubevirt-csi-node` use the
same host paths: both register `/var/lib/kubelet/plugins/csi.kubevirt.io/csi.sock`. They must not run at the same time
on a node.

The `csi-kubevirt` addon action gets a `supportFn`, which runs before the addon is applied:

1. Delete DaemonSet `kube-system/kubevirt-csi-node`. If it does not exist, skip to step 3.
2. Wait until no pod with label `app: kubevirt-csi-driver` and an owner of kind `DaemonSet` is left in `kube-system`.
   The controller pods have the same label, but are owned by a ReplicaSet.
3. Delete `kube-system/kubevirt-csi-node-sa`, ClusterRole `kubevirt-csi-node-cr`, ClusterRoleBinding
   `kubevirt-csi-node-binding` and ConfigMap `kube-system/driver-config`, ignoring NotFound. Prune would remove these
   too, but only when the addon inventory exists. Clusters last applied by a KubeOne version that used
   `kubectl apply` have none.

The operator creates its DaemonSet after the addon is applied. Between the deletion and the new pod being ready, a
node runs no node plugin. New volume attach, mount and unmount operations on that node wait. Mounted volumes keep
working. The images are the same as before, so the new pods usually start without pulling.

### StorageClass and VolumeSnapshotClass reconciliation

`addonAction` gets a new optional `postFn func() error`, which `Ensure` calls after `EnsureAddonByName` succeeds. The
deletions below must happen after the new `Tenant` is applied: the operator recreates a deleted StorageClass right
away from the `Tenant` it currently sees, and with the old spec it would recreate the old StorageClass.

The `csi-kubevirt` `postFn`:

1. Builds the desired StorageClasses from `cloudProvider.kubevirt.storageClasses`, with the same rules as the operator
   (`getDesiredStorageClass` in `controllers/tenant/storageclass.go`): name `kubevirt-<infraStorageClassName>`,
   provisioner `csi.kubevirt.io`, parameters `infraStorageClassName` and `bus`, `volumeBindingMode`,
   `allowVolumeExpansion`, `reclaimPolicy` (`Delete` when empty), `labels`, and `allowedTopologies` from `zones`
   (`topology.kubernetes.io/zone`) and `regions` (`topology.kubernetes.io/region`).
2. Lists the StorageClasses with a controller ownerReference to the `Tenant` named `tenant`.
3. Deletes each of them that is not desired, or whose `parameters`, `volumeBindingMode`, `reclaimPolicy`,
   `allowVolumeExpansion`, `labels` or `allowedTopologies` differ from the desired one. Every StorageClass is deleted
   before waiting, because the operator stops at the first one it cannot update. Mutable fields are also handled by
   deleting and recreating, so the operator remains the only writer of these StorageClasses.
4. Lists the VolumeSnapshotClasses with a controller ownerReference to the `Tenant`, and deletes those that are not
   desired.
5. Waits, for up to 5 minutes, until every desired StorageClass exists and matches. On a fresh install this includes
   pulling the operator image. On timeout it returns an error that tells the user to check the logs of
   `kubevirt-csi-driver/kubevirt-csi-driver-operator`.
6. Waits the same way, for up to 5 more minutes, until every desired VolumeSnapshotClass exists and has the
   `driver`, `deletionPolicy`, `parameters` and `snapshot.storage.kubernetes.io/is-default-class` annotation the
   operator sets. The operator reconciles VolumeSnapshotClasses after all StorageClasses. The VolumeSnapshotClasses
   are read as unstructured objects, because KubeOne has no typed client for the snapshot API.

The comparison ignores the `storageclass.kubernetes.io/is-default-class` annotation, which the operator updates
itself.

Effects:

- Existing PVs are not affected. A StorageClass is only read when a volume is provisioned or expanded.
- A new PVC of a StorageClass that is being recreated stays Pending until the StorageClass exists again. The
  provisioner retries it.
- StorageClasses that users created themselves are not touched, unless they are named `kubevirt-<name>` for a
  configured entry. The operator adopts those by setting its ownerReference, and the `postFn` may then recreate them.
  The release note mentions this.

### Images

| Name                        | Image | Change |
|-----------------------------|-------|--------|
| `KubeVirtCSIDriverOperator` | `quay.io/kubermatic/kubevirt-csi-driver-operator:v0.5.3` | new |
| `KubeVirtCSISnapshotter`    | `registry.k8s.io/sig-storage/csi-snapshotter:v8.6.0` | new |
| `KubeVirtCSIResizer`        | `registry.k8s.io/sig-storage/csi-resizer:v2.2.1` | new |
| `KubeVirtCSI`               | `quay.io/kubermatic/kubevirt-csi-driver:v0.4.5` | kept, used by the controller |
| `KubeVirtCSINodeDriverRegistrar`, `KubeVirtCSILivenessProbe` | unchanged | kept, see below |

The snapshotter and resizer versions match the ones KubeOne already uses for DigitalOcean and Nutanix.

The node plugin images are not rendered by KubeOne anymore, but `KubeVirtCSI`, `KubeVirtCSINodeDriverRegistrar` and
`KubeVirtCSILivenessProbe` stay in the `kubevirt` image list, so `kubeone config images` still lists everything a
mirror registry needs. A comment in `images.go` says that these three must match the images hard-coded in the
operator release, and must be bumped together with `KubeVirtCSIDriverOperator`.

### API changes

New optional fields in `KubevirtSpec`, added to the internal API (`pkg/apis/kubeone/types.go`), `v1beta2` and
`v1beta3`, followed by `make update-codegen`:

```yaml
cloudProvider:
  kubevirt:
    infraNamespace: tenant-a
    # Rendered into Tenant.spec.storageClasses.
    storageClasses:
      - infraStorageClassName: standard   # required
        isDefaultClass: true
        volumeBindingMode: WaitForFirstConsumer
        allowVolumeExpansion: true
        reclaimPolicy: Delete
        bus: scsi
        zones: []
        regions: []
        labels: {}
    # Rendered into Tenant.spec.volumeSnapshotClasses.
    volumeSnapshotClasses:
      - infraVolumeSnapshotClass: csi-snapclass   # required
        isDefaultClass: true
        deletionPolicy: Delete
```

Go types:

```go
type KubevirtSpec struct {
	// ...existing fields

	// StorageClasses to create in the cluster, each backed by a StorageClass of the infra cluster
	StorageClasses []KubevirtStorageClass `json:"storageClasses,omitempty"`

	// VolumeSnapshotClasses to create in the cluster, each backed by a VolumeSnapshotClass of the infra cluster
	VolumeSnapshotClasses []KubevirtVolumeSnapshotClass `json:"volumeSnapshotClasses,omitempty"`
}

type KubevirtStorageClass struct {
	InfraStorageClassName string                                `json:"infraStorageClassName"`
	IsDefaultClass        bool                                  `json:"isDefaultClass,omitempty"`
	VolumeBindingMode     storagev1.VolumeBindingMode           `json:"volumeBindingMode,omitempty"`
	AllowVolumeExpansion  bool                                  `json:"allowVolumeExpansion,omitempty"`
	ReclaimPolicy         corev1.PersistentVolumeReclaimPolicy  `json:"reclaimPolicy,omitempty"`
	Bus                   string                                `json:"bus,omitempty"`
	Zones                 []string                              `json:"zones,omitempty"`
	Regions               []string                              `json:"regions,omitempty"`
	Labels                map[string]string                     `json:"labels,omitempty"`
}

type KubevirtVolumeSnapshotClass struct {
	InfraVolumeSnapshotClass string `json:"infraVolumeSnapshotClass"`
	IsDefaultClass           bool   `json:"isDefaultClass,omitempty"`
	DeletionPolicy           string `json:"deletionPolicy,omitempty"`
}
```

Both lists are optional. When both are empty, the `Tenant` has empty lists, and no StorageClass or VolumeSnapshotClass
is created, as today.

The `Tenant` template omits empty fields, so the operator and the API server apply their defaults: `volumeBindingMode`
`Immediate`, `reclaimPolicy` `Delete`, `deletionPolicy` `Delete`. The `postFn` uses the same defaults when it compares.

Validation (`pkg/apis/kubeone/validation`):

- `infraStorageClassName` and `infraVolumeSnapshotClass` are required, valid DNS subdomain names after the `kubevirt-`
  prefix is added, and unique within their list.
- At most one entry of each list has `isDefaultClass: true`.
- `bus` is empty, `virtio`, `scsi` or `sata`.
- `volumeBindingMode` is empty, `Immediate` or `WaitForFirstConsumer`.
- `reclaimPolicy` is empty, `Delete` or `Retain`.
- `deletionPolicy` is empty, `Delete` or `Retain`.
- Setting either list together with `disableBundledCSIDrivers: true` is an error, because nothing would create the
  classes.

## Migration of existing clusters

`kubeone apply` on an existing KubeVirt cluster:

1. `supportFn`: deletes the old node plugin DaemonSet, waits for its pods, and deletes its RBAC and `driver-config`.
2. Applies the addon: CRD, operator, the controller with the new sidecars and arguments, and the `Tenant`. The
   controller rolls out with a new pod template.
3. The operator creates the `CSIDriver`, the new node plugin DaemonSet and the configured classes.
4. `postFn`: reconciles StorageClasses and VolumeSnapshotClasses, and waits for them.

Existing PVs keep working, because the driver name `csi.kubevirt.io` and the infra labels do not change. Before this
change no `CSIDriver` object existed, and kubelet used the defaults for a driver without one (`attachRequired: true`,
`podInfoOnMount: false`). The operator's `CSIDriver` has `attachRequired: true` and `podInfoOnMount: true`. With
`podInfoOnMount`, kubelet adds pod information to the volume context of `NodePublishVolume`. KKP has run the driver
with this `CSIDriver` for a long time, and the upgrade test below covers volumes that are already published.

StorageClasses that users created themselves keep working.

## Reset

`kubeone reset` needs no change. The `Tenant`, and the objects the operator owns, are removed together with the
cluster.

## Testing

- Unit tests for validation, in `validation_test.go`.
- Unit tests that render the addon for a KubeVirt cluster: with empty lists, with several StorageClasses and
  VolumeSnapshotClasses, and with `overwriteRegistry`. They check the `Tenant` spec and the operator arguments.
- Unit test for the `supportFn` with a fake client: the DaemonSet exists, the DaemonSet does not exist, and the old
  RBAC exists without the DaemonSet.
- Manual test on a KubeVirt cluster, as there is no KubeVirt e2e test in `test/tests.yml`:
  - a fresh install with two StorageClasses and a VolumeSnapshotClass, then create a PVC, expand it and snapshot it;
  - an upgrade from the previous release with a running pod that has a PVC attached;
  - changing `bus` and `allowVolumeExpansion` of a configured StorageClass.

## Tasks & Efforts

Each step is a separate git commit.

1. API fields, validation and codegen.
2. `postFn` in `addonAction` and `Ensure`.
3. New images, the rewritten `csi-kubevirt` addon (CRD, operator, controller, `Tenant`), and the `supportFn` that
   replaces the node plugin.
4. StorageClass and VolumeSnapshotClass reconciliation in the `postFn`.
5. Docs: the new fields, the infra permissions needed for snapshots, and the release note.
