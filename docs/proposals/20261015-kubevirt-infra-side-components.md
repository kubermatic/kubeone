# KubeVirt: run infra-facing components in the infra cluster

**Status:** **Draft**
**Created:** 2026-10-15
**Last updated:** 2026-10-15
**Author:** Artiom Diomin ([@kron4eg](https://github.com/kron4eg))

## Abstract

On the KubeVirt cloud provider, the KubeOne cluster (the *tenant* cluster) runs inside an existing Kubernetes cluster
with KubeVirt installed (the *infra* cluster). Today, every component that has to talk to the infra cluster runs inside
the tenant cluster and receives the infra kubeconfig (`KUBEVIRT_KUBECONFIG`) through Secrets in the tenant's
`kube-system`:

| Component               | Tenant Secret holding the infra kubeconfig     |
|-------------------------|------------------------------------------------|
| KubeVirt CCM            | `kubeone-ccm-credentials`                      |
| KubeVirt CSI controller | `kubeone-ccm-credentials`                      |
| machine-controller      | `kubeone-machine-controller-credentials`       |
| operating-system-manager| `kubeone-operating-system-manager-credentials` |

Anyone with read access to Secrets in the tenant's `kube-system` gets the user-provided infra kubeconfig. That
kubeconfig often has wide permissions on the infra cluster.

This proposal moves every infra-facing component into the infra cluster's `InfraNamespace`. The tenant cluster keeps
only identities that grant access to the tenant cluster itself. It also replaces the existing `csi-kubevirt` addon with
the new `kubevirt-csi-driver-operator` addon, based on
[kubermatic/kubevirt-csi-driver-operator](https://github.com/kubermatic/kubevirt-csi-driver-operator).

This follows the topology KKP already uses, with the infra cluster's `InfraNamespace` in the role of the KKP seed.

## Goals

- No infra cluster credentials are stored in the tenant cluster.
- Components that run in the infra cluster use least-privilege ServiceAccounts scoped to `InfraNamespace`, not the
  user-provided `KUBEVIRT_KUBECONFIG`.
- Ship the KubeVirt CSI driver through `kubevirt-csi-driver-operator`, and remove `csi-kubevirt`.
- Existing KubeVirt clusters migrate on `kubeone apply`, without downtime of provisioning, storage or load balancing.
- Several KubeOne clusters can share one `InfraNamespace`. Nothing prevents this today, and the current `csi-kubevirt`
  addon already breaks in that setup: its fixed `--infra-cluster-labels=csi-driver/cluster=tenant` cannot tell
  DataVolumes of different clusters apart.

## Non-goals

- Changes to providers other than KubeVirt.
- Removing `KUBEVIRT_KUBECONFIG` from KubeOne itself. KubeOne still uses it locally to create control plane VMs, the API
  server Service, and the infra-side objects described below.
- Rotating ServiceAccount tokens on a schedule.

## Background

### kubevirt-csi-driver-operator

The operator (latest release `v0.5.3`, image `quay.io/kubermatic/kubevirt-csi-driver-operator`) runs in the tenant
cluster. It only talks to the tenant API server and never needs infra credentials. It reconciles a cluster-scoped
`Tenant` (`csiprovisioner.kubevirt.io/v1alpha1`) that must be named `tenant`, and creates the following:

- the `CSIDriver` object `csi.kubevirt.io`;
- the node plugin DaemonSet `kubevirt-csi-node` in the hard-coded namespace `kubevirt-csi-driver`, with its RBAC;
- one StorageClass `kubevirt-<infraStorageClassName>` per `spec.storageClasses[]` entry;
- one VolumeSnapshotClass per `spec.volumeSnapshotClasses[]` entry.

It also copies zone and region labels from the selected node onto the `nodeAffinity` of bound PVs.

The operator does **not** deploy the CSI controller (provisioner/attacher). It also deletes any Deployment
`kubevirt-csi-driver/kubevirt-csi-controller`, which is a leftover from older versions. The integrator has to run the
controller. The driver supports this with `--tenant-cluster-kubeconfig` and `--infra-cluster-kubeconfig`, and the CSI
sidecars support `--kubeconfig`.

The upstream kustomize output is incomplete:
- it does not create the `kubevirt-csi-driver` namespace;
- its ClusterRole lacks `volumesnapshotclasses`.

The KKP addon `addons/csi/kubevirt/csi-driver-operator.yaml` fixes both and is the reference for the tenant-side
manifests.

### machine-controller

Both `machine-controller` and its webhook accept `-kubeconfig`, so they can run outside the cluster they manage. The
KubeVirt provider reads the infra kubeconfig from the `auth.kubeconfig` field, or from the `KUBEVIRT_KUBECONFIG`
environment variable. It has no in-cluster fallback.

The webhook needs infra access too. `Validate` performs a VMI `GET` against the infra cluster.

### Applying objects to the infra cluster

The addon pipeline applies manifests with `kubectl apply` over SSH on the leader node, so it can only target the tenant
cluster. KubeOne already builds a controller-runtime client for the infra cluster (`kubevirtInfraClient` in
`pkg/cloudprovider/kubevirt/provider.go`) and uses it for the API server Service.

## Design

### Topology

```
infra cluster, InfraNamespace                      tenant cluster
─────────────────────────────────────────────      ─────────────────────────────────────
<cluster>-machine-controller         ──kubeconfig──▶ Machines, MachineDeployments, Nodes
<cluster>-machine-controller-webhook ◀──url webhook── kube-apiserver
<cluster>-kubevirt-ccm               ──kubeconfig──▶ Nodes, Services
<cluster>-kubevirt-csi-controller    ──kubeconfig──▶ PVs, PVCs, VolumeAttachments
        │ infra SA (in-cluster / kubeconfig)          kubevirt-csi-driver-operator
        ▼                                             kubevirt-csi-node (DaemonSet)
VMs, VMIs, DataVolumes, Services                      ServiceAccounts + RBAC only
```

### Identities

Each component moved to the infra cluster gets two credentials.

**Access to the infra cluster.**
- KubeOne creates, in `InfraNamespace`:
  - a ServiceAccount `<cluster>-<component>`;
  - a Role and RoleBinding with only the permissions that component needs.
- The CSI controller and the CCM use this ServiceAccount directly through in-cluster configuration: the CSI driver falls
  back to it when `--infra-cluster-kubeconfig` is empty, the CCM when the `kubeconfig` key of its `cloud-config` is
  empty.
- machine-controller needs a kubeconfig: its KubeVirt provider has no in-cluster fallback. For it, KubeOne also:
  1. creates a `kubernetes.io/service-account-token` Secret;
  2. generates a kubeconfig from it, with server
     `https://kubernetes.default.svc` and the token's `ca.crt` as the CA;
  3. stores the kubeconfig in a Secret in the same namespace.
- This kubeconfig only works from inside the infra cluster.

**Access to the tenant cluster.**
- The component's tenant ServiceAccount stays in tenant `kube-system`: `machine-controller`, `cloud-controller-manager`,
  and the new `kubevirt-csi-controller`.
- It keeps its existing ClusterRoles. A `kubernetes.io/service-account-token` Secret is added for it.
- After the tenant manifests are applied, KubeOne:
  1. waits for the token to be populated;
  2. builds a kubeconfig from the token, the cluster CA and `APIEndpoint`;
  3. stores it in a Secret in `InfraNamespace`.
- `APIEndpoint` is already in the API server certificate. It resolves from inside the infra cluster, for example
  `<lb-name>.<InfraNamespace>.svc.cluster.local` for a ClusterIP API server Service.

All infra-side objects are named with a `<cluster>-` prefix and labelled `kubeone_cluster_name=<cluster>`. They are
built as typed Go objects in a new `pkg/cloudprovider/kubevirt/infra` package and applied with the infra client. They do
not go through the addon pipeline.

### CSI: `kubevirt-csi-driver-operator`

**Tenant side** (embedded addon `addons/kubevirt-csi-driver-operator/`):
- Namespace `kubevirt-csi-driver`.
- The `tenants.csiprovisioner.kubevirt.io` CRD.
- The operator's ServiceAccount, ClusterRole and ClusterRoleBinding. The ClusterRole is upstream's plus
  `snapshot.storage.k8s.io/volumesnapshotclasses`.
- The operator Deployment, with `--leader-elect`, and with `--overwrite-registry` when
  `registryConfiguration.overwriteRegistry` is set. The `Tenant.spec.imageRepository` and `imageTag` fields are ignored
  by the operator and are not used.
- ServiceAccount `kubevirt-csi-controller` in `kube-system`, with:
  - a ClusterRole based on today's `kubevirt-csi-controller-cr`, extended with the permissions the resizer and
    snapshotter need;
  - a token Secret.

The `Tenant` CR named `tenant`, rendered from the new API fields (see [API changes](#api-changes)), is not part of the
addon: a single `kubectl apply` cannot create a CRD and an object of that CRD before the CRD is established. KubeOne
applies it from Go after the addon, once the CRD is established, together with the infra side.

Snapshot CRDs and the snapshot controller are already deployed by `csi-external-snapshotter` for KubeVirt and are not
duplicated.

**Infra side:** Deployment `<cluster>-kubevirt-csi-controller`.

| Container    | Image                         | Notes |
|--------------|-------------------------------|-------|
| csi-driver   | `KubeVirtCSI` (`v0.4.5`)      | `--run-controller-service=true`, `--run-node-service=false`, `--tenant-cluster-kubeconfig`, no `--infra-cluster-kubeconfig` (in-cluster ServiceAccount), `--infra-cluster-namespace=<InfraNamespace>`, `--infra-cluster-labels=kubeone_cluster_name=<cluster>` |
| provisioner  | `KubeVirtCSIProvisioner`      | `--kubeconfig=<tenant>` |
| attacher     | `KubeVirtCSIAttacher`         | `--kubeconfig=<tenant>` |
| snapshotter  | `KubeVirtCSISnapshotter` (new)| `--kubeconfig=<tenant>` |
| resizer      | `KubeVirtCSIResizer` (new)    | `--kubeconfig=<tenant>`, `--handle-volume-inuse-error=false` |
| livenessprobe| `KubeVirtCSILivenessProbe`    | |

The driver version matches the one the operator deploys in its node DaemonSet.

Infra Role:

```yaml
- apiGroups: [cdi.kubevirt.io]
  resources: [datavolumes]
  verbs: [get, create, delete]
- apiGroups: [kubevirt.io]
  resources: [virtualmachines, virtualmachineinstances]
  verbs: [get, list]
- apiGroups: [subresources.kubevirt.io]
  resources:
    - virtualmachines/addvolume
    - virtualmachines/removevolume
    - virtualmachineinstances/addvolume
    - virtualmachineinstances/removevolume
  verbs: [update]
- apiGroups: [snapshot.storage.k8s.io]
  resources: [volumesnapshots]
  verbs: [get, create, delete]
- apiGroups: [""]
  resources: [persistentvolumeclaims]
  verbs: [get, list, watch, update, patch]
```

**Removing `csi-kubevirt`.**
- Delete `addons/csi-kubevirt/`.
- Remove `AddonCSIKubeVirt` from `CloudAddons()` and `embeddedAddons`.
- In the KubeVirt case of `ensureCSIAddons`, the new addon replaces it. Its `supportFn` deletes the objects of the old
  addon by name (`DeleteAddonByName` needs the old manifests, which are removed), so two node plugins never register
  `csi.kubevirt.io` at the same time. This includes the tenant `driver-config` ConfigMap. No `CSIDriver` object existed
  before, the operator creates it.
- A new `postFn` of the addon action, run after the addon is applied, creates the `Tenant` CR and the infra side.
- The addon is deployed only when `cloudProvider.kubevirt` is set and `cloudProvider.disableBundledCSIDrivers` is false.

### CCM

This follows the upstream `config/isolated` overlay of `cloud-provider-kubevirt`.

**Tenant side** (`addons/ccm-kubevirt`, now hand-maintained; the kustomize generation is removed):
- The `cloud-controller-manager` ServiceAccount and a token Secret.
- ClusterRole `kccm`: nodes, `nodes/status`, services and `services/status`, events, endpointslices (read), leases.
  The previous ClusterRole only covered nodes and leases.
- RoleBinding to `extension-apiserver-authentication-reader` and ClusterRoleBinding to `system:auth-delegator`, for
  delegated authentication of the CCM secure port.
- Removed: the Deployment, the `cloud-config` ConfigMap, the `kubeone-ccm-credentials` mount, and the `kccm` Role, which
  granted KubeVirt permissions in the tenant cluster and had no effect.

**Infra side:**
- Deployment `<cluster>-kubevirt-ccm` with `--kubeconfig`, `--authentication-kubeconfig` and
  `--authorization-kubeconfig` all pointing at the tenant. Leader election stays on the tenant lease. It no longer runs
  privileged. A hash of the `cloud-config` on the pod template rolls the Deployment when the config changes.
- ConfigMap `<cluster>-kubevirt-ccm-cloud-config` containing:
  - `namespace: <InfraNamespace>`;
  - `instancesV2.enabled: true` and `instancesV2.zoneAndRegionEnabled`;
  - `loadBalancer.enabled`, always explicit, because the CCM defaults it to true;
  - `infraLabels: {kubeone_cluster_name: <cluster>}`, set on the LoadBalancer Services the CCM creates. They do not get
    `kubeone_component`, so the infra cleanup leaves them alone;
  - no `kubeconfig`, so the CCM uses in-cluster configuration.
- The infra Role follows upstream `isolated`: VMs read, VMIs read and update, pods read, and create, read, update and
  delete on `services` only when `loadBalancerEnabled` is set. The EPS controller is disabled by default, so no
  `endpointslices` permissions are needed in the infra cluster.

**Migration:** a `postFn` deploys the infra side, then deletes the old tenant Deployment, `cloud-config` ConfigMap,
`kccm` Role and `kccm-sa` RoleBinding. `kubectl apply --prune` removes the Deployment and ConfigMap already, when the
addon is applied, so the CCM is down for a short time between the addon apply and the `postFn`. Node initialization and
LoadBalancer updates are delayed by that much.

The CCM no longer depends on a schedulable tenant control-plane node. This removes the bootstrap dependency on the
`node.cloudprovider.kubernetes.io/uninitialized` toleration.

### machine-controller

**Tenant side** (`addons/machinecontroller`, KubeVirt branch):
- Keep the CRDs, the ServiceAccounts and the RBAC, and add token Secrets (`addons/machinecontroller/kubevirt-tokens.yaml`).
- Skip both Deployments, the webhook Service, the serving certificate Secret, the PodDisruptionBudgets and the
  `MutatingWebhookConfiguration`. Other providers are unchanged.
- The `MutatingWebhookConfiguration` is created from Go instead, with `clientConfig.url` pointing at the infra-side
  webhook Service: its address is only known once that Service exists.

**Infra side** (`pkg/cloudprovider/kubevirt/machinecontroller`):
- Two identities, `machine-controller` and `machine-controller-webhook`, each with an infra ServiceAccount, Role, token,
  infra kubeconfig Secret and tenant kubeconfig Secret.
- Deployment `<cluster>-machine-controller` with `-kubeconfig=<tenant>`. It reads `KUBEVIRT_KUBECONFIG` from the Secret
  holding the infra SA kubeconfig, and `POD_NAMESPACE` is the `InfraNamespace`. The other flags stay as today, including
  the `JOIN_TIMEOUT` addon param. Leader election stays on the tenant lease.
- Deployment `<cluster>-machine-controller-webhook` with `-kubeconfig=<tenant>` and the same environment.
- The tenant proxy settings (`HTTPS_PROXY`, `NO_PROXY`) are not passed: they are for the tenant network, and these pods
  run in the infra cluster.
- A custom CA bundle is put into a per-cluster infra ConfigMap `<cluster>-machine-controller-ca-bundle` and mounted the
  same way as in the tenant cluster.
- Service `<cluster>-machine-controller-webhook`. Its type is configurable (see [API changes](#api-changes)).
- The webhook serving certificate is signed by the tenant cluster CA, which KubeOne has locally, for the Service IP or
  LoadBalancer hostname. It is stored in `<cluster>-machine-controller-webhook-serving-cert` and reused while it is valid
  for at least 30 more days, matches the address and is signed by the current CA, so the webhook does not restart on
  every apply. The `caBundle` of the `MutatingWebhookConfiguration` stays the tenant cluster CA.
- The machine-controller infra Role covers VirtualMachines (read and write), VirtualMachineInstances (read),
  DataVolumes (read, create, delete), `datavolumes/source` (create, for cloning a PVC in the same namespace), Secrets
  (cloud-init user data) and namespaced instancetypes and preferences (read). The webhook Role is read-only.
- **Both also need a read-only ClusterRole in the infra cluster.** The KubeVirt provider reads the primary disk
  StorageClass on every machine create and in the webhook defaulting, the CDI StorageProfile when no access mode is set,
  and cluster instancetypes and preferences. These are cluster-scoped, so a Role cannot grant them. The ClusterRole and
  ClusterRoleBinding are named `kubeone:<InfraNamespace>:<cluster>-<component>` and labelled with
  `kubeone_infra_namespace`, so the cleanup selects them per namespace. The user-provided `KUBEVIRT_KUBECONFIG` therefore
  needs permission to create ClusterRoles and ClusterRoleBindings, and has to hold these read permissions itself.
- Cloning a PVC from another namespace for the VM disk needs `datavolumes/source` permission in that namespace, which
  KubeOne does not grant: the infra cluster admin has to.

`machinecontroller.WaitReady` waits for the infra Deployments when the provider is KubeVirt.

**Webhook reachability.** The tenant kube-apiserver must reach the webhook Service in the infra cluster.
- When the control plane runs on KubeVirt (`cloudProvider.kubevirt.controlPlane` is set), the control-plane VMs use
  masquerade pod networking and can reach infra ClusterIPs. A ClusterIP Service, addressed by IP, works. This is the
  default then.
- Otherwise the Service must be of type `LoadBalancer`, the default then.

**Migration:** a `postFn` deploys the infra side, then deletes the old tenant Deployments, Service, serving certificate
Secret and PodDisruptionBudgets. `kubectl apply --prune` removes all of them except the PodDisruptionBudgets when the
addon is applied already, so machine-controller and its webhook are down between the addon apply and the `postFn`. The
old service-based `MutatingWebhookConfiguration` stays in place during that time (it is not pruned), so Machine and
MachineDeployment changes are rejected until the `postFn` replaces it.

### operating-system-manager

OSM stays in the tenant cluster. It does not use the KubeVirt credentials, so the OSM credentials Secret is no longer
created for KubeVirt.

### Credentials Secrets

In `pkg/credentials/secret.go`, for KubeVirt, stop creating:
- `kubeone-ccm-credentials`;
- `kubeone-machine-controller-credentials`;
- `kubeone-operating-system-manager-credentials`.

`MachineControllerCredentialsEnvVars` and `MachineControllerCredentialsHash` are no longer rendered for KubeVirt.

## API changes

New optional fields in `KubevirtSpec`, added to the internal API (`pkg/apis/kubeone/types.go`), `v1beta3` and `v1beta2`,
followed by `make update-codegen`:

```yaml
cloudProvider:
  kubevirt:
    infraNamespace: tenant-a
    # Rendered into the Tenant CR of kubevirt-csi-driver-operator.
    storageClasses:
      - infraStorageClassName: standard
        isDefaultClass: true
        volumeBindingMode: WaitForFirstConsumer
        allowVolumeExpansion: true
        reclaimPolicy: Delete
        bus: scsi
        zones: []
        regions: []
        labels: {}
    volumeSnapshotClasses:
      - infraVolumeSnapshotClass: csi-snapclass
        isDefaultClass: true
        deletionPolicy: Delete
    machineController:
      webhook:
        # ClusterIP or LoadBalancer.
        # Default: ClusterIP when controlPlane is set, LoadBalancer otherwise.
        serviceType: ClusterIP
        annotations: {}
```

## Migration of existing clusters

`kubeone apply` on an existing KubeVirt cluster runs these steps in order:

1. Reconcile the infra-side identities and Secrets.
2. Create the infra-side Deployments. They use the same tenant leader-election leases as the in-tenant pods, so only one
   replica of each controller is active while both are running.
3. Wait for the infra-side Deployments to become ready.
4. Switch the machine-controller `MutatingWebhookConfiguration` to the URL form.
5. Delete the tenant-side Deployments of machine-controller, its webhook and the CCM.
6. Delete the `csi-kubevirt` addon and apply `kubevirt-csi-driver-operator`.
7. Delete the three credentials Secrets from tenant `kube-system`.

Existing PVs keep working: the driver name `csi.kubevirt.io` is unchanged. Existing user-created StorageClasses that use
`csi.kubevirt.io` are left untouched.

## Reset

`kubeone reset` destroys the workers through machine-controller, which now runs in the infra cluster. Infra-side objects
are therefore deleted only after the workers are gone. They are deleted by the label `kubeone_cluster_name=<cluster>` in
`InfraNamespace`, alongside `CleanupLoadBalancer`.

## Security considerations

- The tenant cluster holds no infra credentials. Tenant-side tokens grant only the existing ClusterRoles of each
  component in the tenant cluster.
- Infra-side ServiceAccounts are scoped to `InfraNamespace` with Roles. The user-provided `KUBEVIRT_KUBECONFIG` never
  leaves the machine that runs KubeOne.
- The user-provided `KUBEVIRT_KUBECONFIG` must allow creating ServiceAccounts, Roles, RoleBindings, Secrets, ConfigMaps,
  Deployments and Services in `InfraNamespace`, and ClusterRoles and ClusterRoleBindings for machine-controller's
  read-only access to StorageClasses, StorageProfiles and cluster instancetypes. This becomes a documented requirement.
- Tokens are long-lived `service-account-token` Secrets. The legacy token cleanup controller does not remove manually
  created token Secrets. Every `kubeone apply` recreates missing tokens and refreshes the kubeconfig Secrets.

## Images

New entries in `pkg/templates/images/images.go`:

| Name                        | Image |
|-----------------------------|-------|
| `KubeVirtCSIDriverOperator` | `quay.io/kubermatic/kubevirt-csi-driver-operator:v0.5.3` |
| `KubeVirtCSISnapshotter`    | `registry.k8s.io/sig-storage/csi-snapshotter:v8.7.0` |
| `KubeVirtCSIResizer`        | `registry.k8s.io/sig-storage/csi-resizer:v2.3.0` |

## Tasks & Efforts

Each step is a separate git commit.

1. Infra identity and kubeconfig helpers, plus the infra cleanup on reset, with unit tests.
2. `kubevirt-csi-driver-operator` addon, the infra-side CSI controller, the API fields and codegen, removal of
   `csi-kubevirt`, and golden fixtures.
3. Move the CCM to the infra cluster.
4. Move machine-controller and its webhook to the infra cluster, including the readiness checks and the webhook Service
   type.
5. Stop creating the tenant credentials Secrets, and add the migration steps. Run the KubeVirt e2e tests
   (`test/tests.yml`) for a fresh install and for an upgrade from the previous release.
