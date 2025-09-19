# [v1.10.4](https://github.com/kubermatic/kubeone/releases/tag/v1.10.4) - 2025-09-19

## Changelog since v1.10.3

## Changes by Kind

### Chore

- Upgrade machine-controller version to [v1.61.4](https://github.com/kubermatic/machine-controller/releases/tag/v1.61.4) and operating-system-manager version to [v1.6.9](https://github.com/kubermatic/operating-system-manager/releases/tag/v1.6.9) ([#3819](https://github.com/kubermatic/kubeone/pull/3819), [@archups](https://github.com/archups))

### Bug or Regression

- Fix Nutanix credentials ([#3789](https://github.com/kubermatic/kubeone/pull/3789), [@kubermatic-bot](https://github.com/kubermatic-bot))
- Fix validation to pass when ChartURL is given ([#3825](https://github.com/kubermatic/kubeone/pull/3825), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.10.3](https://github.com/kubermatic/kubeone/releases/tag/v1.10.3) - 2025-07-24

## Changelog since v1.10.2

## Changes by Kind

### Bug or Regression

- Fix CSI snapshot webhook name for Nutanix [#3761](https://github.com/kubermatic/kubeone/pull/3761), [@kron4eg](https://github.com/kron4eg))

# [v1.10.2](https://github.com/kubermatic/kubeone/releases/tag/v1.10.2) - 2025-07-21

## Changelog since v1.10.1

## Changes by Kind

### Updates

#### machine-controller

- Update machine-controller to v1.61.3 ([#3673](https://github.com/kubermatic/kubeone/pull/3673), [@xmudrii](https://github.com/xmudrii))

#### operating-system-manager

- Update operating-system-manager to v1.6.7 ([#3739](https://github.com/kubermatic/kubeone/pull/3739), [@archups](https://github.com/archups))
- Update operating-system-manager to v1.6.6 ([#3737](https://github.com/kubermatic/kubeone/pull/3737), [@kron4eg](https://github.com/kron4eg))

#### Others

- Update Helm client to v3.17.4 ([#3752](https://github.com/kubermatic/kubeone/pull/3752), [@kron4eg](https://github.com/kron4eg))

# [v1.10.1](https://github.com/kubermatic/kubeone/releases/tag/v1.10.1) - 2025-06-13

## Changelog since v1.10.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- KubeVirt Cloud Controller Manager (CCM) is now deployed by default for all KubeVirt clusters. Two new fields are added to the API used to configure the CCM (`.cloudProvider.kubevirt.zoneAndRegionEnabled` and `.cloudProvider.kubevirt.loadBalancerEnabled`). `.cloudProvider.kubevirt.infraNamespace` is now a required field and KubeOne will fail validation if not set ([#3661](https://github.com/kubermatic/kubeone/pull/3661), [@moadqassem](https://github.com/moadqassem))
- [ACTION REQUIRED] The KubeVirt CCM requires some permissions to be added to the ServiceAccount that is bound to the infrastructure cluster kubeconfig in order to perform some tasks on the infrastructure side. For more information about the required roles please check [this file](https://github.com/kubevirt/cloud-provider-kubevirt/blob/v0.5.1/config/rbac/kccm_role.yaml)
- [ACTION REQUIRED] The `.cloudProvider.kubevirt.infraClusterKubeconfig` field has been removed from the KubeOneCluster type. Users must remove this field from their KubeOneCluster manifests otherwise the runtime validation will fail. The kubeconfig file provided via the `KUBEVIRT_KUBECONFIG` environment variable is used as a kubeconfig file for the infrastructure cluster ([#3675](https://github.com/kubermatic/kubeone/pull/3675), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### API Changes

- Add a new `annotations` field to `HostConfig` used to annotate control plane and static worker nodes ([#3658](https://github.com/kubermatic/kubeone/pull/3658), [@kron4eg](https://github.com/kron4eg))

### Bug or Regression

- Fix incorrect CABundle flag in the operating-system-manager (OSM) Deployment ([#3644](https://github.com/kubermatic/kubeone/pull/3644), [@kubermatic-bot](https://github.com/kubermatic-bot))

# [v1.10.0](https://github.com/kubermatic/kubeone/releases/tag/v1.10.0) - 2025-04-15

We're happy to announce a new KubeOne minor release — KubeOne 1.10! Please
consult the changelog below, as well as, the following two documents before
upgrading:

- [Upgrading from KubeOne 1.9 to 1.10 guide](https://docs.kubermatic.com/kubeone/v1.10/tutorials/upgrading/upgrading-from-1.9-to-1.10/)
- [Known Issues in KubeOne 1.10](https://docs.kubermatic.com/kubeone/v1.10/known-issues/)

## Changelog since v1.9.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Disallow using machine-controller and operating-system-manager with the cloud provider `none` (`.cloudProvider.none`). If you're affected by this change, you have to either disable machine-controller and/or operating-system-manager, or switch from the cloud provider `none` to a supported cloud provider ([#3369](https://github.com/kubermatic/kubeone/pull/3369), [@kron4eg](https://github.com/kron4eg))
- The Calico VXLAN optional addon has been removed from KubeOne. This addon has been non-functional for the past several releases. If you still need and use this addon, we advise using the [addons mechanism](https://docs.kubermatic.com/kubeone/v1.9/guides/addons/) to deploy it ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- The minimum kernel version for Kubernetes 1.32+ clusters is 4.19. Trying to provision a cluster with Kubernetes 1.32 or upgrade an existing cluster to Kubernetes 1.32, where nodes are not satisfying this requirement, will result in a pre-flight check failure ([#3590](https://github.com/kubermatic/kubeone/pull/3590), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### API Change

- Add `.cloudProvider.kubevirt.infraNamespace` field to the KubeOneCluster API used to control what namespace will be used by the KubeVirt provider to create and manage resources in the infra cluster, such as VirtualMachines and VirtualMachineInstances ([#3487](https://github.com/kubermatic/kubeone/pull/3487), [@moadqassem](https://github.com/moadqassem))
- Add a new optional field, `.cloudProvider.kubevirt.infraClusterKubeconfig`, to the KubeOneCluster API used to provide a kubeconfig file for a KubeVirt infra cluster (a cluster where KubeVirt is installed). This kubeconfig can be used by the CSI driver for provisioning volumes. ([#3499](https://github.com/kubermatic/kubeone/pull/3499), [@moadqassem](https://github.com/moadqassem))
- Change `kubeProxyReplacement` Type in `CiliumSpec` into boolean ([#3535](https://github.com/kubermatic/kubeone/pull/3535), [@mohamed-rafraf](https://github.com/mohamed-rafraf))

### CLI Change

- Always upgrade MachineDeployments to the target Kubernetes version upon running `kubeone apply` when the `--upgrade-machine-deployments` flag is set. Previously, MachineDeployments were upgraded only if the control plane was being upgraded ([#3528](https://github.com/kubermatic/kubeone/pull/3528), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Display `ensure machinedeployment` message (action) upon running `kubeone apply` only if `--create-machine-deployments` flag is set (`true` by default) ([#3477](https://github.com/kubermatic/kubeone/pull/3477), [@MaximilianMeister](https://github.com/MaximilianMeister))
- Add Fish shell auto-completion support ([#3471](https://github.com/kubermatic/kubeone/pull/3471), [@chenrui333](https://github.com/chenrui333))

### Feature

- Add support for Kubernetes 1.32 ([#3565](https://github.com/kubermatic/kubeone/pull/3565), [@kron4eg](https://github.com/kron4eg))
- Update Helm client to v3.17.2. This update allows users to pull Helm charts directly from OCI-compliant repositories ([#3587](https://github.com/kubermatic/kubeone/pull/3587), [@kron4eg](https://github.com/kron4eg))
- Add support for the KubeVirt CSI driver. The CSI driver is deployed automatically for all KubeVirt clusters (unless `.cloudProvider.disableBundledCSIDrivers` is set to `true`) ([#3499](https://github.com/kubermatic/kubeone/pull/3499), [@moadqassem](https://github.com/moadqassem))
- Label the control plane nodes before applying addons and Helm charts to allow addons and Helm charts to utilize the control plane label selectors ([#3544](https://github.com/kubermatic/kubeone/pull/3544), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Add `insecure` parameter to the `backups-restic` addon used to disable/skip the TLS verification ([#3522](https://github.com/kubermatic/kubeone/pull/3522), [@steled](https://github.com/steled))

### Bug or Regression

- Restart kubelet only after other upgrade tasks has been successfully completed. This fixes an issue where pods fail to start after the node is upgraded due to the `CreateContainerConfigError` error ([#3583](https://github.com/kubermatic/kubeone/pull/3583), [@kron4eg](https://github.com/kron4eg))
- KubeOne will remove orphaned etcd members when the control plane count is less than the number of etcd ring members ([#3584](https://github.com/kubermatic/kubeone/pull/3584), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Resolve the `clusterID` conflicts in cloud-config for AWS by prioritizing the cluster name from the Terraform configuration ([#3534](https://github.com/kubermatic/kubeone/pull/3534), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Drop trailing slash from the `VSPHERE_SERVER` variable to ensure compatibility with machine-controller and vSphere CCM and CSI ([#3537](https://github.com/kubermatic/kubeone/pull/3537), [@kron4eg](https://github.com/kron4eg))
- Cleanup the stale objects from the `unattended-upgrades` addon removed in KubeOne 1.8 ([#3538](https://github.com/kubermatic/kubeone/pull/3538), [@kron4eg](https://github.com/kron4eg))
- Fix an error message appearing in the KubeOne UI for clusters that don't have any Machine/MachineDeployment ([#3476](https://github.com/kubermatic/kubeone/pull/3476), [@soer3n](https://github.com/soer3n))
- Add `caBundle` volumeMounts to the `backups-restic` addon ([#3560](https://github.com/kubermatic/kubeone/pull/3560), [@kron4eg](https://github.com/kron4eg))

### Other (Cleanup or Flake)

- CNI plugins (`kubernetes-cni`) version on Flatcar now depends on the Kubernetes version ([#3632](https://github.com/kubermatic/kubeone/pull/3632), [@kron4eg](https://github.com/kron4eg))
- Use the GPG key from the latest Kubernetes package repository to fix failures to install older versions of Kubernetes packages ([#3524](https://github.com/kubermatic/kubeone/pull/3524), [@mohamed-rafraf](https://github.com/mohamed-rafraf))
- Use a dedicated keyring for Docker repositories to solve `apt-key` deprecation warning upon installing/upgrading containerd ([#3482](https://github.com/kubermatic/kubeone/pull/3482), [@kron4eg](https://github.com/kron4eg))
- Configure the `POD_NAMESPACE` environment variable for machine-controller-webhook on the KubeVirt clusters ([#3548](https://github.com/kubermatic/kubeone/pull/3548), [@moadqassem](https://github.com/moadqassem))
- Add 1 minute wait to the example Terraform configs for DigitalOcean to give enough time to freshly-created nodes to get upgraded and avoid issues with `apt-get` failing due to the `dpkg` lock file being present ([#3634](https://github.com/kubermatic/kubeone/pull/3634), [@kron4eg](https://github.com/kron4eg))

### Updates

#### machine-controller

- Update machine-controller to v1.61.1 ([#3630](https://github.com/kubermatic/kubeone/pull/3630), [@kron4eg](https://github.com/kron4eg))
- Update machine-controller to v1.61.0 ([#3546](https://github.com/kubermatic/kubeone/pull/3546), [@mohamed-rafraf](https://github.com/mohamed-rafraf))

#### operating-system-manager

- Update operating-system-manager to v1.6.4 ([#3630](https://github.com/kubermatic/kubeone/pull/3630), [@kron4eg](https://github.com/kron4eg))

#### CNIs

- Update Canal to v3.29.2 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update Cilium to v1.17.1 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update NodeLocalDNSCache to v1.25.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))

#### Cloud Provider integrations

- Update AWS CCM to v1.32.1, v1.31.5, and v1.30.7 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update Azure CCM to v1.32.1, v1.32.2, v1.30.8 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update AzureDisk CSI driver to v1.32.1 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update AzureFile CSI driver to v1.32.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CCM to v1.23.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update Hetzner CSI to v2.13.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update OpenStack CCM to v1.32.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update OpenStack CCM and CSI driver to v1.31.2 and v1.30.2 ([#3484](https://github.com/kubermatic/kubeone/pull/3484), [@rajaSahil](https://github.com/rajaSahil))
- Update Nutanix CCM to v0.5.0 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Downgrade csi-external-snapshotter to v8.1.0 ([#3622](https://github.com/kubermatic/kubeone/pull/3622), [@kron4eg](https://github.com/kron4eg))

#### Others

- KubeOne is now built with Go 1.23.4 ([#3509](https://github.com/kubermatic/kubeone/pull/3509), [@xmudrii](https://github.com/xmudrii))
- Update Helm client to v3.17.3 ([#3633](https://github.com/kubermatic/kubeone/pull/3633))
- Update cluster-autoscaler to v1.32.0, v1.31.1, and v1.30.3 ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
- Update Restic to v0.17.3 in the `backups-restic` addon ([#3568](https://github.com/kubermatic/kubeone/pull/3568), [@kron4eg](https://github.com/kron4eg))
