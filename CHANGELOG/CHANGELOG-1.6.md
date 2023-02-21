# [v1.6.0](https://github.com/kubermatic/kubeone/releases/tag/v1.6.0) - 2023-02-23

We're happy to announce a new KubeOne minor release — KubeOne 1.6! Please
consult the changelog below before upgrading to this minor release.

## Changelog since v1.5.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- The minimum Kubernetes version is increased to v1.24.0. Please use KubeOne 1.5 to upgrade your clusters to Kubernetes 1.24 prior to upgrading to KubeOne 1.6. If your clusters are running Docker, please [migrate them to containerd using KubeOne 1.5](https://docs.kubermatic.com/kubeone/v1.5/guides/containerd-migration/). ([#2599](https://github.com/kubermatic/kubeone/pull/2599), [@xmudrii](https://github.com/xmudrii))
- Forbid Kubernetes 1.26 and newer for OpenStack clusters with the in-tree cloud provider. The in-tree cloud provider for OpenStack is removed with Kubernetes 1.26.0. Make sure to [migrate to the external CCM/CSI](https://docs.kubermatic.com/kubeone/v1.5/guides/ccm-csi-migration/) before upgrading to Kubernetes 1.26. ([#2573](https://github.com/kubermatic/kubeone/pull/2573), [@xmudrii](https://github.com/xmudrii))
- Add support for Ubuntu 22.04. Example Terraform configs for all providers are now using Ubuntu 22.04 by default. If you're using the latest Terraform configs with an existing cluster, make sure to bind the operating system/image to the image that you're currently using, otherwise your instances will get recreated ([#2367](https://github.com/kubermatic/kubeone/pull/2367), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- `control_plane_replicas` variable in Terraform configs for Hetzner is renamed to `control_plane_vm_count`. If you set the old variable explicitly, make sure to migrate to the new variable before migrating to the new configs ([#2550](https://github.com/kubermatic/kubeone/pull/2550), [@xmudrii](https://github.com/xmudrii))
- Forbid PodSecurityPolicy feature for Kubernetes clusters running 1.25 and newer. PodSecurityPolicies got removed from Kubernetes in 1.25. For more details, see [the official blog post](https://kubernetes.io/blog/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/) ([#2594](https://github.com/kubermatic/kubeone/pull/2594), [@xmudrii](https://github.com/xmudrii))
- Image references are changed from `k8s.gcr.io` to `registry.k8s.io`. This is done to keep up with [the latest upstream changes](https://github.com/kubernetes/enhancements/tree/master/keps/sig-release/3000-artifact-distribution). Please ensure that any mirrors you use are able to host `registry.k8s.io` and/or that firewall rules are going to allow access to `registry.k8s.io` to pull images. This change has been already introduced as part of KubeOne 1.5.4 and 1.4.12 patch releases ([#2501](https://github.com/kubermatic/kubeone/pull/2501), [@xmudrii](https://github.com/xmudrii))

## Changes by Kind

### API Change

- Stop applying `node-role.kubernetes.io/master` taint for Kubernetes 1.25+ nodes. The taint will be removed from existing nodes upon upgrading to Kubernetes 1.25 ([#2604](https://github.com/kubermatic/kubeone/pull/2604), [#2688](https://github.com/kubermatic/kubeone/pull/2688), [@xmudrii](https://github.com/xmudrii))
- Add a new `NodeLocalDNS` field to the KubeOneCluster API used to control should the NodeLocalDNSCache component be deployed or not. Run `kubeone config print --full` for details on how to use this field ([#2356](https://github.com/kubermatic/kubeone/pull/2356), [@kron4eg](https://github.com/kron4eg))
- Introduce a new `.addons.addons[*].disableTemplating` field to the KubeOneCluster API which can be used to disable templatization for an addon ([#2630](https://github.com/kubermatic/kubeone/pull/2630), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- `.cloudProvider.csiConfig` is now a mandatory field for vSphere clusters using the external cloud provider (`.cloudProvider.external: true`) ([#2430](https://github.com/kubermatic/kubeone/pull/2430), [@xmudrii](https://github.com/xmudrii))
- `.cloudProvider.csiConfig` can be specified for vSphere clusters even if the in-tree provider is used, but the provided CSIConfig is ignored in such cases (a warning about this is printed) ([#2430](https://github.com/kubermatic/kubeone/pull/2430), [@xmudrii](https://github.com/xmudrii))
- Allow overriding image repository for CoreDNS via `.features.coreDNS.imageRepository` field ([#2394](https://github.com/kubermatic/kubeone/pull/2394), [@xmudrii](https://github.com/xmudrii))

### Feature

#### General

- Add support for Helm-based addon ([#2498](https://github.com/kubermatic/kubeone/pull/2498), [@kron4eg](https://github.com/kron4eg))
  - Upgrade Helm releases only if it's differs from the already deployed release ([#2571](https://github.com/kubermatic/kubeone/pull/2571), [@kron4eg](https://github.com/kron4eg))
  - Uninstall Helm release that was installed by KubeOne but is not listed anymore in the KubeOneCluster manifest ([#2522](https://github.com/kubermatic/kubeone/pull/2522), [@kron4eg](https://github.com/kron4eg))
  - Fix Helm deployment of multiple charts ([#2515](https://github.com/kubermatic/kubeone/pull/2515), [@kron4eg](https://github.com/kron4eg))
- Implement a new `kubeone init` command used to generate the KubeOneCluster manifest and example Terraform configurations ([#2396](https://github.com/kubermatic/kubeone/pull/2396), [@kron4eg](https://github.com/kron4eg))
- Implement an interactive mode for the `kubeone init` subcommand ([#2552](https://github.com/kubermatic/kubeone/pull/2552), [@xmudrii](https://github.com/xmudrii))
- Add support for SSH Host Public key verification ([#2391](https://github.com/kubermatic/kubeone/pull/2391), [@kron4eg](https://github.com/kron4eg))
- Enable etcd compact hash checks as per [the recommendations from etcd for detecting data corruption](https://etcd.io/docs/v3.5/op-guide/data_corruption/#enabling-data-corruption-detection) ([#2497](https://github.com/kubermatic/kubeone/pull/2497), [@xmudrii](https://github.com/xmudrii))
- Schedule CSI Snapshot Validation webhook for OpenStack on the control plane nodes ([#2427](https://github.com/kubermatic/kubeone/pull/2427), [@xmudrii](https://github.com/xmudrii))
- Run kubeadm with increased verbosity unconditionally. This only changes the behavior if KubeOne is run without the verbose flag but kubeadm fails, in which case kubeadm is going to print more information about the issue ([#2556](https://github.com/kubermatic/kubeone/pull/2556), [@xmudrii](https://github.com/xmudrii))
- Expose machine-controller metrics port (8080/TCP), so Prometheus ServiceMonitor can be used for scraping ([#2421](https://github.com/kubermatic/kubeone/pull/2421), [@sphr2k](https://github.com/sphr2k))
- Migrate `ebs.csi.aws.com` CSIDriver to set `fsGroupPolicy: File` ([#2424](https://github.com/kubermatic/kubeone/pull/2424), [@xmudrii](https://github.com/xmudrii))

#### Terraform

- Add `allow_insecure` variable (default `false`) to Terraform configs for vSphere. The value of this variable is propagated to the MachineDeployment template in `output.tf` ([#2432](https://github.com/kubermatic/kubeone/pull/2432), [@xmudrii](https://github.com/xmudrii))
- Add `cluster_autoscaler_min_replicas` and `cluster_autoscaler_max_replicas` variables to Terraform configs. Those variables control the minimum and the maximum number of replicas for MachineDeployments. cluster-autoscaler must be enabled for those variables to have an effect ([#2551](https://github.com/kubermatic/kubeone/pull/2551), [@xmudrii](https://github.com/xmudrii))
- Add `control_plane_vm_count` variable to Terraform configs for DigitalOcean, Equinix Metal, GCE, Nutanix, OpenStack, and VMware Cloud Director (defaults to 3) ([#2546](https://github.com/kubermatic/kubeone/pull/2546), [@xmudrii](https://github.com/xmudrii))
- Add `os` variable to Terraform configs for DigitalOcean, Equinix Metal, and Hetzner (defaults to `ubuntu`) ([#2546](https://github.com/kubermatic/kubeone/pull/2546), [@xmudrii](https://github.com/xmudrii))
- Make volume size for worker nodes configurable in Terraform configs for AWS (50 GB by default) ([#2415](https://github.com/kubermatic/kubeone/pull/2415), [@xmudrii](https://github.com/xmudrii))
- Update Terraform provider for VMware Cloud Director to v3.8.1 ([#2583](https://github.com/kubermatic/kubeone/pull/2583), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Add support for insecure HTTPS connection to VMware Cloud Director API in the Terraform example configs ([#2583](https://github.com/kubermatic/kubeone/pull/2583), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

#### Addons

- Add a new addon parameter called `HubbleIPv6` (`true`/`false`, default: `true`) for Cilium CNI used to enable/disable Hubble UI listening on an IPv6 interface ([#2448](https://github.com/kubermatic/kubeone/pull/2448), [@xmudrii](https://github.com/xmudrii))

#### Experimental

- Add Experimental Dual-Stack IPv6 support for AWS with Canal/Calico and Cilium ([#2414](https://github.com/kubermatic/kubeone/pull/2414), [@PratikDeoghare](https://github.com/PratikDeoghare))

#### Kubernetes Version Support

- Add support for Kubernetes 1.26 ([#2568](https://github.com/kubermatic/kubeone/pull/2568), [@xmudrii](https://github.com/xmudrii))
- Add support for Kubernetes 1.25 ([#2405](https://github.com/kubermatic/kubeone/pull/2405), [@xmudrii](https://github.com/xmudrii))
- Add support for Kubernetes 1.25.5, 1.24.9, and 1.23.15. Upgrading to the latest Kubernetes 1.25 or 1.24 patch release is strongly advised because those releases are built with Go 1.19.4+ which includes fixes for [CVE-2022-41720 and CVE-2022-41717](https://groups.google.com/g/golang-announce/c/L_3rmdT0BMU/m/yZDrXjIiBQAJ) ([#2531](https://github.com/kubermatic/kubeone/pull/2531), [@xmudrii](https://github.com/xmudrii))
- Add support for Kubernetes 1.25.4, 1.24.8, and 1.23.14. Those Kubernetes patch releases fix CVE-2022-3162 and CVE-2022-3294, both in kube-apiserver:
    - [CVE-2022-3162: Unauthorized read of Custom Resources](https://groups.google.com/g/kubernetes-announce/c/oR2PUBiODNA/m/tShPgvpUDQAJ)
    - [CVE-2022-3294: Node address isn't always verified when proxying](https://groups.google.com/g/kubernetes-announce/c/eR0ghAXy2H8/m/sCuQQZlVDQAJ)
  We strongly recommend upgrading to the latest Kubernetes patch releases as soon as possible. ([#2466](https://github.com/kubermatic/kubeone/pull/2466), [@xmudrii](https://github.com/xmudrii))

### Updates

#### General

- Update containerd to 1.6. This change affects control plane nodes, static worker nodes, and nodes managed by machine-controller/operating-system-manager ([#2382](https://github.com/kubermatic/kubeone/pull/2382), [@kron4eg](https://github.com/kron4eg))
- Update containerd to 1.6 on Amazon Linux 2 ([#2601](https://github.com/kubermatic/kubeone/pull/2601), [@xmudrii](https://github.com/xmudrii))
- Update kubernetes-cni to v1.2.0 and cri-tools to v1.26.0 ([#2606](https://github.com/kubermatic/kubeone/pull/2606), [@xmudrii](https://github.com/xmudrii))
- Update kubernetes-cni to v1.1.1 to allow installation of Kubernetes v1.24.5+ ([#2353](https://github.com/kubermatic/kubeone/pull/2353), [@kron4eg](https://github.com/kron4eg))

#### Etcd

- Update etcd to 3.5.6 which includes a fix for [the reported data inconsistency issue for a case when etcd crashes during processing defragmentation operation](https://groups.google.com/a/kubernetes.io/g/dev/c/sEVopPxKPDo/m/9ME3CzicBwAJ) ([#2497](https://github.com/kubermatic/kubeone/pull/2497), [@xmudrii](https://github.com/xmudrii))
- Update etcd to 3.5.5 or use the version provided by kubeadm if it's newer ([#2419](https://github.com/kubermatic/kubeone/pull/2419), [@xmudrii](https://github.com/xmudrii))

#### CNI

- Update Canal to v3.23.5. This Canal release is supposed to fix an issue where Calico pods are crashing after upgrading from an older Calico version to a newer one (see the [Known Issues](https://docs.kubermatic.com/kubeone/v1.5/known-issues/) document for more details) ([#2538](https://github.com/kubermatic/kubeone/pull/2538), [@xmudrii](https://github.com/xmudrii))
- Update Cilium from v1.12.3 to v1.12.5 ([#2582](https://github.com/kubermatic/kubeone/pull/2582), [@xmudrii](https://github.com/xmudrii))
- Update Cilium to v1.12.3 ([#2478](https://github.com/kubermatic/kubeone/pull/2478), [@xmudrii](https://github.com/xmudrii))
- Upgrade Cilium to v1.12.2 ([#2359](https://github.com/kubermatic/kubeone/pull/2359), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

#### General Addons

- Update metrics-server to v0.6.2 ([#2580](https://github.com/kubermatic/kubeone/pull/2580), [@xmudrii](https://github.com/xmudrii))
- Update NodeLocalDNSCache to v1.22.15 ([#2580](https://github.com/kubermatic/kubeone/pull/2580), [@xmudrii](https://github.com/xmudrii))
- Update NodeLocalDNSCache to v1.22.13 ([#2477](https://github.com/kubermatic/kubeone/pull/2477), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler to v1.26.1 for Kubernetes 1.26+ clusters ([#2580](https://github.com/kubermatic/kubeone/pull/2580), [@xmudrii](https://github.com/xmudrii))
- Update cluster-autoscaler to v1.25.0 for Kubernetes 1.25 clusters ([#2476](https://github.com/kubermatic/kubeone/pull/2476), [@xmudrii](https://github.com/xmudrii))
- Update `backup-restic`, `unattended-updates`, `csi-vault-secret-provider`, `secrets-store-csi-driver` addons ([#2579](https://github.com/kubermatic/kubeone/pull/2579), [@kron4eg](https://github.com/kron4eg))

#### machine-controller and operating-system-manager

- Update machine-controller to v1.56.0 ([#2640](https://github.com/kubermatic/kubeone/pull/2640), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update operating-system-manager to v1.2.0 ([#2640](https://github.com/kubermatic/kubeone/pull/2640), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Upgrade to operating-system-manager v1.1.1 ([#2387](https://github.com/kubermatic/kubeone/pull/2387), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))

#### AWS

- Update AWS CCM to v1.26.0 and v1.24.3 ([#2569](https://github.com/kubermatic/kubeone/pull/2569), [@xmudrii](https://github.com/xmudrii))
- Update AWS CCM to v1.25.1 ([#2420](https://github.com/kubermatic/kubeone/pull/2420), [@xmudrii](https://github.com/xmudrii))
- Update AWS EBS CSI driver to v1.14.0 ([#2569](https://github.com/kubermatic/kubeone/pull/2569), [@xmudrii](https://github.com/xmudrii))
- Update AWS EBS CSI driver to v1.12.1 ([#2420](https://github.com/kubermatic/kubeone/pull/2420), [@xmudrii](https://github.com/xmudrii))
- Update Azure CCM to v1.26.0, v1.25.5, v1.24.11, v1.23.24 ([#2572](https://github.com/kubermatic/kubeone/pull/2572), [@xmudrii](https://github.com/xmudrii))

#### Azure

- Update Azure CCM to v1.25.3, v1.24.8, v1.23.21, v1.1.24 (for Kubernetes 1.22) ([#2422](https://github.com/kubermatic/kubeone/pull/2422), [@xmudrii](https://github.com/xmudrii))
- Update AzureDisk CSI to v1.25.0 ([#2572](https://github.com/kubermatic/kubeone/pull/2572), [@xmudrii](https://github.com/xmudrii))
- Update AzureDisk CSI driver to v1.23.0 ([#2422](https://github.com/kubermatic/kubeone/pull/2422), [@xmudrii](https://github.com/xmudrii))
- Update AzureFile CSI to v1.24.0 ([#2572](https://github.com/kubermatic/kubeone/pull/2572), [@xmudrii](https://github.com/xmudrii))
- Update AzureFile CSI driver to v1.22.0 ([#2422](https://github.com/kubermatic/kubeone/pull/2422), [@xmudrii](https://github.com/xmudrii))

#### DigitalOcean

- Update DigitalOcean CCM to v0.1.41 ([#2576](https://github.com/kubermatic/kubeone/pull/2576), [@xmudrii](https://github.com/xmudrii))
- Update DigitalOcean CCM to v0.1.40 ([#2475](https://github.com/kubermatic/kubeone/pull/2475), [@xmudrii](https://github.com/xmudrii))
- Update DigitalOcean CSI to v4.5.0 ([#2590](https://github.com/kubermatic/kubeone/pull/2590), [@xmudrii](https://github.com/xmudrii))
- Update DigitalOcean CSI to v4.4.1 ([#2474](https://github.com/kubermatic/kubeone/pull/2474), [@xmudrii](https://github.com/xmudrii))

#### Equinix Metal

- Update Equinix Metal CCM to v3.5.0 ([#2425](https://github.com/kubermatic/kubeone/pull/2425), [@xmudrii](https://github.com/xmudrii))

#### Google Cloud (GCP/GCE)

- Update GCP CSI driver to v1.8.1 ([#2576](https://github.com/kubermatic/kubeone/pull/2576), [@xmudrii](https://github.com/xmudrii))
- Update GCP CSI driver to v1.8.0 and external-snapshotter for GCP CSI to v6.1.0 ([#2471](https://github.com/kubermatic/kubeone/pull/2471), [@xmudrii](https://github.com/xmudrii))

#### Hetzner

- Update Hetzner CCM to v1.13.2 ([#2426](https://github.com/kubermatic/kubeone/pull/2426), [@xmudrii](https://github.com/xmudrii))
- Update Hetzner CSI to v2.1.0 ([#2578](https://github.com/kubermatic/kubeone/pull/2578), [@xmudrii](https://github.com/xmudrii))

#### Nutanix

- Update Nutanix CSI driver to v2.6.1 ([#2473](https://github.com/kubermatic/kubeone/pull/2473), [@xmudrii](https://github.com/xmudrii))

#### OpenStack

- Update OpenStack CCM to v1.26.0 for Kubernetes 1.26+ clusters ([#2582](https://github.com/kubermatic/kubeone/pull/2582), [@xmudrii](https://github.com/xmudrii))
- Update OpenStack Cinder CSI to v1.26.0 for Kubernetes 1.26+ clusters ([#2582](https://github.com/kubermatic/kubeone/pull/2582), [@xmudrii](https://github.com/xmudrii))
- Update OpenStack CCM and CSI to v1.25.3, v1.24.5, v1.22.2 ([#2427](https://github.com/kubermatic/kubeone/pull/2427), [@xmudrii](https://github.com/xmudrii))

#### vSphere

- Update vSphere CCM to v1.25.0, v1.24.2, v1.23.2 ([#2429](https://github.com/kubermatic/kubeone/pull/2429), [@xmudrii](https://github.com/xmudrii))
- Update vSphere CSI driver to v2.7.0 ([#2429](https://github.com/kubermatic/kubeone/pull/2429), [@xmudrii](https://github.com/xmudrii))

#### VMware Cloud Director (VCD)

- Update VMware Cloud Director (VCD) CSI driver from v1.2.0 to v1.3.1 ([#2576](https://github.com/kubermatic/kubeone/pull/2576), [@xmudrii](https://github.com/xmudrii))

#### Go

- KubeOne is now built using Go 1.19.6 ([#2649](https://github.com/kubermatic/kubeone/pull/2649), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built using Go 1.19.4 ([#2525](https://github.com/kubermatic/kubeone/pull/2525), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built using Go 1.19.3 ([#2461](https://github.com/kubermatic/kubeone/pull/2461), [@xmudrii](https://github.com/xmudrii))
- KubeOne is now built using Go 1.19.2 ([#2418](https://github.com/kubermatic/kubeone/pull/2418), [@xmudrii](https://github.com/xmudrii))

### Bug or Regression

- Automatically delete the CoreDNS PodDistruptionBudget if the feature is disabled ([#2364](https://github.com/kubermatic/kubeone/pull/2364), [@xmudrii](https://github.com/xmudrii))
- Use `vmware-system-csi` namespace when generating certs for the vSphere CSI webhooks ([#2366](https://github.com/kubermatic/kubeone/pull/2366), [@xmudrii](https://github.com/xmudrii))
- Recreate SSH connection in case of errors with the session ([#2345](https://github.com/kubermatic/kubeone/pull/2345), [@kron4eg](https://github.com/kron4eg))
- Fix SSH client failing to re-establish a broken SSH connection ([#2647](https://github.com/kubermatic/kubeone/pull/2647), [@kron4eg](https://github.com/kron4eg))
- Remove the leftover `/tmp/k1-etc-environment` file. This fixes an issue with `kubeone apply` failing if the username is changed ([#2560](https://github.com/kubermatic/kubeone/pull/2560), [@xmudrii](https://github.com/xmudrii))
- Use the pause image from `registry.k8s.io` for all Kubernetes releases ([#2528](https://github.com/kubermatic/kubeone/pull/2528), [@xmudrii](https://github.com/xmudrii))
- Fix Azure CCM failing to start because of unknown port flag ([#2647](https://github.com/kubermatic/kubeone/pull/2647), [@kron4eg](https://github.com/kron4eg))
- Fix an issue where the custom CA bundle was not being propagated to machine-controller-webhook ([#2586](https://github.com/kubermatic/kubeone/pull/2586), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Fix an issue where the custom CA bundle was not being propagated to operating-system-manager ([#2588](https://github.com/kubermatic/kubeone/pull/2588), [@FalcoSuessgott](https://github.com/FalcoSuessgott))
- Fix a panic (NPE) in cluster probes ([#2483](https://github.com/kubermatic/kubeone/pull/2483), [@kron4eg](https://github.com/kron4eg))
- Fix a panic (NPE) when determining if it is safe to repair a cluster when there's no kubelet or kubelet systemd unit on the node ([#2494](https://github.com/kubermatic/kubeone/pull/2494), [@xmudrii](https://github.com/xmudrii))
- Fix a panic (NPE) when the v1beta1 API is used ([#2349](https://github.com/kubermatic/kubeone/pull/2349), [@kron4eg](https://github.com/kron4eg))
- Fix a panic (NPE) when machine-controller deployment is disabled ([#2344](https://github.com/kubermatic/kubeone/pull/2344), [@kron4eg](https://github.com/kron4eg))
- Fix a panic (NPE) in case when building dynamic Kubernetes client failed on a previous try ([#2643](https://github.com/kubermatic/kubeone/pull/2643), [@WeirdMachine](https://github.com/WeirdMachine))
- Fix AMI filter for CentOS 7 in Terraform configs for AWS ([#2555](https://github.com/kubermatic/kubeone/pull/2555), [@xmudrii](https://github.com/xmudrii))
- Force-disable operating-system-manager (OSM) when the KubeOneCluster v1beta1 API is used ([#2354](https://github.com/kubermatic/kubeone/pull/2354), [@kron4eg](https://github.com/kron4eg))

### Other (Cleanup or Flake)

- The installation script (`install.sh`) has been modified to match only the stable releases ([#2355](https://github.com/kubermatic/kubeone/pull/2355), [@xmudrii](https://github.com/xmudrii))
- Change default branch from `master` to `main` ([#2400](https://github.com/kubermatic/kubeone/pull/2400), [@xrstf](https://github.com/xrstf))
- The `kubeone-e2e` image is moved from Docker Hub to Quay (`quay.io/kubermatic/kubeone-e2e`) ([#2463](https://github.com/kubermatic/kubeone/pull/2463), [@xmudrii](https://github.com/xmudrii))
- Remove the Kubernetes test binaries from the `kubeone-e2e` image because the new KubeOne E2E tests are using Sonobuoy instead ([#2404](https://github.com/kubermatic/kubeone/pull/2404), [@xmudrii](https://github.com/xmudrii))
- Rename `generate-internal-groups` Make target to `update-codegen` ([#2433](https://github.com/kubermatic/kubeone/pull/2433), [@xmudrii](https://github.com/xmudrii))
