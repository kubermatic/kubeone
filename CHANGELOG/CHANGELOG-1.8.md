# [v1.8.0](https://github.com/kubermatic/kubeone/releases/tag/v1.8.0) - 2024-05-14

## Changelog since v1.7.0

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

- Refactor example Terraform configs for Hetzner to randomly generate the private network subnet in order to support creating multiple KubeOne clusters ([#3152](https://github.com/kubermatic/kubeone/pull/3152), [@xmudrii](https://github.com/xmudrii))
- Credentials defined in the credentials file now have precedence over credentials defined via environment variables. This change is made to match the behavior that's already documented in the KubeOne docs. If you use both the credentials file and the environment variables, we recommend double-checking your credentials file to make sure the credentials are up to date, as those credentials will be applied on the next `kubeone apply` run ([#2991](https://github.com/kubermatic/kubeone/pull/2991), [@kron4eg](https://github.com/kron4eg))
- kured has been removed, you have to re-enable it back in form of `helmRelease` ([#3024](https://github.com/kubermatic/kubeone/pull/3024), [@kron4eg](https://github.com/kron4eg))
- OSM: The latest Ubuntu 22.04 images on Azure have modified the configuration for `cloud-init` and how it accesses its datasource in Azure, in a breaking way. If you're having an Azure cluster, it's required to [refresh your machines](https://docs.kubermatic.com/kubeone/v1.7/cheat-sheets/rollout-machinedeployment/) with the latest provided OSPs to ensure that a system-wide package update doesn't result in broken machines. ([#3172](https://github.com/kubermatic/kubeone/pull/3172), [@xrstf](https://github.com/xrstf))
- The example Terraform configs for Azure have been migrated to use the Standard SKU for IP addresses. This is a breaking change for existing setups; in which case you should continue using your current SKU. Manual migration is possible by dissociating IP from the VM and LB, the migrating it, and assigning it back, however please consider all potential risks before doing this migration ([#3149](https://github.com/kubermatic/kubeone/pull/3149), [@kron4eg](https://github.com/kron4eg))
- Support for Docker is removed; `containerRuntime.docker` became a no-op. ([#3008](https://github.com/kubermatic/kubeone/pull/3008), [@kron4eg](https://github.com/kron4eg))

## Changes by Kind

### Feature

- Add GCP CCM ([#3038](https://github.com/kubermatic/kubeone/pull/3038), [@kron4eg](https://github.com/kron4eg))
- Add Nutanix CCM addon ([#3034](https://github.com/kubermatic/kubeone/pull/3034), [@kron4eg](https://github.com/kron4eg))
- Add `certOption` to the `hostConfig` ([#3020](https://github.com/kubermatic/kubeone/pull/3020), [@AhmadAlEdlbi](https://github.com/AhmadAlEdlbi))
- Add support for Kubernetes 1.28 ([#2948](https://github.com/kubermatic/kubeone/pull/2948), [@xmudrii](https://github.com/xmudrii))
- Add support for kubernetes 1.29 ([#3048](https://github.com/kubermatic/kubeone/pull/3048), [@kron4eg](https://github.com/kron4eg))
- Allow setting `CCM_CONCURRENT_SERVICE_SYNCS` parameter on CCM addons to configure number of concurrent `LoadBalancer` service reconciles ([#2916](https://github.com/kubermatic/kubeone/pull/2916), [@embik](https://github.com/embik))
- Canal CNI: Add `IFACE` and `IFACE_REGEX` parameters to allow explicitly selecting network interface to be used for inter-node communication and VXLAN ([#3152](https://github.com/kubermatic/kubeone/pull/3152), [@xmudrii](https://github.com/xmudrii))
- Hostnames are checked against Kubernetes node name requirements ([#3091](https://github.com/kubermatic/kubeone/pull/3091), [@SimonTheLeg](https://github.com/SimonTheLeg))
- Improve error messaging when working with remote files over SSH ([#3052](https://github.com/kubermatic/kubeone/pull/3052), [@kron4eg](https://github.com/kron4eg))
- Make v1.29 the default stable Kubernetes version ([#3073](https://github.com/kubermatic/kubeone/pull/3073), [@kron4eg](https://github.com/kron4eg))
- New API to configure TLS cipher suites for kube-apiserver, etcd and kubelet ([#3081](https://github.com/kubermatic/kubeone/pull/3081), [@kron4eg](https://github.com/kron4eg))
- Set `cloudProvider.external` = `true` by default for supported cloud providers in kubernetes 1.29+ ([#3048](https://github.com/kubermatic/kubeone/pull/3048), [@kron4eg](https://github.com/kron4eg))
- Support for customizing `vAppName` for VMware Cloud Director CSI driver ([#2932](https://github.com/kubermatic/kubeone/pull/2932), [@JamesClonk](https://github.com/JamesClonk))
- Support for passing additional args to the kube-apiserver, kube-controller-manager, and kube-scheduler ([#3162](https://github.com/kubermatic/kubeone/pull/3162), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Update to Go 1.22.1 ([#3072](https://github.com/kubermatic/kubeone/pull/3072), [@xrstf](https://github.com/xrstf))


### Bug or Regression

- Add escaping of registry name for the case when registry configured as wildcard. ([#2927](https://github.com/kubermatic/kubeone/pull/2927), [@kron4eg](https://github.com/kron4eg))
- Bind `FLANNELD_IFACE` statically to status.hostIP ([#3157](https://github.com/kubermatic/kubeone/pull/3157), [@kron4eg](https://github.com/kron4eg))
- Clean yum cache upon configuring Kubernetes repos. This fixes an issue with cluster upgrades failing on nodes with an older yum version ([#3146](https://github.com/kubermatic/kubeone/pull/3146), [@xmudrii](https://github.com/xmudrii))
- Delete AzureDisk's `csi-azuredisk-node-secret-binding` ClusterRoleBinding if RoleRef's name is `csi-azuredisk-node-sa` to allow upgrading KubeOne from 1.6 to 1.7 ([#2972](https://github.com/kubermatic/kubeone/pull/2972), [@xmudrii](https://github.com/xmudrii))
- Deploy user defined addons before the external CCM initialization. This fixes an issue with cluster provisioning for users that use both external CCM and external CNI ([#3065](https://github.com/kubermatic/kubeone/pull/3065), [@kron4eg](https://github.com/kron4eg))
- Don't use the deprecated path for GPG keys for Kubernetes and Docker repositories ([#2919](https://github.com/kubermatic/kubeone/pull/2919), [@xmudrii](https://github.com/xmudrii))
- Download cri-tools from the Kubernetes repos instead of the Amazon Linux 2 repos on instances running Amazon Linux 2 ([#2950](https://github.com/kubermatic/kubeone/pull/2950), [@xmudrii](https://github.com/xmudrii))
- Drop `containerRuntimeEndpoint` field from KubeletConfiguration to fix warning from `kubeadm init` and `kubeadm join` for clusters running Kubernetes prior to 1.27 ([#2939](https://github.com/kubermatic/kubeone/pull/2939), [@xmudrii](https://github.com/xmudrii))
- Fix Azure CCM crashlooping ([#3154](https://github.com/kubermatic/kubeone/pull/3154), [@kron4eg](https://github.com/kron4eg))
- Fix AzureDisk CSI `hostNetwork` ([#3150](https://github.com/kubermatic/kubeone/pull/3150), [@kron4eg](https://github.com/kron4eg))
- Fix DigitalOcean CSI driver ([#3139](https://github.com/kubermatic/kubeone/pull/3139), [@kron4eg](https://github.com/kron4eg))
- Fix Helm deploying resources in the wrong namespace ([#3000](https://github.com/kubermatic/kubeone/pull/3000), [@kron4eg](https://github.com/kron4eg))
- Fix a bug with the VMware Cloud Director CSI driver addon where it would crash if no `VCD_API_TOKEN` is set ([#2932](https://github.com/kubermatic/kubeone/pull/2932), [@JamesClonk](https://github.com/JamesClonk))
- Fix a globbing issue for `apt-get install` causing KubeOne to install wrong Kubernetes version in some circumstances ([#2958](https://github.com/kubermatic/kubeone/pull/2958), [@xmudrii](https://github.com/xmudrii))
- Fix cluster upgrades on Debian hosts with deprecated Kubernetes repositories ([#3076](https://github.com/kubermatic/kubeone/pull/3076), [@cnvergence](https://github.com/cnvergence))
- Fix file permissions setting on Flatcar ([#3138](https://github.com/kubermatic/kubeone/pull/3138), [@kron4eg](https://github.com/kron4eg))
- Fix incorrect validation that made `VCD_API_TOKEN` unusable for VMware Cloud Director ([#2945](https://github.com/kubermatic/kubeone/pull/2945), [@embik](https://github.com/embik))
- Fix indentation for manifests of csi-vsphere-ks addon ([#2905](https://github.com/kubermatic/kubeone/pull/2905), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Fix kubelet file permissions according to CIS 1.8 ([#3100](https://github.com/kubermatic/kubeone/pull/3100), [@kron4eg](https://github.com/kron4eg))
- Fix support for Flatcar stable channel 3815.2.0 ([#3040](https://github.com/kubermatic/kubeone/pull/3040), [@4ch3los](https://github.com/4ch3los))
- Fix: preserve containerd runtime config ([#3030](https://github.com/kubermatic/kubeone/pull/3030), [@tahajahangir](https://github.com/tahajahangir))
- Force `node-role.kubernetes.io/control-plane` label on control-plane Nodes ([#3099](https://github.com/kubermatic/kubeone/pull/3099), [@kron4eg](https://github.com/kron4eg))
- Initialize `host.Labels` to avoid NPE ([#3106](https://github.com/kubermatic/kubeone/pull/3106), [@kron4eg](https://github.com/kron4eg))
- Propagate CA Bundle to vSphere CSI driver ([#2906](https://github.com/kubermatic/kubeone/pull/2906), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- `registryConfiguration.OverrideRegistry` is correctly applied to the pause image configured in static nodes (control plane and static workers) ([#2925](https://github.com/kubermatic/kubeone/pull/2925), [@embik](https://github.com/embik))
- Update CRDs for operating-system-manager addon ([#2933](https://github.com/kubermatic/kubeone/pull/2933), [@ahmedwaleedmalik](https://github.com/ahmedwaleedmalik))
- Use the `CLUSTER_NAME` environment variable from the OpenStack CCM pods to determine the current cluster name used by the OpenStack CCM. Fixes a regression where cluster name was changed to `kubernetes` upon running `kubeone apply` two or more times after upgrading from KubeOne 1.6 to KubeOne 1.7. Please check the known issues document to find if you're affected by this issue and what steps you need to take if you're affected ([#2978](https://github.com/kubermatic/kubeone/pull/2978), [@xmudrii](https://github.com/xmudrii))

### Other (Cleanup or Flake)

- Increase the memory requests and limits from 300Mi to 600Mi for cluster-autoscaler ([#2978](https://github.com/kubermatic/kubeone/pull/2978), [@xmudrii](https://github.com/xmudrii))
- Extract csi-external-snapshotter into its own addon ([#3016](https://github.com/kubermatic/kubeone/pull/3016), [@kron4eg](https://github.com/kron4eg))
- Replace JSON6902 with Strategic Merge in Nutanix CSI driver ([#3035](https://github.com/kubermatic/kubeone/pull/3035), [@kron4eg](https://github.com/kron4eg))
- Use `DisableCloudProviders` feature gate as a replacement for `InTreePluginXXXUnregister` for each former in-tree provider ([#3075](https://github.com/kubermatic/kubeone/pull/3075), [@kron4eg](https://github.com/kron4eg))

### Updates

#### machine-controller

- - Update machine-controller to v1.59.1 ([#3184](https://github.com/kubermatic/kubeone/pull/3184), [@xmudrii](https://github.com/xmudrii))

#### operating-system-manager

- Update operating-system-manager to v1.5.1 ([#3165](https://github.com/kubermatic/kubeone/pull/3165), [@xrstf](https://github.com/xrstf))

#### Cloud Provider integrations

- Add Nutanix CCM and GCP CCM to the update-images template ([#3033](https://github.com/kubermatic/kubeone/pull/3033), [@kron4eg](https://github.com/kron4eg))
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
