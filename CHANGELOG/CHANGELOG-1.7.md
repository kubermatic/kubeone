# [v1.7.2](https://github.com/kubermatic/kubeone/releases/tag/v1.7.2) - 2024-01-05

## Changelog since v1.7.1

## Urgent Upgrade Notes 

### (No, really, you MUST read this before you upgrade)

- Fix determining the `CLUSTER_NAME` environment variable for OpenStack CCM DaemonSet and OpenStack Cinder CSI Deployment. This fixes a regression where the cluster name was changed to `kubernetes` upon running `kubeone apply` two or more times after upgrading from KubeOne 1.6 to KubeOne 1.7. This regression might affect Octavia Load Balancers, making them fail to reconcile. Please check the [known issues document](https://docs.kubermatic.com/kubeone/v1.7/known-issues/) to find if you're affected by this issue and what steps you need to take if you're affected ([#2979](https://github.com/kubermatic/kubeone/pull/2979), [@kubermatic-bot](https://github.com/kubermatic-bot))

## Changes by Kind

### Bug or Regression

- Delete AzureDisk's `csi-azuredisk-node-secret-binding` ClusterRoleBinding if RoleRef's name is `csi-azuredisk-node-sa` to allow upgrading KubeOne from 1.6 to 1.7 ([#2984](https://github.com/kubermatic/kubeone/pull/2984), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Increase the memory requests and limits from 300Mi to 600Mi for cluster-autoscaler to prevent cluster-autoscaler from being `OOMKilled` ([#2979](https://github.com/kubermatic/kubeone/pull/2979), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Updates

#### machine-controller

- Update machine-controller to v1.57.4 ([#2989](https://github.com/kubermatic/kubeone/pull/2989), [@xmudrii](https://github.com/xmudrii))

#### Addons

- Update backup-restic addon to use etcd 3.5.11 for creating etcd snapshots ([#2982](https://github.com/kubermatic/kubeone/pull/2982), [@kubermatic-bot](https://github.com/kubermatic-bot))

#### Go

- KubeOne is now built with Go 1.21.5 ([#2989](https://github.com/kubermatic/kubeone/pull/2989), [@xmudrii](https://github.com/xmudrii))

# [v1.7.1](https://github.com/kubermatic/kubeone/releases/tag/v1.7.1) - 2023-11-10

## Changelog since v1.7.0

## Changes by Kind

### Feature

- Allow setting `CCM_CONCURRENT_SERVICE_SYNCS` parameter on CCM addons to configure number of concurrent `LoadBalancer` service reconciles ([#2918](https://github.com/kubermatic/kubeone/pull/2918), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Support for customizing `vAppName` for VMware Cloud Director CSI driver ([#2956](https://github.com/kubermatic/kubeone/pull/2956), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Bug or Regression

- Fix a bug with the VMware Cloud Director CSI driver addon where it would crash if no `VCD_API_TOKEN` is set ([#2956](https://github.com/kubermatic/kubeone/pull/2956), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix a bug where wildcard registry name provided to the machine-controller-webhook Deployment is not properly escaped ([#2928](https://github.com/kubermatic/kubeone/pull/2928), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix a bug where `kubeadm init` and `kubeadm join` report that the `containerRuntimeEndpoint` field in `KubeletConfiguration` doesn't exist for clusters running Kubernetes versions prior to 1.27 ([#2940](https://github.com/kubermatic/kubeone/pull/2940), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix a bug where `registryConfiguration.OverrideRegistry` wasn't correctly applied to the pause image configured for control plane and static worker nodes ([#2930](https://github.com/kubermatic/kubeone/pull/2930), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix a bug where the CA bundle was not propagated to the vSphere CSI driver ([#2907](https://github.com/kubermatic/kubeone/pull/2907), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix a globbing issue for `apt-get install` causing KubeOne to install wrong Kubernetes version in some circumstances ([#2959](https://github.com/kubermatic/kubeone/pull/2959), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix incorrect validation that made `VCD_API_TOKEN` unusable for VMware Cloud Director ([#2946](https://github.com/kubermatic/kubeone/pull/2946), [@kubermatic-bot](https://github.com/kubermatic-bot))

### Updates

#### Core Components

- Update metrics-server and cluster-autoscaler to the latest patch releases ([#2957](https://github.com/kubermatic/kubeone/pull/2957), [@xmudrii](https://github.com/xmudrii))
- Update Canal to v3.26.3 ([#2957](https://github.com/kubermatic/kubeone/pull/2957), [@xmudrii](https://github.com/xmudrii))
- Update Calico VXLAN to v3.26.3 ([#2957](https://github.com/kubermatic/kubeone/pull/2957), [@xmudrii](https://github.com/xmudrii))
- Update Cilium to v1.14.3 ([#2957](https://github.com/kubermatic/kubeone/pull/2957), [@xmudrii](https://github.com/xmudrii))

#### Go

- KubeOne is now built with Go 1.21.3 ([#2954](https://github.com/kubermatic/kubeone/pull/2954), [@xmudrii](https://github.com/xmudrii))

### Other (Cleanup or Flake)

- Store GPG keys for package repositories in `/etc/apt/keyrings` instead of deprecated `apt-key-add` command ([#2920](https://github.com/kubermatic/kubeone/pull/2920), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Download cri-tools from the Kubernetes package repositories instead of the Amazon Linux 2 package repositories on instances running Amazon Linux 2 ([#2951](https://github.com/kubermatic/kubeone/pull/2951), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Update CRDs for the operating-system-manager (OSM) addon ([#2935](https://github.com/kubermatic/kubeone/pull/2935), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.7.0](https://github.com/kubermatic/kubeone/releases/tag/v1.7.0) - 2023-09-08

## Changelog since v1.6.0

## Urgent Upgrade Notes 

### (No, really, you MUST read this before you upgrade)

- Migrate from the legacy package repositories (`apt.kubernetes.io` and `yum.kubernetes.io`) to the Kubernetes community-hosted package repositories (`pkgs.k8s.io`). The legacy repositories [have been deprecated as of August 31, 2023 and will be frozen starting from September 13, 2023](https://kubernetes.io/blog/2023/08/31/legacy-package-repository-deprecation/). Upgrading to KubeOne v1.7.0+ or v1.6.3+ is required in order to install or upgrade to Kubernetes version newer than v1.27.6, v1.26.9, and v1.25.14. **If IP-based or URL-based filtering is in place**, you may need to mirror the release packages to a local package repository that you have strict control over. See [the official announcement](https://kubernetes.io/blog/2023/08/15/pkgs-k8s-io-introduction/) for more details ([#2873](https://github.com/kubermatic/kubeone/pull/2873), [@xmudrii](https://github.com/xmudrii))
- Migrate from the Kubernetes release bucket (`https://storage.googleapis.com/kubernetes-release/release`) to `dl.k8s.io` for downloading binaries. This change only affects Flatcar-based clusters. **If IP-based or URL-based filtering is in place**, you need to allow the appropriate IP addresses and domains as described in [the official `dl.k8s.io` announcement](https://kubernetes.io/blog/2023/06/09/dl-adopt-cdn/) ([#2873](https://github.com/kubermatic/kubeone/pull/2873), [@xmudrii](https://github.com/xmudrii))
- Use OpenStack native Load Balancer for the Kubernetes API in the example Terraform configs for OpenStack. Do **not** apply this change for existing clusters as that will **completely break** the control plane. Existing clusters must continue using the GoBetween Load Balancer or whatever solution is in place ([#2869](https://github.com/kubermatic/kubeone/pull/2869), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### API Change

- The minimum Kubernetes version has been increased to v1.25.0. If you're using an earlier Kubernetes version, please upgrade to at least v1.25 before upgrading to KubeOne v1.7. See [the Kubernetes compatibility matrix](https://docs.kubermatic.com/kubeone/v1.7/architecture/compatibility/supported-versions/) for more information about supported Kubernetes versions ([#2897](https://github.com/kubermatic/kubeone/pull/2897), [@xmudrii](https://github.com/xmudrii))
- Add `.cloudProvider.disableBundledCSIDrivers` boolean field to the API. When set to `true`, the built-in CSI driver will not be deployed to the cluster. If enabled for an existing cluster, the CSI driver and relevant volumes must be removed manually ([#2784](https://github.com/kubermatic/kubeone/pull/2784), [@kron4eg](https://github.com/kron4eg))
- Add support for referencing credentials exposed via environment variables or credentials file in cloudConfig (`.cloudProvider.cloudConfig`). Credentials are referenced like `{{ .Credentials.ENVIRONMENT_VARIABLE_NAME }}` ([#2789](https://github.com/kubermatic/kubeone/pull/2789), [@kron4eg](https://github.com/kron4eg))
- Add `.helmReleases.*.chartURL` field to the API. This field can be used to provide a direct chart URL location ([#2836](https://github.com/kubermatic/kubeone/pull/2836), [@kron4eg](https://github.com/kron4eg))
- Make `.helmReleases.*.repoURL` an optional field ([#2715](https://github.com/kubermatic/kubeone/pull/2715), [@kron4eg](https://github.com/kron4eg))

### Feature

- Add support for Kubernetes 1.27 ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))
  - **Important:** AWS-based clusters require using external cloud controller manager (CCM) with Kubernetes 1.27 and newer. Existing clusters running in-tree cloud provider must [migrate to the external CCM](https://docs.kubermatic.com/kubeone/v1.7/guides/ccm-csi-migration/) before upgrading to Kubernetes 1.27
- Add IPv4/IPv6 dual-stack support for vSphere ([#2806](https://github.com/kubermatic/kubeone/pull/2806), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Add experimental support for Debian ([#2732](https://github.com/kubermatic/kubeone/pull/2732), [@madalinignisca](https://github.com/madalinignisca))
- Add support for API token authentication for VMware Cloud Director ([#2751](https://github.com/kubermatic/kubeone/pull/2751), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Add an optional `CLUSTER_AUTOSCALER_SKIP_LOCAL_STORAGE` parameter for the `cluster-autoscaler` addon used to enable/disable skipping local storage when downscaling nodes (see https://github.com/kubermatic/kubeone/tree/release/v1.7/addons/cluster-autoscaler for more details) ([#2872](https://github.com/kubermatic/kubeone/pull/2872), [@c4tz](https://github.com/c4tz))
- Add an optional `clusterid` parameter for the VMware Cloud Director CSI driver addon used to customize the Cluster ID value used by the CSI driver ([#2730](https://github.com/kubermatic/kubeone/pull/2730), [@JamesClonk](https://github.com/JamesClonk))
- Provide the explicit list of safe ciphersuites to kubelet to fix the issue reported by the CIS benchmark ([#2814](https://github.com/kubermatic/kubeone/pull/2814), [@kron4eg](https://github.com/kron4eg))

### Updates

#### General

- Upgrade cri-tools to v1.27.1 for clusters running Kubernetes 1.27 ([#2873](https://github.com/kubermatic/kubeone/pull/2873), [@xmudrii](https://github.com/xmudrii))
- Update base image for KubeOne container image to `alpine:3.17` ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))

#### CNI

- Update Canal CNI to v3.26.1 and Cilium to v1.14.1 ([#2860](https://github.com/kubermatic/kubeone/pull/2860), [@WeirdMachine](https://github.com/WeirdMachine))
- Update Canal CNI to v3.26.0 and Cilium to v1.13.3 ([#2799](https://github.com/kubermatic/kubeone/pull/2799), [@WeirdMachine](https://github.com/WeirdMachine))
- Update Calico VXLAN CNI addon to v3.26.1 ([#2844](https://github.com/kubermatic/kubeone/pull/2844), [@kron4eg](https://github.com/kron4eg))

#### General Addons

- Update NodeLocalDNSCache to 1.22.23 ([#2813](https://github.com/kubermatic/kubeone/pull/2813), [@xmudrii](https://github.com/xmudrii))
- Update metrics-server to v0.6.3 ([#2813](https://github.com/kubermatic/kubeone/pull/2813), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler to v1.27.2, v1.26.3, v1.25.2, v1.24.2 ([#2842](https://github.com/kubermatic/kubeone/pull/2842), [@xmudrii](https://github.com/xmudrii))
- Update images in `backups-restic` and `unattended-upgrades` addons ([#2845](https://github.com/kubermatic/kubeone/pull/2845), [@kron4eg](https://github.com/kron4eg))

#### machine-controller and operating-system-manager

- Update machine-controller to v1.57.3 ([#2861](https://github.com/kubermatic/kubeone/pull/2861), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to v1.57.2 ([#2833](https://github.com/kubermatic/kubeone/pull/2833), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to v1.57.0 ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))
- Update operating-system-manager to v1.3.2 ([#2861](https://github.com/kubermatic/kubeone/pull/2861), [@kron4eg](https://github.com/kron4eg))
- Update operating-system-manager to v1.3.0 ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))
- Update operating-system-manager to 1.2.2 ([#2762](https://github.com/kubermatic/kubeone/pull/2762), [@pkprzekwas](https://github.com/pkprzekwas))

#### AWS

- Update AWS CCM to v1.27.1, v1.26.1, v1.25.3, v1.24.4 ([#2820](https://github.com/kubermatic/kubeone/pull/2820), [@xmudrii](https://github.com/xmudrii))
- Update AWS EBS CSI driver to v1.22.0 ([#2859](https://github.com/kubermatic/kubeone/pull/2859), [@kron4eg](https://github.com/kron4eg))
- Update AWS EBS CSI driver to v1.20.0 ([#2820](https://github.com/kubermatic/kubeone/pull/2820), [@xmudrii](https://github.com/xmudrii))
- Update CSI Snapshotter for AWS EBS CSI driver to v6.2.1 ([#2820](https://github.com/kubermatic/kubeone/pull/2820), [@xmudrii](https://github.com/xmudrii))

#### Azure

- Update Azure CCM to v1.27.6 ([#2830](https://github.com/kubermatic/kubeone/pull/2830), [@kron4eg](https://github.com/kron4eg))
- Update AzureDisk CSI and AzureFile CSI to v1.27.1 ([#2831](https://github.com/kubermatic/kubeone/pull/2831), [@kron4eg](https://github.com/kron4eg))

#### DigitalOcean

- Update DigitalOcean CCM to v0.1.43 ([#2840](https://github.com/kubermatic/kubeone/pull/2840), [@kron4eg](https://github.com/kron4eg))
- Update DigitalOcean CSI to v4.6.1 ([#2840](https://github.com/kubermatic/kubeone/pull/2840), [@kron4eg](https://github.com/kron4eg))

#### Equinix Metal

- Update Equinix Metal CCM to v3.6.2 ([#2841](https://github.com/kubermatic/kubeone/pull/2841), [@kron4eg](https://github.com/kron4eg))

#### Google Cloud (GCP/GCE)

- Update GCP CSI driver to v1.10.1 ([#2843](https://github.com/kubermatic/kubeone/pull/2843), [@kron4eg](https://github.com/kron4eg))

#### Hetzner

- Update Hetzner CCM to v1.17.1 ([#2825](https://github.com/kubermatic/kubeone/pull/2825), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CCM to v1.16.0 ([#2816](https://github.com/kubermatic/kubeone/pull/2816), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CCM to v1.15.0 to support the new ARM instances ([#2774](https://github.com/kubermatic/kubeone/pull/2774), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CSI to v2.2.0 ([#2722](https://github.com/kubermatic/kubeone/pull/2722), [@xmudrii](https://github.com/xmudrii))

#### Nutanix

- Update Nutanix CSI to v2.6.3 ([#2817](https://github.com/kubermatic/kubeone/pull/2817), [@kron4eg](https://github.com/kron4eg))

#### OpenStack

- Update Openstack CCM and CSI to v1.27.1 ([#2819](https://github.com/kubermatic/kubeone/pull/2819), [@kron4eg](https://github.com/kron4eg))

#### vSphere

- Update vSphere CCM to v1.27.0 ([#2826](https://github.com/kubermatic/kubeone/pull/2826), [@kron4eg](https://github.com/kron4eg))
- Update vSphere CSI to v3.0.2 ([#2826](https://github.com/kubermatic/kubeone/pull/2826), [@kron4eg](https://github.com/kron4eg))

#### VMware Cloud Director (VCD)

- Update VMWare Cloud Director CSI driver to v1.4.0 ([#2827](https://github.com/kubermatic/kubeone/pull/2827), [@kron4eg](https://github.com/kron4eg))
- Update VMware Cloud Director CSI driver to v1.3.2 ([#2747](https://github.com/kubermatic/kubeone/pull/2747), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

#### Go

- KubeOne is now built with Go 1.20.5 ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built with Go 1.20.4 ([#2783](https://github.com/kubermatic/kubeone/pull/2783), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built with Go 1.20.3 ([#2756](https://github.com/kubermatic/kubeone/pull/2756), [@pkprzekwas](https://github.com/pkprzekwas))

### Bug or Regression


- Fix installing Helm charts containing CRDs ([#2839](https://github.com/kubermatic/kubeone/pull/2839), [@kron4eg](https://github.com/kron4eg))
- Fix defaulting for `vpc_id` in the example Terraform configs for AWS with dual-stack networking ([#2815](https://github.com/kubermatic/kubeone/pull/2815), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Fix some of issues reported by the CIS benchmark for the control plane nodes ([#2797](https://github.com/kubermatic/kubeone/pull/2797), [@kron4eg](https://github.com/kron4eg))
- Explicitly start Docker in the example Terraform configs for vSphere ([#2744](https://github.com/kubermatic/kubeone/pull/2744), [@kron4eg](https://github.com/kron4eg))

### Other (Cleanup or Flake)

- `net.ipv4.conf.all.rp_filter` sysctl config is now managed by Cilium instead of KubeOne ([#2894](https://github.com/kubermatic/kubeone/pull/2894), [@xmudrii](https://github.com/xmudrii))
- Apply the external CCM addon before applying user-provided addons ([#2861](https://github.com/kubermatic/kubeone/pull/2861), [@kron4eg](https://github.com/kron4eg))
- Redeploy AWS EBS CSI driver upon upgrading from earlier KubeOne versions to KubeOne 1.7 to update PodSelector labels ([#2824](https://github.com/kubermatic/kubeone/pull/2824), [@xmudrii](https://github.com/xmudrii))
- Redeploy OpenStack CCM and Cinder CSI driver upon upgrading from earlier KubeOne versions to KubeOne 1.7 to update PodSelector labels ([#2824](https://github.com/kubermatic/kubeone/pull/2824), [@xmudrii](https://github.com/xmudrii))
- Explicitly bind the pause image (version depends on Kubernetes version) to avoid version drift between kubeadm/kubelet and containerd ([#2812](https://github.com/kubermatic/kubeone/pull/2812), [@xmudrii](https://github.com/xmudrii))
- Run `kubeadm` preflight checks to validate that the cluster requirements are satisfied before initializing and provisioning a cluster ([#2759](https://github.com/kubermatic/kubeone/pull/2759), [@kron4eg](https://github.com/kron4eg))
- Ignore some `kubeadm` preflight checks when validating cluster requirements to account for adding new static worker nodes ([#2803](https://github.com/kubermatic/kubeone/pull/2803), [@xmudrii](https://github.com/xmudrii))
- Default to Basic SKU for Azure Load Balancers in the example Terraform configs for Azure ([#2858](https://github.com/kubermatic/kubeone/pull/2858), [@kron4eg](https://github.com/kron4eg))
- Rename anti-affinity rule for the control plane nodes in the example Terraform configs for vSphere to include the cluster name ([#2794](https://github.com/kubermatic/kubeone/pull/2794), [@WeirdMachine](https://github.com/WeirdMachine))
- Use `buildx` instead of Buildah to create multi-architecture KubeOne container images ([#2807](https://github.com/kubermatic/kubeone/pull/2807), [@xmudrii](https://github.com/xmudrii))

### Deprecation

- Remove `quay.io/kubermatic/kubeone-e2e` image and replace it with `quay.io/kubermatic/build` image ([#2783](https://github.com/kubermatic/kubeone/pull/2783), [@xmudrii](https://github.com/xmudrii))
