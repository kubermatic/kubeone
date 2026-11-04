# [v1.15.0](https://github.com/kubermatic/kubeone/releases/tag/v1.15.0) - 2026-11-09

## Changelog since v1.14.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Unattended-upgrades addon: the Flatcar Linux Update Operator is now deployed only if the cluster has at least one Flatcar Linux node (control plane, static workers, dynamic workers or MachineDeployments in the cluster). Action required for existing clusters without Flatcar nodes: on the next `kubeone apply` the `reboot-coordinator` namespace and its workloads are pruned, but the cluster-scoped `flatcar-linux-update-agent` and `flatcar-linux-update-operator` ClusterRoles and ClusterRoleBindings are not removed automatically and can be deleted manually. ([#4264](https://github.com/kubermatic/kubeone/pull/4264), [@steled](https://github.com/steled))

## Changes by Kind

### Chore

- Update Cilium version to v1.19.7 ([#4209](https://github.com/kubermatic/kubeone/pull/4209), [@archups](https://github.com/archups))
- Update Hetzner CCM to v1.39.2
  - Update Hetzner CSI Driver to v2.24.1
  - Update OpenStack CSI driver to v1.37.1
  - Update OpenStack Cloud Controller Manager to v1.37.1
  - Update AWS EBS CSI driver to v2.67.0
  - Update external CSI snapshot controller to v8.7.0
  - Update AzureFile CSI driver to v1.35.10
  - Update Cilium to v1.20.2
  - Update vSphere CSI driver to v3.8.0
  - Update vSphere CPI to v1.36.0
  - Update Nutanix CSI Storage to v3.9.0
  - Update Nutanix Cloud Provider to v0.6.3
  - Update GCP Cloud Controller Manager to v37.0.1
  - Update Canal to v3.33.0
  - Update GCP Compute Persistent Disk CSI driver to v1.27.1
  - Update AzureDisk CSI driver to v1.36.1
  - Update DigitalOcean CSI driver to v4.19.0
  - Update DigitalOcean CCM to v0.1.71
  - Update DNSNodeCache to v1.27.1
  - Update metrics-server to v0.9.0 ([#4260](https://github.com/kubermatic/kubeone/pull/4260), [@kron4eg](https://github.com/kron4eg))
- Update Machine-controller to [v1.66.1](https://github.com/kubermatic/machine-controller/releases/tag/v1.66.1) ([#4157](https://github.com/kubermatic/kubeone/pull/4157), [@archups](https://github.com/archups))
- Update OSM version to [v1.11.1](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.11.1)
  - Update Go version to v1.26.5 ([#4180](https://github.com/kubermatic/kubeone/pull/4180), [@archups](https://github.com/archups))
- Update cilium to v1.19.6 ([#4160](https://github.com/kubermatic/kubeone/pull/4160), [@kron4eg](https://github.com/kron4eg))
- Update flatcar-linux-update-operator to v0.10.0 to fix the update agent crashing on Flatcar >= 4757 due to quoted /etc/os-release values ([#4229](https://github.com/kubermatic/kubeone/pull/4229), [@steled](https://github.com/steled))

### Feature

- Add managed control plane support for DigitalOcean. ([#4262](https://github.com/kubermatic/kubeone/pull/4262), [@rajaSahil](https://github.com/rajaSahil))
- Add managed control plane support for GCP. ([#4269](https://github.com/kubermatic/kubeone/pull/4269), [@rajaSahil](https://github.com/rajaSahil))
- Add structured AuthenticationConfiguration support ([#4267](https://github.com/kubermatic/kubeone/pull/4267), [@kron4eg](https://github.com/kron4eg))
- Added `journald.systemMaxUse` to configure journald's `SystemMaxUse` on all nodes (default `5G`). ([#4278](https://github.com/kubermatic/kubeone/pull/4278), [@rajaSahil](https://github.com/rajaSahil))
- Add AWS control-plane provider ([#4204](https://github.com/kubermatic/kubeone/pull/4204), [@kron4eg](https://github.com/kron4eg))
- Add `--provider` flag to `kubeone config images list` and `kubeone mirror-images` to include CCM/CSI images only for the given cloud provider(s) ([#4112](https://github.com/kubermatic/kubeone/pull/4112), [@kron4eg](https://github.com/kron4eg))
- Add `kubeone addons apply` command to install, update and remove user addons and helm releases without running a full `kubeone apply`. The `--name` flag limits it to specific addons or helm releases.
  Fix deleting user addons with `delete: true`, and uninstalling helm releases that have more than one revision. ([#4263](https://github.com/kubermatic/kubeone/pull/4263), [@steled](https://github.com/steled))
- Add a `--destroy-control-plane` flag to `kubeone reset` to destroy control plane machines at the cloud provider instead of resetting them ([#4210](https://github.com/kubermatic/kubeone/pull/4210), [@kron4eg](https://github.com/kron4eg))
- Added support for an `addonParams` section in the credentials file, allowing addon secret parameters (e.g. the built-in `backups-restic` addon's `resticPassword`) to be provided via `credentials.yaml` instead of the KubeOneCluster manifest. ([#4206](https://github.com/kubermatic/kubeone/pull/4206), [@steled](https://github.com/steled))
- Added support for deploying openstack clusters with dualstack ([#4259](https://github.com/kubermatic/kubeone/pull/4259), [@4ch3los](https://github.com/4ch3los))
- Addons are applied with server-side apply instead of `kubectl apply` ([#4272](https://github.com/kubermatic/kubeone/pull/4272), [@kron4eg](https://github.com/kron4eg))
- Async dashboard loading ([#4165](https://github.com/kubermatic/kubeone/pull/4165), [@kron4eg](https://github.com/kron4eg))
- Azure control-plane provider ([#4185](https://github.com/kubermatic/kubeone/pull/4185), [@kron4eg](https://github.com/kron4eg))
- Implement reset --destroy-control-plane-load-balancer ([#4215](https://github.com/kubermatic/kubeone/pull/4215), [@kron4eg](https://github.com/kron4eg))
- Kubevirt CCM, CSI controller and machine-controller now run in the infra cluster's `InfraNamespace`.
  - `csi-kubevirt` addon is replaced by `kubevirt-csi-driver-operator`. Existing clusters are migrated automatically on `kubeone apply` ([#4261](https://github.com/kubermatic/kubeone/pull/4261), [@kron4eg](https://github.com/kron4eg))
- Paremetrize schedule and flags in backups-restic addon ([#4207](https://github.com/kubermatic/kubeone/pull/4207), [@kron4eg](https://github.com/kron4eg))
- Static worker upgrade retries now retry only the failed worker instead of repeating upgrades on workers that already succeeded. ([#4192](https://github.com/kubermatic/kubeone/pull/4192), [@taaahaaa](https://github.com/taaahaaa))
- The Hubble certgen job of the cni-cilium addon now fails when the Cilium CA (`kube-system/cilium-ca`) would expire before the newly issued certificates, instead of succeeding until the CA has expired. Rotate the CA when the `hubble-generate-certs` job fails. ([#4232](https://github.com/kubermatic/kubeone/pull/4232), [@steled](https://github.com/steled))
- Update supported Kubernetes version to 1.37 ([#4249](https://github.com/kubermatic/kubeone/pull/4249), [@kron4eg](https://github.com/kron4eg))

### Documentation

- Fix generated apidocs links to upstream objects to always be up to date. ([#4266](https://github.com/kubermatic/kubeone/pull/4266), [@archups](https://github.com/archups))

### Failing Test

- Migrate stale VolumeGroupSnapshot CRDs ([#4299](https://github.com/kubermatic/kubeone/pull/4299), [@kron4eg](https://github.com/kron4eg))
- Rename variable `name` to `cluster_name` OpenStack Terraform configs ([#4297](https://github.com/kubermatic/kubeone/pull/4297), [@kron4eg](https://github.com/kron4eg))

### Bug or Regression

- Fix worker nodes provisioned by operating-system-manager getting a non-existent DNS server (`10.10.10.10`) when NodeLocalDNS is disabled, including clusters using Cilium Local Redirect Policy. With Cilium Local Redirect Policy, kubelets now use the node-local DNS address with CoreDNS as fallback. The CoreDNS IP is now computed from the IPv6 service subnet on IPv6-primary clusters.
  Add `clusterNetwork.clusterDNS` to override the DNS servers that kubelets configure for pods on all nodes. It can't be combined with `features.nodeLocalDNS.deploy`. ([#4251](https://github.com/kubermatic/kubeone/pull/4251), [@kron4eg](https://github.com/kron4eg))
- Add lease permissions for Cilium and enable l2announcements in helm values ([#4257](https://github.com/kubermatic/kubeone/pull/4257), [@kron4eg](https://github.com/kron4eg))
- Add subnet ID retrieval for OpenStack load balancer members ([#4242](https://github.com/kubermatic/kubeone/pull/4242), [@kron4eg](https://github.com/kron4eg))
- Fix Cilium `tlsinterception-secrets` Roles/RoleBindings being deployed to `kube-system` instead of `cilium-secrets` ([#4276](https://github.com/kubermatic/kubeone/pull/4276), [@kron4eg](https://github.com/kron4eg))
- Fix KubeOne retrying PodDisruptionBudget-blocked pod evictions without a delay during node drain. ([#4181](https://github.com/kubermatic/kubeone/pull/4181), [@taaahaaa](https://github.com/taaahaaa))
- Fix PodDisruptionBudget selectors that match no pods in addons ([#4281](https://github.com/kubermatic/kubeone/pull/4281), [@rajaSahil](https://github.com/rajaSahil))
- Fix overridePath handling in ContainerRuntimeConfig.MachineControllerFlags ([#4174](https://github.com/kubermatic/kubeone/pull/4174), [@kron4eg](https://github.com/kron4eg))
- Fix provisioning with cloud provider `none` failing on CSI webhook certificate generation when no CSI addon is deployed ([#4234](https://github.com/kubermatic/kubeone/pull/4234), [@mvanhorn](https://github.com/mvanhorn))
- Fix the Cilium LRP ([#4189](https://github.com/kubermatic/kubeone/pull/4189), [@kron4eg](https://github.com/kron4eg))
- Fix the `ccm-hetzner` add-on PodDisruptionBudget selector so it matches the hcloud-cloud-controller-manager pods ([#4280](https://github.com/kubermatic/kubeone/pull/4280), [@damienbenon](https://github.com/damienbenon))
- Fixed OCI Helm chart authentication so `HelmRelease.Auth` credentials are correctly passed to private OCI registries.
  - Fixed `mirror-images` failing to retag digest-pinned images by preserving the source digest instead of incorrectly using it as a tag. ([#4221](https://github.com/kubermatic/kubeone/pull/4221), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Fixed a bug where upgrading a vSphere cluster from an older KubeOne version left behind a stale `validation.csi.vsphere.vmware.com` ValidatingWebhookConfiguration, blocking StorageClass changes and all new dynamic volume provisioning/PVC deletion. ([#4195](https://github.com/kubermatic/kubeone/pull/4195), [@steled](https://github.com/steled))
- Update Helm dry run strategy on release upgrade ([#4168](https://github.com/kubermatic/kubeone/pull/4168), [@kron4eg](https://github.com/kron4eg))
- Update golang.org/x/net v0.60.0 ([#4252](https://github.com/kubermatic/kubeone/pull/4252), [@kron4eg](https://github.com/kron4eg))
- Update oras-go dependency to v2.6.2 ([#4233](https://github.com/kubermatic/kubeone/pull/4233), [@kron4eg](https://github.com/kron4eg))
- Upgrade golang.org/x/crypto to v0.56.0 ([#4217](https://github.com/kubermatic/kubeone/pull/4217), [@kron4eg](https://github.com/kron4eg))

### Other (Cleanup or Flake)

- Add control-plane cloud-providers interface ([#4179](https://github.com/kubermatic/kubeone/pull/4179), [@kron4eg](https://github.com/kron4eg))
- Remove default AWS prowjob ([#4298](https://github.com/kubermatic/kubeone/pull/4298), [@kron4eg](https://github.com/kron4eg))
- Upgrade Go to 1.27 ([#4208](https://github.com/kubermatic/kubeone/pull/4208), [@kron4eg](https://github.com/kron4eg))
- Use upstream machine-controller cloudprovider types ([#4245](https://github.com/kubermatic/kubeone/pull/4245), [@kron4eg](https://github.com/kron4eg))
