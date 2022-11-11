# [v1.4.11](https://github.com/kubermatic/kubeone/releases/tag/v1.4.11) - 2022-11-11

## Important Registry Change Information

For the next series of KubeOne and KKP patch releases, image references will move from `k8s.gcr.io` to `registry.k8s.io`. This will be done to keep up with [the latest upstream changes](https://github.com/kubernetes/enhancements/tree/master/keps/sig-release/3000-artifact-distribution). Please ensure that any mirrors you use are able to host `registry.k8s.io` and/or that firewall rules are going to allow access to `registry.k8s.io` to pull images before applying the next KubeOne patch releases. **This is not included in this patch release but just a notification of future changes.**

## Important Security Information

**Kubernetes releases prior to 1.25.4, 1.24.8, 1.23.14, and 1.22.16 are affected by two Medium CVEs in kube-apiserver**: [CVE-2022-3162 (Unauthorized read of Custom Resources)](https://groups.google.com/g/kubernetes-announce/c/oR2PUBiODNA/m/tShPgvpUDQAJ) and [CVE-2022-3294 (Node address isn't always verified when proxying)](https://groups.google.com/g/kubernetes-announce/c/eR0ghAXy2H8/m/sCuQQZlVDQAJ). We **strongly recommend** upgrading to 1.25.4, 1.24.8, 1.23.14, or 1.22.16 **as soon as possible**.

## Changelog since v1.4.10

## Changes by Kind

### Feature

- Update etcd to 3.5.5 for Kubernetes 1.22+ clusters or use the version provided by kubeadm if it's newer ([#2444](https://github.com/kubermatic/kubeone/pull/2444), [@xmudrii](https://github.com/xmudrii))

### Other (Cleanup or Flake)

- Expose machine-controller metrics port (8080/TCP), so Prometheus ServiceMonitor can be used for scraping ([#2440](https://github.com/kubermatic/kubeone/pull/2440), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Chore

- KubeOne is now built using Go 1.18.8 ([#2465](https://github.com/kubermatic/kubeone/pull/2465), [@xmudrii](https://github.com/xmudrii))
- The `kubeone-e2e` image is moved from Docker Hub to Quay (`quay.io/kubermatic/kubeone-e2e`) ([#2465](https://github.com/kubermatic/kubeone/pull/2465), [@xmudrii](https://github.com/xmudrii))

# [v1.4.10](https://github.com/kubermatic/kubeone/releases/tag/v1.4.10) - 2022-10-20

## Changelog since v1.4.9

## Changes by Kind

### Bug or Regression

- Update `golang.org/x/crypto` dependency to a newer version to fix issues with SSH authentication on instances with newer OpenSSH versions ([#2390](https://github.com/kubermatic/kubeone/pull/2390), [@xmudrii](https://github.com/xmudrii))

# [v1.4.9](https://github.com/kubermatic/kubeone/releases/tag/v1.4.9) - 2022-09-26

## Changelog since v1.4.8

## Changes by Kind

### Feature

- Update the `kubernetes-cni` package from 0.8.7 to 1.1.1 to support the latest Kubernetes patch releases ([#2358](https://github.com/kubermatic/kubeone/pull/2358), [@xmudrii](https://github.com/xmudrii))

# [v1.4.8](https://github.com/kubermatic/kubeone/releases/tag/v1.4.8) - 2022-08-29

## Changelog since v1.4.7

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Update machine-controller to v1.43.7. This update fixes several issues for RHEL clusters on Azure. If you have RHEL-based MachineDeployments on Azure, we **strongly recommend** upgrading to KubeOne 1.4.8 and rotating those MachineDeployments **BEFORE** upgrading to KubeOne 1.5. **If not done, the Canal CNI update might break the cluster networking when upgrading to KubeOne 1.5.** ([#2333](https://github.com/kubermatic/kubeone/pull/2333), [@xmudrii](https://github.com/xmudrii))

## Changes by Kind

### Bug or Regression

- Mount `/etc/pki` to the OpenStack CCM container to fix CrashLoopBackoff on clusters running CentOS 7 ([#2303](https://github.com/kubermatic/kubeone/pull/2303), [@xmudrii](https://github.com/xmudrii))
- Explicitly create `/opt/bin` on Flatcar before trying to untar anything to that directory ([#2305](https://github.com/kubermatic/kubeone/pull/2305), [@xmudrii](https://github.com/xmudrii))
- Mount `/etc/pki` to the Azure CCM container to fix CrashLoopBackoff on clusters running CentOS 7 and Rocky Linux ([#2310](https://github.com/kubermatic/kubeone/pull/2310), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Mount `/usr/share/ca-certificates` to the Azure CCM container to fix CrashLoopBackoff on clusters running Flatcar ([#2334](https://github.com/kubermatic/kubeone/pull/2334), [@xmudrii](https://github.com/xmudrii))
- Set iptables backend (`FELIX_IPTABLESBACKEND`) to `NFT` for Canal and Calico VXLAN on clusters running Flatcar Linux and RHEL. For non Flatcar/RHEL clusters, iptables backend is set to Auto, which is the default value and results in Calico determining the iptables backend automatically. The value can be overridden by setting the `iptablesBackend` addon parameter (see the PR description for an example). ([#2334](https://github.com/kubermatic/kubeone/pull/2334), [@xmudrii](https://github.com/xmudrii))

# [v1.4.7](https://github.com/kubermatic/kubeone/releases/tag/v1.4.7) - 2022-08-16

## Changes by Kind

### Bug or Regression

- Enable `nf_conntrack` (`nf_conntrack_ipv4`) module by default on all operating systems. This fixes an issue with pods unable to reach services running on a host on operating systems that are using the NFT backend. ([#2283](https://github.com/kubermatic/kubeone/pull/2283), [@xmudrii](https://github.com/xmudrii))

### Terraform Integration

#### AWS

- Remove defaulting for the Flatcar provisioning utility in example Terraform configs for AWS (defaulted to cloud-init by machine-controller) ([#2286](https://github.com/kubermatic/kubeone/pull/2286), [@xmudrii](https://github.com/xmudrii))

# [v1.4.6](https://github.com/kubermatic/kubeone/releases/tag/v1.4.6) - 2022-08-03

## Changes by Kind

### Feature

- Add missing snapshot controller and webhook for OpenStack Cinder CSI ([#2218](https://github.com/kubermatic/kubeone/pull/2218), [@xmudrii](https://github.com/xmudrii))
- Rollout pods that are using `kubeone-*-credentials` Secrets if credentials are changed ([#2216](https://github.com/kubermatic/kubeone/pull/2216), [@xmudrii](https://github.com/xmudrii))

### Updates

- Update containerd to v1.5. Escape docker/containerd versions to avoid wildcard matching ([#2228](https://github.com/kubermatic/kubeone/pull/2228), [@xmudrii](https://github.com/xmudrii))
- Update Canal to v3.22.4 ([#2189](https://github.com/kubermatic/kubeone/pull/2189), [@xmudrii](https://github.com/xmudrii))
- Update OpenStack CCM and Cinder CSI to v1.23.4 for Kubernetes 1.23 clusters ([#2186](https://github.com/kubermatic/kubeone/pull/2186), [@xmudrii](https://github.com/xmudrii))
- Update machine-controller to v1.43.6 ([#2227](https://github.com/kubermatic/kubeone/pull/2227), [@xmudrii](https://github.com/xmudrii))
- Update machine-controller to v1.43.5 ([#2210](https://github.com/kubermatic/kubeone/pull/2210), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to v1.43.4. This machine-controller release fixes an issue with finding Node objects by ProviderID ([#2193](https://github.com/kubermatic/kubeone/pull/2193), [@xmudrii](https://github.com/xmudrii))

### Bug or Regression

- Disable `--configure-cloud-routes` on Azure CCM to fix errors when starting the CCM ([#2185](https://github.com/kubermatic/kubeone/pull/2185), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Force regenerating CSRs for Kubelet serving certificates after CCM is deployed. This fixes an issue with Kubelet generating CSRs that are stuck in Pending. ([#2204](https://github.com/kubermatic/kubeone/pull/2204), [@xmudrii](https://github.com/xmudrii))
- Properly propagate external cloud provider and CSI migration options to OSM ([#2203](https://github.com/kubermatic/kubeone/pull/2203), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Replace `operator: Exists` toleration with the control plane tolerations for metrics-server. This fixes an issue with metrics-server pods breaking eviction ([#2206](https://github.com/kubermatic/kubeone/pull/2206), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Tenant ID or Name is not required when using application credentials ([#2201](https://github.com/kubermatic/kubeone/pull/2201), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

# [v1.4.5](https://github.com/kubermatic/kubeone/releases/tag/v1.4.5) - 2022-07-12

## Changes by Kind

### Feature

- Add GCP Compute Persistent Disk CSI driver. The CSI driver is deployed by default for all GCE clusters running Kubernetes 1.23 or newer ([#2141](https://github.com/kubermatic/kubeone/pull/2141), [@xmudrii](https://github.com/xmudrii))
- Migrate GCE `standard` default StorageClass to set volumeBindingMode to WaitForFirstConsumer. The StorageClass will be automatically recreated the next time you run `kubeone apply` ([#2141](https://github.com/kubermatic/kubeone/pull/2141), [@xmudrii](https://github.com/xmudrii))

### Bug or Regression

- Disable node IPAM in Azure CCM ([#2107](https://github.com/kubermatic/kubeone/pull/2107), [@rastislavs](https://github.com/rastislavs))
- Disable preserveUnknownFields in all Canal CRDs. This fixes an issue preventing upgrading Canal to v3.22 for KubeOne clusters created with KubeOne 1.2 and older ([#2105](https://github.com/kubermatic/kubeone/pull/2105), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix wrong maxPods value on follower control plane nodes and static worker nodes ([#2128](https://github.com/kubermatic/kubeone/pull/2128), [@xmudrii](https://github.com/xmudrii))
- Set rp_filter=0 on all interfaces when Cilium is used. This fixes an issue with Cilium clusters losing pod connectivity after upgrading the cluster ([#2108](https://github.com/kubermatic/kubeone/pull/2108), [@xmudrii](https://github.com/xmudrii))

# [v1.4.4](https://github.com/kubermatic/kubeone/releases/tag/v1.4.4) - 2022-06-02

## Changes by Kind

### Feature

- Add MaxPods field to the KubeletConfig used to control the maximum number of pods per node ([#2080](https://github.com/kubermatic/kubeone/pull/2080), [@xmudrii](https://github.com/xmudrii))
- Update machine-controller to v1.43.3 ([#2080](https://github.com/kubermatic/kubeone/pull/2080), [@xmudrii](https://github.com/xmudrii))
- Add `machineObjectAnnotations` field to DynamicWorkerNodes used to apply annotations to resulting Machine objects. Add `nodeAnnotations` field to DynamicWorkerNodes Config as a replacement for deprecated `machineAnnotations` field ([#2077](https://github.com/kubermatic/kubeone/pull/2077), [@xmudrii](https://github.com/xmudrii))
- Update Canal and Calico VXLAN addons to v3.22.2. This allows users to use kube-proxy in IPVS mode on AMD64 clusters running Kubernetes 1.23 and newer. It currently remains impossible to use kube-proxy in IPVS mode on ARM64 clusters running Kubernetes 1.23 and newer. ([#2042](https://github.com/kubermatic/kubeone/pull/2042), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Update Terraform integration for Azure with new fields ([#2085](https://github.com/kubermatic/kubeone/pull/2085), [@xmudrii](https://github.com/xmudrii))
- Update vSphere CCM to v1.23.0 for Kubernetes 1.23 clusters. Add support for Kubernetes 1.23 on vSphere ([#2069](https://github.com/kubermatic/kubeone/pull/2069), [@xmudrii](https://github.com/xmudrii))

### Bug or Regression

- Migrate AzureDisk CSIDriver to set fsGroupPolicy to File ([#2086](https://github.com/kubermatic/kubeone/pull/2086), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.4.3](https://github.com/kubermatic/kubeone/releases/tag/v1.4.3) - 2022-05-11

## Changes by Kind

### Bug or Regression

- Add missing VolumeAttachments permissions to machine-controller ([#2032](https://github.com/kubermatic/kubeone/pull/2032), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Provide registry configuration to kubeadm when pre-pulling images ([#2028](https://github.com/kubermatic/kubeone/pull/2028), [@kron4eg](https://github.com/kron4eg))

# [v1.4.2](https://github.com/kubermatic/kubeone/releases/tag/v1.4.2) - 2022-04-26

## Attention Needed

This patch releases updates etcd to v3.5.3 which includes a fix for the data inconsistency issues reported earlier (https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ). To upgrade etcd for an existing cluster, you need to [force upgrade the cluster as described here](https://docs.kubermatic.com/kubeone/v1.4/guides/etcd_corruption/#enabling-etcd-corruption-checks). If you're running Kubernetes 1.22 or newer, we strongly recommend upgrading etcd **as soon as possible**.

## Changes by Kind

### Feature

- Domain is not required when using application credentials ([#1938](https://github.com/kubermatic/kubeone/pull/1938), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

### Bug or Regression

- Bump flannel image to v0.15.1 ([#1993](https://github.com/kubermatic/kubeone/pull/1993), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Deploy etcd v3.5.3 for clusters running Kubernetes 1.22 or newer. etcd v3.5.3 includes a fix for [the data inconsistency issues announced by the etcd maintainers](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ. To upgrade etcd) for an existing cluster, you need to [force upgrade the cluster as described here](https://docs.kubermatic.com/kubeone/v1.4/guides/etcd_corruption/#enabling-etcd-corruption-checks) ([#1953](https://github.com/kubermatic/kubeone/pull/1953)
- Fixes containerd upgrade on deb based distros ([#1935](https://github.com/kubermatic/kubeone/pull/1935), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Show "Ensure MachineDeployments" as an action to be taken only when provisioning a cluster for the first time ([#1931](https://github.com/kubermatic/kubeone/pull/1931), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Update machine-controller to v1.43.2 ([#2001](https://github.com/kubermatic/kubeone/pull/2001), [@kron4eg](https://github.com/kron4eg))
  - Fixes an issue where the machine-controller would not wait for the volumeAttachments deletion before deleting the node
  - Fixes an issue where masked services on Flatcar are not properly stopped when provisioning a Flatcar node

# [v1.4.1](https://github.com/kubermatic/kubeone/releases/tag/v1.4.1) - 2022-04-04

## Attention Needed

This patch release enables the etcd corruption checks on every etcd member that is running etcd 3.5 (which applies to all Kubernetes 1.22+ clusters). This change is a [recommendation from the etcd maintainers](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ) due to issues in etcd 3.5 that can cause data consistency issues. The changes in this patch release will prevent corrupted etcd members from joining or staying in the etcd ring.

## Changes by Kind

### Bug or Regression

- Regenerate container runtime configurations based on kubeone.yaml during control-plane upgrades on Flatcar Linux nodes, not only on the initial installation. ([#1918](https://github.com/kubermatic/kubeone/pull/1918))
- Approve pending CSRs when upgrading control plane and static worker nodes ([#1888](https://github.com/kubermatic/kubeone/pull/1888))
- Enable the etcd integrity checks (on startup and every 4 hours) for Kubernetes 1.22+ clusters. See [the official etcd announcement for more details](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ). ([#1909](https://github.com/kubermatic/kubeone/pull/1909))
- Fix CSR approving issue for existing nodes with already approved and GCed CSRs ([#1897](https://github.com/kubermatic/kubeone/pull/1897))
- Fix missing snapshot CRDs for Openstack CSI ([#1913](https://github.com/kubermatic/kubeone/pull/1913))
- Ensure old machine-controller MutatingWebhookConfiguration is deleted ([#1913](https://github.com/kubermatic/kubeone/pull/1913))
- Fix overwriteRegistry not overwriting the Kubernetes control plane images ([#1885](https://github.com/kubermatic/kubeone/pull/1885))
- Mount /usr/share/ca-certificates to the OpenStack CCM pod to fix the OpenStack CCM pod CrashLooping on Flatcar Linux ([#1905](https://github.com/kubermatic/kubeone/pull/1905))
- Fix the GoBetween script failing to install the zip package on Flatcar Linux ([#1905](https://github.com/kubermatic/kubeone/pull/1905))
- Expand path to SSH private key file ([#1859](https://github.com/kubermatic/kubeone/pull/1859))
- Fix an issue with `kubeone config migrate` failing to migrate configs with the `containerRuntime` block ([#1861](https://github.com/kubermatic/kubeone/pull/1861))

# [v1.4.0](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0) - 2022-02-16

## Attention Needed

Check out the [Upgrading from 1.3 to 1.4 tutorial](https://docs.kubermatic.com/kubeone/v1.4/tutorials/upgrading/upgrading_from_1.3_to_1.4/) for more details about the breaking changes and how to mitigate them.

* KubeOne 1.4.0-beta.0 introduces the new KubeOneCluster v1beta2 API
  * The existing KubeOneCluster v1beta1 manifests can be migrated by using the `kubeone config migrate` command
  * The `kubeone config print` command now uses the new v1beta2 API
  * The existing KubeOneCluster v1beta1 API is considered deprecated and will be removed in KubeOne 1.6+
  * Highlights:
    * The API group has been changed from `kubeone.io` to `kubeone.k8c.io`
    * The AssetConfiguration API has been removed from the v1beta2 API. The AssetConfiguration API can still be used with the v1beta1 API, but we highly recommend migrating away because the v1beta1 API is deprecated
    * The PodPresets feature has been removed from the v1beta2 API because Kubernetes removed support for PodPresets in Kubernetes 1.20
    * Packet (`packet`) cloud provider has been rebranded to Equinix Metal (`equinixmetal`). The existing Packet cluster will work with `equinixmetal` cloud provider, however, [manual migration steps](https://docs.kubermatic.com/kubeone/v1.4/tutorials/upgrading/upgrading_from_1.3_to_1.4/#packet-to-equinix-metal-rebranding) are required if you want to use new Terraform configs for Equinix Metal
    * A new ContainerRuntime API has been added to the v1beta2 API in order to support configuring mirror registries
* `kubeone install` and `kubeone upgrade` commands are considered deprecated in favor of `kubeone apply`
  * `install` and `upgrade` commands will be removed in KubeOne 1.6+
  * We highly encourage switching to `kubeone apply`. The `apply` command has the same semantics and works in the same way as `install`/`upgrade`, with some additional checks to ensure each requested operation is safe for the cluster
* Unconditionally deploy AWS, AzureDisk, AzureFile, and vSphere CSI drivers if the Kubernetes version is 1.23 or newer ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * Those providers have the CSI migration enabled by default in Kubernetes 1.23, so the CSI driver will be used for all volumes operations
* Unconditionally deploy DigitalOcean, Hetzner, Nutanix, and OpenStack Cinder CSI drivers ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * OpenStack has the CSI migration enabled by default since Kubernetes 1.18, so the CSI driver will be used for all operations
* CentOS 8 has reached End-Of-Life (EOL) on January 31st, 2022. It will no longer receive any updates (including security updates). Support for CentOS 8 in KubeOne is deprecated and will be removed in a future release. We strongly recommend migrating to another operating system or RHEL/CentOS distribution as soon as possible.

### Breaking changes / Action Required

* The default AMI for CentOS in Terraform configs for AWS has been changed to Rocky Linux. If you use the new Terraform configs with an existing cluster, make sure to bind the AMI as described in [the production recommendations document](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/) ([#1809](https://github.com/kubermatic/kubeone/pull/1809))
* The `cloud-provider-credentials` Secret is removed by KubeOne because KubeOne does not use it any longer. If you have any workloads **NOT** created by KubeOne that use this Secret, please migrate before upgrading KubeOne. Instead, KubeOne now creates `kubeone-machine-controller-credentials` and `kubeone-ccm-credentials` Secrets used by machine-controller and external CCM
* Support for Amazon EKS-D clusters has been removed starting from this release
* GCP: Default operating system for control plane instances is now Ubuntu 20.04 ([#1576](https://github.com/kubermatic/kubeone/pull/1576))
  * Make sure to bind `control_plane_image_family` to the image you're currently using or Terraform might recreate all your control plane instances
* Azure: Default VM type is changed to `Standard_F2` ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
  * Make sure to bind `control_plane_vm_size` and `worker_vm_size` to the VM size you're currently using or Terraform might recreate all your instances

## Known Issues

* It's not possible to run kube-proxy in IPVS mode on Kubernetes 1.23 clusters using Canal/Calico CNI. Trying to upgrade existing 1.22 clusters using IPVS to 1.23 will result in a validation error from KubeOne

## Added

### API

* Add the KubeOneCluster v1beta2 API and change the API group to `kubeone.k8c.io` ([#1649](https://github.com/kubermatic/kubeone/pull/1649))
  * Make `kubeone config print` command use the new `kubeone.k8c.io/v1beta2` API ([#1651](https://github.com/kubermatic/kubeone/pull/1651))
  * Add the new ContainerRuntime API with support for mirror registries ([#1674](https://github.com/kubermatic/kubeone/pull/1674))
  * Add the Registry Credentials configuration to the RegistryConfiguration API ([#1724](https://github.com/kubermatic/kubeone/pull/1724))
  * Add ability to change the container log maximum size (defaults to 100Mi) ([#1644](https://github.com/kubermatic/kubeone/pull/1644))
  * Add ability to change the container log maximum files (defaults to 5) ([#1759](https://github.com/kubermatic/kubeone/pull/1759))
  * Addons directory path (`.addons.path`) is not required when using only embedded addons ([#1668](https://github.com/kubermatic/kubeone/pull/1668))
  * Addons directory path (`.addons.path`) is not defaulted to `./addons` any longer ([#1668](https://github.com/kubermatic/kubeone/pull/1668))
  * Add the KubeletConfig API used to configure `systemReserved`, `kubeReserved`, and `evictionHard` Kubelet options ([#1698](https://github.com/kubermatic/kubeone/pull/1698))
  * Allow providing operating system via the API ([#1809](https://github.com/kubermatic/kubeone/pull/1809))
  * Remove the PodPresets feature ([#1662](https://github.com/kubermatic/kubeone/pull/1662))
  * Remove the AssetConfiguration API ([#1699](https://github.com/kubermatic/kubeone/pull/1699))
  * Rebrand Packet (`packet`) to Equinix Metal (`equinixmetal`) and support migrating existing Packet clusters to Equinix Metal
  clusters ([#1663](https://github.com/kubermatic/kubeone/pull/1663))

### Features

* Add experimental/alpha-level support for [Kubermatic Operating System Manager (OSM)](https://github.com/kubermatic/operating-system-manager) ([#1748](https://github.com/kubermatic/kubeone/pull/1748))
* Add support for Kubernetes 1.23 ([#1678](https://github.com/kubermatic/kubeone/pull/1678))
* Add support for Cilium CNI ([#1560](https://github.com/kubermatic/kubeone/pull/1560), [#1629](https://github.com/kubermatic/kubeone/pull/1629))
* Add experimental/alpha support for Nutanix ([#1723](https://github.com/kubermatic/kubeone/pull/1723), [#1725](https://github.com/kubermatic/kubeone/pull/1725), [#1733](https://github.com/kubermatic/kubeone/pull/1733))
  * Support for Nutanix is experimental, so implementation and relevant addons might be changed until it doesn't graduate to beta/stable
* Add CCM/CSI migration support for clusters with the static worker nodes ([#1544](https://github.com/kubermatic/kubeone/pull/1544))
* Add CCM/CSI migration support for the Azure clusters ([#1610](https://github.com/kubermatic/kubeone/pull/1610))
* Automatically create cloud-config Secret for all providers if external cloud controller manager (`.cloudProvider.external`) is enabled ([#1575](https://github.com/kubermatic/kubeone/pull/1575))
* Source `.cloudProvider.csiConfig` from the credentials file if present ([#1739](https://github.com/kubermatic/kubeone/pull/1739))
* Fetch containerd auth config from the credentials file if present ([#1745](https://github.com/kubermatic/kubeone/pull/1745))
* Add support for different credentials for machine-controller and CCM. Environment variables can be prefixed with `MC_` for machine-controller credentials and `CCM_` for CCM credentials ([#1717](https://github.com/kubermatic/kubeone/pull/1717))
* Add support for OpenStack Application Credentials ([#1666](https://github.com/kubermatic/kubeone/pull/1666))
* Add support for additional Subject Alternative Names (SANs) for the Kubernetes API server ([#1599](https://github.com/kubermatic/kubeone/pull/1599), [#1603](https://github.com/kubermatic/kubeone/pull/1603), [#1606](https://github.com/kubermatic/kubeone/pull/1606))
* Addon parameters can be resolved into environment variable contents if the `env:` prefix is set in the parameter value ([#1691](https://github.com/kubermatic/kubeone/pull/1691))
* Add `kubeone addons list` command used to list available and enabled addons ([#1642](https://github.com/kubermatic/kubeone/pull/1642))
* Add a new `--kubernetes-version` flag to the `kubeone config images` command ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
  * This flag is used to filter images for a particular Kubernetes version. The flag cannot be used along with the KubeOneCluster manifest (`--manifest` flag)
* Add a new `MachineAnnotations` field in the API used to define annotations in `MachineDeployment.Spec.Template.Spec.Annotations` ([#1601](https://github.com/kubermatic/kubeone/pull/1601))
* Add a new `--create-machine-deployments` flag to the `kubeone apply` command used to control should KubeOne create initial MachineDeployment objects when provisioning the cluster (default is `true`) ([#1617](https://github.com/kubermatic/kubeone/pull/1617))
* Generate and approve CSRs for control plane and static workers nodes. Enable the server TLS bootstrap for control plane and static worker nodes ([#1750](https://github.com/kubermatic/kubeone/pull/1750), [#1758](https://github.com/kubermatic/kubeone/pull/1758))

### Addons

* Integrate the AWS CCM addon with KubeOne ([#1585](https://github.com/kubermatic/kubeone/pull/1585))
  * The AWS CCM is now deployed if the external cloud provider (`.cloudProvider.external`) is enabled
  * This option cannot be enabled for existing AWS clusters running in-tree cloud provider, instead, those clusters must go through the CCM/CSI migration process
* Add the AWS EBS CSI driver addon ([#1597](https://github.com/kubermatic/kubeone/pull/1597))
  * Automatically deploy the AWS EBS CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AWS EBS CSI driver to the `default-storage-class` embedded addon
* Integrate the Azure CCM addon with KubeOne ([#1561](https://github.com/kubermatic/kubeone/pull/1561), [#1579](https://github.com/kubermatic/kubeone/pull/1579))
  * The Azure CCM is now deployed if the external cloud provider (`.cloudProvider.external`) is enabled
  * This option cannot be enabled for existing Azure clusters running in-tree cloud provider, instead, those clusters must go through the CCM/CSI migration process
* Add the AzureFile CSI driver addon ([#1575](https://github.com/kubermatic/kubeone/pull/1575), [#1579](https://github.com/kubermatic/kubeone/pull/1579))
  * Automatically deploy the AzureFile CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AzureFile CSI driver to the `default-storage-class` embedded addon
* Add the AzureDisk CSI driver addon ([#1577](https://github.com/kubermatic/kubeone/pull/1577))
  * Automatically deploy the AzureDisk CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AzureDisk CSI driver to the `default-storage-class` embedded addon
* Add the DigitalOcean CSI driver ([#1754](https://github.com/kubermatic/kubeone/pull/1754))
  * The CSI driver is deployed automatically if `.cloudProvider.external` is enabled
  * Add the default StorageClass and VolumeSnapshotClass for the DigitalOcean CSI driver. The StorageClass and VolumeSnapshotClass can be deployed by enabling the default-storage-class embedded addon
* Add the Nutanix CSI driver addon ([#1733](https://github.com/kubermatic/kubeone/pull/1733), [#1734](https://github.com/kubermatic/kubeone/pull/1734))
  * The addon is deployed manually, on-demand, by enabling the `csi-nutanix` embedded addon (see the PR description for more details and examples)
  * Add the default StorageClass for the Nutanix CSI driver. The StorageClass can be deployed by enabling the `default-storage-class` embedded addon (see the PR description for more details and examples)

### Other

* Include darwin/arm64 and linux/arm64 builds in release artifacts ([#1821](https://github.com/kubermatic/kubeone/pull/1821))
* Add a deprecation warning for PodSecurityPolicies ([#1595](https://github.com/kubermatic/kubeone/pull/1595))

## Changed

### General

* Increase the minimum Kubernetes version to 1.20 ([#1818](https://github.com/kubermatic/kubeone/pull/1818))
* Validate the Kubernetes version against supported versions constraints ([#1808](https://github.com/kubermatic/kubeone/pull/1808))
* Allow Docker as a container runtime up to Kubernetes v1.24 (previously up to v1.22) ([#1826](https://github.com/kubermatic/kubeone/pull/1826))
* Validate the cluster name to ensure it's a correct DNS subdomain (RFC 1123) ([#1641](https://github.com/kubermatic/kubeone/pull/1641), [#1646](https://github.com/kubermatic/kubeone/pull/1646), [#1648](https://github.com/kubermatic/kubeone/pull/1648))
* **[BREAKING]** The `cloud-provider-credentials` Secret is removed by KubeOne because KubeOne does not use it any longer. If you have any workloads NOT created by KubeOne that use this Secret, please migrate before upgrading KubeOne. Instead, KubeOne now creates `kubeone-machine-controller-credentials` and `kubeone-ccm-credentials` Secrets used by machine-controller and external CCM ([#1717](https://github.com/kubermatic/kubeone/pull/1717), [#1718](https://github.com/kubermatic/kubeone/pull/1718))
* Unconditionally deploy AWS, AzureDisk, AzureFile, and vSphere CSI drivers if the Kubernetes version is 1.23 or newer ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * Those providers have the CSI migration enabled by default in Kubernetes 1.23, so the CSI driver will be used for all volumes operations
* Unconditionally deploy DigitalOcean, Hetzner, Nutanix, and OpenStack Cinder CSI drivers ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * OpenStack has the CSI migration enabled by default since Kubernetes 1.18, so the CSI driver will be used for all operations
* Improve installation scripts used to install container runtime ([#1664](https://github.com/kubermatic/kubeone/pull/1664))
* Create MachineDeployments only for newly-provisioned clusters ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Show warning about LBs on CCM migration for OpenStack clusters ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Change default Kubernetes version in the example configuration to v1.22.3 ([#1605](https://github.com/kubermatic/kubeone/pull/1605))

### Fixed

* Change baseurl to `vault.centos.org` for CentOS 8 ([#1767](https://github.com/kubermatic/kubeone/pull/1767))
* Fix Docker to containerd migration on non-Flatcar operating systems ([#1743](https://github.com/kubermatic/kubeone/pull/1743))
* Fix propagation of proxy config to machines and Kubernetes components ([#1746](https://github.com/kubermatic/kubeone/pull/1746))
* Fix a bug with the addons applier applying all files when the addons path is not provided ([#1733](https://github.com/kubermatic/kubeone/pull/1733))
* Fix issues when disabling nm-cloud-setup on RHEL ([#1706](https://github.com/kubermatic/kubeone/pull/1706))
* cri-tools is now installed automatically as a dependency of kubeadm on Amazon Linux 2. This fixes provisioning issues on Amazon Linux 2 with newer Kubernetes versions. ([#1701](https://github.com/kubermatic/kubeone/pull/1701))
* Fix the image loader script to support KubeOne 1.3+ and Kubernetes 1.22+ ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
* The `kubeone config images` command now shows images for the latest Kubernetes version (instead of for the oldest) ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
* Allow pods with the seccomp profile defined to get scheduled if the PodSecurityPolicy (PSP) feature is enabled ([#1686](https://github.com/kubermatic/kubeone/pull/1686))
* Force drain nodes to remove standalone pods ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Check for minor version when choosing kubeadm API version ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Provide `--cluster-name` flag to the OpenStack external CCM (read PR description for more details) ([#1619](https://github.com/kubermatic/kubeone/pull/1619))
* Enable ip_tables related kernel modules and disable `nm-cloud-setup` tool on AWS for RHEL machines ([#1607](https://github.com/kubermatic/kubeone/pull/1607))
* Properly pass machine-controllers args ([#1594](https://github.com/kubermatic/kubeone/pull/1594))
  * This fixes the issue causing machine-controller and machine-controller-webhook deployments to run with incorrect flags
  * If you created your cluster with KubeOne 1.2 or older, and already upgraded to KubeOne 1.3, we recommend running kubeone apply again with KubeOne 1.3.2 or newer to properly reconcile machine-controller deployments
* Fix `yum versionlock delete containerd.io` error ([#1600](https://github.com/kubermatic/kubeone/pull/1600))
* Ensure containerd/docker be upgraded automatically when running kubeone apply ([#1589](https://github.com/kubermatic/kubeone/pull/1589))
* Edit SELinux config file only if file exists ([#1532](https://github.com/kubermatic/kubeone/pull/1532))
* Restore missing addons deploy after containerd migration ([#1824](https://github.com/kubermatic/kubeone/pull/1824))
* Select correct CSR to approve ([#1813](https://github.com/kubermatic/kubeone/pull/1813))
* Don't upgrade MachineDeployments when refreshing resources ([#1840](https://github.com/kubermatic/kubeone/pull/1840))
* Provide the correct path to the vSphere CSI webhook certs ([#1845](https://github.com/kubermatic/kubeone/pull/1845))

### Addons

* Replace Hubble static certificate with CronJob generation ([#1752](https://github.com/kubermatic/kubeone/pull/1752))
* Make template function `required` available to addons manifest templates ([#1737](https://github.com/kubermatic/kubeone/pull/1737))
* Ensure unattended-upgrades in dpkg is active ([#1756](https://github.com/kubermatic/kubeone/pull/1756))
* Fix control plane tolerations in Azure CCM and CSI addons (`node-role.kubernetes.io/master` doesn't have a value) ([#1733](https://github.com/kubermatic/kubeone/pull/1733))
* Add node affinity to the cluster-autoscaler addon ([#1716](https://github.com/kubermatic/kubeone/pull/1716))
* Update the cluster-autoscaler addon to match the upstream manifest ([#1713](https://github.com/kubermatic/kubeone/pull/1713))
* Add new "required" addons template function ([#1618](https://github.com/kubermatic/kubeone/pull/1618))
* Replace critical-pod annotation with priorityClassName ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Default image in the cluster-autoscaler addon and allow the image to be overridden using addon parameters ([#1552](https://github.com/kubermatic/kubeone/pull/1552))
* Minor improvements to OpenStack CCM and CSI addons. OpenStack CSI controller can now be scheduled on control plane nodes ([#1531](https://github.com/kubermatic/kubeone/pull/1531))
* Deploy default StorageClass for GCP clusters if the `default-storage-class` addon is enabled ([#1638](https://github.com/kubermatic/kubeone/pull/1638))
* Deploy the StorageClass based on the CSI driver if the CSI driver is deployed (requires the `default-storage-class` addon to be enabled) ([#1853](https://github.com/kubermatic/kubeone/pull/1853))
* Always deploy the StorageClass based on the in-tee provider (requires the `default-storage-class` addon to be enabled). This StorageClass will use the CSI driver if the CSI migration is enabled (default for all providers if the Kubernetes version is >= 1.23 or if the external CCM is used) ([#1853](https://github.com/kubermatic/kubeone/pull/1853))

### Terraform Configs

* **[BREAKING]** GCP: Default operating system for control plane instances is now Ubuntu 20.04 ([#1576](https://github.com/kubermatic/kubeone/pull/1576))
  * Make sure to bind `control_plane_image_family` to the image you're currently using or Terraform might recreate all your control plane instances
* **[BREAKING]** Azure: Default VM type is changed to `Standard_F2` ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
  * Make sure to bind `control_plane_vm_size` and `worker_vm_size` to the VM size you're currently using or Terraform might recreate all your instances
* **[BREAKING]** The default AMI for CentOS in Terraform configs for AWS has been changed to Rocky Linux. If you use the new Terraform configs with an existing cluster, make sure to bind the AMI as described in [the production recommendations document](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/) ([#1809](https://github.com/kubermatic/kubeone/pull/1809))
* Add the `control_plane_vm_count` variable to the AWS configs used to control the number of control plane nodes (defaults to 3) ([#1810](https://github.com/kubermatic/kubeone/pull/1810))
* Update the Terraform provider for OpenStack to version 1.47.0 ([#1816](https://github.com/kubermatic/kubeone/pull/1816))
* Set Ubuntu 20.04 as the default image for OpenStack ([#1816](https://github.com/kubermatic/kubeone/pull/1816))
* Add example Terraform configs for Flatcar on vSphere ([#1838](https://github.com/kubermatic/kubeone/pull/1838))
* Create a placement group for control plane nodes in Terraform configs for Hetzner ([#1762](https://github.com/kubermatic/kubeone/pull/1762))
* Remove `centos` choice from the GCE Terraform example configs as it's unsupported ([#1712](https://github.com/kubermatic/kubeone/pull/1712))
* Automatically determine GCE zone for the initial MachineDeployment ([#1703](https://github.com/kubermatic/kubeone/pull/1703))
* Fix AMI filter in Terraform configs for AWS to always use `x86_64` images ([#1692](https://github.com/kubermatic/kubeone/pull/1692))
* Automatically determine the SSH username in Terraform configs for AWS if the username is not provided ([#1844](https://github.com/kubermatic/kubeone/pull/1844))
* Terraform configs for AWS are now using `amzn` instead of `amzn2` as an identifier for Amazon Linux 2. `worker_os` is still using `amzn2` as the identifier ([#1855](https://github.com/kubermatic/kubeone/pull/1855))
* OpenStack: Open NodePorts by default ([#1530](https://github.com/kubermatic/kubeone/pull/1530))
* AWS: Open NodePorts by default ([#1535](https://github.com/kubermatic/kubeone/pull/1535))
* GCE: Open NodePorts by default ([#1529](https://github.com/kubermatic/kubeone/pull/1529))
* Hetzner: Create Firewall by default ([#1533](https://github.com/kubermatic/kubeone/pull/1533))
* Azure: Open NodePorts by default ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
* Fix keepalived script in Terraform configs for vSphere to assume yes when updating repos ([#1537](https://github.com/kubermatic/kubeone/pull/1537))
* Add additional Availability Set used for worker nodes to Terraform configs for Azure ([#1556](https://github.com/kubermatic/kubeone/pull/1556))
  * Make sure to check the [production recommendations for Azure clusters](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/#azure) for more information about how this additional availability set is used

### Updated

* Update Canal CNI to v3.22.0 ([#1797](https://github.com/kubermatic/kubeone/pull/1797))
* Update Cilium to v1.11.1 ([#1752](https://github.com/kubermatic/kubeone/pull/1752))
* Update Calico VXLAN addon to v3.22.0 ([#1797](https://github.com/kubermatic/kubeone/pull/1797))
* Update images in order to support Kubernetes 1.23 ([#1751](https://github.com/kubermatic/kubeone/pull/1751), [#1753](https://github.com/kubermatic/kubeone/pull/1753), [#1820](https://github.com/kubermatic/kubeone/pull/1820))
  * Update AWS External Cloud Controller Manager (CCM) to v1.23.0-alpha.0 for Kubernetes 1.23 clusters
  * Update Azure External Cloud Controller Manager (CCM) to v1.23.2 for Kubernetes 1.23 clusters
  * Update AWS EBS CSI driver to v1.5.0
  * Update AzureFile CSI driver to v1.9.0
  * Update AzureDisk CSI driver to v1.10.0
  * Update OpenStack External Cloud Controller Manager (CCM) to v1.23.0 for Kubernetes 1.23 clusters
  * Update the DigitalOcean External Cloud Controller Manager (CCM) to v0.1.36
  * Update DigitalOcean CSI to v4.0.0
  * Update the Hetzner External Cloud Controller Manager (CCM) to v1.12.1
* Update machine-controller to v1.43.0 ([#1834](https://github.com/kubermatic/kubeone/pull/1834))
  * machine-controller is now using Ubuntu 20.04 instead of 18.04 by default for all newly-created Machines on AWS, Azure, DO, GCE, Hetzner, Openstack and Equinix Metal
* Update vSphere CSI driver addon to v2.4.0. This change introduces Kubernetes 1.22 support for vSphere clusters ([#1675](https://github.com/kubermatic/kubeone/pull/1675))
* Update Go to 1.17.5 ([#1689](https://github.com/kubermatic/kubeone/pull/1689))

## Removed

* **[BREAKING]** Remove support for Amazon EKS-D clusters ([#1699](https://github.com/kubermatic/kubeone/pull/1699))
* Remove the PodPresets feature ([#1593](https://github.com/kubermatic/kubeone/pull/1593))
  * If you're still using this feature, make sure to migrate away before upgrading to this KubeOne release
* Remove Ansible examples ([#1633](https://github.com/kubermatic/kubeone/pull/1633))

# [v1.4.0-rc.1](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0-rc.1) - 2022-02-11

## Attention Needed

* Unconditionally deploy AWS, AzureDisk, AzureFile, and vSphere CSI drivers if the Kubernetes version is 1.23 or newer ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * Those providers have the CSI migration enabled by default in Kubernetes 1.23, so the CSI driver will be used for all volumes operations
* Unconditionally deploy DigitalOcean, Hetzner, Nutanix, and OpenStack Cinder CSI drivers ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * OpenStack has the CSI migration enabled by default since Kubernetes 1.18, so the CSI driver will be used for all operations
* **[BREAKING]** The default AMI for CentOS in Terraform configs for AWS has been changed to Rocky Linux. If you use the new Terraform configs with an existing cluster, make sure to bind the AMI as described in [the production recommendations document](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/) ([#1809](https://github.com/kubermatic/kubeone/pull/1809))

## Added

* Include darwin/arm64 and linux/arm64 builds in release artifacts ([#1821](https://github.com/kubermatic/kubeone/pull/1821))
* Allow providing operating system via the API ([#1809](https://github.com/kubermatic/kubeone/pull/1809))

## Changed

### General

* Increase the minimum Kubernetes version to 1.20 ([#1818](https://github.com/kubermatic/kubeone/pull/1818))
* Validate the Kubernetes version against supported versions constraints ([#1808](https://github.com/kubermatic/kubeone/pull/1808))
* Allow Docker as a container runtime up to Kubernetes v1.24 (previously up to v1.22) ([#1826](https://github.com/kubermatic/kubeone/pull/1826))
* Unconditionally deploy AWS, AzureDisk, AzureFile, and vSphere CSI drivers if the Kubernetes version is 1.23 or newer ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * Those providers have the CSI migration enabled by default in Kubernetes 1.23, so the CSI driver will be used for all volumes operations
* Unconditionally deploy DigitalOcean, Hetzner, Nutanix, and OpenStack Cinder CSI drivers ([#1831](https://github.com/kubermatic/kubeone/pull/1831))
  * OpenStack has the CSI migration enabled by default since Kubernetes 1.18, so the CSI driver will be used for all operations

### Fixed

* Restore missing addons deploy after containerd migration ([#1824](https://github.com/kubermatic/kubeone/pull/1824))
* Select correct CSR to approve ([#1813](https://github.com/kubermatic/kubeone/pull/1813))

### Terraform Configs

* **[BREAKING]** The default AMI for CentOS in Terraform configs for AWS has been changed to Rocky Linux. If you use the new Terraform configs with an existing cluster, make sure to bind the AMI as described in [the production recommendations document](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/) ([#1809](https://github.com/kubermatic/kubeone/pull/1809))
* Add the `control_plane_vm_count` variable to the AWS configs used to control the number of control plane nodes (defaults to 3) ([#1810](https://github.com/kubermatic/kubeone/pull/1810))
* Update the Terraform provider for OpenStack to version 1.47.0 ([#1816](https://github.com/kubermatic/kubeone/pull/1816))
* Set Ubuntu 20.04 as the default image for OpenStack ([#1816](https://github.com/kubermatic/kubeone/pull/1816))
* Add example Terraform configs for Flatcar on vSphere ([#1838](https://github.com/kubermatic/kubeone/pull/1838), [#1846](https://github.com/kubermatic/kubeone/pull/1846))

### Updated

* Update DigitalOcean CSI to v4.0.0 ([#1820](https://github.com/kubermatic/kubeone/pull/1820))
* Update machine-controller to v1.43.0 ([#1834](https://github.com/kubermatic/kubeone/pull/1834))

# [v1.4.0-rc.0](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0-rc.0) - 2022-02-03

## Attention Needed

* CentOS 8 has reached End-Of-Life (EOL) on January 31st, 2022. It will no longer receive any updates (including security updates). Support for CentOS 8 in KubeOne is deprecated and will be removed in a future release. We strongly recommend migrating to another operating system or CentOS distribution as soon as possible.

## Added

* Add experimental/alpha-level support for [Kubermatic Operating System Manager (OSM)](https://github.com/kubermatic/operating-system-manager) ([#1748](https://github.com/kubermatic/kubeone/pull/1748))
* Add ability to change the container log maximum size (defaults to 100Mi) ([#1644](https://github.com/kubermatic/kubeone/pull/1644))
* Add ability to change the container log maximum files (defaults to 5) ([#1759](https://github.com/kubermatic/kubeone/pull/1759))
* Add the DigitalOcean CSI driver. The CSI driver is deployed automatically if `.cloudProvider.external` is enabled ([#1754](https://github.com/kubermatic/kubeone/pull/1754))
* Add the default StorageClass and VolumeSnapshotClass for the DigitalOcean CSI driver. The StorageClass and VolumeSnapshotClass can be deployed by enabling the default-storage-class embedded addon ([#1754](https://github.com/kubermatic/kubeone/pull/1754))
* Generate and approve CSRs for control plane and static workers nodes. Enable the server TLS bootstrap for control plane and static worker nodes ([#1750](https://github.com/kubermatic/kubeone/pull/1750), [#1758](https://github.com/kubermatic/kubeone/pull/1758))
* Source `.cloudProvider.csiConfig` from the credentials file if present ([#1739](https://github.com/kubermatic/kubeone/pull/1739))
* Fetch containerd auth config from the credentials file if present ([#1745](https://github.com/kubermatic/kubeone/pull/1745))

## Changed

### Fixed

* Change baseurl to `vault.centos.org` for CentOS 8 ([#1767](https://github.com/kubermatic/kubeone/pull/1767))
* Fix Docker to containerd migration on non-Flatcar operating systems ([#1743](https://github.com/kubermatic/kubeone/pull/1743))
* Fix propagation of proxy config to machines and Kubernetes components ([#1746](https://github.com/kubermatic/kubeone/pull/1746))

### Addons

* Replace Hubble static certificate with CronJob generation ([#1752](https://github.com/kubermatic/kubeone/pull/1752))
* Make template function `required` available to addons manifest templates ([#1737](https://github.com/kubermatic/kubeone/pull/1737))
* Ensure unattended-upgrades in dpkg is active ([#1756](https://github.com/kubermatic/kubeone/pull/1756))

### Terraform Configs

* Create a placement group for control plane nodes in Terraform configs for Hetzner ([#1762](https://github.com/kubermatic/kubeone/pull/1762))

### Updated

* Update Canal CNI to v3.22.0 ([#1797](https://github.com/kubermatic/kubeone/pull/1797))
* Update Cilium to v1.11.1 ([#1752](https://github.com/kubermatic/kubeone/pull/1752))
* Update Calico VXLAN addon to v3.22.0 ([#1797](https://github.com/kubermatic/kubeone/pull/1797))
* Update images in order to support Kubernetes 1.23 ([#1751](https://github.com/kubermatic/kubeone/pull/1751), [#1753](https://github.com/kubermatic/kubeone/pull/1753))
  * Update AWS External Cloud Controller Manager (CCM) to v1.23.0-alpha.0 for Kubernetes 1.23 clusters
  * Update Azure External Cloud Controller Manager (CCM) to v1.23.2 for Kubernetes 1.23 clusters
  * Update AWS EBS CSI driver to v1.5.0
  * Update AzureFile CSI driver to v1.9.0
  * Update AzureDisk CSI driver to v1.10.0
  * Update OpenStack External Cloud Controller Manager (CCM) to v1.23.0 for Kubernetes 1.23 clusters
  * Update the DigitalOcean External Cloud Controller Manager (CCM) to v0.1.36
  * Update the Hetzner External Cloud Controller Manager (CCM) to v1.12.1
* Update machine-controller to v1.42.2 ([#1748](https://github.com/kubermatic/kubeone/pull/1748))

# [v1.4.0-beta.1](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0-beta.1) - 2022-01-14

## Attention Needed

* **[BREAKING]** The `cloud-provider-credentials` Secret is removed by KubeOne because KubeOne does not use it any longer. If you have any workloads NOT created by KubeOne that use this Secret, please migrate before upgrading KubeOne. Instead, KubeOne now creates `kubeone-machine-controller-credentials` and `kubeone-ccm-credentials` Secrets used by machine-controller and external CCM ([#1717](https://github.com/kubermatic/kubeone/pull/1717), [#1718](https://github.com/kubermatic/kubeone/pull/1718))

## Added

* Add experimental/alpha support for Nutanix ([#1723](https://github.com/kubermatic/kubeone/pull/1723), [#1725](https://github.com/kubermatic/kubeone/pull/1725), [#1733](https://github.com/kubermatic/kubeone/pull/1733))
  * Support for Nutanix is experimental, so implementation and relevant addons might be changed until it doesn't graduate to beta/stable
* Add the Nutanix CSI driver addon. The addon is deployed manually, on-demand, by enabling the `csi-nutanix` embedded addon (see the PR description for more details and examples) ([#1733](https://github.com/kubermatic/kubeone/pull/1733), [#1734](https://github.com/kubermatic/kubeone/pull/1734))
* Add the default StorageClass for the Nutanix CSI driver. The StorageClass can be deployed by enabling the `default-storage-class` embedded addon (see the PR description for more details and examples) ([#1733](https://github.com/kubermatic/kubeone/pull/1733))
* Add the Registry Credentials configuration to the RegistryConfiguration API ([#1724](https://github.com/kubermatic/kubeone/pull/1724))
* Add support for different credentials for machine-controller and CCM. Environment variables can be prefixed with `MC_` for machine-controller credentials and `CCM_` for CCM credentials ([#1717](https://github.com/kubermatic/kubeone/pull/1717))

## Changed

### General

* **[BREAKING]** The `cloud-provider-credentials` Secret is removed by KubeOne because KubeOne does not use it any longer. If you have any workloads NOT created by KubeOne that use this Secret, please migrate before upgrading KubeOne. Instead, KubeOne now creates `kubeone-machine-controller-credentials` and `kubeone-ccm-credentials` Secrets used by machine-controller and external CCM ([#1717](https://github.com/kubermatic/kubeone/pull/1717), [#1718](https://github.com/kubermatic/kubeone/pull/1718))

### Fixed

* Fix a bug with the addons applier applying all files when the addons path is not provided ([#1733](https://github.com/kubermatic/kubeone/pull/1733))

### Addons

* Fix control plane tolerations in Azure CCM and CSI addons (`node-role.kubernetes.io/master` doesn't have a value) ([#1733](https://github.com/kubermatic/kubeone/pull/1733))
* Add node affinity to the cluster-autoscaler addon ([#1716](https://github.com/kubermatic/kubeone/pull/1716))

### Terraform Configs

* Remove `centos` choice from the GCE Terraform example configs as it's unsupported ([#1712](https://github.com/kubermatic/kubeone/pull/1712))

### Updated

* Update machine-controller to v1.42.0 ([#1733](https://github.com/kubermatic/kubeone/pull/1733))

# [v1.4.0-beta.0](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0-beta.0) - 2022-01-04

## Attention Needed

* KubeOne 1.4.0-beta.0 introduces the new KubeOneCluster v1beta2 API
  * The new v1beta2 API is still under-development and might be changed before the KubeOne 1.4.0 release
  * We recommend and highly encourage testing the new API, but considering that the API might be changed before the final release, we don't recommend migrating production clusters to the new API yet
  * The migration for existing KubeOneCluster manifests is not yet available
  * The `kubeone config print` command now uses the new v1beta2 API
  * The existing KubeOneCluster v1beta1 API is considered as deprecated and will be removed in KubeOne 1.6+
  * Highlights:
    * The API group has been changed from `kubeone.io` to `kubeone.k8c.io`
    * The AssetConfiguration API has been removed from the v1beta2 API. The AssetConfiguration API can still be used with the v1beta1 API, but we highly recommend migrating away because the v1beta1 API is deprecated
    * The PodPresets feature has been removed from the v1beta2 API because Kubernetes removed support for PodPresets in Kubernetes 1.20
    * Packet (`packet`) cloud provider has been rebranded to Equinix Metal (`equinixmetal`). The existing Packet cluster will work with `equinixmetal` cloud provider, however, manual migration steps are required if you want to use new Terraform configs for Equinix Metal
    * A new ContainerRuntime API has been added to the v1beta2 API in order to support configuring mirror registries. This API is still work-in-progress and will mostly like be extended before the final release
* `kubeone install` and `kubeone upgrade` commands are considered as deprecated in favor of `kubeone apply`
  * `install` and `upgrade` commands will be removed in KubeOne 1.6+
  * We highly encourage switching to `kubeone apply`. The `apply` command has the same semantics and works in the same way as `install`/`upgrade`, with some additional checks to ensure each requested operation is safe for the cluster
* Support for Amazon EKS-D clusters has been removed starting from this release

## Known Issues

* It's not possible to run kube-proxy in IPVS mode on Kubernetes 1.23 clusters using Canal/Calico CNI. Trying to upgrade existing 1.22 clusters using IPVS to 1.23 will result in a validation error from KubeOne
  * More information about this issue can be found in the following Calico ticket: https://github.com/projectcalico/calico/issues/5011

## Added

### API

* Add the KubeOneCluster v1beta2 API and change the API group to `kubeone.k8c.io` ([#1649](https://github.com/kubermatic/kubeone/pull/1649))
  * Make `kubeone config print` command use the new `kubeone.k8c.io/v1beta2` API ([#1651](https://github.com/kubermatic/kubeone/pull/1651))
  * Add the new ContainerRuntime API with support for mirror registries ([#1674](https://github.com/kubermatic/kubeone/pull/1674))
  * Addons directory path (`.addons.path`) is not required when using only embedded addons ([#1668](https://github.com/kubermatic/kubeone/pull/1668))
  * Addons directory path (`.addons.path`) is not defaulted to `./addons` any longer ([#1668](https://github.com/kubermatic/kubeone/pull/1668))
  * Add the KubeletConfig API used to configure `systemReserved`, `kubeReserved`, and `evictionHard` Kubelet options ([#1698](https://github.com/kubermatic/kubeone/pull/1698))
  * Remove the PodPresets feature ([#1662](https://github.com/kubermatic/kubeone/pull/1662))
  * Remove the AssetConfiguration API ([#1699](https://github.com/kubermatic/kubeone/pull/1699))
  * Rebrand Packet (`packet`) to Equinix Metal (`equinixmetal`) and support migrating existing Packet clusters to Equinix Metal
  clusters ([#1663](https://github.com/kubermatic/kubeone/pull/1663))

### Features

* Add support for Kubernetes 1.23 ([#1678](https://github.com/kubermatic/kubeone/pull/1678))
* Add `kubeone addons list` command used to list available and enabled addons ([#1642](https://github.com/kubermatic/kubeone/pull/1642))
* Add support for OpenStack Application Credentials ([#1666](https://github.com/kubermatic/kubeone/pull/1666))
* Add a new `--kubernetes-version` flag to the `kubeone config images` command ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
  * This flag is used to filter images for a particular Kubernetes version. The flag cannot be used along with the KubeOneCluster manifest (`--manifest` flag)
* Addon parameters can be resolved into environment variable contents if the `env:` prefix is set in the parameter value ([#1691](https://github.com/kubermatic/kubeone/pull/1691))

## Changed

### General

* Improve installation scripts used to install container runtime ([#1664](https://github.com/kubermatic/kubeone/pull/1664))

### Fixed

* Fix issues when disabling nm-cloud-setup on RHEL ([#1706](https://github.com/kubermatic/kubeone/pull/1706))
* cri-tools is now installed automatically as a dependency of kubeadm on Amazon Linux 2. This fixes provisioning issues on Amazon Linux 2 with newer Kubernetes versions. ([#1701](https://github.com/kubermatic/kubeone/pull/1701))
* Fix the image loader script to support KubeOne 1.3+ and Kubernetes 1.22+ ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
* The `kubeone config images` command now shows images for the latest Kubernetes version (instead of for the oldest) ([#1671](https://github.com/kubermatic/kubeone/pull/1671))
* Allow pods with the seccomp profile defined to get scheduled if the PodSecurityPolicy (PSP) feature is enabled ([#1686](https://github.com/kubermatic/kubeone/pull/1686))

### Addons

* Update the cluster-autoscaler addon to match the upstream manifest ([#1713](https://github.com/kubermatic/kubeone/pull/1713))

### Terraform Configs

* Automatically determine GCE zone for the initial MachineDeployment ([#1703](https://github.com/kubermatic/kubeone/pull/1703))
* Fix AMI filter in Terraform configs for AWS to always use `x86_64` images ([#1692](https://github.com/kubermatic/kubeone/pull/1692))

### Updated

* Update Cilium CNI addon to v1.11.0 ([#1681](https://github.com/kubermatic/kubeone/pull/1681))
* Update vSphere CSI driver addon to v2.4.0. This change introduces Kubernetes 1.22 support for vSphere clusters ([#1675](https://github.com/kubermatic/kubeone/pull/1675))
* Update Go to 1.17.5 ([#1689](https://github.com/kubermatic/kubeone/pull/1689))

## Removed

* Remove support for Amazon EKS-D clusters ([#1699](https://github.com/kubermatic/kubeone/pull/1699))

# [v1.4.0-alpha.0](https://github.com/kubermatic/kubeone/releases/tag/v1.4.0-alpha.0) - 2021-11-29

## Attention Needed

* [**BREAKING**] GCP: Default operating system for control plane instances is now Ubuntu 20.04 ([#1576](https://github.com/kubermatic/kubeone/pull/1576))
  * Make sure to bind `control_plane_image_family` to the image you're currently using or Terraform might recreate all your control plane instances
* [**BREAKING**] Azure: Default VM type is changed to `Standard_F2` ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
  * Make sure to bind `control_plane_vm_size` and `worker_vm_size` to the VM size you're currently using or Terraform might recreate all your instances

## Added

### Features

* Add CCM/CSI migration support for clusters with the static worker nodes ([#1544](https://github.com/kubermatic/kubeone/pull/1544))
* Add CCM/CSI migration support for the Azure clusters ([#1610](https://github.com/kubermatic/kubeone/pull/1610))
* Automatically create cloud-config Secret for all providers if external cloud controller manager (`.cloudProvider.external`) is enabled ([#1575](https://github.com/kubermatic/kubeone/pull/1575))
* Add support for Cilium CNI ([#1560](https://github.com/kubermatic/kubeone/pull/1560), [#1629](https://github.com/kubermatic/kubeone/pull/1629))
* Add support for additional Subject Alternative Names (SANs) for the Kubernetes API server ([#1599](https://github.com/kubermatic/kubeone/pull/1599), [#1603](https://github.com/kubermatic/kubeone/pull/1603), [#1606](https://github.com/kubermatic/kubeone/pull/1606))
* Add a new `MachineAnnotations` field in the API used to define annotations in `MachineDeployment.Spec.Template.Spec.Annotations` ([#1601](https://github.com/kubermatic/kubeone/pull/1601))
* Add a new `--create-machine-deployments` flag to the `kubeone apply` command used to control should KubeOne create initial MachineDeployment objects when provisioning the cluster (default is `true`) ([#1617](https://github.com/kubermatic/kubeone/pull/1617))

### Addons

* Integrate the AWS CCM addon with KubeOne ([#1585](https://github.com/kubermatic/kubeone/pull/1585))
  * The AWS CCM is now deployed if the external cloud provider (`.cloudProvider.external`) is enabled
  * This option cannot be enabled for existing AWS clusters running in-tree cloud provider, instead, those clusters must go through the CCM/CSI migration process
* Add the AWS EBS CSI driver addon ([#1597](https://github.com/kubermatic/kubeone/pull/1597))
  * Automatically deploy the AWS EBS CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AWS EBS CSI driver to the `default-storage-class` embedded addon
* Integrate the Azure CCM addon with KubeOne ([#1561](https://github.com/kubermatic/kubeone/pull/1561), [#1579](https://github.com/kubermatic/kubeone/pull/1579))
  * The Azure CCM is now deployed if the external cloud provider (`.cloudProvider.external`) is enabled
  * This option cannot be enabled for existing Azure clusters running in-tree cloud provider, instead, those clusters must go through the CCM/CSI migration process
* Add the AzureFile CSI driver addon ([#1575](https://github.com/kubermatic/kubeone/pull/1575), [#1579](https://github.com/kubermatic/kubeone/pull/1579))
  * Automatically deploy the AzureFile CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AzureFile CSI driver to the `default-storage-class` embedded addon
* Add the AzureDisk CSI driver addon ([#1577](https://github.com/kubermatic/kubeone/pull/1577))
  * Automatically deploy the AzureDisk CSI driver addon if external cloud controller manager (`.cloudProvider.external`) is enabled
  * Add default StorageClass for AzureDisk CSI driver to the `default-storage-class` embedded addon

### Other

* Add a deprecation warning for PodSecurityPolicies ([#1595](https://github.com/kubermatic/kubeone/pull/1595))

## Changed

### General

* Validate the cluster name to ensure it's a correct DNS subdomain (RFC 1123) ([#1641](https://github.com/kubermatic/kubeone/pull/1641), [#1646](https://github.com/kubermatic/kubeone/pull/1646), [#1648](https://github.com/kubermatic/kubeone/pull/1648))
* Create MachineDeployments only for newly-provisioned clusters ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Show warning about LBs on CCM migration for OpenStack clusters ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Change default Kubernetes version in the example configuration to v1.22.3 ([#1605](https://github.com/kubermatic/kubeone/pull/1605))

### Fixed

* Force drain nodes to remove standalone pods ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Check for minor version when choosing kubeadm API version ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Provide `--cluster-name` flag to the OpenStack external CCM (read PR description for more details) ([#1619](https://github.com/kubermatic/kubeone/pull/1619))
* Enable ip_tables related kernel modules and disable `nm-cloud-setup` tool on AWS for RHEL machines ([#1607](https://github.com/kubermatic/kubeone/pull/1607))
* Properly pass machine-controllers args ([#1594](https://github.com/kubermatic/kubeone/pull/1594))
  * This fixes the issue causing machine-controller and machine-controller-webhook deployments to run with incorrect flags
  * If you created your cluster with KubeOne 1.2 or older, and already upgraded to KubeOne 1.3, we recommend running kubeone apply again with KubeOne 1.3.2 or newer to properly reconcile machine-controller deployments
* Fix `yum versionlock delete containerd.io` error ([#1600](https://github.com/kubermatic/kubeone/pull/1600))
* Ensure containerd/docker be upgraded automatically when running kubeone apply ([#1589](https://github.com/kubermatic/kubeone/pull/1589))
* Edit SELinux config file only if file exists ([#1532](https://github.com/kubermatic/kubeone/pull/1532))

### Addons

* Add new "required" addons template function ([#1618](https://github.com/kubermatic/kubeone/pull/1618))
* Replace critical-pod annotation with priorityClassName ([#1627](https://github.com/kubermatic/kubeone/pull/1627))
* Default image in the cluster-autoscaler addon and allow the image to be overridden using addon parameters ([#1552](https://github.com/kubermatic/kubeone/pull/1552))
* Minor improvements to OpenStack CCM and CSI addons. OpenStack CSI controller can now be scheduled on control plane nodes ([#1531](https://github.com/kubermatic/kubeone/pull/1531))
* Deploy default StorageClass for GCP clusters if the `default-storage-class` addon is enabled ([#1638](https://github.com/kubermatic/kubeone/pull/1638))

### Terraform Configs

* [**BREAKING**] GCP: Default operating system for control plane instances is now Ubuntu 20.04 ([#1576](https://github.com/kubermatic/kubeone/pull/1576))
  * Make sure to bind `control_plane_image_family` to the image you're currently using or Terraform might recreate all your control plane instances
* [**BREAKING**] Azure: Default VM type is changed to `Standard_F2` ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
  * Make sure to bind `control_plane_vm_size` and `worker_vm_size` to the VM size you're currently using or Terraform might recreate all your instances
* OpenStack: Open NodePorts by default ([#1530](https://github.com/kubermatic/kubeone/pull/1530))
* AWS: Open NodePorts by default ([#1535](https://github.com/kubermatic/kubeone/pull/1535))
* GCE: Open NodePorts by default ([#1529](https://github.com/kubermatic/kubeone/pull/1529))
* Hetzner: Create Firewall by default ([#1533](https://github.com/kubermatic/kubeone/pull/1533))
* Azure: Open NodePorts by default ([#1528](https://github.com/kubermatic/kubeone/pull/1528))
* Fix keepalived script in Terraform configs for vSphere to assume yes when updating repos ([#1537](https://github.com/kubermatic/kubeone/pull/1537))
* Add additional Availability Set used for worker nodes to Terraform configs for Azure ([#1556](https://github.com/kubermatic/kubeone/pull/1556))
  * Make sure to check the [production recommendations for Azure clusters](https://docs.kubermatic.com/kubeone/main/cheat_sheets/production_recommendations/#azure) for more information about how this additional availability set is used

### Updated

* Update machine-controller to v1.37.0 ([#1647](https://github.com/kubermatic/kubeone/pull/1647))
  * machine-controller is now using Ubuntu 20.04 instead of 18.04 by default for all newly-created Machines on AWS, Azure, DO, GCE, Hetzner, Openstack and Equinix Metal
* Update Hetzner Cloud Controller Manager to v1.12.0 ([#1583](https://github.com/kubermatic/kubeone/pull/1583))
* Update Go to 1.17.1 ([#1534](https://github.com/kubermatic/kubeone/pull/1534), [#1541](https://github.com/kubermatic/kubeone/pull/1541), [#1542](https://github.com/kubermatic/kubeone/pull/1542), [#1545](https://github.com/kubermatic/kubeone/pull/1545))

## Removed

* Remove the PodPresets feature ([#1593](https://github.com/kubermatic/kubeone/pull/1593))
  * If you're still using this feature, make sure to migrate away before upgrading to this KubeOne release
* Remove Ansible examples ([#1633](https://github.com/kubermatic/kubeone/pull/1633))
