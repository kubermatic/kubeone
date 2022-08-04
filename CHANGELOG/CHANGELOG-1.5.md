# [v1.5.0-beta.0](https://github.com/kubermatic/kubeone/releases/tag/v1.5.0-beta.0) - 2022-08-04

## Changelog since v1.4.0

## Urgent Upgrade Notes 

### (No, really, you MUST read this before you upgrade)

- Automatically apply the `node-role.kubernetes.io/control-plane` taint to nodes running Kubernetes 1.24. The taint is also applied when upgrading nodes from Kubernetes 1.23 to 1.24. You might need to adjust your workloads to tolerate the `node-role.kubernetes.io/control-plane` taint (in addition to the `node-role.kubernetes.io/master` taint). Workloads deployed by KubeOne will be adjusted automatically. ([#2019](https://github.com/kubermatic/kubeone/pull/2019), [@xmudrii](https://github.com/xmudrii))
- Kubeadm is now applying the `node-role.kubernetes.io/control-plane` label for Kubernetes 1.24 nodes. The old label (`node-role.kubernetes.io/master`) will be removed when upgrading the cluster to Kubernetes 1.24. All addons are updated to use the `node-role.kubernetes.io/control-plane` label selector instead. All addons now have toleration for `node-role.kubernetes.io/control-plane` taint in addition to toleration for `node-role.kubernetes.io/master` taint. If you are overriding addons, make sure to apply those changes before upgrading to Kubernetes 1.24. ([#2017](https://github.com/kubermatic/kubeone/pull/2017), [@xmudrii](https://github.com/xmudrii))
- Operating System Manager is enabled by default and is responsible for generating and managing user-data used for provisioning worker nodes
  - Existing worker machines will not be migrated to use OSM automatically. The user needs to manually rollout all MachineDeployments to start using OSM. This can be done by following the steps described in [Rolling Restart MachineDeploments document](https://docs.kubermatic.com/kubeone/v1.4/cheat_sheets/rollout_machinedeployment/)
  - The user can opt-out from OSM by setting `.operatingSystemManager.deploy` to `false` in their KubeOneCluster manifest. ([#2157](https://github.com/kubermatic/kubeone/pull/2157), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- `workers_replicas` variable has been renamed to `initial_machinedeployment_replicas` in example Terraform configs for Hetzner ([#2115](https://github.com/kubermatic/kubeone/pull/2115), [@adeniyistephen](https://github.com/adeniyistephen))
- Change default instance size in example Terraform configs for Equinix Metal to `c3.small.x86` because `t1.small.x86` is not available any longer. If you're using the latest Terraform configs for Equinix Metal with an existing cluster, make sure to explicitly set the instance size (`device_type` and `lb_device_type`) in `terraform.tfvars` or otherwise your instances might get recreated ([#2054](https://github.com/kubermatic/kubeone/pull/2054), [@xmudrii](https://github.com/xmudrii))
- Update secret name for `backup-restic` addon to `kubeone-backups-credentials`. Manual migration steps are needed for users running KKP on top of a KubeOne installation and using both `backup-restic` addon from KubeOne and `s3-exporter` from KKP. Ensure that the `s3-credentials` Secret with keys `ACCESS_KEY_ID` and `SECRET_ACCESS_KEY` exists in `kube-system` namespace and doesn't have the label `kubeone.io/addon:`. Remove the label if it exists. Otherwise, `s3-exporter` won't be functional. ([#1880](https://github.com/kubermatic/kubeone/pull/1880), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

## Known Issues

* Calico VXLAN addon has an issue with broken network connectivity for pods running on the same node. If you're using Calico VXLAN, we recommend staying on KubeOne 1.4 until the issue is not fixed. Follow [#2192](https://github.com/kubermatic/kubeone/issues/2192) for updates.

## Changes by Kind

### API Change

- Extend KubeOneCluster API with the `CoreDNS` feature allowing users to configure the number of CoreDNS replicas and whether should KubeOne create a PodDistruptionBudget for CoreDNS. Default values are 2 replicas and create PDB. Run `kubeone config print --full` for more details
  - Add Pod Anti Affinity to the CoreDNS deployment to avoid having multiple CoreDNS pods on the same node ([#2165](https://github.com/kubermatic/kubeone/pull/2165), [@xmudrii](https://github.com/xmudrii))
- Add `MaxPods` field to the KubeletConfig used to control the maximum number of pods per node ([#2075](https://github.com/kubermatic/kubeone/pull/2075), [@xmudrii](https://github.com/xmudrii))
- Add `machineObjectAnnotations` field to `DynamicWorkerNodes` used to apply annotations to resulting Machine objects
  Add `nodeAnnotations` field to DynamicWorkerNodes Config as a replacement for deprecated `machineAnnotations` field ([#2074](https://github.com/kubermatic/kubeone/pull/2074), [@xmudrii](https://github.com/xmudrii))
- Add new `HostConfig.Labels` map to manage custom labels on the static worker nodes ([#2130](https://github.com/kubermatic/kubeone/pull/2130), [@kron4eg](https://github.com/kron4eg))
- Allow having no OIDC GroupsPrefix ([#1942](https://github.com/kubermatic/kubeone/pull/1942), [@kron4eg](https://github.com/kron4eg))

### Feature

#### General

- Add support for Rocky Linux operating system ([#2121](https://github.com/kubermatic/kubeone/pull/2121), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Enable the etcd integrity checks (on startup and every 4 hours) for Kubernetes 1.22+ clusters. See the official etcd announcement for more details (https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ). ([#1907](https://github.com/kubermatic/kubeone/pull/1907), [@xmudrii](https://github.com/xmudrii))
- Add `kubeone local` subcommand used to provision single-node Kubernetes cluster on current machine ([#2125](https://github.com/kubermatic/kubeone/pull/2125), [@kron4eg](https://github.com/kron4eg))
- Implement the `kubeone config dump` command used to merge the KubeOneCluster manifest with the Terraform output. The resulting (merged) manifest is printed to stdout. ([#1874](https://github.com/kubermatic/kubeone/pull/1874), [@xmudrii](https://github.com/xmudrii))
- Rollout pods that are using `kubeone-*-credentials` Secrets if credentials are changed ([#2214](https://github.com/kubermatic/kubeone/pull/2214), [@xmudrii](https://github.com/xmudrii))
- Error reporting in CLI now exists with different codes for different error reasons ([#1882](https://github.com/kubermatic/kubeone/pull/1882), [@kron4eg](https://github.com/kron4eg))
- More error handling with new error types ([#1890](https://github.com/kubermatic/kubeone/pull/1890), [@kron4eg](https://github.com/kron4eg))
- Add dedicated error type (and error code) for exec adapter ([#2139](https://github.com/kubermatic/kubeone/pull/2139), [@kron4eg](https://github.com/kron4eg))
- Strict Terraform output reading ([#1833](https://github.com/kubermatic/kubeone/pull/1833), [@kron4eg](https://github.com/kron4eg))
- `--log-format` flag is introduced to choose between text and JSON formatted logging ([#2060](https://github.com/kubermatic/kubeone/pull/2060), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- [EXPERIMENTAL] Add the KubeOne container image. This image should NOT be used in the production. ([#1875](https://github.com/kubermatic/kubeone/pull/1875), [@xmudrii](https://github.com/xmudrii))

#### Cloud Providers

- Add support and Terraform integration for VMware Cloud Director ([#2006](https://github.com/kubermatic/kubeone/pull/2006), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik), [#2059](https://github.com/kubermatic/kubeone/pull/2059), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- OpenStack: Domain is not required when using application credentials ([#1896](https://github.com/kubermatic/kubeone/pull/1896), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Equinix Metal: Replace Facilities with Metro in Terraform configs ([#2158](https://github.com/kubermatic/kubeone/pull/2158), [@xmudrii](https://github.com/xmudrii))

#### Addons

- Add CSI snapshot controller and webhook to the Cinder CSI driver ([#2067](https://github.com/kubermatic/kubeone/pull/2067), [@xmudrii](https://github.com/xmudrii))
- Add missing Snapshot CRDs for Openstack CSI ([#1871](https://github.com/kubermatic/kubeone/pull/1871), [@WeirdMachine](https://github.com/WeirdMachine))
- Add default VolumeSnapshotClass for OpenStack Cinder CSI ([#2217](https://github.com/kubermatic/kubeone/pull/2217), [@xmudrii](https://github.com/xmudrii))
- Add CSI snapshot controller and webhook to the vSphere CSI driver. Add the default VolumeSnapshotClass for vSphere ([#2050](https://github.com/kubermatic/kubeone/pull/2050), [@xmudrii](https://github.com/xmudrii))
- Add GCP Compute Persistent Disk CSI driver. The CSI driver is deployed by default for all GCE clusters running Kubernetes 1.23 or newer. ([#2137](https://github.com/kubermatic/kubeone/pull/2137), [@xmudrii](https://github.com/xmudrii))
- Add the VMware Cloud Director CSI driver addon. Add default StorageClass for the VMware Cloud Director CSI driver. ([#2092](https://github.com/kubermatic/kubeone/pull/2092), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Add Secrets Store CSI driver and Hashicorp Vault provider as optional addons. See addons' README files for more information on how to activate and use those addons. ([#2022](https://github.com/kubermatic/kubeone/pull/2022), [@kron4eg](https://github.com/kron4eg))
- Add `.Params.RequestsCPU` parameter to `cni-canal` addon ([#1925](https://github.com/kubermatic/kubeone/pull/1925), [@kron4eg](https://github.com/kron4eg))
- Create PodDistruptionBudget objects for all Deployments created by KubeOne addons ([#1906](https://github.com/kubermatic/kubeone/pull/1906), [@kron4eg](https://github.com/kron4eg))

### Updates

#### Go

- KubeOne is now built using Go 1.19.0 ([#2226](https://github.com/kubermatic/kubeone/pull/2226), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built using Go 1.18.4 ([#2179](https://github.com/kubermatic/kubeone/pull/2179), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built using Go 1.18.1 ([#2018](https://github.com/kubermatic/kubeone/pull/2018), [@xmudrii](https://github.com/xmudrii))

#### etcd

- Deploy etcd v3.5.3 for clusters running Kubernetes 1.22 or newer. etcd v3.5.3 includes a fix for the data inconsistency issues announced by the etcd maintainers: https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ
  To upgrade etcd for an existing cluster, you need to force upgrade the cluster as described here: https://docs.kubermatic.com/kubeone/v1.4/guides/etcd_corruption/#enabling-etcd-corruption-checks ([#1951](https://github.com/kubermatic/kubeone/pull/1951), [@xmudrii](https://github.com/xmudrii))

#### containerd

- Update containerd to 1.5. Amazon Linux 2 is still using containerd 1.4 because 1.5 is not available. ([#2020](https://github.com/kubermatic/kubeone/pull/2020), [@xmudrii](https://github.com/xmudrii))

#### machine-controller

- Update machine-controller to v1.53.0 ([#2207](https://github.com/kubermatic/kubeone/pull/2207), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update machine-controller to v1.52.0 ([#2126](https://github.com/kubermatic/kubeone/pull/2126), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update machine-controller to v1.51.0 ([#2078](https://github.com/kubermatic/kubeone/pull/2078), [@xmudrii](https://github.com/xmudrii))
- Update machine-controller to v1.49.0. machine-controller images are now hosted on Quay instead of Docker Hub. ([#2025](https://github.com/kubermatic/kubeone/pull/2025), [@xmudrii](https://github.com/xmudrii))
- Update machine-controller to v1.47.0 ([#1979](https://github.com/kubermatic/kubeone/pull/1979), [@kron4eg](https://github.com/kron4eg))

#### Operating System Manager (OSM)

- Update operating-system-manager to v0.6.0 ([#2207](https://github.com/kubermatic/kubeone/pull/2207), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update operating-system-manager to v0.5.0 ([#2126](https://github.com/kubermatic/kubeone/pull/2126), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update operating-system-manager to v0.4.2 ([#1903](https://github.com/kubermatic/kubeone/pull/1903), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))


#### CNI

- Update Canal and Calico VXLAN to v3.23.3. This allows users to use kube-proxy in IPVS mode on ARM64 clusters running Kubernetes 1.23 and newer ([#2188](https://github.com/kubermatic/kubeone/pull/2188), [@xmudrii](https://github.com/xmudrii))
- Update Canal and Calico VXLAN to v3.22.2. This allows users to use kube-proxy in IPVS mode on AMD64 clusters running Kubernetes 1.23 and newer ([#2041](https://github.com/kubermatic/kubeone/pull/2041), [@xmudrii](https://github.com/xmudrii))
- Update Flannel to v0.15.1 to fix an issue with Flannel causing `iptables` segfaults ([#1986](https://github.com/kubermatic/kubeone/pull/1986), [@mfranczy](https://github.com/mfranczy))
- Switching to `quay.io` from `docker.io` for Calico CNI images ([#2043](https://github.com/kubermatic/kubeone/pull/2043), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update Cilium to v1.12.0 ([#2220](https://github.com/kubermatic/kubeone/pull/2220), [@xmudrii](https://github.com/xmudrii))
- Update Cilium to v1.11.5 ([#2049](https://github.com/kubermatic/kubeone/pull/2049), [@xmudrii](https://github.com/xmudrii))

#### AWS

- Update AWS CCM to the latest releases for all supported Kubernetes versions. Update AWS EBS CSI driver to v1.9.0 ([#2171](https://github.com/kubermatic/kubeone/pull/2171), [@xmudrii](https://github.com/xmudrii))
- Update AWS CCM to v1.24.0, v1.23.1, v1.22.2, v1.21.1, v1.20.1. Update AWS EBS CSI driver to v1.6.2 ([#2055](https://github.com/kubermatic/kubeone/pull/2055), [@xmudrii](https://github.com/xmudrii))

#### Azure

- Update Azure CCM to the latest releases for all supported Kubernetes versions. Update AzureDisk CSI driver to v1.21.0. Update AzureFile CSI driver to v1.20.0 ([#2172](https://github.com/kubermatic/kubeone/pull/2172), [@xmudrii](https://github.com/xmudrii))
- Update Azure CCM to v1.24.0, v1.23.11, v1.1.14 (for Kubernetes 1.22), v1.0.18 (for Kubernetes 1.21), v0.7.21 (for Kubernetes 1.20). Update AzureDisk CSI driver to v1.18.0. Update AzureFile CSI driver to v1.18.0 ([#2058](https://github.com/kubermatic/kubeone/pull/2058), [@xmudrii](https://github.com/xmudrii))

#### DigitalOcean

- Update DigitalOcean CSI driver to v4.2.0 ([#2173](https://github.com/kubermatic/kubeone/pull/2173), [@xmudrii](https://github.com/xmudrii))
- Update the DigitalOcean CCM to v0.1.37 ([#2053](https://github.com/kubermatic/kubeone/pull/2053), [@xmudrii](https://github.com/xmudrii))

#### Equinix Metal

- Update Equinix Metal CCM to v3.4.3 ([#2174](https://github.com/kubermatic/kubeone/pull/2174), [@xmudrii](https://github.com/xmudrii))

#### Nutanix

- Update the Nutanix CSI driver to v2.5.1 ([#2182](https://github.com/kubermatic/kubeone/pull/2182), [@xmudrii](https://github.com/xmudrii))

#### GCP

- Update GCP PD CSI driver to v1.7.2 ([#2176](https://github.com/kubermatic/kubeone/pull/2176), [@xmudrii](https://github.com/xmudrii))


#### OpenStack

- Update OpenStack CCM and Cinder CSI to v1.24.2 for Kubernetes 1.24 clusters and v1.23.4 for Kubernetes 1.23 clusters ([#2195](https://github.com/kubermatic/kubeone/pull/2195), [@xmudrii](https://github.com/xmudrii))
- Update OpenStack CCM and Cinder CSI to v1.24.0 for Kubernetes 1.24 clusters ([#2061](https://github.com/kubermatic/kubeone/pull/2061), [@xmudrii](https://github.com/xmudrii))

#### vSphere

- Update vSphere CSI driver to v2.6.0 ([#2169](https://github.com/kubermatic/kubeone/pull/2169), [@xmudrii](https://github.com/xmudrii))
- Update vSphere CCM to v1.24.0 for Kubernetes 1.24+ clusters. Update vSphere CCM to v1.23.1 for Kubernetes 1.23 clusters ([#2169](https://github.com/kubermatic/kubeone/pull/2169), [@xmudrii](https://github.com/xmudrii))
- Update the vSphere CCM to v1.23.0, v1.22.6, v1.21.3, v1.20.1. Update the vSphere CSI driver to v2.5.1
  - The maximum Kubernetes version for vSphere clusters has been increased from 1.22 to 1.25
  - Apply credentials and cloud-config Secrets before applying addons. This ensures that addons depending on those Secrets are applied properly ([#2050](https://github.com/kubermatic/kubeone/pull/2050), [@xmudrii](https://github.com/xmudrii))

#### Other Addons

- Update metrics-server to v0.6.1. The listen port for metrics-server has been changed from 443 to 4443. This change shouldn't affect you if you see the metrics-server Service ([#2079](https://github.com/kubermatic/kubeone/pull/2079), [@xmudrii](https://github.com/xmudrii))
- Update NodeLocalDNS Cache to v1.21.1 ([#2079](https://github.com/kubermatic/kubeone/pull/2079), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler to the latest available releases ([#2175](https://github.com/kubermatic/kubeone/pull/2175), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler to v1.24.0, v1.23.0, v1.22.2, v1.21.2, v1.20.2 ([#2052](https://github.com/kubermatic/kubeone/pull/2052), [@xmudrii](https://github.com/xmudrii))

### Terraform Integration

#### General

- Automate generating terraform configs README files ([#2117](https://github.com/kubermatic/kubeone/pull/2117), [@kron4eg](https://github.com/kron4eg))
- `initial_machinedeployment_operating_system_profile` was added to specify operating system profile for initial MachineDeployments. ([#2097](https://github.com/kubermatic/kubeone/pull/2097), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

#### AWS

- Introduce `initial_machinedeployment_spotinstances_max_price` in example Terraform configs for AWS. When set, spot instances will be used for initial MachineDeployments ([#1924](https://github.com/kubermatic/kubeone/pull/1924), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Example Terraform configs for AWS are now using Ignition instead of cloud-init for Flatcar worker nodes ([#2157](https://github.com/kubermatic/kubeone/pull/2157), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Let OSM default the OperatingSystemProfiles (OSPs) in the example Terraform configs for AWS ([#2198](https://github.com/kubermatic/kubeone/pull/2198), [@kron4eg](https://github.com/kron4eg))

#### Other providers

- Increases default MachineDeployment replicas to 2 for all non-AWS Terraform configs ([#2159](https://github.com/kubermatic/kubeone/pull/2159), [@xmudrii](https://github.com/xmudrii))
- Update Terraform integration for Azure with new fields ([#2081](https://github.com/kubermatic/kubeone/pull/2081), [@xmudrii](https://github.com/xmudrii))
- Terraform configs for GCP are now using the default network instead of creating a new one. For production usage, it's recommended to modify configs to create a dedicated network for your cluster. ([#2143](https://github.com/kubermatic/kubeone/pull/2143), [@kron4eg](https://github.com/kron4eg))
- Add vSphere anti-affinity rule for the control plane to avoid a single point of failure. ([#2124](https://github.com/kubermatic/kubeone/pull/2124), [@mihiragrawal](https://github.com/mihiragrawal))

### Bug or Regression

#### General

- Set `rp_filter=0` on all interfaces when Cilium is used. This fixes an issue with Cilium clusters losing pod connectivity after upgrading the cluster ([#2089](https://github.com/kubermatic/kubeone/pull/2089), [@xmudrii](https://github.com/xmudrii))
- Approve pending CSRs when upgrading control plane and static worker nodes ([#1887](https://github.com/kubermatic/kubeone/pull/1887), [@xmudrii](https://github.com/xmudrii))
- Force regenerating CSRs for Kubelet serving certificates after CCM is deployed. This fixes an issue with Kubelet generating CSRs that are stuck in Pending. ([#2199](https://github.com/kubermatic/kubeone/pull/2199), [@xmudrii](https://github.com/xmudrii))
- Fix CSR approving issue for existing nodes with already approved and GCed CSRs ([#1894](https://github.com/kubermatic/kubeone/pull/1894), [@kron4eg](https://github.com/kron4eg))
- Fix wrong maxPods value on follower control plane nodes and static worker nodes ([#2112](https://github.com/kubermatic/kubeone/pull/2112), [@xmudrii](https://github.com/xmudrii))
- Fix KubeletConfiguration and KubeProxyConfiguration for Kubernetes prior v1.23.x ([#2138](https://github.com/kubermatic/kubeone/pull/2138), [@kron4eg](https://github.com/kron4eg))
- Fix missing reading of the static workers defined in Terraform ([#2015](https://github.com/kubermatic/kubeone/pull/2015), [@kron4eg](https://github.com/kron4eg))
- Fix containerd upgrade on Debian-based distros ([#1930](https://github.com/kubermatic/kubeone/pull/1930), [@kron4eg](https://github.com/kron4eg))
- Fix NPE on SSH connection close ([#2154](https://github.com/kubermatic/kubeone/pull/2154), [@kron4eg](https://github.com/kron4eg))
- Fix the GoBetween script failing to install the zip package on Flatcar Linux ([#1904](https://github.com/kubermatic/kubeone/pull/1904), [@xmudrii](https://github.com/xmudrii))
- Fix issue with `installer.sh` on mac (BSD sed) ([#2161](https://github.com/kubermatic/kubeone/pull/2161), [@dermorz](https://github.com/dermorz))
- Fix "latest version" in `install.sh`. ([#1949](https://github.com/kubermatic/kubeone/pull/1949), [@dermorz](https://github.com/dermorz))
- Fix an issue with `kubeone config migrate` failing to migrate configs with the `containerRuntime` block ([#1860](https://github.com/kubermatic/kubeone/pull/1860), [@xmudrii](https://github.com/xmudrii))
- Fix overwriteRegistry not overwriting the Kubernetes control plane images ([#1884](https://github.com/kubermatic/kubeone/pull/1884), [@xmudrii](https://github.com/xmudrii))
- Fix pre-pull images ([#2029](https://github.com/kubermatic/kubeone/pull/2029), [@kron4eg](https://github.com/kron4eg))
- Use kubeadm config when pre-pulling images ([#2026](https://github.com/kubermatic/kubeone/pull/2026), [@kron4eg](https://github.com/kron4eg))
- Add missing `volumeattachments` permissions to machine-controller ([#2031](https://github.com/kubermatic/kubeone/pull/2031), [@kron4eg](https://github.com/kron4eg))
- Avoid creating and validating MC credentials when MC is disabled ([#1939](https://github.com/kubermatic/kubeone/pull/1939), [@kron4eg](https://github.com/kron4eg))
- Ensure old machine-controller MutatingWebhookConfiguration is deleted ([#1900](https://github.com/kubermatic/kubeone/pull/1900), [@kron4eg](https://github.com/kron4eg))
- Escape docker/containerd versions to avoid wildcard matching ([#1941](https://github.com/kubermatic/kubeone/pull/1941), [@kron4eg](https://github.com/kron4eg))
- Expand path to SSH private key file ([#1849](https://github.com/kubermatic/kubeone/pull/1849), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Add missing `systemctl daemon-reload` when removing binaries ([#2064](https://github.com/kubermatic/kubeone/pull/2064), [@kron4eg](https://github.com/kron4eg))
- Regenerate container runtime configurations based on KubeOneCluster manifest during control plane upgrades on Flatcar Linux nodes, not only on the initial installation. ([#1910](https://github.com/kubermatic/kubeone/pull/1910), [@dermorz](https://github.com/dermorz))
- Remove the `--network-plugin` Kubelet flag when migrating from Docker to containerd and when upgrading from Kubernetes 1.23.x to 1.24.x ([#2024](https://github.com/kubermatic/kubeone/pull/2024), [@xmudrii](https://github.com/xmudrii))
- Restart kubelet after upgrading containerd ([#1944](https://github.com/kubermatic/kubeone/pull/1944), [@kron4eg](https://github.com/kron4eg))
- Update `kubeadm-flags.env` file when upgrading static worker nodes ([#2123](https://github.com/kubermatic/kubeone/pull/2123), [@xmudrii](https://github.com/xmudrii))
- Don't ignore clientset error when resetting cluster ([#1950](https://github.com/kubermatic/kubeone/pull/1950), [@xmudrii](https://github.com/xmudrii))
- Show "Ensure MachineDeployments" as an action to be taken only when provisioning a cluster for the first time ([#1927](https://github.com/kubermatic/kubeone/pull/1927), [@xmudrii](https://github.com/xmudrii))
- Lower exponential backoff times ([#2231](https://github.com/kubermatic/kubeone/pull/2231), [@kron4eg](https://github.com/kron4eg))

#### Addons

- Properly propagate external cloud provider and CSI migration options to OSM ([#2202](https://github.com/kubermatic/kubeone/pull/2202), [@xmudrii](https://github.com/xmudrii))
- Replace `operator: Exists` toleration with the control plane tolerations for metrics-server. This fixes an issue with metrics-server pods breaking eviction ([#2205](https://github.com/kubermatic/kubeone/pull/2205), [@xmudrii](https://github.com/xmudrii))
- Fix the logic for determining if the CSI driver is deployed in the default-storage-class addon. This fixes an issue with deploying the default-storage-class addon on vSphere clusters using the in-tree cloud provider ([#2167](https://github.com/kubermatic/kubeone/pull/2167), [@xmudrii](https://github.com/xmudrii))
- Azure: Migrate AzureDisk CSIDriver to set fsGroupPolicy to File ([#2082](https://github.com/kubermatic/kubeone/pull/2082), [@xmudrii](https://github.com/xmudrii))
- Azure: Disable `--configure-cloud-routes` on Azure CCM to fix errors when starting the CCM ([#2184](https://github.com/kubermatic/kubeone/pull/2184), [@xmudrii](https://github.com/xmudrii))
- Azure: Disable node IPAM in Azure CCM ([#2106](https://github.com/kubermatic/kubeone/pull/2106), [@rastislavs](https://github.com/rastislavs))
- GCE: Migrate GCE `standard` default StorageClass to set volumeBindingMode to WaitForFirstConsumer. The StorageClass will be automatically recreated the next time you run `kubeone apply` ([#2142](https://github.com/kubermatic/kubeone/pull/2142), [@xmudrii](https://github.com/xmudrii))
- Hetzner: Disable Node IPAM in Hetzner CCM. This fixes network connectivity issues on the worker nodes. ([#2200](https://github.com/kubermatic/kubeone/pull/2200), [@xmudrii](https://github.com/xmudrii))
- OpenStack: Tenant ID or Tenant Name is not required when using application credentials ([#2196](https://github.com/kubermatic/kubeone/pull/2196), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- OpenStack: Mount `/usr/share/ca-certificates` to the OpenStack CCM pod to fix the OpenStack CCM pod CrashLooping on Flatcar Linux ([#1904](https://github.com/kubermatic/kubeone/pull/1904), [@xmudrii](https://github.com/xmudrii))
- Disable `preserveUnknownFields` in all Canal CRDs. This fixes an issue preventing upgrading Canal to v3.22 for KubeOne clusters created with KubeOne 1.2 and older ([#2103](https://github.com/kubermatic/kubeone/pull/2103), [@xmudrii](https://github.com/xmudrii))

### Other

- Remove changelog from the release archive. Changelogs can be found on GitHub in the [CHANGELOG directory](https://github.com/kubermatic/kubeone/tree/master/CHANGELOG) ([#2213](https://github.com/kubermatic/kubeone/pull/2213), [@xmudrii](https://github.com/xmudrii))
