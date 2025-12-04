# [v1.12.2](https://github.com/kubermatic/kubeone/releases/tag/v1.12.2) - 2025-12-04

## Changelog since v1.12.1

## Changes by Kind

### Chore

- Update cloud components versions [#3915](https://github.com/kubermatic/kubeone/pull/3915) [@kron4eg](https://github.com/kron4eg)
  - Update metrics-server helm chart to v3.13.0
  - Update vSphere CSI driver to v3.6.0
  - Update OpenStack Cinder CSI driver to v2.34.1
  - Update DigitalOcean CSI driver to v4.15.0
  - Update AzureFile CSI driver to v1.34.2
  - Update Azure Disk CSI driver to v1.33.7
  - Update AWS EBS CSI driver to v2.53.0
  - Update Cilium to v1.18.4
  - Update Canal to v3.31.2
  - Update OpenStack CCM to v1.34.1
  - Update Azure CCM to v1.34.2
  - Update AWS CCM to v0.0.10

### Fixes of Bugs or Regressions

- Fix error applying cluster-autoscaler addon [#3916](https://github.com/kubermatic/kubeone/pull/3916) [@steled](https://github.com/steled)
- Respect customized Addons manifests [#3920](https://github.com/kubermatic/kubeone/pull/3920) [@appiepollo14](https://github.com/appiepollo14)

# [v1.12.1](https://github.com/kubermatic/kubeone/releases/tag/v1.12.1) - 2025-11-21

## Changelog since v1.12.0

## Changes by Kind

### Fixes of Bugs or Regressions

- Fix cilium-envoy image reference [#3910](https://github.com/kubermatic/kubeone/pull/3910) [@peschmae](https://github.com/peschmae)

# [v1.12.0](https://github.com/kubermatic/kubeone/releases/tag/v1.12.0) - 2025-11-21

## Changelog since v1.11.0

## Urgent and BREAKING Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Update RockyLinux 8 -> 9 and RHEL 8 -> 9 versions for the supported providers. ([#3822](https://github.com/kubermatic/kubeone/pull/3822), [@rajaSahil](https://github.com/rajaSahil)).
  RockyLinux 8 and RHEL 8 are not supported anymore because of their too old kernel version fall off minimal required version by Kubernetes.

### Known Issues

- rocky-9 image on hetzner doesn't work as of time of the release, since it only has IPv6 NS servers configured, regardless of the stack.

## Changes by Kind

### Feature

- Add `--all` flag to `config images list` showing all images independent of Kubernetes version ([#3782](https://github.com/kubermatic/kubeone/pull/3782), [@peschmae](https://github.com/peschmae))
- Add `remove-volumes` and `remove-lb-services` flags to delete dynamically provisioned and unretained PersistentVolumes and LoadBalancer Services before resetting the cluster ([#3507](https://github.com/kubermatic/kubeone/pull/3507), [@rajaSahil](https://github.com/rajaSahil))
- Add bastion SSH private key file setting in host config ([#3814](https://github.com/kubermatic/kubeone/pull/3814), [@kron4eg](https://github.com/kron4eg))
- Add overridePath API, to configure containerd override_path mirrors parameter ([#3843](https://github.com/kubermatic/kubeone/pull/3843), [@kron4eg](https://github.com/kron4eg))
- Add support for k8s version 1.34 ([#3823](https://github.com/kubermatic/kubeone/pull/3823), [@archups](https://github.com/archups))
- Cleanup /etc/kubernetes/tmp after upgrades ([#3775](https://github.com/kubermatic/kubeone/pull/3775), [@kron4eg](https://github.com/kron4eg))
- Cluster wide KubeletConfig ([#3845](https://github.com/kubermatic/kubeone/pull/3845), [@kron4eg](https://github.com/kron4eg))
- Export NewRoot() function ([#3809](https://github.com/kubermatic/kubeone/pull/3809), [@kron4eg](https://github.com/kron4eg))
- Make machine-controller -join-cluster-timeout configurable ([#3779](https://github.com/kubermatic/kubeone/pull/3779), [@kron4eg](https://github.com/kron4eg))
- Non-root device usage on non-static worker nodes can now be enabled for containerd runtime by setting the value `operatingSystemManager.enableNonRootDeviceOwnership` to `true` when OSM is enabled. ([#3793](https://github.com/kubermatic/kubeone/pull/3793), [@soer3n](https://github.com/soer3n))
- TBD ([#3835](https://github.com/kubermatic/kubeone/pull/3835), [@archups](https://github.com/archups))
- `kubeone certificates renew` command can be used to renew control plane certificates in a KubeOne cluster ([#3773](https://github.com/kubermatic/kubeone/pull/3773), [@kron4eg](https://github.com/kron4eg))

### Fixes of Bugs or Regressions

- Default canal_iface_regex only for hetzner ([#3797](https://github.com/kubermatic/kubeone/pull/3797), [@kron4eg](https://github.com/kron4eg))
- Don't install software-properties-common on deb systems ([#3833](https://github.com/kubermatic/kubeone/pull/3833), [@ttuellmann](https://github.com/ttuellmann))
- Enable_disk_uuid in vsphere terraform ([#3772](https://github.com/kubermatic/kubeone/pull/3772), [@kron4eg](https://github.com/kron4eg))
- Fix CSI snapshot webhook name for Nutanix ([#3761](https://github.com/kubermatic/kubeone/pull/3761), [@kron4eg](https://github.com/kron4eg))
- Fix Nutanix credentials ([#3776](https://github.com/kubermatic/kubeone/pull/3776), [@kron4eg](https://github.com/kron4eg))
- Fix upgrading OCI helm releases and uninstalling them without reason ([#3849](https://github.com/kubermatic/kubeone/pull/3849), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Fix validation to pass when ChartURL is given ([#3821](https://github.com/kubermatic/kubeone/pull/3821), [@kron4eg](https://github.com/kron4eg))
- Fixed an invalid image reference for the GCE Persistent Disk CSI Driver and update associated images. ([#3884](https://github.com/kubermatic/kubeone/pull/3884), [@rajaSahil](https://github.com/rajaSahil))
- Fixed defaulting of LoggingConfig ([#3881](https://github.com/kubermatic/kubeone/pull/3881), [@kron4eg](https://github.com/kron4eg))
- Fixes the Hubbele Relay Connection Issues with the Cilium Agent, SSL Connection is fixed by mounting the Server Certificates in the Cilium Agent Container ([#3795](https://github.com/kubermatic/kubeone/pull/3795), [@tobstone](https://github.com/tobstone))
- Make it possible to configure FLANNELD_IFACE ([#3790](https://github.com/kubermatic/kubeone/pull/3790), [@kron4eg](https://github.com/kron4eg))
- Restart kubelets sequentially ([#3770](https://github.com/kubermatic/kubeone/pull/3770), [@kron4eg](https://github.com/kron4eg))
- Terraform configs for Hetzner are now using `cx23` instead of `cx22` instance type by default. The `cx22` server type is deprecated and will no longer be available for order as of January 1, 2026. Make sure to override the instance type if you are using the new Terraform configs with an existing cluster. ([#3871](https://github.com/kubermatic/kubeone/pull/3871), [@adoi](https://github.com/adoi))
- Upgrade helm v3.18.5 ([#3781](https://github.com/kubermatic/kubeone/pull/3781), [@kron4eg](https://github.com/kron4eg))

### Chore

- Add RHEL and RockyLinux 9.6 test scenarios for v1.34 ([#3851](https://github.com/kubermatic/kubeone/pull/3851), [@kron4eg](https://github.com/kron4eg))
- Bump machine-controller version to [v1.63.1](https://github.com/kubermatic/machine-controller/releases/tag/v1.63.1) and operating-system-manager version to [v1.7.6](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.7.6) ([#3817](https://github.com/kubermatic/kubeone/pull/3817), [@archups](https://github.com/archups))
- Cluster-autoscaler addon now supports new variable CLUSTER_AUTOSCALER_SCALE_DOWN_UTIL_THRESHOLD to control `--scale-down-utilization-threshold` parameter. ([#3780](https://github.com/kubermatic/kubeone/pull/3780), [@dharapvj](https://github.com/dharapvj))
- Update Azure CCM to v1.34.1
  Update DigitalOcean CCM to v0.1.64
  Update Hetzner CCM and CSI to v2.18.0
  Update AWS EBS CSI to v1.51.0
  Update ClusterAutoscaler to v1.34.1 ([#3847](https://github.com/kubermatic/kubeone/pull/3847), [@archups](https://github.com/archups))
- Update OpenStack CCM and CSI version to 1.34.0 ([#3846](https://github.com/kubermatic/kubeone/pull/3846), [@archups](https://github.com/archups))
- Update machine-controller and operating-system-manager images to v1.64.0 and v1.8.0 respectively ([#3848](https://github.com/kubermatic/kubeone/pull/3848), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to v1.63.0 ([#3799](https://github.com/kubermatic/kubeone/pull/3799), [@archups](https://github.com/archups))
- Upgrade nutanix CSI driver to 3.3.4 ([#3808](https://github.com/kubermatic/kubeone/pull/3808), [@kron4eg](https://github.com/kron4eg))
- Use flatcar-container-linux-corevm-amd64 for flatcar Azure terraform example ([#3806](https://github.com/kubermatic/kubeone/pull/3806), [@kron4eg](https://github.com/kron4eg))

### Other (Cleanup or Flake)

- Drop centos from azure terraform example ([#3805](https://github.com/kubermatic/kubeone/pull/3805), [@kron4eg](https://github.com/kron4eg))
- Improve OS package handling on deb systems ([#3840](https://github.com/kubermatic/kubeone/pull/3840), [@ttuellmann](https://github.com/ttuellmann))
- Remove everything regarding amzn2 linux ([#3842](https://github.com/kubermatic/kubeone/pull/3842), [@kron4eg](https://github.com/kron4eg))
