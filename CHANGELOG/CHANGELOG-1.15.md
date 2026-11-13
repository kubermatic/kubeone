# [v1.15.0](https://github.com/kubermatic/kubeone/releases/tag/v1.15.0) - 2026-11-13

## Changelog since v1.14.0

## Urgent and BREAKING Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Support for Kubernetes **1.34** has been removed. KubeOne v1.15 supports Kubernetes versions **1.35, 1.36, and 1.37**. Before upgrading KubeOne, ensure your clusters are running Kubernetes v1.35 or newer. ([#4249](https://github.com/kubermatic/kubeone/pull/4249), [@kron4eg](https://github.com/kron4eg))
- `kubeone apply` and `kubeone local` no longer create the PKI backup `<cluster-name>.tar.gz` by default. The archive contains sensitive credentials (e.g. the cluster CA key) and is now only created when `--backup <path>` is set explicitly. ([#4304](https://github.com/kubermatic/kubeone/pull/4304), [@steled](https://github.com/steled))
- unattended-upgrades addon: the Flatcar Linux Update Operator is now deployed only if the cluster has at least one Flatcar Linux node (control plane, static workers, dynamic workers or MachineDeployments in the cluster). On existing clusters without Flatcar nodes, the next `kubeone apply` prunes the `reboot-coordinator` namespace and its workloads, but the cluster-scoped `flatcar-linux-update-agent` and `flatcar-linux-update-operator` ClusterRoles and ClusterRoleBindings are not removed automatically and can be deleted manually. ([#4264](https://github.com/kubermatic/kubeone/pull/4264), [@steled](https://github.com/steled))
- KubeVirt: the `csi-kubevirt` addon now deploys the KubeVirt CSI Driver operator. The next `kubeone apply` on an existing KubeVirt cluster replaces the `kube-system` node plugin with the operator's DaemonSet. StorageClasses and VolumeSnapshotClasses are now configured via `cloudProvider.kubevirt.storageClasses` and `cloudProvider.kubevirt.volumeSnapshotClasses`. ([#4307](https://github.com/kubermatic/kubeone/pull/4307), [@kron4eg](https://github.com/kron4eg))
- Addons are now applied with server-side apply instead of `kubectl apply`. KubeOne takes over ownership of fields previously set by client-side apply, and prunes objects of the applied kinds that were removed from an addon. Changes made to addon-managed fields outside of KubeOne will be overwritten on the next apply. ([#4272](https://github.com/kubermatic/kubeone/pull/4272), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### API Change

- Add `clusterNetwork.clusterDNS` to override the DNS servers that kubelets configure for pods on all nodes. It can't be combined with `features.nodeLocalDNS.deploy`. ([#4251](https://github.com/kubermatic/kubeone/pull/4251), [@kron4eg](https://github.com/kron4eg))
- Add structured AuthenticationConfiguration support via `features.authenticationConfiguration` (kube-apiserver `--authentication-config`). It is mutually exclusive with `features.openidConnect`. ([#4267](https://github.com/kubermatic/kubeone/pull/4267), [@kron4eg](https://github.com/kron4eg))
- Add `journald.systemMaxUse` to configure journald's `SystemMaxUse` on all nodes (default `5G`). ([#4278](https://github.com/kubermatic/kubeone/pull/4278), [@rajaSahil](https://github.com/rajaSahil))

### Feature

- Add support for Kubernetes v1.37 ([#4249](https://github.com/kubermatic/kubeone/pull/4249), [@kron4eg](https://github.com/kron4eg))
- Managed control plane: add the AWS control-plane provider ([#4204](https://github.com/kubermatic/kubeone/pull/4204), [@kron4eg](https://github.com/kron4eg))
- Managed control plane: add the Azure control-plane provider ([#4185](https://github.com/kubermatic/kubeone/pull/4185), [@kron4eg](https://github.com/kron4eg))
- Managed control plane: add support for DigitalOcean ([#4262](https://github.com/kubermatic/kubeone/pull/4262), [@rajaSahil](https://github.com/rajaSahil))
- Managed control plane: add support for GCP ([#4269](https://github.com/kubermatic/kubeone/pull/4269), [@rajaSahil](https://github.com/rajaSahil))
- Add a `--destroy-control-plane` flag to `kubeone reset` to destroy control plane machines at the cloud provider instead of resetting them ([#4210](https://github.com/kubermatic/kubeone/pull/4210), [@kron4eg](https://github.com/kron4eg))
- Add a `--destroy-control-plane-load-balancer` flag to `kubeone reset` ([#4215](https://github.com/kubermatic/kubeone/pull/4215), [@kron4eg](https://github.com/kron4eg))
- Add `kubeone addons apply` command to install, update and remove user addons and helm releases without running a full `kubeone apply`. The `--name` flag limits it to specific addons or helm releases. ([#4263](https://github.com/kubermatic/kubeone/pull/4263), [@steled](https://github.com/steled))
- `kubeone addons list` shows the source (embedded, overridden, custom) of each addon ([#4305](https://github.com/kubermatic/kubeone/pull/4305), [@steled](https://github.com/steled))
- Add `--provider` flag to `kubeone config images list` and `kubeone mirror-images` to include CCM/CSI images only for the given cloud provider(s) ([#4112](https://github.com/kubermatic/kubeone/pull/4112), [@kron4eg](https://github.com/kron4eg))
- Added support for an `addonParams` section in the credentials file, allowing addon secret parameters (e.g. the built-in `backups-restic` addon's `resticPassword`) to be provided via `credentials.yaml` instead of the KubeOneCluster manifest ([#4206](https://github.com/kubermatic/kubeone/pull/4206), [@steled](https://github.com/steled))
- Parametrize schedule and flags in the `backups-restic` addon ([#4207](https://github.com/kubermatic/kubeone/pull/4207), [@kron4eg](https://github.com/kron4eg))
- Added support for deploying OpenStack clusters with dual-stack networking ([#4259](https://github.com/kubermatic/kubeone/pull/4259), [@4ch3los](https://github.com/4ch3los))
- Static worker upgrade retries now retry only the failed worker instead of repeating upgrades on workers that already succeeded ([#4192](https://github.com/kubermatic/kubeone/pull/4192), [@taaahaaa](https://github.com/taaahaaa))
- The Hubble certgen job of the `cni-cilium` addon now fails when the Cilium CA (`kube-system/cilium-ca`) would expire before the newly issued certificates, instead of succeeding until the CA has expired. Rotate the CA when the `hubble-generate-certs` job fails. ([#4232](https://github.com/kubermatic/kubeone/pull/4232), [@steled](https://github.com/steled))
- Async dashboard loading ([#4165](https://github.com/kubermatic/kubeone/pull/4165), [@kron4eg](https://github.com/kron4eg))

### Bug or Regression

- Fix worker nodes provisioned by operating-system-manager getting a non-existent DNS server (`10.10.10.10`) when NodeLocalDNS is disabled, including clusters using Cilium Local Redirect Policy. With Cilium Local Redirect Policy, kubelets now use the node-local DNS address with CoreDNS as fallback. The CoreDNS IP is now computed from the IPv6 service subnet on IPv6-primary clusters. ([#4251](https://github.com/kubermatic/kubeone/pull/4251), [@kron4eg](https://github.com/kron4eg))
- Fix the Cilium Local Redirect Policy ([#4189](https://github.com/kubermatic/kubeone/pull/4189), [@kron4eg](https://github.com/kron4eg))
- Add lease permissions for Cilium and enable `l2announcements` in helm values ([#4257](https://github.com/kubermatic/kubeone/pull/4257), [@kron4eg](https://github.com/kron4eg))
- Fix Cilium `tlsinterception-secrets` Roles/RoleBindings being deployed to `kube-system` instead of `cilium-secrets` ([#4276](https://github.com/kubermatic/kubeone/pull/4276), [@kron4eg](https://github.com/kron4eg))
- Fixed a bug where upgrading a vSphere cluster from an older KubeOne version left behind a stale `validation.csi.vsphere.vmware.com` ValidatingWebhookConfiguration, blocking StorageClass changes and all new dynamic volume provisioning/PVC deletion ([#4195](https://github.com/kubermatic/kubeone/pull/4195), [@steled](https://github.com/steled))
- Migrate stale VolumeGroupSnapshot CRDs ([#4299](https://github.com/kubermatic/kubeone/pull/4299), [@kron4eg](https://github.com/kron4eg))
- Fix provisioning with cloud provider `none` failing on CSI webhook certificate generation when no CSI addon is deployed ([#4234](https://github.com/kubermatic/kubeone/pull/4234), [@mvanhorn](https://github.com/mvanhorn))
- Fix the `ccm-hetzner` addon PodDisruptionBudget selector so it matches the hcloud-cloud-controller-manager pods ([#4280](https://github.com/kubermatic/kubeone/pull/4280), [@damienbenon](https://github.com/damienbenon))
- Fix PodDisruptionBudget selectors that match no pods in addons ([#4281](https://github.com/kubermatic/kubeone/pull/4281), [@rajaSahil](https://github.com/rajaSahil))
- Fix KubeVirt CSI reconcile and add missing RBAC to the KubeVirt CCM. The KubeVirt StorageClass bus now defaults to `scsi`. ([#4309](https://github.com/kubermatic/kubeone/pull/4309), [@kron4eg](https://github.com/kron4eg))
- Add subnet ID retrieval for OpenStack load balancer members ([#4242](https://github.com/kubermatic/kubeone/pull/4242), [@kron4eg](https://github.com/kron4eg))
- Fix KubeOne retrying PodDisruptionBudget-blocked pod evictions without a delay during node drain ([#4181](https://github.com/kubermatic/kubeone/pull/4181), [@taaahaaa](https://github.com/taaahaaa))
- Errors updating the kubeadm-config, kubelet-config or kube-proxy ConfigMaps are no longer ignored ([#4308](https://github.com/kubermatic/kubeone/pull/4308), [@sarankumar1325](https://github.com/sarankumar1325))
- Fix `overridePath` handling in `ContainerRuntimeConfig.MachineControllerFlags` ([#4174](https://github.com/kubermatic/kubeone/pull/4174), [@kron4eg](https://github.com/kron4eg))
- Update Helm dry run strategy on release upgrade ([#4168](https://github.com/kubermatic/kubeone/pull/4168), [@kron4eg](https://github.com/kron4eg))
- Fix OCI registry auth for helm charts and digest-pinned image retagging ([#4221](https://github.com/kubermatic/kubeone/pull/4221), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
  - Fixed OCI Helm chart authentication so `HelmRelease.Auth` credentials are correctly passed to private OCI registries.
  - Fixed `mirror-images` failing to retag digest-pinned images by preserving the source digest instead of incorrectly using it as a tag.
- Update flatcar-linux-update-operator to v0.10.0 to fix the update agent crashing on Flatcar >= 4757 due to quoted `/etc/os-release` values ([#4229](https://github.com/kubermatic/kubeone/pull/4229), [@steled](https://github.com/steled))

### Documentation

- Fix generated apidocs links to upstream objects to always be up to date ([#4266](https://github.com/kubermatic/kubeone/pull/4266), [@archups](https://github.com/archups))

### Chore

- Update system addons to support Kubernetes 1.37 ([#4260](https://github.com/kubermatic/kubeone/pull/4260), [@kron4eg](https://github.com/kron4eg))
  - Update Hetzner CCM to v1.39.2
  - Update Hetzner CSI Driver to v2.24.1
  - Update OpenStack CSI driver to v1.37.1
  - Update OpenStack Cloud Controller Manager to v1.37.1
  - Update AWS EBS CSI driver to v2.67.0
  - Update external CSI snapshot controller to v8.7.0
  - Update AzureFile CSI driver to v1.35.10
  - Update AzureDisk CSI driver to v1.36.1
  - Update Cilium to v1.20.2
  - Update vSphere CSI driver to v3.8.0
  - Update vSphere CPI to v1.36.0
  - Update Nutanix CSI Storage to v3.9.0
  - Update Nutanix Cloud Provider to v0.6.3
  - Update GCP Cloud Controller Manager to v37.0.1
  - Update GCP Compute Persistent Disk CSI driver to v1.27.1
  - Update DigitalOcean CSI driver to v4.19.0
  - Update DigitalOcean CCM to v0.1.71
  - Update Canal to v3.33.0
  - Update DNSNodeCache to v1.27.1
  - Update metrics-server to v0.9.0
- Update Cilium to v1.19.7 ([#4209](https://github.com/kubermatic/kubeone/pull/4209), [@archups](https://github.com/archups))
- Update Cilium to v1.19.6 ([#4163](https://github.com/kubermatic/kubeone/pull/4163), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to [v1.66.1](https://github.com/kubermatic/machine-controller/releases/tag/v1.66.1) ([#4157](https://github.com/kubermatic/kubeone/pull/4157), [@archups](https://github.com/archups))
- Update OSM to [v1.11.1](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.11.1) and Go to v1.26.5 ([#4180](https://github.com/kubermatic/kubeone/pull/4180), [@archups](https://github.com/archups))
- Upgrade Go to 1.27 ([#4208](https://github.com/kubermatic/kubeone/pull/4208), [@kron4eg](https://github.com/kron4eg))
- Upgrade golang.org/x/crypto to v0.56.0 ([#4217](https://github.com/kubermatic/kubeone/pull/4217), [@kron4eg](https://github.com/kron4eg))
- Update golang.org/x/net to v0.60.0 ([#4252](https://github.com/kubermatic/kubeone/pull/4252), [@kron4eg](https://github.com/kron4eg))
- Update oras-go dependency to v2.6.2 ([#4233](https://github.com/kubermatic/kubeone/pull/4233), [@kron4eg](https://github.com/kron4eg))
- Update MachineController to v1.67.0, update Operating System Manager to v1.12.0 ([#4312](https://github.com/kubermatic/kubeone/pull/4312), [@kron4eg](https://github.com/kron4eg))
