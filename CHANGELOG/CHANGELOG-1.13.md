# [v1.13.5](https://github.com/kubermatic/kubeone/releases/tag/v1.13.5) - 2026-05-13

## Changelog since v1.13.4

## Changes by Kind

### Fixes of Bugs or Regressions

- Update golang.org/x/net to v0.54.0 [#4076](https://github.com/kubermatic/kubeone/pull/4076), [@kron4eg](https://github.com/kron4eg)
- Update OSM to v1.10.5 with the Flatcar fix [#4068](https://github.com/kubermatic/kubeone/pull/4068), [@kron4eg](https://github.com/kron4eg)
- Removed HonorPVReclaimPolicy feature flag from csi-azuredisk addon [#4064](https://github.com/kubermatic/kubeone/pull/4064), @[bastianpaetzold](https://github.com/bastianpaetzold)
- Fix bastion host can have different ssh key [#4063](https://github.com/kubermatic/kubeone/pull/4063), @[mohamed-rafraf](https://github.com/mohamed-rafraf)

### Updates

- Use fixed Go minor version 1.26 in CI workflows #4056

# [v1.13.4](https://github.com/kubermatic/kubeone/releases/tag/v1.13.4) - 2026-04-17

## Changelog since v1.13.3

## Changes by Kind

### Updates

- Upgrade AzureFile CSI Driver to v1.35.2 [#4055](https://github.com/kubermatic/kubeone/pull/4055), [@kron4eg](https://github.com/kron4eg)
- Upgrade AzureDisk CSI Driver to v1.34.3 [#4055](https://github.com/kubermatic/kubeone/pull/4055), [@kron4eg](https://github.com/kron4eg)
- Upgrade DigitalOcean CCM to v0.1.66  [#4054](https://github.com/kubermatic/kubeone/pull/4054), [@kron4eg](https://github.com/kron4eg)
- Upgrade DigitalOceam CSI Driver to v4.16.0 [#4054](https://github.com/kubermatic/kubeone/pull/4054), [@kron4eg](https://github.com/kron4eg)
- Upgrade GCP CCM to v35.0.2 [#4054](https://github.com/kubermatic/kubeone/pull/4054), [@kron4eg](https://github.com/kron4eg)
- Upgrade GCP CSI compute-persistent driver to v1.23.3 [#4054](https://github.com/kubermatic/kubeone/pull/4054), [@kron4eg](https://github.com/kron4eg)

# [v1.13.3](https://github.com/kubermatic/kubeone/releases/tag/v1.13.3) - 2026-04-14

## Changelog since v1.13.2

## Changes by Kind

### Fixes of Bugs or Regressions

- Fix release formats, return zip files in release assets back [#4049](https://github.com/kubermatic/kubeone/pull/4049), [@kron4eg](https://github.com/kron4eg)

# [v1.13.2](https://github.com/kubermatic/kubeone/releases/tag/v1.13.2) - 2026-04-13

## Changelog since v1.13.1

## Changes by Kind

### Fixes of Bugs or Regressions

- Fix typo in cilium config [#4044](https://github.com/kubermatic/kubeone/pull/4044), [@kron4eg](https://github.com/kron4eg)

# [v1.13.1](https://github.com/kubermatic/kubeone/releases/tag/v1.13.1) - 2026-04-13

## Changelog since v1.13.0

## Changes by Kind

### Fixes of Bugs or Regressions

- Fix Azure CCM and CNM image versions [#4042](https://github.com/kubermatic/kubeone/pull/4042), [@kron4eg](https://github.com/kron4eg)
- Bump helm.sh/helm/v3 to 3.20.2 [#4041](https://github.com/kubermatic/kubeone/pull/4041)

# [v1.13.0](https://github.com/kubermatic/kubeone/releases/tag/v1.13.0) - 2026-04-09

## Changelog since v1.12.0

## Urgent and BREAKING Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Support for Kubernetes **1.31 and 1.32** has been removed. KubeOne v1.13 supports Kubernetes versions **1.33, 1.34, and 1.35**. Before upgrading KubeOne, ensure your clusters are running Kubernetes v1.33 or newer. ([#3973](https://github.com/kubermatic/kubeone/pull/3973), [@kron4eg](https://github.com/kron4eg))
- Delete long deprecated MachineAnnotations ([#3936](https://github.com/kubermatic/kubeone/pull/3936), [@kron4eg](https://github.com/kron4eg))
- **REQUIRES FIPS-140 ENABLED VCENTER!** Upgrade vSphere CSI driver to v3.7.0

## Changes by Kind

### Feature

- Add **Terraform-free Hetzner control plane provisioning** (beta): A new `controlPlane.nodeSets` API field combined with `cloudProvider.hetzner.controlPlane.loadBalancer` configuration allows KubeOne to provision and manage Hetzner VMs and a load balancer for the control plane directly from the KubeOne manifest, without requiring Terraform for provisioning VMs/loadbalancer. - THIS IS BETA, DO NOT USE FOR PRODUCTION! ([#3895](https://github.com/kubermatic/kubeone/pull/3895), [@kron4eg](https://github.com/kron4eg))
- Add `kubeone etcd` command group with subcommands for operating on the etcd cluster of a KubeOne-managed Kubernetes cluster: `members` (list members and alarms), `defragment` (defragment a member's storage), `disarm` (disarm alarms on one or all members), `snapshot` (take an etcd snapshot from a member). etcd `controlPlaneComponents.etcd` configuration options (`quotaBackendBytes`, `autoCompactionRetention`, `autoCompactionMode`) are also now supported. ([#3998](https://github.com/kubermatic/kubeone/pull/3998), [@kron4eg](https://github.com/kron4eg))
- Add support for Kubernetes 1.35. ([#3973](https://github.com/kubermatic/kubeone/pull/3973), [@kron4eg](https://github.com/kron4eg))
- Add `features.alwaysPullImages` API field to enable the `AlwaysPullImages` admission plugin on the Kubernetes API server. ([#4027](https://github.com/kubermatic/kubeone/pull/4027), [@adoi](https://github.com/adoi))
- Add `features.eventRateLimit` API field to enable the `EventRateLimit` admission plugin with a configurable config file path. ([#4029](https://github.com/kubermatic/kubeone/pull/4029), [@adoi](https://github.com/adoi))
- `NodeRestriction` admission plugin is now enabled by default. ([#4012](https://github.com/kubermatic/kubeone/pull/4012), [@adoi](https://github.com/adoi))
- Add `clusterNetwork.cni.cilium.enableL2Announcements` option to enable Cilium Layer 2 announcement feature. ([#3991](https://github.com/kubermatic/kubeone/pull/3991), [@rguhr](https://github.com/rguhr))
- Add insecure field in Helm release. ([#3921](https://github.com/kubermatic/kubeone/pull/3921), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Add helm authentication in HelmRelease. ([#3922](https://github.com/kubermatic/kubeone/pull/3922), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Add registry authentication support for both source registry and mirror hosts in `containerRuntime.containerd.registries`. ([#4014](https://github.com/kubermatic/kubeone/pull/4014), [@rajaSahil](https://github.com/rajaSahil))
- Remove validation of mutual exclusivity between `ContainerdRegistry` and `RegistryConfiguration`. Both can now be configured simultaneously. ([#3993](https://github.com/kubermatic/kubeone/pull/3993), [@kron4eg](https://github.com/kron4eg))
- Upgrade containerd from v1.7.x to v2.2.x.
  Note: The deprecated CRI-based registry authentication configuration is still being used with containerd v2. It is recommended to use Kubernetes ImagePullSecrets for registry authentication instead. ([#4006](https://github.com/kubermatic/kubeone/pull/4006), [@rajaSahil](https://github.com/rajaSahil))
- Use `certificateAuthority.bundle` field consistently across all configuration paths that previously used `caBundle`. ([#3925](https://github.com/kubermatic/kubeone/pull/3925), [@kron4eg](https://github.com/kron4eg))
- Skip `aznfs` apt package installation on Azure when the addon is not needed. ([#3949](https://github.com/kubermatic/kubeone/pull/3949), [@dharapvj](https://github.com/dharapvj))
- Update install script to support ARM architecture on Linux and macOS. ([#3914](https://github.com/kubermatic/kubeone/pull/3914), [@scheeles](https://github.com/scheeles))
- Add support for ECDSA CA key ([#4004](https://github.com/kubermatic/kubeone/pull/4004), [@kron4eg](https://github.com/kron4eg))


### Fixes of Bugs or Regressions


- Remove CPU/memory limits from machine-controller and operating-system-manager deployments. ([#3979](https://github.com/kubermatic/kubeone/pull/3979), [@kron4eg](https://github.com/kron4eg))
- Restore Cilium CIDR match policy that was missing from the Cilium configmap. ([#4036](https://github.com/kubermatic/kubeone/pull/4036), [@kron4eg](https://github.com/kron4eg))
- Add permission for services in KubeVirt CCM. ([#4035](https://github.com/kubermatic/kubeone/pull/4035), [@rajaSahil](https://github.com/rajaSahil))
- Set the infra namespace annotation on the control plane nodes for KubeVirt. ([#4034](https://github.com/kubermatic/kubeone/pull/4034), [@rajaSahil](https://github.com/rajaSahil))
- Fix cilium-envoy image reference ([#3910](https://github.com/kubermatic/kubeone/pull/3910), [@peschmae](https://github.com/peschmae))
- Run file permission reconciliation across all SSH-managed nodes, not just the leader. ([#4030](https://github.com/kubermatic/kubeone/pull/4030), [@adoi](https://github.com/adoi))
- Enables policy-cidr-match-mode: nodes in the Cilium CNI addon configuration. ([#4005](https://github.com/kubermatic/kubeone/pull/4005), [@rajaSahil](https://github.com/rajaSahil))
- Fix kernel version parsing to correctly ignore `+` suffix present in some kernel version strings (e.g., on Flatcar). ([#4009](https://github.com/kubermatic/kubeone/pull/4009), [@ttuellmann](https://github.com/ttuellmann))
- Add `allowVolumeExpansion: true` to the OpenStack Cinder CSI StorageClass to allow volume expansion. ([#4001](https://github.com/kubermatic/kubeone/pull/4001), [@jan-di](https://github.com/jan-di))
- Fix incorrect cluster name passed to KubeVirt CCM arguments. ([#3980](https://github.com/kubermatic/kubeone/pull/3980), [@kron4eg](https://github.com/kron4eg))
- Mirror CoreDNS image when containerd mirrors or `overwriteRegistry` are configured. ([#3929](https://github.com/kubermatic/kubeone/pull/3929), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Fix missing sandbox (pause) image when mirroring images. ([#3926](https://github.com/kubermatic/kubeone/pull/3926), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Respect customized addon manifests when applying addons. ([#3920](https://github.com/kubermatic/kubeone/pull/3920), [@appiepollo14](https://github.com/appiepollo14))
- Fix GCP CCM addon being applied twice when provided as a user-managed addon. ([#3919](https://github.com/kubermatic/kubeone/pull/3919), [@appiepollo14](https://github.com/appiepollo14))
- Fixed an issue in the OpenStack Terraform Quickstart configs that Neutron can not assign the floating IP to the basion host. ([#3943](https://github.com/kubermatic/kubeone/pull/3943), [@kleini](https://github.com/kleini))
- Fix `kubernetes-apt-keyring.gpg` file permissions to be set explicitly. ([#3940](https://github.com/kubermatic/kubeone/pull/3940), [@piotr1212](https://github.com/piotr1212))
- Fix `/etc/kubeone/proxy-env` file permissions to be set explicitly. ([#3939](https://github.com/kubermatic/kubeone/pull/3939), [@piotr1212](https://github.com/piotr1212))
- Fix cluster-autoscaler deployment not being migrated when `matchLabels` changed. ([#3958](https://github.com/kubermatic/kubeone/pull/3958), [@kron4eg](https://github.com/kron4eg))

### Updates

- Update machine-controller to [v1.65.0](https://github.com/kubermatic/machine-controller/releases/tag/v1.65.0) and operating-system-manager to [v1.9.0](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.9.0). ([#3979](https://github.com/kubermatic/kubeone/pull/3979), [#3982](https://github.com/kubermatic/kubeone/pull/3982), [#3983](https://github.com/kubermatic/kubeone/pull/3983), [@kron4eg](https://github.com/kron4eg))
- Update KubeVirt CSI image to v0.4.5 ([#3981](https://github.com/kubermatic/kubeone/pull/3981), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CSI driver to v2.18.3 ([#3934](https://github.com/kubermatic/kubeone/pull/3934), [@kron4eg](https://github.com/kron4eg))
- Update component versions ([#4013](https://github.com/kubermatic/kubeone/pull/4013), [#4017](https://github.com/kubermatic/kubeone/pull/4017), [@kron4eg](https://github.com/kron4eg)):
  - Cilium updated to v1.19.2
  - Canal (Calico) updated to v3.31.4
  - Hetzner CCM updated to v1.30.1 (now uses watch-based route reconciliation instead of polling)
  - Hetzner CSI driver updated to v2.20.0
  - vSphere CSI driver updated to v3.7.0
  - KubeVirt CSI driver updated to v0.4.5
  - metrics-server updated to v0.8.1
  - AWS EBS CSI driver updated to v1.57.1
  - AWS CCM: v1.33.2 / v1.34.0 / v1.35.0 (per Kubernetes version)
  - Azure CCM: v1.33.3 / v1.34.2 / v1.35.0 (per Kubernetes version)
  - OpenStack CCM: v1.33.1 / v1.34.1 / v1.35.0 (per Kubernetes version)
  - OpenStack Cinder CSI: v1.33.1 / v1.34.1 / v1.35.0 (per Kubernetes version)
  - vSphere CPI: v1.33.0 / v1.34.0 / v1.35.1 (per Kubernetes version)
  - ClusterAutoscaler: v1.33.4 / v1.34.3 / v1.35.0 (per Kubernetes version)
  - Equinix Metal CCM updated to v3.8.1
  - GCP CCM updated to v33.1.1
  - GCP Compute Persistent Disk CSI driver updated to v1.17.4
- Rename `cluster-autoscaler-values.yaml` addon values file to `cluster-autoscaler-values` (without extension). ([#3916](https://github.com/kubermatic/kubeone/pull/3916), [@steled](https://github.com/steled))
- Update KubeOne container base image to alpine:3.23. ([#3957](https://github.com/kubermatic/kubeone/pull/3957), [@archups](https://github.com/archups))
