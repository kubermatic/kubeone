# [v1.8.5](https://github.com/kubermatic/kubeone/releases/tag/v1.8.5) - 2024-09-23

## Changelog since v1.8.4

**Note: the v1.8.4 release has been abandoned due to an issue with the [deprecated `goreleaser` flags](https://github.com/kubermatic/kubeone/pull/3519).**

## Changelog since v1.8.3

## Changes by Kind

### Feature

- Add `disable_auto_update` option to example Terraform configs for AWS, Azure, Equinix Metal, OpenStack, and vSphere, used to disable automatic updates for all Flatcar nodes ([#3393](https://github.com/kubermatic/kubeone/pull/3393), [@xmudrii](https://github.com/xmudrii))
- Update OpenStack CCM and CSI driver to v1.30.2, v1.29.1 and v1.28.3 ([#3488](https://github.com/kubermatic/kubeone/pull/3488), [@rajaSahil](https://github.com/rajaSahil))

### Other (Cleanup or Flake)

- Use dedicated keyring for Docker repositories to solve `apt-key` deprecation warning upon installing/upgrading containerd ([#3486](https://github.com/kubermatic/kubeone/pull/3486), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Updates

#### operating-system-manager

- Update operating-system-manager (OSM) to v1.5.3 ([#3390](https://github.com/kubermatic/kubeone/pull/3390), [@kron4eg](https://github.com/kron4eg))

#### Others

- KubeOne is now built with Go 1.22.10 ([#3514](https://github.com/kubermatic/kubeone/pull/3514), [@xmudrii](https://github.com/xmudrii))


# [v1.8.3](https://github.com/kubermatic/kubeone/releases/tag/v1.8.3) - 2024-09-17

## Changelog since v1.8.2

## Urgent Upgrade Notes 

### (No, really, you MUST read this before you upgrade)

- Fix vSphere CCM and CSI images. The CCM images for versions starting with v1.28.0 are pulled from the new community-owned image repository. The CCM images for versions prior to v1.28.0, and the CSI images, are pulled from the Kubermatic-managed mirror on `quay.io`. If you have a vSphere cluster, we strongly recommend upgrading to the latest KubeOne patch release and running `kubeone apply` **as soon as possible**, because the old image repository (`gcr.io/cloud-provider-vsphere`) is not available anymore, hence it's not possible to pull the needed images from that repository ([#3378](https://github.com/kubermatic/kubeone/pull/3378), [@xmudrii](https://github.com/xmudrii))
- Example Terraform configs for Hetzner are now using `cx22` instead of `cx21` instance type by default. If you use the new Terraform configs with an existing cluster, make sure to override the instance type as needed, otherwise your instances/cluster will be destroyed ([#3371](https://github.com/kubermatic/kubeone/pull/3371), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.8.2](https://github.com/kubermatic/kubeone/releases/tag/v1.8.2) - 2024-08-08

## Changelog since v1.8.1

## Changes by Kind

### Feature

- Allow the configuration of the upstream cluster-autoscaler flags  `--enforce-node-group-min-size` and `--balance-similar-node-groups` ([#3306](https://github.com/kubermatic/kubeone/pull/3306), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Bug or Regression

- Do not put multiple identical tolerations on the CoreDNS deployment ([#3298](https://github.com/kubermatic/kubeone/pull/3298), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Use the RHEL-based upstream Docker package repository instead of the CentOS package repository as it's not maintained any longer ([#3336](https://github.com/kubermatic/kubeone/pull/3336), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Updates

#### CNI

- Update Calico CNI to fix the CPU usage spike issues ([#3326](https://github.com/kubermatic/kubeone/pull/3326), [@kron4eg](https://github.com/kron4eg))

#### machine-controller

- Update machine-controller to 1.59.3. This update includes support for IMDSv2 API on AWS for the worker nodes managed by machine-controller ([#3323](https://github.com/kubermatic/kubeone/pull/3323), [@xrstf](https://github.com/xrstf))

### Terraform Configs

- Set `HttpPutResponseHopLimit` to 3 in the example Terraform configs for AWS for the control plane nodes and the static worker nodes in order to support the IMSD v2 API ([#3329](https://github.com/kubermatic/kubeone/pull/3329), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.8.1](https://github.com/kubermatic/kubeone/releases/tag/v1.8.1) - 2024-07-01

## Changelog since v1.8.0

## Changes by Kind

### Feature

- Add support for Kubernetes 1.30 ([#3215](https://github.com/kubermatic/kubeone/pull/3215), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Refactor the cluster upgrade process to adhere to the [Kubernetes recommendations](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-reconfigure/) by updating ConfigMaps used by Kubeadm instead of providing the full config to Kubeadm itself. This change should not have any effect to cluster upgrades, but if you encounter any issue, please [create an issue](https://github.com/kubermatic/kubeone/issues/new/choose) in the KubeOne repository ([#3253](https://github.com/kubermatic/kubeone/pull/3253), [@kubermatic-bot](https://github.com/kubermatic-bot))
- KubeOne now runs `kubeadm upgrade apply` without the `--certificate-renewal=true` flag. This change should not have any effect to the upgrade process, but if you discover any issue, please [create a new issue](https://github.com/kubermatic/kubeone/issues/new/choose) in the KubeOne repository ([#3242](https://github.com/kubermatic/kubeone/pull/3242), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Add default VolumeSnapshotClass for all supported providers as part of the `default-storage-class` addon ([#3275](https://github.com/kubermatic/kubeone/pull/3275), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Bug or Regression

- Fix snapshot-webhook admitting non-supported objects (`VolumeSnapshots` and `VolumeSnapshotContents`). This fixes an issue that caused inability to create new `VolumeSnapshots` ([#3275](https://github.com/kubermatic/kubeone/pull/3275), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Ensure `apparmor-utils` package is installed on Ubuntu as it's required for `kubelet` to function properly ([#3235](https://github.com/kubermatic/kubeone/pull/3235), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Load the CA bundle before any addon installations to resolve issues with untrusted  TLS connections in environments with self-signed certificates ([#3247](https://github.com/kubermatic/kubeone/pull/3247), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix deletion issues for local Helm charts ([#3268](https://github.com/kubermatic/kubeone/pull/3268), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Updates

- Upgrade control plane components:
  - Update NodeLocalDNS to v1.23.1
  - Update AWS CCM to v1.30.1, v1.29.3, v1.28.6, and v1.27.7
  - Update CSI snapshot controller and webhook to v8.0.1
  - Update AWS EBS CSI driver to v1.31.0
  - Update Azure CCM to v1.30.3 for Kubernetes 1.30 clusters
  - Update AzureFile CSI driver to v1.30.2
  - Update AzureDisk CSI driver to v1.30.1
  - Update DigitalOcean CCM to v0.1.53
  - Update DigitalOcean CSI to v4.10.0
  - Update Hetzner CSI to v2.7.0
  - Update OpenStack CCM and CSI to v1.30.0 for Kubernetes 1.30 clusters
  - Update vSphere CCM to v1.30.1 for Kubernetes 1.30 clusters
  - Update vSphere CSI driver to v3.2.0
  - Update GCP Compute CSI driver to v1.13.2
  - Update Cilium to v1.15.6
  - Update cluster-autoscaler to v1.30.1, v1.29.3, v1.28.5, and v1.27.8 ([#3214](https://github.com/kubermatic/kubeone/pull/3214), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Update GCP CCM to v30.0.0 (Kubernetes 1.30), v29.0.0 (Kubernetes 1.29), v28.2.1 (Kubernetes 1.28 and 1.27) ([#3241](https://github.com/kubermatic/kubeone/pull/3241), [#3284](https://github.com/kubermatic/kubeone/pull/3284), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Update Canal CNI to v3.27.3 ([#3200](https://github.com/kubermatic/kubeone/pull/3200), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Bind the `csi-snapshotter` image to v8.0.1 for all providers that are supporting snapshotting volumes ([#3270](https://github.com/kubermatic/kubeone/pull/3270), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Terraform Configs

- Fix the default Rocky Linux EC2 image filter query in the example Terraform configs for AWS ([#3262](https://github.com/kubermatic/kubeone/pull/3262), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Add bastion host support to the example Terraform configs for VMware Virtual Cloud Director (VCD) ([#3278](https://github.com/kubermatic/kubeone/pull/3278), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.8.0](https://github.com/kubermatic/kubeone/releases/tag/v1.8.0) - 2024-05-14

We're happy to announce a new KubeOne minor release — KubeOne 1.8! Please
consult the changelog below, as well as, the following two documents before
upgrading:

- [Upgrading from KubeOne 1.7 to 1.8 guide](https://docs.kubermatic.com/kubeone/v1.9/tutorials/upgrading/upgrading-from-1.7-to-1.8/)
- [Known Issues in KubeOne 1.8](https://docs.kubermatic.com/kubeone/v1.9/known-issues/)

## Changelog since v1.7.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Refactor example Terraform configs for Hetzner to randomly generate the private network subnet in order to support creating multiple KubeOne clusters ([#3152](https://github.com/kubermatic/kubeone/pull/3152), [@xmudrii](https://github.com/xmudrii))
- The example Terraform configs for Azure have been migrated to use the Standard SKU for IP addresses. This is a breaking change for existing setups; in which case you should continue using your current SKU. Manual migration is possible by dissociating IP from the VM and LB, the migrating it, and assigning it back, however please consider all potential risks before doing this migration ([#3149](https://github.com/kubermatic/kubeone/pull/3149), [@kron4eg](https://github.com/kron4eg))
- Credentials defined in the credentials file now have precedence over credentials defined via environment variables. This change is made to match the behavior that's already documented in the KubeOne docs. If you use both the credentials file and the environment variables, we recommend double-checking your credentials file to make sure the credentials are up to date, as those credentials will be applied on the next `kubeone apply` run ([#2991](https://github.com/kubermatic/kubeone/pull/2991), [@kron4eg](https://github.com/kron4eg))
- kured has been removed, you have to re-enable it back in form of `helmRelease` ([#3024](https://github.com/kubermatic/kubeone/pull/3024), [@kron4eg](https://github.com/kron4eg))
- OSM: The latest Ubuntu 22.04 images on Azure have modified the configuration for `cloud-init` and how it accesses its datasource in Azure, in a breaking way. If you're having an Azure cluster, it's required to [refresh your machines](https://docs.kubermatic.com/kubeone/v1.7/cheat-sheets/rollout-machinedeployment/) with the latest provided OSPs to ensure that a system-wide package update doesn't result in broken machines. ([#3172](https://github.com/kubermatic/kubeone/pull/3172), [@xrstf](https://github.com/xrstf))
- Support for Docker is removed; `containerRuntime.docker` became a no-op. ([#3008](https://github.com/kubermatic/kubeone/pull/3008), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### API Changes

- Set `cloudProvider.external` = `true` by default for supported cloud providers in kubernetes 1.29+ ([#3048](https://github.com/kubermatic/kubeone/pull/3048), [@kron4eg](https://github.com/kron4eg))
- Check hostnames against Kubernetes node name requirements ([#3091](https://github.com/kubermatic/kubeone/pull/3091), [@SimonTheLeg](https://github.com/SimonTheLeg))
- Force `node-role.kubernetes.io/control-plane` label on control-plane Nodes ([#3099](https://github.com/kubermatic/kubeone/pull/3099), [@kron4eg](https://github.com/kron4eg))

### Feature

- Add support for Kubernetes 1.28 ([#2948](https://github.com/kubermatic/kubeone/pull/2948), [@xmudrii](https://github.com/xmudrii))
- Add support for kubernetes 1.29 ([#3048](https://github.com/kubermatic/kubeone/pull/3048), [@kron4eg](https://github.com/kron4eg))
- Make Kubernetes v1.29 the default stable Kubernetes version ([#3073](https://github.com/kubermatic/kubeone/pull/3073), [@kron4eg](https://github.com/kron4eg))
- Add GCP CCM addon ([#3038](https://github.com/kubermatic/kubeone/pull/3038), [@kron4eg](https://github.com/kron4eg))
- Add Nutanix CCM addon ([#3034](https://github.com/kubermatic/kubeone/pull/3034), [@kron4eg](https://github.com/kron4eg))
- Add `certOption` to the `hostConfig` API ([#3020](https://github.com/kubermatic/kubeone/pull/3020), [@AhmadAlEdlbi](https://github.com/AhmadAlEdlbi))
- Add a new API to configure TLS cipher suites for kube-apiserver, etcd and kubelet ([#3081](https://github.com/kubermatic/kubeone/pull/3081), [@kron4eg](https://github.com/kron4eg))
- Add support for customizing `vAppName` for VMware Cloud Director CSI driver ([#2932](https://github.com/kubermatic/kubeone/pull/2932), [@JamesClonk](https://github.com/JamesClonk))
- Add support for passing additional args to the kube-apiserver, kube-controller-manager, and kube-scheduler ([#3162](https://github.com/kubermatic/kubeone/pull/3162), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Allow setting `CCM_CONCURRENT_SERVICE_SYNCS` parameter on CCM addons to configure number of concurrent `LoadBalancer` service reconciles ([#2916](https://github.com/kubermatic/kubeone/pull/2916), [@embik](https://github.com/embik))
- Improve error messaging when working with remote files over SSH ([#3052](https://github.com/kubermatic/kubeone/pull/3052), [@kron4eg](https://github.com/kron4eg))
- Canal CNI: Add `IFACE` and `IFACE_REGEX` parameters to allow explicitly selecting network interface to be used for inter-node communication and VXLAN ([#3152](https://github.com/kubermatic/kubeone/pull/3152), [@xmudrii](https://github.com/xmudrii))
- Update to Go 1.22.1 ([#3072](https://github.com/kubermatic/kubeone/pull/3072), [@xrstf](https://github.com/xrstf))

### Bug or Regression

- Escape the registry name when the registry is configured as a wildcard ([#2927](https://github.com/kubermatic/kubeone/pull/2927), [@kron4eg](https://github.com/kron4eg))
- Bind `FLANNELD_IFACE` statically to status.hostIP ([#3157](https://github.com/kubermatic/kubeone/pull/3157), [@kron4eg](https://github.com/kron4eg))
- Clean yum cache upon configuring Kubernetes repos. This fixes an issue with cluster upgrades failing on nodes with an older yum version ([#3146](https://github.com/kubermatic/kubeone/pull/3146), [@xmudrii](https://github.com/xmudrii))
- Deploy user defined addons before the external CCM initialization. This fixes an issue with cluster provisioning for users that use both external CCM and external CNI ([#3065](https://github.com/kubermatic/kubeone/pull/3065), [@kron4eg](https://github.com/kron4eg))
- Don't use the deprecated path for GPG keys for Kubernetes and Docker repositories ([#2919](https://github.com/kubermatic/kubeone/pull/2919), [@xmudrii](https://github.com/xmudrii))
- Download cri-tools from the Kubernetes repos instead of the Amazon Linux 2 repos on instances running Amazon Linux 2 ([#2950](https://github.com/kubermatic/kubeone/pull/2950), [@xmudrii](https://github.com/xmudrii))
- Drop `containerRuntimeEndpoint` field from KubeletConfiguration to fix warning from `kubeadm init` and `kubeadm join` for clusters running Kubernetes prior to 1.27 ([#2939](https://github.com/kubermatic/kubeone/pull/2939), [@xmudrii](https://github.com/xmudrii))
- Fix Helm deploying resources in the wrong namespace ([#3000](https://github.com/kubermatic/kubeone/pull/3000), [@kron4eg](https://github.com/kron4eg))
- Fix a bug with the VMware Cloud Director CSI driver addon where it would crash if no `VCD_API_TOKEN` is set ([#2932](https://github.com/kubermatic/kubeone/pull/2932), [@JamesClonk](https://github.com/JamesClonk))
- Fix a globbing issue for `apt-get install` causing KubeOne to install wrong Kubernetes version in some circumstances ([#2958](https://github.com/kubermatic/kubeone/pull/2958), [@xmudrii](https://github.com/xmudrii))
- Fix cluster upgrades on Debian hosts with deprecated Kubernetes repositories ([#3076](https://github.com/kubermatic/kubeone/pull/3076), [@cnvergence](https://github.com/cnvergence))
- Fix file permissions setting on Flatcar ([#3138](https://github.com/kubermatic/kubeone/pull/3138), [@kron4eg](https://github.com/kron4eg))
- Fix incorrect validation that made `VCD_API_TOKEN` unusable for VMware Cloud Director ([#2945](https://github.com/kubermatic/kubeone/pull/2945), [@embik](https://github.com/embik))
- Fix indentation for manifests of csi-vsphere-ks addon ([#2905](https://github.com/kubermatic/kubeone/pull/2905), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Fix kubelet file permissions according to CIS 1.8 ([#3100](https://github.com/kubermatic/kubeone/pull/3100), [@kron4eg](https://github.com/kron4eg))
- Fix support for Flatcar stable channel 3815.2.0 ([#3040](https://github.com/kubermatic/kubeone/pull/3040), [@4ch3los](https://github.com/4ch3los))
- Propagate CA Bundle to vSphere CSI driver ([#2906](https://github.com/kubermatic/kubeone/pull/2906), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- `registryConfiguration.OverrideRegistry` is correctly applied to the pause image configured in static nodes (control plane and static workers) ([#2925](https://github.com/kubermatic/kubeone/pull/2925), [@embik](https://github.com/embik))
- Update CRDs for operating-system-manager addon ([#2933](https://github.com/kubermatic/kubeone/pull/2933), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

### Other (Cleanup or Flake)

- Increase the memory requests and limits from 300Mi to 600Mi for cluster-autoscaler ([#2978](https://github.com/kubermatic/kubeone/pull/2978), [@xmudrii](https://github.com/xmudrii))
- Extract csi-external-snapshotter into its own addon ([#3016](https://github.com/kubermatic/kubeone/pull/3016), [@kron4eg](https://github.com/kron4eg))
- Replace JSON6902 with Strategic Merge in Nutanix CSI driver ([#3035](https://github.com/kubermatic/kubeone/pull/3035), [@kron4eg](https://github.com/kron4eg))
- Use `DisableCloudProviders` feature gate as a replacement for `InTreePluginXXXUnregister` for each former in-tree provider ([#3075](https://github.com/kubermatic/kubeone/pull/3075), [@kron4eg](https://github.com/kron4eg))

### Updates

#### machine-controller

- Update machine-controller to v1.59.1 ([#3184](https://github.com/kubermatic/kubeone/pull/3184), [@xmudrii](https://github.com/xmudrii))

#### operating-system-manager

- Update operating-system-manager to v1.5.1 ([#3165](https://github.com/kubermatic/kubeone/pull/3165), [@xrstf](https://github.com/xrstf))

#### Cloud Provider integrations

- Update AWS CCM ([#3056](https://github.com/kubermatic/kubeone/pull/3056), [@kron4eg](https://github.com/kron4eg))
- Update AWS CSI driver, add snapshot webhook ([#3013](https://github.com/kubermatic/kubeone/pull/3013), [@kron4eg](https://github.com/kron4eg))
- Update Azure CCM ([#3019](https://github.com/kubermatic/kubeone/pull/3019), [@kron4eg](https://github.com/kron4eg))
- Update AzureDrive CSI ([#3019](https://github.com/kubermatic/kubeone/pull/3019), [@kron4eg](https://github.com/kron4eg))
- Update AzureFile CSI ([#3019](https://github.com/kubermatic/kubeone/pull/3019), [@kron4eg](https://github.com/kron4eg))
- Update DigitalOcean CCM ([#3027](https://github.com/kubermatic/kubeone/pull/3027), [@kron4eg](https://github.com/kron4eg))
- Update DigitalOcean CSI driver ([#3026](https://github.com/kubermatic/kubeone/pull/3026), [@kron4eg](https://github.com/kron4eg))
- Update Equinix Metal CCM ([#3028](https://github.com/kubermatic/kubeone/pull/3028), [@kron4eg](https://github.com/kron4eg))
- Update GCP CSI Driver ([#3023](https://github.com/kubermatic/kubeone/pull/3023), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CCM & CSI ([#3022](https://github.com/kubermatic/kubeone/pull/3022), [@kron4eg](https://github.com/kron4eg))
- Update Nutanix CSI ([#3029](https://github.com/kubermatic/kubeone/pull/3029), [@kron4eg](https://github.com/kron4eg))
- Update OpenStack CCM / CSI driver versions, drop unsupported versions ([#3014](https://github.com/kubermatic/kubeone/pull/3014), [@kron4eg](https://github.com/kron4eg))
- Update VMware Cloud Director CSI Driver to v1.6.0 ([#3094](https://github.com/kubermatic/kubeone/pull/3094), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update vSphere CPI (CCM) ([#3018](https://github.com/kubermatic/kubeone/pull/3018), [@kron4eg](https://github.com/kron4eg))
- Update vSphere CSI driver ([#3018](https://github.com/kubermatic/kubeone/pull/3018), [@kron4eg](https://github.com/kron4eg))

#### Others

- Update Canal / Calico VXLAN addon to v3.26.3 ([#2949](https://github.com/kubermatic/kubeone/pull/2949), [@xmudrii](https://github.com/xmudrii))
- Update Canal CNI to v3.27.2 ([#3055](https://github.com/kubermatic/kubeone/pull/3055), [@kron4eg](https://github.com/kron4eg))
- Update Cilium to v1.15 ([#3089](https://github.com/kubermatic/kubeone/pull/3089), [@kron4eg](https://github.com/kron4eg))
- Update Flatcar Linux Update Operator ([#3024](https://github.com/kubermatic/kubeone/pull/3024), [@kron4eg](https://github.com/kron4eg))
- Update Helm to v3.14.2 ([#3045](https://github.com/kubermatic/kubeone/pull/3045), [@kron4eg](https://github.com/kron4eg))
- Update and kustomize csi-azuredisk addon ([#3144](https://github.com/kubermatic/kubeone/pull/3144), [@kron4eg](https://github.com/kron4eg))
- Update and kustomize nodelocaldns ([#3039](https://github.com/kubermatic/kubeone/pull/3039), [@kron4eg](https://github.com/kron4eg))
- Update backup-restic addon to use etcd 3.5.11 for creating etcd snapshots ([#2981](https://github.com/kubermatic/kubeone/pull/2981), [@embik](https://github.com/embik))
- Update cluster-autoscaler to v1.27.3, v1.26.4, v1.25.3, add support for v1.28 ([#2949](https://github.com/kubermatic/kubeone/pull/2949), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler with scale from zero instructions ([#3086](https://github.com/kubermatic/kubeone/pull/3086), [@kron4eg](https://github.com/kron4eg))
- Update etcd to v3.5.10 ([#3002](https://github.com/kubermatic/kubeone/pull/3002), [@kron4eg](https://github.com/kron4eg))
- Update Kubernetes libs to v0.29.2 ([#3045](https://github.com/kubermatic/kubeone/pull/3045), [@kron4eg](https://github.com/kron4eg))
- Update metrics-server to v0.7.0 ([#3046](https://github.com/kubermatic/kubeone/pull/3046), [@kron4eg](https://github.com/kron4eg))
- Update restic addon ([#3025](https://github.com/kubermatic/kubeone/pull/3025), [@kron4eg](https://github.com/kron4eg))
