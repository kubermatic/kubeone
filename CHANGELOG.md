# Changelog

# [v1.2.1](https://github.com/kubermatic/kubeone/releases/tag/v1.2.1) - 2021-03-23

**Check out the changelog for the [v1.2.0 release](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0) for more information about what changes were introduced in the 1.2 release.**

## Changed

### Bug Fixes

* Install `cri-tools` (`crictl`) on Amazon Linux 2. This fixes the issue with provisioning Kubernetes and Amazon EKS-D clusters on Amazon Linux 2 ([#1282](https://github.com/kubermatic/kubeone/pull/1282))

# [v1.2.0](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0) - 2021-03-18

## Attention Needed

* [**BREAKING/ACTION REQUIRED**] Starting with the KubeOne 1.3 release, the `kubeone reset` command will require an explicit confirmation like the `apply` command
  * Running the `reset` command will require typing `yes` to confirm the intention to unprovision/reset the cluster
  * The command can be automatically approved by using the `--auto-approve` flag
  * The `--auto-approve` flag has been already implemented as a no-op flag in this release
  * Starting with this release, running `kubeone reset` will show a warning about this change each time the `reset` command is used
* [**BREAKING/ACTION REQUIRED**] Disallow and deprecate the PodPresets feature
  * If you're upgrading a cluster that uses the PodPresets feature from Kubernetes 1.19 to 1.20, you have to disable the PodPresets feature in the KubeOne configuration manifest
  * The PodPresets feature has been removed from Kubernetes 1.20 with no built-in replacement
  * It's not possible to use the PodPresets feature starting with Kubernetes 1.20, however, it currently remains possible to use it for older Kubernetes versions
  * The PodPresets feature will be removed from the KubeOneCluster API once Kubernetes 1.19 reaches End-of-Life (EOL)
  * As an alternative to the PodPresets feature, Kubernetes recommends using the MutatingAdmissionWebhooks.
* [**BREAKING/ACTION REQUIRED**] Support for CoreOS has been removed from KubeOne and machine-controller
  * CoreOS has reached End-of-Life on May 26, 2020
  * As an alternative to CoreOS, KubeOne supports Flatcar Linux
  * We recommend migrating your CoreOS clusters to the Flatcar Linux or other supported operating system
* [**BREAKING/ACTION REQUIRED**] Default values for OpenIDConnect has been corrected to match what's advised by the example configuration
  * Previously, there were no default values for the OpenIDConnect fields
  * This might only affect users using the OpenIDConnect feature
* Kubernetes has announced deprecation of the Docker (dockershim) support in
  the Kubernetes 1.20 release. It's expected that Docker support will be
  removed in Kubernetes 1.22
  * All newly created clusters running Kubernetes 1.21+ will be provisioned
    with containerd instead of Docker
  * Automated migration from Docker to containerd is currently not available,
    but is planned for one of the upcoming KubeOne releases
  * We highly recommend using containerd instead of Docker for all newly
    created clusters. You can opt-in to use containerd instead of Docker by
    adding `containerRuntime` configuration to your KubeOne configuration
    manifest:
    ```yaml
    containerRuntime:
      containerd: {}
    ```
    For the configuration file reference, run `kubeone config print --full`.

## Known Issues

* Provisioning a Kubernetes or Amazon EKS-D cluster on Amazon Linux 2 will fail due to missing `crictl` binary. This bug has been fixed in the [v1.2.1 release](https://github.com/kubermatic/kubeone/releases/tag/v1.2.1).
* Upgrading an Amazon EKS-D cluster will fail due to kubeadm preflight checks failing. We're investigating the issue and you can follow the progress by checking the issue [#1284](https://github.com/kubermatic/kubeone/issues/1284).

## Added

* Add support for Kubernetes 1.20
* Add support for containerd container runtime ([#1180](https://github.com/kubermatic/kubeone/pull/1180), [#1188](https://github.com/kubermatic/kubeone/pull/1188), [#1190](https://github.com/kubermatic/kubeone/pull/1190), [#1205](https://github.com/kubermatic/kubeone/pull/1205), [#1227](https://github.com/kubermatic/kubeone/pull/1227), [#1229](https://github.com/kubermatic/kubeone/pull/1229))
  * Kubernetes has announced deprecation of the Docker (dockershim) support in
    the Kubernetes 1.20 release. It's expected that Docker support will be
    removed in Kubernetes 1.22 or 1.23
  * All newly created clusters running Kubernetes 1.21+ will use
    containerd instead of Docker by default
  * Automated migration from Docker to containerd for existing clusters is
    currently not available, but is planned for one of the upcoming KubeOne
    releases
* Add support for Debian on control plane and static worker nodes ([#1233](https://github.com/kubermatic/kubeone/pull/1233))
  * Debian is currently not supported by machine-controller, so it's not
    possible to use it on worker nodes managed by Kubermatic machine-controller
* Add alpha-level support for Amazon Linux 2 ([#1167](https://github.com/kubermatic/kubeone/pull/1167), [#1173](https://github.com/kubermatic/kubeone/pull/1173), [#1175](https://github.com/kubermatic/kubeone/pull/1175), [#1176](https://github.com/kubermatic/kubeone/pull/1176))
  * Currently, all Kubernetes packages are installed by downloading binaries instead of using packages. Therefore, users are required to provide URLs using the new AssetConfiguration API to the CNI tarball, the Kubernetes Node binaries tarball (can be found in the Kubernetes CHANGELOG), and to the kubectl binary. Support for package managers is planned for the future.
* Add alpha-level AssetConfiguration API ([#1170](https://github.com/kubermatic/kubeone/pull/1170), [#1171](https://github.com/kubermatic/kubeone/pull/1171))
  * The AssetConfiguration API controls how assets are pulled
  * You can use it to specify custom images for containers or custom URLs for binaries
  * Currently-supported assets are CNI, Kubelet and Kubeadm (by providing a node binaries tarball), Kubectl, the control plane images, and the metrics-server image
  * Changing the binary assets (CNI, Kubelet, Kubeadm and Kubectl) currently works only on Amazon Linux 2. Changing the image assets works on all supported operating systems
* Add `Annotations` field to the `ProviderSpec` API used to add annotations to MachineDeployment objects ([#1174](https://github.com/kubermatic/kubeone/pull/1174))
* Add support for defining Static Worker nodes in Terraform ([#1166](https://github.com/kubermatic/kubeone/pull/1166))
* Add scrape Prometheus headless service for NodeLocalDNS ([#1165](https://github.com/kubermatic/kubeone/pull/1165))

## Changed

### API Changes

* [**BREAKING/ACTION REQUIRED**] Default values for OpenIDConnect has been corrected to match what's advised by the example configuration ([#1235](https://github.com/kubermatic/kubeone/pull/1235))
  * Previously, there were no default values for the OpenIDConnect fields
  * This might only affect users using the OpenIDConnect feature
* [**BREAKING/ACTION REQUIRED**] Disallow and deprecate the PodPresets feature ([#1236](https://github.com/kubermatic/kubeone/pull/1236))
  * If you're upgrading a cluster that uses the PodPresets feature from Kubernetes 1.19 to 1.20, you have to disable the PodPresets feature in the KubeOne configuration manifest
  * The PodPresets feature has been removed from Kubernetes 1.20 with no built-in replacement
  * It's not possible to use the PodPresets feature starting with Kubernetes 1.20, however, it currently remains possible to use it for older Kubernetes versions
  * The PodPresets feature will be removed from the KubeOneCluster API once Kubernetes 1.19 reaches End-of-Life (EOL)
  * As an alternative to the PodPresets feature, Kubernetes recommends using the MutatingAdmissionWebhooks.

### General

* Warn about `kubeone reset` requiring explicit confirmation starting with KubeOne 1.3 ([#1252](https://github.com/kubermatic/kubeone/pull/1252))
* Build KubeOne using Go 1.16.1 ([#1268](https://github.com/kubermatic/kubeone/pull/1268), [#1267](https://github.com/kubermatic/kubeone/pull/1267))
* Stop Kubelet and reload systemd when removing binaries on CoreOS/Flatcar ([#1176](https://github.com/kubermatic/kubeone/pull/1176))
* Add rsync on CentOS and Amazon Linux ([#1240](https://github.com/kubermatic/kubeone/pull/1240))

### Bug Fixes

* Drop mounting Flexvolume plugins into the OpenStack CCM. This fixes the issue with deploying the OpenStack CCM on the clusters running Flatcar Linux ([#1234](https://github.com/kubermatic/kubeone/pull/1234))
* Ensure all credentials are available to be used in addons. This fixes the issue with the Backups addon not working on non-AWS providers ([#1248](https://github.com/kubermatic/kubeone/pull/1248))
* Fix wrong legacy Docker version on RPM systems ([#1191](https://github.com/kubermatic/kubeone/pull/1191))

### Updated

* Update machine-controller to v1.25.0 ([#1238](https://github.com/kubermatic/kubeone/pull/1238))
* Update Calico CNI to v3.16.5 ([#1163](https://github.com/kubermatic/kubeone/pull/1163))

### Terraform Configs

* Replace GoBetween load-balancer in vSphere Terraform example by keepalived ([#1217](https://github.com/kubermatic/kubeone/pull/1217))

### Addons

* Fix DNS resolution issues for the Backups addon ([#1179](https://github.com/kubermatic/kubeone/pull/1179))

## Removed

* [**BREAKING/ACTION REQUIRED**] Support for CoreOS has been removed from KubeOne and machine-controller ([#1232](https://github.com/kubermatic/kubeone/pull/1232))
  * CoreOS has reached End-of-Life on May 26, 2020
  * As an alternative to CoreOS, KubeOne supports Flatcar Linux
  * We recommend migrating your CoreOS clusters to the Flatcar Linux or other supported operating system

# [v1.2.0-rc.1](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0-rc.1) - 2021-03-12

## Changed

### General

* Build KubeOne using Go 1.16.1 ([#1268](https://github.com/kubermatic/kubeone/pull/1268), [#1267](https://github.com/kubermatic/kubeone/pull/1267))

# [v1.2.0-rc.0](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0-rc.0) - 2021-03-08

## Attention Needed

* [**BREAKING/ACTION REQUIRED**] Starting with the KubeOne 1.3 release, the `kubeone reset` command will require an explicit confirmation like the `apply` command
  * Running the `reset` command will require typing `yes` to confirm the intention to unprovision/reset the cluster
  * The command can be automatically approved by using the `--auto-approve` flag
  * The `--auto-approve` flag has been already implemented as a no-op flag in this release
  * Starting with this release, running `kubeone reset` will show a warning about this change each time the `reset` command is used

## Changed

### General

* Warn about `kubeone reset` requiring explicit confirmation starting with KubeOne 1.3 ([#1252](https://github.com/kubermatic/kubeone/pull/1252))

# [v1.2.0-beta.1](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0-beta.1) - 2021-02-17

## Attention Needed

* [**Breaking**] Support for CoreOS has been removed from KubeOne and machine-controller
  * CoreOS has reached End-of-Life on May 26, 2020
  * As an alternative to CoreOS, KubeOne supports Flatcar Linux
  * We recommend migrating your CoreOS clusters to the Flatcar Linux or other supported operating system
* [**Breaking**] Default values for OpenIDConnect has been corrected to match what's advised by the example configuration
  * Previously, there were no default values for the OpenIDConnect fields
  * This might only affect users using the OpenIDConnect feature
* [**Breaking**] Disallow and deprecate the PodPresets feature
  * [**Action Required**] If you're upgrading a cluster that uses the PodPresets feature from Kubernetes 1.19 to 1.20, you have to disable the PodPresets feature in the KubeOne configuration manifest
  * The PodPresets feature has been removed from Kubernetes 1.20 with no built-in replacement
  * It's not possible to use the PodPresets feature starting with Kubernetes 1.20, however, it currently remains possible to use it for older Kubernetes versions
  * The PodPresets feature will be removed from the KubeOneCluster API once Kubernetes 1.19 reaches End-of-Life (EOL)
  * As an alternative to the PodPresets feature, Kubernetes recommends using the MutatingAdmissionWebhooks.

## Added

* Add support for Kubernetes 1.20
  * Previously, we've shared that there is an issue affecting newly created clusters where the first control plane node is unhealthy/broken for the first 5-10 minutes. We've investigated the issue and found out that the issue can be successfully mitigated by restarting the first API server. We've implemented a task that automatically restarts the API server if it's affected by the issue ([#1243](https://github.com/kubermatic/kubeone/pull/1243), [#1245](https://github.com/kubermatic/kubeone/pull/1245))
* Add support for Debian on control plane and static worker nodes ([#1233](https://github.com/kubermatic/kubeone/pull/1233))
  * Debian is currently not supported by machine-controller, so it's not possible to use it on worker nodes managed by machine-controller

## Changed

### API Changes

* [**Breaking**] Default values for OpenIDConnect has been corrected to match what's advised by the example configuration ([#1235](https://github.com/kubermatic/kubeone/pull/1235))
  * Previously, there were no default values for the OpenIDConnect fields
  * This might only affect users using the OpenIDConnect feature
* [**Breaking**] Disallow and deprecate the PodPresets feature ([#1236](https://github.com/kubermatic/kubeone/pull/1236))
  * [**Action Required**] If you're upgrading a cluster that uses the PodPresets feature from Kubernetes 1.19 to 1.20, you have to disable the PodPresets feature in the KubeOne configuration manifest
  * The PodPresets feature has been removed from Kubernetes 1.20 with no built-in replacement
  * It's not possible to use the PodPresets feature starting with Kubernetes 1.20, however, it currently remains possible to use it for older Kubernetes versions
  * The PodPresets feature will be removed from the KubeOneCluster API once Kubernetes 1.19 reaches End-of-Life (EOL)
  * As an alternative to the PodPresets feature, Kubernetes recommends using the MutatingAdmissionWebhooks.

### General

* Add rsync on CentOS and Amazon Linux ([#1240](https://github.com/kubermatic/kubeone/pull/1240))

### Bug Fixes

* Drop mounting Flexvolume plugins into the OpenStack CCM. This fixes the issue with deploying the OpenStack CCM on the clusters running Flatcar Linux ([#1234](https://github.com/kubermatic/kubeone/pull/1234))
* Ensure all credentials are available to be used in addons. This fixes the issue with the Backups addon not working on non-AWS providers ([#1248](https://github.com/kubermatic/kubeone/pull/1248))

### Updated

* Update machine-controller to v1.25.0 ([#1238](https://github.com/kubermatic/kubeone/pull/1238))

## Removed

* [**Breaking**] Support for CoreOS has been removed from KubeOne and machine-controller ([#1232](https://github.com/kubermatic/kubeone/pull/1232))
  * CoreOS has reached End-of-Life on May 26, 2020
  * As an alternative to CoreOS, KubeOne supports Flatcar Linux
  * We recommend migrating your CoreOS clusters to the Flatcar Linux or other supported operating system

# [v1.2.0-beta.0](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0-beta.0) - 2021-01-27

## Attention Needed

* Kubernetes has announced deprecation of the Docker (dockershim) support in
  the Kubernetes 1.20 release. It's expected that Docker support will be
  removed in Kubernetes 1.22
  * All newly created clusters running Kubernetes 1.21+ will be provisioned
    with containerd instead of Docker
  * Automated migration from Docker to containerd is currently not available,
    but is planned for one of the upcoming KubeOne releases
  * We highly recommend using containerd instead of Docker for all newly
    created clusters. You can opt-in to use containerd instead of Docker by
    adding `containerRuntime` configuration to your KubeOne configuration
    manifest:
    ```yaml
    containerRuntime:
      containerd: {}
    ```
    For the configuration file reference, run `kubeone config print --full`.


## Known Issues

* Provisioning Kubernetes 1.20 clusters results with one of the control plane
  nodes being unhealthy/broken for the first 5-10 minutes after provisioning
  the cluster. This causes KubeOne to fail to create MachineDeployment objects
  because the `machine-controller-webhook` service can't be found. Also, one of
  the NodeLocalDNS pods might get stuck in the crash loop.
  * KubeOne currently still doesn't support Kubernetes 1.20. We do **not**
    recommend provisioning 1.20 clusters or upgrading existing clusters to
    Kubernetes 1.20
  * We're currently investigating the issue. You can follow the progress
    in the issue [#1222](https://github.com/kubermatic/kubeone/issues/1222)

## Added

* Add support for containerd container runtime ([#1180](https://github.com/kubermatic/kubeone/pull/1180), [#1188](https://github.com/kubermatic/kubeone/pull/1188), [#1190](https://github.com/kubermatic/kubeone/pull/1190), [#1205](https://github.com/kubermatic/kubeone/pull/1205), [#1227](https://github.com/kubermatic/kubeone/pull/1227), [#1229](https://github.com/kubermatic/kubeone/pull/1229))
  * Kubernetes has announced deprecation of the Docker (dockershim) support in
    the Kubernetes 1.20 release. It's expected that Docker support will be
    removed in Kubernetes 1.22
  * All newly created clusters running Kubernetes 1.21+ will default to
    containerd instead of Docker
  * Automated migration from Docker to containerd is currently not available,
    but is planned for one of the upcoming KubeOne releases

## Changed

### Bug Fixes

* Fix wrong legacy Docker version on RPM systems ([#1191](https://github.com/kubermatic/kubeone/pull/1191))

### Terraform Configs

* Replace GoBetween load-balancer in vSphere Terraform example by keepalived ([#1217](https://github.com/kubermatic/kubeone/pull/1217))

### Addons

* Fix DNS resolution issues for the Backups addon ([#1179](https://github.com/kubermatic/kubeone/pull/1179))

# [v1.2.0-alpha.0](https://github.com/kubermatic/kubeone/releases/tag/v1.2.0-alpha.0) - 2020-11-27

## Added

* Add support for Amazon Linux 2 ([#1167](https://github.com/kubermatic/kubeone/pull/1167), [#1173](https://github.com/kubermatic/kubeone/pull/1173), [#1175](https://github.com/kubermatic/kubeone/pull/1175), [#1176](https://github.com/kubermatic/kubeone/pull/1176))
  * Support for Amazon Linux 2 is currently in alpha.
  * Currently, all Kubernetes packages are installed by downloading binaries instead of using packages. Therefore, users are required to provide URLs using the new AssetConfiguration API to the CNI tarball, the Kubernetes Node binaries tarball (can be found in the Kubernetes CHANGELOG), and to the kubectl binary. Support for packages is planned for the future.
* Add the AssetConfiguration API ([#1170](https://github.com/kubermatic/kubeone/pull/1170), [#1171](https://github.com/kubermatic/kubeone/pull/1171))
  * The AssetConfiguration API controls how assets are pulled.
  * You can use it to specify custom images for containers or custom URLs for binaries.
  * Currently-supported assets are CNI, Kubelet and Kubeadm (by providing a node binaries tarball), Kubectl, the control plane images, and the metrics-server image.
  * Changing the binary assets (CNI, Kubelet, Kubeadm and Kubectl) currently works only on Amazon Linux 2. Changing the image assets works on all supported operating systems.
* Add `Annotations` field to the `ProviderSpec` API used to add annotations to MachineDeployment objects ([#1174](https://github.com/kubermatic/kubeone/pull/1174))
* Support for defining Static Worker nodes in Terraform ([#1166](https://github.com/kubermatic/kubeone/pull/1166))
* Add scrape Prometheus headless service for NodeLocalDNS ([#1165](https://github.com/kubermatic/kubeone/pull/1165))

## Changed

### General

* Stop Kubelet and reload systemd when removing binaries on CoreOS/Flatcar ([#1176](https://github.com/kubermatic/kubeone/pull/1176))

### Updated

* Update Calico CNI to v3.16.5 ([#1163](https://github.com/kubermatic/kubeone/pull/1163))

# [v1.1.0](https://github.com/kubermatic/kubeone/releases/tag/v1.1.0) - 2020-11-13

**Changelog since v1.0.5.**

## Attention Needed

* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Hetzner Terraform config ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
  * It's **highly recommended** to bind the image by setting `var.image` to the image you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Added

### General

* Implement the OverwriteRegistry functionality ([#1145](https://github.com/kubermatic/kubeone/pull/1145))
  * This PR adds a new top-level API field `registryConfiguration` which controls how images used for components deployed by KubeOne and kubeadm are pulled from an image registry.
  * The `registryConfiguration.overwriteRegisty` field specifies a custom Docker registry to be used instead of the default one.
  * The `registryConfiguration.insecureRegistry` field configures Docker to consider the registry specified in `registryConfiguration.overwriteRegisty` as an insecure registry.
  * For example, if `registryConfiguration.overwriteRegisty` is set to `127.0.0.1:5000`, image called `k8s.gcr.io/kube-apiserver:v1.19.3` would become `127.0.0.1:5000/kube-apiserver:v1.19.3`.
  * Setting `registryConfiguration.overwriteRegisty` applies to all images deployed by KubeOne and kubeadm, including addons deployed by KubeOne.
  * Setting `registryConfiguration.overwriteRegisty` applies to worker nodes managed by machine-controller and KubeOne as well.
  * You can run `kubeone config print -f` for more details regarding the RegistryConfiguration API.
* Add external cloud controller manager support for VMware vSphere clusters ([#1159](https://github.com/kubermatic/kubeone/pull/1159))

### Addons

* Add the cluster-autoscaler addon ([#1103](https://github.com/kubermatic/kubeone/pull/1103))

## Changed

### Bug Fixes

* Explicitly restart Kubelet when upgrading clusters running on Ubuntu ([#1098](https://github.com/kubermatic/kubeone/pull/1098))
* Merge CloudProvider structs instead of overriding them when the cloud provider is defined via Terraform ([#1108](https://github.com/kubermatic/kubeone/pull/1108))

### Updated

* Update Flannel to v0.13.0 ([#1135](https://github.com/kubermatic/kubeone/pull/1135))
* Update WeaveNet to v2.7.0 ([#1153](https://github.com/kubermatic/kubeone/pull/1153))
* Update Hetzner Cloud Controller Manager (CCM) to v1.8.1 ([#1149](https://github.com/kubermatic/kubeone/pull/1149))
  * This CCM release includes support for external LoadBalacners backed by Hetzner LoadBalancers

### Terraform Configs

* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Hetzner Terraform config ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
  * It's **highly recommended** to bind the image by setting `var.image` to the image you're currently using to prevent the instances from being recreated the next time you run Terraform!
* Ensure the example Hetzner Terraform config support both Terraform v0.12 and v0.13 ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
* Update Azure example Terraform config to work with the latest versions of the Azure provider ([#1059](https://github.com/kubermatic/kubeone/pull/1059))
* Use Hetzner Load Balancers instead of GoBetween in the example Hetzner Terraform config ([#1066](https://github.com/kubermatic/kubeone/pull/1066))

# [v1.1.0-rc.0](https://github.com/kubermatic/kubeone/releases/tag/v1.1.0-rc.0) - 2020-10-27

**Changelog since v1.0.5.**

## Attention Needed

* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Hetzner Terraform config ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
  * It's **highly recommended** to bind the image by setting `var.image` to the image you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Added

### General

* Implement the OverwriteRegistry functionality ([#1145](https://github.com/kubermatic/kubeone/pull/1145))
  * This PR adds a new top-level API field `registryConfiguration` which controls how images used for components deployed by KubeOne and kubeadm are pulled from an image registry.
  * The `registryConfiguration.overwriteRegisty` field can be used to specify a custom Docker registry to be used instead of the default one.
  * For example, if `registryConfiguration.overwriteRegisty` is set to `127.0.0.1:5000`, image called `k8s.gcr.io/kube-apiserver:v1.19.3` would become `127.0.0.1:5000/kube-apiserver:v1.19.3`.
  * Setting `registryConfiguration.overwriteRegisty` applies to all images deployed by KubeOne and kubeadm, including addons deployed by KubeOne.
  * You can run `kubeone config print -f` for more details regarding the RegistryConfiguration API.

### Addons

* Add the cluster-autoscaler addon ([#1103](https://github.com/kubermatic/kubeone/pull/1103))

## Changed

### Bug Fixes

* Explicitly restart Kubelet when upgrading clusters running on Ubuntu ([#1098](https://github.com/kubermatic/kubeone/pull/1098))
* Merge CloudProvider structs instead of overriding them when the cloud provider is defined via Terraform ([#1108](https://github.com/kubermatic/kubeone/pull/1108))

### Updated

* Update Flannel to v0.13.0 ([#1135](https://github.com/kubermatic/kubeone/pull/1135))
* Update Hetzner Cloud Controller Manager (CCM) to v1.7.0 ([#1068](https://github.com/kubermatic/kubeone/pull/1068))
  * This CCM release includes support for external LoadBalacners backed by Hetzner LoadBalancers

### Terraform Configs

* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Hetzner Terraform config ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
  * It's **highly recommended** to bind the image by setting `var.image` to the image you're currently using to prevent the instances from being recreated the next time you run Terraform!
* Ensure the example Hetzner Terraform config support both Terraform v0.12 and v0.13 ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
* Update Azure example Terraform config to work with the latest versions of the Azure provider ([#1059](https://github.com/kubermatic/kubeone/pull/1059))
* Use Hetzner Load Balancers instead of GoBetween in the example Hetzner Terraform config ([#1066](https://github.com/kubermatic/kubeone/pull/1066))

# [v1.0.5](https://github.com/kubermatic/kubeone/releases/tag/v1.0.5) - 2020-10-19

## Changed

### Updated

* Update machine-controller to v1.19.0 ([#1141](https://github.com/kubermatic/kubeone/pull/1141))
  * This machine-controller release uses the Hyperkube Kubelet image for Flatcar worker nodes running Kubernetes 1.18, as the Poseidon Kubelet image repository doesn't publish 1.18 images any longer. This change ensures that you can provision or upgrade to Kubernetes 1.18.8+ on Flatcar.

# [v1.0.4](https://github.com/kubermatic/kubeone/releases/tag/v1.0.4) - 2020-10-16

## Attention Needed

* KubeOne now creates a dedicated secret with vSphere credentials for kube-controller-manager and vSphere Cloud Controller Manager (CCM). This is required because those components require the secret to adhere to the expected format.
  * The new secret is called `vsphere-ccm-credentials` and is deployed in the `kube-system` namespace.
  * This fix ensures that you can use all vSphere provider features, for example, volumes and having cloud provider metadata/labels applied on each node.
  * If you're upgrading an existing vSphere cluster, you should take the following steps:
    * Upgrade the cluster without changing the cloud-config. Upgrading the cluster creates a new Secret for kube-controller-manager and vSphere CCM.
    * Change the cloud-config in your KubeOne Configuration Manifest to refer to the new secret.
    * Upgrade the cluster again to apply the new cloud-config, or change it manually on each control plane node and restart kube-controller-manager and vSphere CCM (if deployed).
    * **Note: the cluster can be force-upgrade without changing the Kubernetes version. To do that, you can use the `--force-upgrade` flag with `kubeone apply`, or the `--force` flag with `kubeone upgrade`.**
* The example Terraform config for vSphere now has `disk.enableUUID` option enabled. This change ensures that you can mount volumes on the control plane nodes.
  * **WARNING: If you're applying the latest Terraform configs on an existing infrastructure/cluster, an in-place upgrade of control plane nodes is required. This means that all control plane nodes will be restarted, which causes a downtime until all instances doesn't come up again.**

## Changed

### Bug Fixes

* Don't stop Kubelet when upgrading Kubeadm on Flatcar ([#1099](https://github.com/kubermatic/kubeone/pull/1099))
* Create a dedicated secret with vSphere credentials for kube-controller-manager and vSphere Cloud Controller Manager (CCM) ([#1128](https://github.com/kubermatic/kubeone/pull/1128))
* Enable `disk.enableUUID` option in Terraform example configs for vSphere ([#1130](https://github.com/kubermatic/kubeone/pull/1130))

# [v1.0.3](https://github.com/kubermatic/kubeone/releases/tag/v1.0.3) - 2020-09-28

## Attention Needed

* This release includes a fix for Kubernetes 1.18.9 clusters failing to provision due to the unmet kubernetes-cni dependency.
* This release includes a fix for CentOS and RHEL clusters failing to provision due to missing Docker versions from the CentOS/RHEL 8 Docker repo.

## Added

* Add the `image` field for the Hetzner worker nodes ([#1105](https://github.com/kubermatic/kubeone/pull/1105))

## Changed

### General

* Use CentOS 7 Docker repo for both CentOS/RHEL 7 and 8 clusters ([#1111](https://github.com/kubermatic/kubeone/pull/1111))

### Updated

* Update machine-controller to v1.18.0 ([#1114](https://github.com/kubermatic/kubeone/pull/1114))
* Update kubernetes-cni to v0.8.7 ([#1100](https://github.com/kubermatic/kubeone/pull/1100))

# [v1.0.2](https://github.com/kubermatic/kubeone/releases/tag/v1.0.2) - 2020-09-02

## Known Issues

* We do **not** recommend upgrading to Kubernetes 1.19.0 due to an [upstream bug in kubeadm affecting Highly-Available clusters](https://github.com/kubernetes/kubeadm/issues/2271). The bug will be fixed in 1.19.1, which is scheduled for Wednesday, September 9th. Until 1.19.1 is not released, we highly recommend staying on 1.18.

## Changed

### Updated

* Update machine-controller to v1.17.1 ([#1086](https://github.com/kubermatic/kubeone/pull/1086))
  * This machine-controller release uses Docker 19.03.12 instead of Docker 18.06 for worker nodes running Kubernetes 1.17 and newer.
  * Due to an [upstream issue](https://github.com/kubernetes/kubernetes/issues/94281), pod metrics are not available on worker nodes running Kubernetes 1.19 with Docker 18.06.
  * If you experience any issues with pod metrics on worker nodes running Kubernetes 1.19, provisioned with an earlier machine-controller version, you might have to update machine-controller to v1.17.1 and then rotate the affected worker nodes.

# [v1.0.1](https://github.com/kubermatic/kubeone/releases/tag/v1.0.1) - 2020-08-31

## Known Issues

* We do **not** recommend upgrading to Kubernetes 1.19.0 due to an [upstream bug in kubeadm affecting Highly-Available clusters](https://github.com/kubernetes/kubeadm/issues/2271). The bug will be fixed in 1.19.1, which is scheduled for Wednesday, September 9th. Until 1.19.1 is not released, we highly recommend staying on 1.18.

## Changed

### General

* Include cloud-config in the generated KubeOne configuration manifest when it's required by validation ([#1062](https://github.com/kubermatic/kubeone/pull/1062))

### Bug Fixes

* Properly apply labels when upgrading components. This fixes the issue with upgrade failures on clusters created with KubeOne v1.0.0-rc.0 and earlier ([#1078](https://github.com/kubermatic/kubeone/pull/1078))
* Fix race condition between kube-proxy and node-local-dns ([#1058](https://github.com/kubermatic/kubeone/pull/1058))

# [v1.0.0](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0) - 2020-08-18

**Changelog since v0.11.2. For changelog since v1.0.0-rc.1, please check the [release notes](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0)**

## Attention Needed

* Upgrading to this release is **highly recommended** as v0.11 release doesn't support Kubernetes versions 1.16.11/1.17.7/1.18.4 and newer. Kubernetes versions older than 1.16.11/1.17.7/1.18.4 are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.

* KubeOne now uses vanity domain `k8c.io/kubeone`.
  * The `go get` command to get KubeOne is now `GO111MODULE=on go get k8c.io/kubeone`.
* The `kubeone` AUR package has been moved to the official Arch Linux repositories. The AUR package has been removed in the favor of the official one.

* This release introduces the new KubeOneCluster v1beta1 API. The v1alpha1 API has been deprecated.
  * It remains possible to use both APIs with all `kubeone` commands
  * The v1alpha1 manifest can be converted to the v1beta1 manifest using the `kubeone config migrate` command
  * All example configurations have been updated to the v1beta1 API
  * More information about migrating to the new API and what has been changed can be found in the [API migration document](https://docs.kubermatic.com/kubeone/master/advanced/api_migration/)
* Default MTU for the Canal CNI depending on the provider
  * AWS - 8951 (9001 AWS Jumbo Frame - 50 VXLAN bytes)
  * GCE - 1410 (GCE specific 1460 bytes - 50 VXLAN bytes)
  * Hetzner - 1400 (Hetzner specific 1450 bytes - 50 VXLAN bytes)
  * OpenStack - 1400 (OpenStack specific 1450 bytes - 50 VXLAN bytes)
  * Default - 1450
  * If you're using KubeOneCluster v1alpha1 API, the default MTU is 1450 regardless of the provider
* RHSMOfflineToken has been removed from the CloudProviderSpec. This and other relevant fields are now located in the OperatingSystemSpec

* The KubeOneCluster manifest (config file) is now provided using the `--manifest` flag, such as `kubeone install --manifest config.yaml`. Providing it as an argument will result in an error
* The paths to the config files for PodNodeSelector feature (`.features.config.configFilePath`) and StaticAuditLog
feature (`.features.staticAuditLog.policyFilePath`) are now relative to the manifest path instead of to the working
directory. This change might be breaking for some users.

* It's now possible to install Kubernetes 1.18.6 and 1.17.9 on CentOS 7, however, only Canal CNI is known to work properly. We are aware that the DNS and networking problems may still be present even with the latest versions. It remains impossible to install older versions of Kubernetes on CentOS 7.

* Example Terraform configs for AWS now Use Ubuntu 20.04 (Focal) instead of Ubuntu 18.04
  * It's **highly recommended** to bind the AMI by setting `var.ami` to the AMI you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Known Issues

* We do **not** recommend upgrading to Kubernetes 1.19.0 due to an [upstream bug in kubeadm affecting Highly-Available clusters](https://github.com/kubernetes/kubeadm/issues/2271). The bug will be fixed in 1.19.1, which is scheduled for Wednesday, September 9th. Until 1.19.1 is not released, we highly recommend staying on 1.18.
* It remains impossible to provision Kubernetes older than 1.18.6/1.17.9 on CentOS 7. CentOS 8 and RHEL are unaffected.
* Upgrading a cluster created with KubeOne v1.0.0-rc.0 or earlier fails to update components deployed by KubeOne (e.g. CNI, machine-controller). Please upgrade to KubeOne v1.0.1 or newer if you have clusters created with v1.0.0-rc.0 or earlier.

## Added

* Implement the `kubeone apply` command
  * The apply command is used to reconcile (install, upgrade, and repair) clusters
  * More details about how to use the apply command can be found in the [Cluster reconciliation (apply) document](https://docs.kubermatic.com/kubeone/master/advanced/cluster_reconciliation/)
* Implement the `kubeone config machinedeployments` command ([#966](https://github.com/kubermatic/kubeone/pull/966))
  * The new command is used to generate a YAML manifest containing all MachineDeployment objects defined in the KubeOne configuration manifest and Terraform output
  * The generated manifest can be used with kubectl if you want to create and modify MachineDeployments once the cluster is created
* Add the `kubeone proxy` command ([#1035](https://github.com/kubermatic/kubeone/pull/1035))
  * The `kubeone proxy` command launches a local HTTPS capable proxy (CONNECT method) to tunnel HTTPS request via SSH. This is especially useful in cases when the Kubernetes API endpoint is not accessible from the public internets.
* Add the KubeOneCluster v1beta1 API ([#894](https://github.com/kubermatic/kubeone/pull/894))
  * Implemented automated conversion between v1alpha1 and v1beta1 APIs. It remains possible to use all `kubeone` commands with both v1alpha1 and v1beta1 manifests, however, migration to the v1beta1 manifest is recommended
  * Implement the Terraform integration for the v1beta1 API. Currently, the Terraform integration output format is the same for both APIs, but that might change in the future
  * The kubeone config migrate command has been refactored to migrate v1alpha1 to v1beta1 manifests. The manifest is now provided using the --manifest flag instead of providing it as an argument. It's not possible to convert pre-v0.6.0 manifest to v1alpha1 anymore
  * The example configurations are updated to the v1beta1 API
  * Drop the leading 'v' in the Kubernetes version if it's provided. This fixes a bug causing provisioning to fail if the Kubernetes version starts with 'v'
* Automatic cluster repairs ([#888](https://github.com/kubermatic/kubeone/pull/888))
  * Detect and delete broken etcd members
  * Detect and delete outdated corev1.Node objects
* Add ability to provision static worker nodes ([#834](https://github.com/kubermatic/kubeone/pull/834))
  * Check out [the documentation](https://docs.kubermatic.com/kubeone/master/workers/static_workers/) to learn more about static worker nodes
* Add ability to skip cluster provisioning when running the `install` command using the `--no-init` flag ([#871](https://github.com/kubermatic/kubeone/pull/871))
* Add support for Kubernetes 1.16.11, 1.17.7, 1.18.4 releases ([#925](https://github.com/kubermatic/kubeone/pull/925))
* Add support for Ubuntu 20.04 (Focal) ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Add support for CentOS 8 ([#981](https://github.com/kubermatic/kubeone/pull/981))
* Add RHEL support ([#918](https://github.com/kubermatic/kubeone/pull/918))
* Add support for Flatcar Linux ([#879](https://github.com/kubermatic/kubeone/pull/879))
* Support for vSphere resource pools ([#883](https://github.com/kubermatic/kubeone/pull/883))
* Support for Azure AZs ([#883](https://github.com/kubermatic/kubeone/pull/883))
* Support for Flexvolumes on CoreOS and Flatcar ([#885](https://github.com/kubermatic/kubeone/pull/885))
* Add the `ImagePlan` field to Azure Terraform integration ([#947](https://github.com/kubermatic/kubeone/pull/947))
* Add support for the PodNodeSelector admission controller ([#920](https://github.com/kubermatic/kubeone/pull/920))
* Add the `PodPresets` feature ([#837](https://github.com/kubermatic/kubeone/pull/837))
* Add support for OpenStack external cloud controller manager (CCM) ([#820](https://github.com/kubermatic/kubeone/pull/820))
* Add ability to change MTU for the Canal CNI using `.clusterNetwork.cni.canal.mtu` field (requires KubeOneCluster v1beta1 API) ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Add the Calico VXLAN addon ([#972](https://github.com/kubermatic/kubeone/pull/972))
  * More information about how to use this addon can be found on the [docs website](https://docs.kubermatic.com/kubeone/master/using_kubeone/calico-vxlan-addon/)
* Add ability to use an external CNI plugin ([#862](https://github.com/kubermatic/kubeone/pull/862))

## Changed

### General

* [**Breaking**] Replace positional argument for the config file with the global `--manifest` flag ([#880](https://github.com/kubermatic/kubeone/pull/880))
* [**Breaking**] Make paths to the config file for PodNodeSelector and StaticAuditLog features relative to the manifest path ([#920](https://github.com/kubermatic/kubeone/pull/920))
  * The default value for the `--manifest` flag is `kubeone.yaml`
* KubeOne now uses vanity domain `k8c.io/kubeone` ([#1008](https://github.com/kubermatic/kubeone/pull/1008))
  * The `go get` command to get KubeOne is now `GO111MODULE=on go get k8c.io/kubeone`.
* The `kubeone` AUR package has been moved to the official Arch Linux repositories. The AUR package has been removed in the favor of the official one ([#971](https://github.com/kubermatic/kubeone/pull/971))
* Default MTU for the Canal CNI depending on the provider ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
  * AWS - 8951 (9001 AWS Jumbo Frame - 50 VXLAN bytes)
  * GCE - 1410 (GCE specific 1460 bytes - 50 VXLAN bytes)
  * Hetzner - 1400 (Hetzner specific 1450 bytes - 50 VXLAN bytes)
  * OpenStack - 1400 (OpenStack specific 1450 bytes - 50 VXLAN bytes)
  * Default - 1450
  * If you're using KubeOneCluster v1alpha1 API, the default MTU is 1450 regardless of the provider ([#1016](https://github.com/kubermatic/kubeone/pull/1016))
* Label all components deployed by KubeOne with the `kubeone.io/component: <component-name>` label ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Increase default number of task retries to 10 ([#1020](https://github.com/kubermatic/kubeone/pull/1020))
* Use the proper package revision for Kubernetes packages ([#933](https://github.com/kubermatic/kubeone/pull/933))
* Hold the `docker-ce-cli` package on Ubuntu ([#902](https://github.com/kubermatic/kubeone/pull/902))
* Hold the `docker-ce`, `docker-ce-cli`, and all Kubernetes packages on CentOS ([#902](https://github.com/kubermatic/kubeone/pull/902))
* Ignore etcd data if it is present on the control plane nodes ([#874](https://github.com/kubermatic/kubeone/pull/874))
  * This allows clusters to be restored from a backup
* machine-controller and machine-controller-webhook are bound to the control plane nodes ([#832](https://github.com/kubermatic/kubeone/pull/832))
* KubeOne is now built using Go 1.15 ([#1048](https://github.com/kubermatic/kubeone/pull/1048))

### Bug Fixes

* Verify that `crd.projectcalico.org` CRDs are established before proceeding with the Canal CNI installation ([#994](https://github.com/kubermatic/kubeone/pull/994))
* Add NodeRegistration object to the kubeadm v1beta2 JoinConfiguration for static worker nodes. Fix the issue with nodes not joining a cluster on AWS ([#969](https://github.com/kubermatic/kubeone/pull/969))
* Unconditionally renew certificates when upgrading the cluster. Due to an upstream bug, kubeadm wasn't automatically renewing certificates for clusters running Kubernetes versions older than v1.17 ([#990](https://github.com/kubermatic/kubeone/pull/990))
* Apply only addons with `.yaml`, `.yml` and `.json` extensions ([#873](https://github.com/kubermatic/kubeone/pull/873))
* Install `curl` before configuring repositories on Ubuntu instances ([#945](https://github.com/kubermatic/kubeone/pull/945))
* Stop copying kubeconfig to the home directory on control plane instances ([#936](https://github.com/kubermatic/kubeone/pull/936))
* Fix the cluster provisioning issues caused by `docker-ce-cli` version mismatch ([#896](https://github.com/kubermatic/kubeone/pull/896))
* Remove hold from the docker-ce-cli package on upgrade ([#941](https://github.com/kubermatic/kubeone/pull/941))
  * This ensures clusters created with KubeOne v0.11 can be upgraded using KubeOne v1.0.0
  * This fixes the issue with upgrading CoreOS/Flatcar clusters
* Force restart Kubelet on CentOS on upgrade ([#988](https://github.com/kubermatic/kubeone/pull/988))
  * Fixes the cluster provisioning for instances that don't have `curl` installed
* Fix CoreOS host architecture detection ([#882](https://github.com/kubermatic/kubeone/pull/882))
* Fix CoreOS/Flatcar cluster provisioning/upgrading: use the correct URL for the CNI plugins ([#929](https://github.com/kubermatic/kubeone/pull/929))
* Fix the Kubelet service unit for CoreOS/Flatcar ([#908](https://github.com/kubermatic/kubeone/pull/908), [#909](https://github.com/kubermatic/kubeone/pull/909))
* Fix the CoreOS install and upgrade scripts ([#904](https://github.com/kubermatic/kubeone/pull/904))
* Fix CoreOS/Flatcar provisioning issues for Kubernetes 1.18 ([#895](https://github.com/kubermatic/kubeone/pull/895))
* Fix Kubelet version detection on Flatcar ([#1032](https://github.com/kubermatic/kubeone/pull/1032))
* Fix the gobetween script failing to install the `tar` package ([#963](https://github.com/kubermatic/kubeone/pull/963))

### Updated

* Update machine-controller to v1.16.1 ([#1043](https://github.com/kubermatic/kubeone/pull/1043))
* Update Canal CNI to v3.15.1 ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Update NodeLocalDNSCache to v1.15.12 ([#872](https://github.com/kubermatic/kubeone/pull/872))
* Update Docker versions. Minimal Docker version is 18.09.9 for clusters older than v1.17 and 19.03.12 for clusters running v1.17 and newer ([#1002](https://github.com/kubermatic/kubeone/pull/1002/files))
  * This fixes the issue preventing users to upgrade Kubernetes clusters created with KubeOne v0.11 or earlier
* Update Packet Cloud Controller Manager (CCM) to v1.0.0 ([#884](https://github.com/kubermatic/kubeone/pull/884))
* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Terraform scripts for AWS ([#1001](https://github.com/kubermatic/kubeone/pull/1001))
  * It's **highly recommended** to bind the AMI by setting `var.ami` to the AMI you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Removed

* Remove RHSMOfflineToken from the CloudProviderSpec ([#883](https://github.com/kubermatic/kubeone/pull/883))
  * RHSM settings are now located in the OperatingSystemSpec

# [v1.0.0-rc.1](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-rc.1) - 2020-08-06

## Attention Needed

* KubeOne now uses vanity domain `k8c.io/kubeone` ([#1008](https://github.com/kubermatic/kubeone/pull/1008))
  * The `go get` command to get KubeOne is now `GO111MODULE=on go get k8c.io/kubeone`.
  * You may be required to specify a tag/branch name until v1.0.0 is not released, such as: `GO111MODULE=on go get k8c.io/kubeone@v1.0.0-rc.1`.
* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Terraform scripts for AWS ([#1001](https://github.com/kubermatic/kubeone/pull/1001))
  * It's **highly recommended** to bind the AMI by setting `var.ami` to the AMI you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Added

* Add ability to change MTU for the Canal CNI using `.clusterNetwork.cni.canal.mtu` field (requires KubeOneCluster v1beta1 API) ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Add support for Ubuntu 20.04 (Focal) ([#1005](https://github.com/kubermatic/kubeone/pull/1005))

## Changed

### General

* KubeOne now uses vanity domain `k8c.io/kubeone` ([#1008](https://github.com/kubermatic/kubeone/pull/1008))
  * The `go get` command to get KubeOne is now `GO111MODULE=on go get k8c.io/kubeone`.
  * You may be required to specify a tag/branch name until v1.0.0 is not released, such as: `GO111MODULE=on go get k8c.io/kubeone@v1.0.0-rc.1`.
* Default MTU for the Canal CNI depending on the provider ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
  * AWS - 8951 (9001 AWS Jumbo Frame - 50 VXLAN bytes)
  * GCE - 1410 (GCE specific 1460 bytes - 50 VXLAN bytes)
  * Hetzner - 1400 (Hetzner specific 1450 bytes - 50 VXLAN bytes)
  * OpenStack - 1400 (OpenStack specific 1450 bytes - 50 VXLAN bytes)
  * Default - 1450
  * If you're using KubeOneCluster v1alpha1 API, the default MTU is 1450 regardless of the provider ([#1016](https://github.com/kubermatic/kubeone/pull/1016))
* Label all components deployed by KubeOne with the `kubeone.io/component: <component-name>` label ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* Increase default number of task retries to 10 ([#1020](https://github.com/kubermatic/kubeone/pull/1020))

### Updated

* Update machine-controller to v1.16.0 ([#1017](https://github.com/kubermatic/kubeone/pull/1017))
* Update Canal CNI to v3.15.1 ([#1005](https://github.com/kubermatic/kubeone/pull/1005))
* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Terraform scripts for AWS ([#1001](https://github.com/kubermatic/kubeone/pull/1001))
  * It's **highly recommended** to bind the AMI by setting `var.ami` to the AMI you're currently using to prevent the instances from being recreated the next time you run Terraform!

# [v1.0.0-rc.0](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-rc.0) - 2020-07-27

## Attention Needed

* This is the first release candidate for the KubeOne v1.0.0 release! We encourage everyone to test it and let us know if you have any questions or if you find any issues or bugs. You can [create an issue on GitHub](https://github.com/kubermatic/kubeone/issues/new/choose), reach out to us on [our forums](http://forum.kubermatic.com/), or on [`#kubeone` channel](https://kubernetes.slack.com/messages/CNEV2UMT7) on [Kubernetes Slack](http://slack.k8s.io/).
* It's recommended to use this release instead of v0.11, as v0.11 doesn't support the latest Kubernetes patch releases. Older Kubernetes releases are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.

## Changed

### Bug Fixes

* Verify that `crd.projectcalico.org` CRDs are established before proceeding with the Canal CNI installation ([#994](https://github.com/kubermatic/kubeone/pull/994))

### Updated

* Update Docker versions. Minimal Docker version is 18.09.9 for clusters older than v1.17 and 19.03.12 for clusters running v1.17 and newer ([#1002](https://github.com/kubermatic/kubeone/pull/1002/files))
  * This fixes the issue preventing users to upgrade Kubernetes clusters created with KubeOne v0.11 or earlier

# [v1.0.0-beta.3](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-beta.3) - 2020-07-22

## Attention Needed

* It's recommended to use this release instead of v0.11, as v0.11 doesn't support the latest Kubernetes patch releases. Older Kubernetes releases are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.
* It's now possible to install Kubernetes 1.18.6/1.17.9/1.16.13 on CentOS 7, however, only Canal CNI is known to work properly. We are aware that the DNS and networking problems may still be present even with the latest versions. It remains impossible to install older versions of Kubernetes on CentOS 7.
* The `kubeone` AUR package has been moved to the official Arch Linux repositories. The AUR package has been removed in the favor of the official one ([#971](https://github.com/kubermatic/kubeone/pull/971)).

## Added

* Implement the `kubeone apply` command
  * The apply command is used to reconcile (install, upgrade, and repair) clusters
  * The apply command is currently in beta, but we're encouraging everyone to test it and report any issues and bugs
  * More details about how to use the apply command can be found in the [Cluster reconciliation (apply) document](https://docs.kubermatic.com/kubeone/master/using_kubeone/cluster_reconciliation/)
* Implement the `kubeone config machinedeployments` command ([#966](https://github.com/kubermatic/kubeone/pull/966))
  * The new command is used to generate a YAML manifest containing all MachineDeployment objects defined in the KubeOne configuration manifest and Terraform output
  * The generated manifest can be used with kubectl if you want to create and modify MachineDeployments once the cluster is created
* Add support for CentOS 8 ([#981](https://github.com/kubermatic/kubeone/pull/981))
* Add the Calico VXLAN addon ([#972](https://github.com/kubermatic/kubeone/pull/972))
  * More information about how to use this addon can be found on the [docs website](https://docs.kubermatic.com/kubeone/master/using_kubeone/calico-vxlan-addon/)

## Changed

### General

* The `kubeone` AUR package has been moved to the official Arch Linux repositories. The AUR package has been removed in the favor of the official one ([#971](https://github.com/kubermatic/kubeone/pull/971))

### Bug Fixes

* Add NodeRegistration object to the kubeadm v1beta2 JoinConfiguration for static worker nodes. Fix the issue with nodes not joining a cluster on AWS ([#969](https://github.com/kubermatic/kubeone/pull/969))
* Unconditionally renew certificates when upgrading the cluster. Due to an upstream bug, kubeadm wasn't automatically renewing certificates for clusters running Kubernetes versions older than v1.17 ([#990](https://github.com/kubermatic/kubeone/pull/990))
* Force restart Kubelet on CentOS on upgrade ([#988](https://github.com/kubermatic/kubeone/pull/988))
* Fix the gobetween script failing to install the `tar` package ([#963](https://github.com/kubermatic/kubeone/pull/963))

### Updated

* Update machine-controller to v1.15.3 ([#995](https://github.com/kubermatic/kubeone/pull/995))
  * This release includes a fix for the machine-controller's NodeCSRApprover controller refusing to approve certificates for the GCP worker nodes
  * This release includes support for CentOS 8 and RHEL 7 worker nodes

# [v1.0.0-beta.2](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-beta.2) - 2020-07-03

## Attention Needed

* It's recommended to use this release instead of v0.11, as v0.11 doesn't support the latest Kubernetes patch releases. Older Kubernetes releases are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.

## Known Issues

* See known issues for the [v1.0.0-beta.1 release](https://github.com/kubermatic/kubeone/blob/master/CHANGELOG.md#known-issues) for more details.

## Added

* Add the `ImagePlan` field to Azure Terraform integration ([#947](https://github.com/kubermatic/kubeone/pull/947))

## Changed

### Bug Fixes

* Install `curl` before configuring repositories on Ubuntu instances ([#945](https://github.com/kubermatic/kubeone/pull/945))
  * Fixes the cluster provisioning for instances that don't have `curl` installed

### Updated

* Update machine-controller to v1.15.1 ([#947](https://github.com/kubermatic/kubeone/pull/947))

# [v1.0.0-beta.1](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-beta.1) - 2020-07-02

## Attention Needed

* It's recommended to use this release instead of v0.11, as v0.11 doesn't support the latest Kubernetes patch releases. Older Kubernetes releases are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.

## Known Issues

* machine-controller is failing to sign Kubelet CertificateSingingRequests (CSRs) for worker nodes on GCP due to [missing hostname in the Machine object](https://github.com/kubermatic/machine-controller/issues/781). We are currently working on a fix. In meanwhile, you can sign the CSRs manually by following [instructions](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#approving-certificate-signing-requests) from the Kubernetes docs.
* It remains impossible to provision Kubernetes 1.16.10+ clusters on CentOS 7. CentOS 8 and RHEL are unaffected. We are investigating the root cause of the issue.

## Changed

### Bug Fixes

* Remove hold from the docker-ce-cli package on upgrade ([#941](https://github.com/kubermatic/kubeone/pull/941))
  * This ensures clusters created with KubeOne v0.11 can be upgraded using KubeOne v1.0.0-beta.1 release

# [v1.0.0-beta.0](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-beta.0) - 2020-07-01

## Attention Needed

* It's recommended to use this release instead of v0.11, as v0.11 doesn't support the latest Kubernetes patch releases. Older Kubernetes releases are affected by two CVEs and therefore it's strongly advised to use 1.16.11/1.17.7/1.18.4 or newer.

## Known Issues

* machine-controller is failing to sign Kubelet CertificateSingingRequests (CSRs) for worker nodes on GCP due to [missing hostname in the Machine object](https://github.com/kubermatic/machine-controller/issues/781). We are currently working on a fix. In meanwhile, you can sign the CSRs manually by following [instructions](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#approving-certificate-signing-requests) from the Kubernetes docs.
* It remains impossible to provision Kubernetes 1.16.10+ clusters on CentOS 7. CentOS 8 and RHEL are unaffected. We are investigating the root cause of the issue.
* Trying to upgrade cluster created with KubeOne v0.11.2 or older results in an error due to KubeOne failing to upgrade the `docker-ce-cli` package. This issue has been fixed in the v1.0.0-beta.1 release.

## Changed

### General

* Re-introduce the support for the kubernetes-cni package and use the proper revision for Kubernetes packages ([#933](https://github.com/kubermatic/kubeone/pull/933))
  * This change ensures operators can use KubeOne to install the latest Kubernetes releases on CentOS/RHEL

### Bug Fixes

* Stop copying kubeconfig to the home directory on control plane instances ([#936](https://github.com/kubermatic/kubeone/pull/936))
  * This fixes the issue with upgrading CoreOS/Flatcar clusters

### Updated

* Update machine-controller to v1.14.3 ([#937](https://github.com/kubermatic/kubeone/pull/937))

# [v1.0.0-alpha.6](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.6) - 2020-06-19

## Changed

### General

* Fix CoreOS/Flatcar cluster provisioning/upgrading: use the correct URL for the CNI plugins ([#929](https://github.com/kubermatic/kubeone/pull/929))

# [v1.0.0-alpha.5](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.5) - 2020-06-18

## Attention Needed

* This release includes support for Kubernetes releases 1.16.11, 1.17.7, and 1.18.4. Those releases use the CNI v0.8.6,
which includes a fix for the CVE-2020-10749. It's **strongly advised** to upgrade your clusters and rotate worker nodes
as soon as possible.
* The path to the config file for PodNodeSelector feature (`.features.config.configFilePath`) and StaticAuditLog
feature (`.features.staticAuditLog.policyFilePath`) are now relative to the manifest path instead of to the working
directory. This change might be breaking for some users.

## Known Issues

* Currently it's impossible to install Kubernetes versions older than 1.16.11/1.17.7/1.18.4 on CentOS/RHEL, due
to an issue with packages.
* Currently it's impossible to install Kubernetes on CentOS 7. The 1.16.11 release is not working due to an [upstream
issue](https://github.com/kubernetes/kubernetes/issues/92250) with the kube-proxy component, while newer releases are
having DNS problems which we are investigating.
* Currently it's impossible to install Kubernetes on CoreOS/Flatcar due to an incorrect URL
to the CNI plugin. The fix has been already merged and will be included in the upcoming release.

## Added

* Add RHEL support ([#918](https://github.com/kubermatic/kubeone/pull/918))
* Add support for the PodNodeSelector admission controller ([#920](https://github.com/kubermatic/kubeone/pull/920))
* Add support for Kubernetes 1.16.11, 1.17.7, 1.18.4 releases ([#925](https://github.com/kubermatic/kubeone/pull/925))

## Changed

### General

* BREAKING: Make paths to the config file for PodNodeSelector and StaticAuditLog features relative to the manifest path ([#920](https://github.com/kubermatic/kubeone/pull/920))

### Updated

* Update machine-controller to v1.14.2 ([#918](https://github.com/kubermatic/kubeone/pull/918))

# [v1.0.0-alpha.4](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.4) - 2020-05-21

## Changed

### General

* Fix the Kubelet service unit for CoreOS/Flatcar ([#908](https://github.com/kubermatic/kubeone/pull/908), [#909](https://github.com/kubermatic/kubeone/pull/909))

# [v1.0.0-alpha.3](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.3) - 2020-05-21

## Attention Needed

* In the v1.0.0-alpha.2 release, we didn't explicitly hold the `docker-ce-cli` package, which means it can be upgraded to a newer version. Upgrading the `docker-ce-cli` package would render the Kubernetes cluster unusable because of the version mismatch between Docker daemon and CLI.
* If you already provisioned a cluster using the v1.0.0-alpha.2 release, please hold the `docker-ce-cli` package to prevent it from being upgraded:
  * Ubuntu: run the following command over SSH on all control plane instances: `sudo apt-mark hold docker-ce-cli`
  * CentOS: we've started using `yum versionlock` to handle package locks. The best way to set it up on your control plane instances is to run `kubeone upgrade -f` with the exact same config that's currently running.

## Known Issues

* The CoreOS cluster provisioning has been broken in this release due to the incorrect formatted Kubelet service file. This issue has been fixed in the v1.0.0-alpha.4 release.

## Changed

### General

* Hold the `docker-ce-cli` package on Ubuntu ([#902](https://github.com/kubermatic/kubeone/pull/902))
* Hold the `docker-ce`, `docker-ce-cli`, and all Kubernetes packages on CentOS ([#902](https://github.com/kubermatic/kubeone/pull/902))
* Fix the CoreOS install and upgrade scripts ([#904](https://github.com/kubermatic/kubeone/pull/904))

# [v1.0.0-alpha.2](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.2) - 2020-05-20

## Attention Needed

* This alpha version fixes the provisioning failures caused by `docker-ce-cli` version mismatch. The older alpha release are not working anymore ([#896](https://github.com/kubermatic/kubeone/pull/896))
* `machine-controller` must be updated to v1.14.0 or newer on existing clusters or otherwise newly created worker nodes will not work properly. The `machine-controller` can be updated on one of the following ways:
  * (Recommended) Run `kubeone upgrade -f` with the exact same config that's currently running 
  * Run `kubeone install` with the exact same config that's currently running
  * Update the `machine-controller` and `machine-controller-webhook` deployments manually
* This release introduces the new KubeOneCluster v1beta1 API. The v1alpha1 API has been deprecated.
  * It remains possible to use both APIs with all `kubeone` commands
  * The v1alpha1 manifest can be converted to the v1beta1 manifest using the `kubeone config migrate` command
  * All example configurations have been updated to the v1beta1 API

## Known Issues

* The `docker-ce-cli` package is not put on hold after it's installed. Upgrading the `docker-ce-cli` package would render the Kubernetes cluster unusable because of the version mismatch between Docker daemon and CLI. It's highly recommended to upgrade to KubeOne v1.0.0-alpha.3.
* If you already provisioned a cluster using the v1.0.0-alpha.2 release, please hold the `docker-ce-cli` package to prevent it from being upgraded:
  * Ubuntu: run the following command over SSH on all control plane instances: `sudo apt-mark hold docker-ce-cli`
  * CentOS: we've started using `yum versionlock` to handle package locks. The best way to set it up on your control plane instances is to run `kubeone upgrade -f` with the exact same config that's currently running.
* The CoreOS cluster provisioning has been broken in this release due to the incorrect formatted Kubelet service file. This issue has been fixed in the v1.0.0-alpha.4 release.

## Added

* Add the KubeOneCluster v1beta1 API ([#894](https://github.com/kubermatic/kubeone/pull/894))
  * Implemented automated conversion between v1alpha1 and v1beta1 APIs. It remains possible to use all `kubeone` commands with both v1alpha1 and v1beta1 manifests, however, migration to the v1beta1 manifest is recommended
  * Implement the Terraform integration for the v1beta1 API. Currently, the Terraform integration output format is the same for both APIs, but that might change in the future
  * The kubeone config migrate command has been refactored to migrate v1alpha1 to v1beta1 manifests. The manifest is now provided using the --manifest flag instead of providing it as an argument. It's not possible to convert pre-v0.6.0 manifest to v1alpha1 anymore
  * The example configurations are updated to the v1beta1 API
  * Drop the leading 'v' in the Kubernetes version if it's provided. This fixes a bug causing provisioning to fail if the Kubernetes version starts with 'v'

## Changed

### General

* Fix the cluster provisioning issues caused by `docker-ce-cli` version mismatch ([#896](https://github.com/kubermatic/kubeone/pull/896))
* Fix CoreOS/Flatcar provisioning issues for Kubernetes 1.18 ([#895](https://github.com/kubermatic/kubeone/pull/895))

### Updated

* Update machine-controller to v1.14.0 ([#899](https://github.com/kubermatic/kubeone/pull/899))

# [v1.0.0-alpha.1](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.1) - 2020-05-11

## Attention Needed

* The KubeOneCluster manifest (config file) is now provided using the `--manifest` flag, such as `kubeone install --manifest config.yaml`. Providing it as an argument will result in an error
* RHSMOfflineToken has been removed from the CloudProviderSpec. This and other relevant fields are now located in the OperatingSystemSpec
* Packet Cloud Controller Manager (CCM) has been updated to v1.0.0. This update fixes the issue preventing users to provision new Packet clusters

## Added

* Initial support for Flatcar Linux ([#879](https://github.com/kubermatic/kubeone/pull/879))
  * Currently, Flatcar Linux is supported only on AWS worker nodes
* Support for vSphere resource pools ([#883](https://github.com/kubermatic/kubeone/pull/883))
* Support for Azure AZs ([#883](https://github.com/kubermatic/kubeone/pull/883))
* Support for Flexvolumes on CoreOS and Flatcar ([#885](https://github.com/kubermatic/kubeone/pull/885))
* Automatic cluster repairs ([#888](https://github.com/kubermatic/kubeone/pull/888))
  * Detect and delete broken etcd members
  * Detect and delete outdated corev1.Node objects

## Changed

### General

* [Breaking] Replace positional argument for the config file with the global `--manifest` flag ([#880](https://github.com/kubermatic/kubeone/pull/880))
  * The default value for the `--manifest` flag is `kubeone.yaml`
* Ignore etcd data if it is present on the control plane nodes ([#874](https://github.com/kubermatic/kubeone/pull/874))
  * This allows clusters to be restored from a backup

### Bug Fixes

* Fix CoreOS host architecture detection ([#882](https://github.com/kubermatic/kubeone/pull/882))

### Updated

* Update machine-controller to v1.13.1 ([#885](https://github.com/kubermatic/kubeone/pull/885))
* Update Packet Cloud Controller Manager (CCM) to v1.0.0 ([#884](https://github.com/kubermatic/kubeone/pull/884))

## Removed

* Remove RHSMOfflineToken from the CloudProviderSpec ([#883](https://github.com/kubermatic/kubeone/pull/883))
  * RHSM settings are now located in the OperatingSystemSpec

# [v1.0.0-alpha.0](https://github.com/kubermatic/kubeone/releases/tag/v1.0.0-alpha.0) - 2020-04-22

## Added

* Add support for OpenStack external cloud controller manager (CCM) ([#820](https://github.com/kubermatic/kubeone/pull/820))
* Add `Untaint` API field to remove default taints from the control plane nodes ([#823](https://github.com/kubermatic/kubeone/pull/823))
* Add the `PodPresets` feature ([#837](https://github.com/kubermatic/kubeone/pull/837))
* Add ability to provision static worker nodes ([#834](https://github.com/kubermatic/kubeone/pull/834))
* Add ability to use an external CNI plugin ([#862](https://github.com/kubermatic/kubeone/pull/862))
* Add ability to skip cluster provisioning when running the `install` command using the `--no-init` flag ([#871](https://github.com/kubermatic/kubeone/pull/871))

## Changed

### General

* machine-controller and machine-controller-webhook are bound to the control plane nodes ([#832](https://github.com/kubermatic/kubeone/pull/832))

### Bug Fixes

* Apply only addons with `.yaml`, `.yml` and `.json` extensions ([#873](https://github.com/kubermatic/kubeone/pull/873))

### Updated

* Update machine-controller to v1.11.2 ([#861](https://github.com/kubermatic/kubeone/pull/861))
* Update NodeLocalDNSCache to v1.15.12 ([#872](https://github.com/kubermatic/kubeone/pull/872))

# [v0.11.2](https://github.com/kubermatic/kubeone/releases/tag/v0.11.2) - 2020-05-21

## Attention Needed

* This version fixes the provisioning failures caused by `docker-ce-cli` version mismatch. The older releases are not working anymore ([#907](https://github.com/kubermatic/kubeone/pull/907))
* `machine-controller` must be updated to v1.11.3 on existing clusters or otherwise newly created worker nodes will not work properly. The `machine-controller` can be updated on one of the following ways:
  * (Recommended) Run `kubeone upgrade -f` with the exact same config that's currently running 
  * Run `kubeone install` with the exact same config that's currently running
  * Update the `machine-controller` and `machine-controller-webhook` deployments manually


## Changed

### General

* Bind `docker-ce-cli` to the same version as `docker-ce` ([#907](https://github.com/kubermatic/kubeone/pull/907))
* Fix CoreOS install and upgrade scripts ([#907](https://github.com/kubermatic/kubeone/pull/907))

### Updated

* Update machine-controller to v1.11.3 ([#907](https://github.com/kubermatic/kubeone/pull/907))

# [v0.11.1](https://github.com/kubermatic/kubeone/releases/tag/v0.11.1) - 2020-04-08

## Added

* Add support for Kubernetes 1.18 ([#841](https://github.com/kubermatic/kubeone/pull/841))

## Changed

### Bug Fixes

* Ensure machine-controller CRDs are Established before deploying MachineDeployments ([#824](https://github.com/kubermatic/kubeone/pull/824))
* Fix `leader_ip` parsing from Terraform ([#819](https://github.com/kubermatic/kubeone/pull/819))

### Updated

* Update machine-controller to v1.11.1 ([#808](https://github.com/kubermatic/kubeone/pull/808))
  * NodeCSRApprover controller is enabled by default to automatically approve CSRs for kubelet serving certificates
  * machine-controller types are updated to include recently added fields
  * Terraform example scripts are updated with the new fields

# [v0.11.0](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0) - 2020-03-05

**Changelog since v0.10.0. For changelog since v0.11.0-beta.3, please check the [release notes](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0)**

## Attention Needed

* Kubernetes 1.14 clusters are not supported as of this release because 1.14 isn't supported by the upstream anymore
  * It remains possible and is advisable to upgrade 1.14 clusters to 1.15
  * Currently, it also remains possible to provision 1.14 clusters, but that can be dropped at any time and it'll not be fixed if it stops working
* As of this release, it is not possible to upgrade 1.13 clusters to 1.14
  * Please use an older version of KubeOne in the case you need to upgrade 1.13 clusters
* The AWS Terraform configuration has been refactored a in backward-incompatible way ([#729](https://github.com/kubermatic/kubeone/issues/729))
  * Terraform now handles setting up subnets
  * All resources are tagged ensuring all cluster features offered by AWS CCM are supported
  * The security of the setup has been increased
  * Access to nodes and the Kubernetes API is now going over a bastion host
  * The `aws-private` configuration has been removed
  * Check out the [new Terraform configuration](https://github.com/kubermatic/kubeone/tree/v0.11.0-beta.0/examples/terraform/aws) for more details

## Added

* Add support for Kubernetes 1.17
  * Fix cluster upgrade failures when upgrading from 1.16 to 1.17 ([#764](https://github.com/kubermatic/kubeone/pull/764))
* Add support for ARM64 clusters ([#783](https://github.com/kubermatic/kubeone/pull/783))
* Add ability to deploy Kubernetes manifests on the provisioning time (KubeOne Addons) ([#782](https://github.com/kubermatic/kubeone/pull/782))
* Add the `kubeone status` command which checks the health of the cluster, API server and `etcd` ([#734](https://github.com/kubermatic/kubeone/issues/734))
* Add support for NodeLocalDNSCache ([#704](https://github.com/kubermatic/kubeone/issues/704))
* Add ability to divert access to the Kubernetes API over SSH tunnel ([#714](https://github.com/kubermatic/kubeone/issues/714))
* Add support for sourcing proxy settings from Terraform output ([#698](https://github.com/kubermatic/kubeone/issues/698))
* Persist configured proxy in the system package managers ([#749](https://github.com/kubermatic/kubeone/pull/749))

## Changed

### General

* [Breaking] The AWS Terraform configuration has been refactored ([#729](https://github.com/kubermatic/kubeone/issues/729))
* The KubeOneCluster manifests is now parsed strictly ([#802](https://github.com/kubermatic/kubeone/pull/802))
* The leader instance can be defined declarative using API ([#790](https://github.com/kubermatic/kubeone/pull/790))
* Make vSphere Cloud Controller Manager read credentials from a Secret instead from `cloud-config` ([#724](https://github.com/kubermatic/kubeone/issues/724))

### Bug fixes

* Fix CentOS cluster provisioning ([#770](https://github.com/kubermatic/kubeone/pull/770))
* Fix AWS shared credentials file handling ([#806](https://github.com/kubermatic/kubeone/pull/806))
* Fix credentials handling if `.cloudProvider.Name` is `none` ([#696](https://github.com/kubermatic/kubeone/issues/696))
* Fix upgrades not determining hostname correctly causing upgrades to fail ([#708](https://github.com/kubermatic/kubeone/issues/708))
* Fix `kubeone reset` failing to reset the cluster ([#727](https://github.com/kubermatic/kubeone/issues/727))
* Fix `configure-cloud-routes` bug for AWS causing `kube-controller-manager` to log warnings ([#725](https://github.com/kubermatic/kubeone/issues/725))
* Disable validation of replicas count in the workers definition ([#775](https://github.com/kubermatic/kubeone/pull/775))
* Proxy settings defined in the config have precedence over those defined in Terraform ([#760](https://github.com/kubermatic/kubeone/pull/760))

### Updates

* Update machine-controller to v1.9.0 ([#774](https://github.com/kubermatic/kubeone/pull/774))
* Update Canal CNI to v3.10 ([#718](https://github.com/kubermatic/kubeone/issues/718))
* Update metrics-server to v0.3.6 ([#720](https://github.com/kubermatic/kubeone/issues/720))
* Update DigitalOcean Cloud Controller Manager to v0.1.21 ([#722](https://github.com/kubermatic/kubeone/issues/722))
* Update Hetzner Cloud Controller Manager to v1.5.0 ([#726](https://github.com/kubermatic/kubeone/issues/726))

### Removed

* Remove ability to upgrade 1.13 clusters to 1.14 ([#764](https://github.com/kubermatic/kubeone/pull/764))
* Removed FlexVolume support from the Canal CNI ([#756](https://github.com/kubermatic/kubeone/pull/756))

### Docs

* GCE clusters must be configured as Regional to work properly ([#732](https://github.com/kubermatic/kubeone/issues/732))


# [v0.11.0-beta.3](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0-beta.3) - 2019-12-20

## Attention Needed

* Kubernetes 1.14 clusters are not supported as of this release because 1.14 isn't supported by the upstream anymore
  * It remains possible and is advisable to upgrade 1.14 clusters to 1.15
  * Currently, it also remains possible to provision 1.14 clusters, but that can be dropped at any time and it'll not be fixed if it stops working
* As of this release, it is not possible to upgrade 1.13 clusters to 1.14
  * Please use an older version of KubeOne in the case you need to upgrade 1.13 clusters

## Added

* Add support for Kubernetes 1.17
  * Fix cluster upgrade failures when upgrading from 1.16 to 1.17 ([#764](https://github.com/kubermatic/kubeone/pull/764))

## Removed

* Remove ability to upgrade 1.13 clusters to 1.14 ([#764](https://github.com/kubermatic/kubeone/pull/764))

# [v0.11.0-beta.2](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0-beta.2) - 2019-12-12

## Changed

### Bug Fixes

* Proxy settings defined in the config have precedence over those defined in Terraform ([#760](https://github.com/kubermatic/kubeone/pull/760))

# [v0.11.0-beta.1](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0-beta.1) - 2019-12-10

## Added

* Persist configured proxy in the system package managers ([#749](https://github.com/kubermatic/kubeone/pull/749))

## Changed

### Bug fixes

* Fix `kubeone status` reporting wrong `etcd` status when the hostname is not provided by Terraform or config ([#753](https://github.com/kubermatic/kubeone/pull/753))

## Removed

* Removed FlexVolume support from the Canal CNI ([#756](https://github.com/kubermatic/kubeone/pull/756))

# [v0.11.0-beta.0](https://github.com/kubermatic/kubeone/releases/tag/v0.11.0-beta.0) - 2019-11-28

## Attention Needed

* The AWS Terraform configuration has been refactored a in backward-incompatible way ([#729](https://github.com/kubermatic/kubeone/issues/729))
  * Terraform now handles setting up subnets
  * All resources are tagged ensuring all cluster features offered by AWS CCM are supported
  * The security of the setup has been increased
  * Access to nodes and the Kubernetes API is now going over a bastion host
  * The `aws-private` configuration has been removed
  * Check out the [new Terraform configuration](https://github.com/kubermatic/kubeone/tree/v0.11.0-beta.0/examples/terraform/aws) for more details

## Added

* Add the `kubeone status` command which checks the health of the cluster, API server and `etcd` ([#734](https://github.com/kubermatic/kubeone/issues/734))
* Add support for NodeLocalDNSCache ([#704](https://github.com/kubermatic/kubeone/issues/704))
* Add ability to divert access to the Kubernetes API over SSH tunnel ([#714](https://github.com/kubermatic/kubeone/issues/714))
* Add support for sourcing proxy settings from Terraform output ([#698](https://github.com/kubermatic/kubeone/issues/698))

## Changed

### General

* [Breaking] The AWS Terraform configuration has been refactored ([#729](https://github.com/kubermatic/kubeone/issues/729))
* Make vSphere Cloud Controller Manager read credentials from a Secret instead from `cloud-config` ([#724](https://github.com/kubermatic/kubeone/issues/724))

### Bug fixes

* Fix credentials handling if `.cloudProvider.Name` is `none` ([#696](https://github.com/kubermatic/kubeone/issues/696))
* Fix upgrades not determining hostname correctly causing upgrades to fail ([#708](https://github.com/kubermatic/kubeone/issues/708))
* Fix `kubeone reset` failing to reset the cluster ([#727](https://github.com/kubermatic/kubeone/issues/727))
* Fix `configure-cloud-routes` bug for AWS causing `kube-controller-manager` to log warnings ([#725](https://github.com/kubermatic/kubeone/issues/725))

### Updates

* Update machine-controller to v1.8.0 ([#736](https://github.com/kubermatic/kubeone/issues/736))
* Update Canal CNI to v3.10 ([#718](https://github.com/kubermatic/kubeone/issues/718))
* Update metrics-server to v0.3.6 ([#720](https://github.com/kubermatic/kubeone/issues/720))
* Update DigitalOcean Cloud Controller Manager to v0.1.21 ([#722](https://github.com/kubermatic/kubeone/issues/722))
* Update Hetzner Cloud Controller Manager to v1.5.0 ([#726](https://github.com/kubermatic/kubeone/issues/726))

### Docs

* GCE clusters must be configured as Regional to work properly ([#732](https://github.com/kubermatic/kubeone/issues/732))

# [v0.10.0](https://github.com/kubermatic/kubeone/releases/tag/v0.10.0) - 2019-10-09

## Attention Needed

* Kubernetes 1.13 clusters are not supported as of this release because 1.13 isn't supported by the upstream anymore
  * It remains possible to upgrade 1.13 clusters to 1.14 and is strongly advised
  * Currently, it also remains possible to provision 1.13 clusters, but that can be dropped at any time and it'll not be fixed if it stops working 
* KubeOne now uses Go modules! :tada: ([#550](https://github.com/kubermatic/kubeone/pull/550))
  * This should not introduce any breaking changes
  * If you're using `go get` to obtain KubeOne, you have to enable support for Go modules by setting the `GO111MODULE` environment variable to `on`
  * You can obtain KubeOne v0.10.0 using the following `go get` command: `go get github.com/kubermatic/kubeone@v0.10.0`

## Known Issues

* `kubectl scale` is not working with `kubectl` v1.15 to due an [upstream issue](https://github.com/kubernetes/kubernetes/issues/80515). Please upgrade to `kubectl` v1.16 if you want to use the `kubectl scale` command.

## Added

* Add support for Kubernetes 1.16
    * Fix cluster upgrade failures when upgrading from 1.15 to 1.16 ([#670](https://github.com/kubermatic/kubeone/pull/670))
    * Update default admission controllers for 1.16 ([#667](https://github.com/kubermatic/kubeone/pull/667))
* Add support for sourcing credentials and `cloud-config` file from a YAML file ([#623](https://github.com/kubermatic/kubeone/pull/623), [#641](https://github.com/kubermatic/kubeone/pull/641))
* Add the `StaticAuditLog` feature used to configure the [audit log backend](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#log-backend) ([#631](https://github.com/kubermatic/kubeone/pull/631))
* Add `SystemPackages` API field used to control configuration of repositories and packages ([#670](https://github.com/kubermatic/kubeone/pull/670))
* Add support for managing clusters over a bastion host ([#567](https://github.com/kubermatic/kubeone/pull/567), [#651](https://github.com/kubermatic/kubeone/pull/651))
* Add support for specifying OpenStack Tenant ID using the `OS_TENANT_ID` environment variable ([#551](https://github.com/kubermatic/kubeone/pull/551))
* Add ability to configure static networking for worker nodes ([#606](https://github.com/kubermatic/kubeone/pull/606))
* Add taints on control plane nodes by default ([#564](https://github.com/kubermatic/kubeone/pull/564))
* Add ability to apply labels on the Node object using the `.workers.providerSpec.Labels` field ([#677](https://github.com/kubermatic/kubeone/pull/677))
* Add ability to apply taints on the worker nodes using the `.workers.providerSpec.Taints` field ([#678](https://github.com/kubermatic/kubeone/pull/678))
* Add an optional `rootDiskSizeGB` field to the worker spec for OpenStack ([#549](https://github.com/kubermatic/kubeone/pull/549))
* Add an optional `nodeVolumeAttachLimit` field to the worker spec for OpenStack ([#572](https://github.com/kubermatic/kubeone/pull/572))
* Add an optional `TrustDevicePath` field to the worker spec for OpenStack ([#686](https://github.com/kubermatic/kubeone/pull/686))
* Add optional `BillingCycle` and `Tags` fields to the worker spec for Packet ([#686](https://github.com/kubermatic/kubeone/pull/686))
* Add ability to use AWS spot instances for worker nodes using the `isSpotInstance` field ([#686](https://github.com/kubermatic/kubeone/pull/686))
* Add support for Hetzner Private Networking ([#596](https://github.com/kubermatic/kubeone/pull/596))
* Add `ShortNames` and `AdditionalPrinterColumns` for Cluster-API CRDs ([#689](https://github.com/kubermatic/kubeone/pull/689))
* Add an example KubeOne Ansible playbook ([#576](https://github.com/kubermatic/kubeone/pull/576))

## Changed

### Bug Fixes

* Fix `kubeone install` and `kubeone upgrade` generating `v1beta1` instead of `v1beta2` `kubeadm` configuration file for 1.15 and 1.16 clusters ([#670](https://github.com/kubermatic/kubeone/pull/670))
* Fix cluster provisioning failures when the `DynamicAuditLog` feature is enabled ([#630](https://github.com/kubermatic/kubeone/pull/630))
* Fix `kubeone reset` not retrying deleting worker nodes on error ([#639](https://github.com/kubermatic/kubeone/pull/639))
* Fix `kubeone reset` not skipping deleting worker nodes if the `machine-controller` CRDs are not deployed ([#683](https://github.com/kubermatic/kubeone/pull/683))
* Fix Terraform integration not respecting multiple workerset definitions from `output.tf` ([#568](https://github.com/kubermatic/kubeone/pull/568))
* Fix `kubeone install` failing if Terraform output is not provided ([#574](https://github.com/kubermatic/kubeone/pull/574))
* Flannel CNI is forced use an internal network if it's available ([#598](https://github.com/kubermatic/kubeone/pull/598))

### Updates

* Update `machine-controller` to v1.5.7 ([#682](https://github.com/kubermatic/kubeone/pull/682))
* Update [DigitalOcean Cloud Controller Manager (CCM)](https://github.com/digitalocean/digitalocean-cloud-controller-manager) to v0.1.16 ([#591](https://github.com/kubermatic/kubeone/pull/591))

### Proxy

* Write proxy configuration to the `/etc/environment` file ([#687](https://github.com/kubermatic/kubeone/pull/687), [#688](https://github.com/kubermatic/kubeone/pull/688))
* Fix proxy configuration file (`/etc/kubeone/proxy-env`) generation ([#650](https://github.com/kubermatic/kubeone/pull/650))
* Fix `machine-controller` and `machine-controller-webhook` deployments not receiving the proxy configuration ([#657](https://github.com/kubermatic/kubeone/pull/657))

### Examples

* Fix GCE provisioning failure when using a longer cluster name with the example Terraform script ([#607](https://github.com/kubermatic/kubeone/pull/607))

## Removed

* Remove `TemplateNetName` field from the vSphere workers spec ([#624](https://github.com/kubermatic/kubeone/pull/624))
* Remove the old KubeOne configuration API ([#626](https://github.com/kubermatic/kubeone/pull/626))
    * This should not affect you unless you're using the pre-v0.6.0 configuration API manually
    * The `kubeone config migrate` command is still available, but might be deleted at any time

# [v0.10.0-alpha.3](https://github.com/kubermatic/kubeone/releases/tag/v0.10.0-alpha.3) - 2019-08-22

## Changed

* Update `machine-controller` to v1.5.3 ([#633](https://github.com/kubermatic/kubeone/pull/633))

# [v0.10.0-alpha.2](https://github.com/kubermatic/kubeone/releases/tag/v0.10.0-alpha.2) - 2019-08-21

## Changed

* Fix cluster provisioning failures when DynamicAuditLog feature is enabled ([#630](https://github.com/kubermatic/kubeone/pull/630))
* Update `machine-controller` to v1.5.2 ([#624](https://github.com/kubermatic/kubeone/pull/624))

## Removed

* Remove `TemplateNetName` field from the vSphere workers spec ([#624](https://github.com/kubermatic/kubeone/pull/624))
* Remove the old KubeOne configuration API ([#626](https://github.com/kubermatic/kubeone/pull/626))

# [v0.10.0-alpha.1](https://github.com/kubermatic/kubeone/releases/tag/v0.10.0-alpha.1) - 2019-08-16

## Added

* Add ability to configure static networking for worker nodes ([#606](https://github.com/kubermatic/kubeone/pull/606))

## Changed

* Flannel CNI is forced use an internal network if it's available ([#598](https://github.com/kubermatic/kubeone/pull/598))
* Update `machine-controller` to v1.5.1 ([#602](https://github.com/kubermatic/kubeone/pull/602))
* Update [DigitalOcean Cloud Controller Manager (CCM)](https://github.com/digitalocean/digitalocean-cloud-controller-manager) to v0.1.16 ([#591](https://github.com/kubermatic/kubeone/pull/591))

# [v0.10.0-alpha.0](https://github.com/kubermatic/kubeone/releases/tag/v0.10.0-alpha.0) - 2019-07-17

## Attention Needed

* KubeOne now uses Go modules! :tada: ([#550](https://github.com/kubermatic/kubeone/pull/550))
  * This should not introduce any breaking change
  * If you're using `go get` to obtain KubeOne, you may have to enable support for Go modules by setting the `GO111MODULE` environment variable to `on`

## Added

* Add support for SSH over a bastion host ([#567](https://github.com/kubermatic/kubeone/pull/567))
* Add an optional `rootDiskSizeGB` field to the worker spec for OpenStack ([#549](https://github.com/kubermatic/kubeone/pull/549))
* Add an optional `nodeVolumeAttachLimit` field to the worker spec for OpenStack ([#572](https://github.com/kubermatic/kubeone/pull/572))
* Add support for specifying OpenStack Tenant ID using the `OS_TENANT_ID` environment variable ([#551](https://github.com/kubermatic/kubeone/pull/551))
* Add an example KubeOne Ansible playbook ([#576](https://github.com/kubermatic/kubeone/pull/576))

## Changed

* Fix Terraform integration not respecting multiple workerset definitions from `output.tf` ([#568](https://github.com/kubermatic/kubeone/pull/568))
* Fix `install` failing if Terraform output is not provided ([#574](https://github.com/kubermatic/kubeone/pull/574))
* Update `machine-controller` to v1.4.2 ([#572](https://github.com/kubermatic/kubeone/pull/572))
* Control plane nodes are now tainted by default ([#564](https://github.com/kubermatic/kubeone/pull/564))

# [v0.9.2](https://github.com/kubermatic/kubeone/releases/tag/v0.9.2) - 2019-07-04

## Changed

* Fix the CNI plugin URL for cluster upgrades on CoreOS ([#554](https://github.com/kubermatic/kubeone/pull/554))
* Fix `kubelet` binary upgrade failure on CoreOS because of binary lock ([#556](https://github.com/kubermatic/kubeone/pull/556))

# [v0.9.1](https://github.com/kubermatic/kubeone/releases/tag/v0.9.1) - 2019-07-03

## Changed

* Fix `.ClusterNetwork.PodSubnet` not being respected when using the Weave-Net CNI plugin ([#540](https://github.com/kubermatic/kubeone/pull/540))
* Fix `kubeadm` preflight check failure (`IsDockerSystemdCheck`) on Ubuntu and CoreOS by making Docker use `systemd` cgroups driver ([#536](https://github.com/kubermatic/kubeone/pull/536), [#541](https://github.com/kubermatic/kubeone/pull/541))
* Fix `kubeadm` preflight check failure on CentOS due to `kubelet` service not being enabled ([#541](https://github.com/kubermatic/kubeone/pull/541))

# [v0.9.0](https://github.com/kubermatic/kubeone/releases/tag/v0.9.0) - 2019-07-02

## Action Required

* The Terraform integration now requires Terraform v0.12+
  * Please see the official [Upgrading to Terraform v0.12](https://www.terraform.io/upgrade-guides/0-12.html)
  document to find out how to update your Terraform scripts for v0.12
  * The example Terraform scripts coming with KubeOne are already updated for v0.12
  * KubeOne is not able to parse output of `terraform output` generated with Terraform
  v0.11 or older anymore
  * The Terraform output template (`output.tf`) has been changed and KubeOne is not able
  to parse the old template anymore. You can check the output template
  used by [example Terraform scripts](https://github.com/kubermatic/kubeone/blob/986b4bc361c6a0ae5d5319990c54e82c15a66cb3/examples/terraform/aws/output.tf) as a
  reference

## Added

* Add support for Kubernetes 1.15 ([#486](https://github.com/kubermatic/kubeone/pull/486))
* Add support for Microsoft Azure ([#469](https://github.com/kubermatic/kubeone/pull/469))
* Add `kubeone completion` command for generating the shell completion scripts for `bash` and `zsh` ([#484](https://github.com/kubermatic/kubeone/pull/484))
* Add `kubeone document` command for generating man pages and KubeOne documentation ([#484](https://github.com/kubermatic/kubeone/pull/484))
* Add support for reading Terraform output directly from the directory ([#495](https://github.com/kubermatic/kubeone/pull/495))
* Add missing fields to the workers specification API ([#499](https://github.com/kubermatic/kubeone/pull/499))

## Changed

* [BREAKING] KubeOne Terraform integration now uses Terraform v0.12+ ([#466](https://github.com/kubermatic/kubeone/pull/466))
* [BREAKING] Change Terraform output template to conform with the KubeOneCluster API ([#503](https://github.com/kubermatic/kubeone/pull/503))
* Fix Docker not starting properly for some providers on Ubuntu ([#512](https://github.com/kubermatic/kubeone/pull/512))
* Fix `kubeone reset` failing if a MachineSet or Machine object has been already deleted and include more details in the error messages ([#508](https://github.com/kubermatic/kubeone/pull/508))
* Fix ability to read Terraform output from the standard input (`stdin`) ([#479](https://github.com/kubermatic/kubeone/pull/479))
* Update `machine-controller` to v1.3.0 ([#499](https://github.com/kubermatic/kubeone/pull/499))
* Update DigitalOcean Cloud Controller Manager to v0.1.15 ([#516](https://github.com/kubermatic/kubeone/pull/516))
* Use Docker 18.09.7 when provisioning new clusters ([#517](https://github.com/kubermatic/kubeone/pull/517))
* Configure proxy for `kubelet` on control plane nodes if proxy settings are provided ([#496](https://github.com/kubermatic/kubeone/pull/496))
* Configure proxy on worker nodes if proxy settings are provided ([#490](https://github.com/kubermatic/kubeone/pull/490))
* Make GoBetween Load balancer configuration script work on all operating systems and fix minor bugs ([#494](https://github.com/kubermatic/kubeone/pull/494))

# [v0.8.0](https://github.com/kubermatic/kubeone/releases/tag/v0.8.0) - 2019-05-30

## Added

* Add support for VMware vSphere ([#428](https://github.com/kubermatic/kubeone/pull/428))

# [v0.7.0](https://github.com/kubermatic/kubeone/releases/tag/v0.7.0) - 2019-05-28

## Added

* Add WeaveNet as an additional CNI plugin ([#432](https://github.com/kubermatic/kubeone/pull/432))
* Add the `--remove-binaries` flag to the `kubeone reset` command that removes Kubernetes binaries ([#450](https://github.com/kubermatic/kubeone/pull/450))

## Changed

* Fix `kubeone reset` failing if no `MachineDeployment` or `Machine` object exist ([#450](https://github.com/kubermatic/kubeone/pull/450))
* Update `machine-controller` to v1.1.8 ([#454](https://github.com/kubermatic/kubeone/pull/454))

# [v0.6.2](https://github.com/kubermatic/kubeone/releases/tag/v0.6.2) - 2019-05-13

## Changed

* Fix a missing JSON tag on the `Name` field, so specifying `name` in lowercase in the manifest works as expected ([#439](https://github.com/kubermatic/kubeone/pull/439))
* Fix failing to disable SELinux on CentOS if it's already disabled ([#443](https://github.com/kubermatic/kubeone/pull/443))
* Fix setting permissions on the remote KubeOne directory when the `$USER` environment variable isn't set ([#443](https://github.com/kubermatic/kubeone/pull/443))

# [v0.6.1](https://github.com/kubermatic/kubeone/releases/tag/v0.6.1) - 2019-05-09

## Changed

* Provide the `--kubelet-preferred-address-types` flag to metrics-server, so it works on all providers ([#424](https://github.com/kubermatic/kubeone/pull/424))

# [v0.6.0](https://github.com/kubermatic/kubeone/releases/tag/v0.6.0) - 2019-05-08

We're excited to announce that as of this release KubeOne is in **beta**! We have the new
[backwards compatibility policy](https://github.com/kubermatic/kubeone/blob/v0.6.0/docs/api_migration.md)
going in effect as of this release.

Check out the [documentation](https://github.com/kubermatic/kubeone/tree/v0.6.0/docs) for this release to find out how to get started with KubeOne.

## Action Required

* This release introduces the new KubeOneCluster API. The new API is supposed to the improve user experience and bring
many new possibilities, like API versioning.
  * **Old KubeOne configuration manifests will not work as of this release!**
  * To continue using KubeOne, you need to migrate your existing manifests to the new KubeOneCluster API. Follow
  [the migration guidelines](https://github.com/kubermatic/kubeone/blob/v0.6.0/docs/api_migration.md) to find out
  how to migrate.

## Added

* [BREAKING] Implement and migrate to the KubeOneCluster API ([#343](https://github.com/kubermatic/kubeone/pull/343), [#353](https://github.com/kubermatic/kubeone/pull/353), [#360](https://github.com/kubermatic/kubeone/pull/360), [#379](https://github.com/kubermatic/kubeone/pull/379), [#390](https://github.com/kubermatic/kubeone/pull/390))
* Implement the `config migrate` command for migrating old configuration manifests to KubeOneCluster manifests ([#408](https://github.com/kubermatic/kubeone/pull/408))
* Implement the `config print` command for printing an example KubeOneCluster manifest ([#412](https://github.com/kubermatic/kubeone/pull/412), [#415](https://github.com/kubermatic/kubeone/pull/415))
* Enable scaling subresource for MachineDeployments ([#334](https://github.com/kubermatic/kubeone/pull/334))
* Deploy `metrics-server` by default ([#338](https://github.com/kubermatic/kubeone/pull/338), [#351](https://github.com/kubermatic/kubeone/pull/351))
* Deploy external cloud controller manager for DigitalOcean, Hetzner, and Packet ([#364](https://github.com/kubermatic/kubeone/pull/364))
* Implement Terraform integration for Hetzner ([#331](https://github.com/kubermatic/kubeone/pull/331))
* Add support for OpenIDConnect configuration ([#344](https://github.com/kubermatic/kubeone/pull/344))
* Add support for Packet provider ([#384](https://github.com/kubermatic/kubeone/pull/384))
* Patch CoreDNS deployment to work with external cloud controller manager taints ([#362](https://github.com/kubermatic/kubeone/pull/362))

## Changed

* Fix a typo in vSphere CloudProviderName ([#339](https://github.com/kubermatic/kubeone/pull/339))
* Expose `ssh_username` variable in the Terraform output ([#350](https://github.com/kubermatic/kubeone/pull/350))
* Pass `-external-cloud-provider` flag to `machine-controller` when external cloud provider is enabled ([#361](https://github.com/kubermatic/kubeone/pull/361))
* Update `machine-controller` to v1.1.5 ([#378](https://github.com/kubermatic/kubeone/pull/378))
* Don't wait for `machine-controller` if it's not deployed ([#392](https://github.com/kubermatic/kubeone/pull/392))
* Deploy instance with a load balancer on OpenStack. General improvements to OpenStack support ([#401](https://github.com/kubermatic/kubeone/pull/401))
* Parse all parameters from Terraform output for DigitalOcean ([#370](https://github.com/kubermatic/kubeone/pull/370))

# [v0.5.0](https://github.com/kubermatic/kubeone/releases/tag/v0.5.0) - 2019-04-03

## Added

* Add support for Kubernetes 1.14
* Add support for upgrading from Kubernetes 1.13 to 1.14
* Add support for Google Compute Engine ([#307](https://github.com/kubermatic/kubeone/pull/307), [#317](https://github.com/kubermatic/kubeone/pull/317))
* Update machine-controller when upgrading the cluster ([#304](https://github.com/kubermatic/kubeone/pull/304))
* Add timeout after upgrading each node to let nodes to settle down ([#316](https://github.com/kubermatic/kubeone/pull/316), [#319](https://github.com/kubermatic/kubeone/pull/319))

## Changed

* Deploy machine-controller v1.1.2 on the new clusters ([#317](https://github.com/kubermatic/kubeone/pull/317))
* Creating MachineDeployments and upgrading nodes tasks are repeated three times on the failure ([#328](https://github.com/kubermatic/kubeone/pull/328))
* Allow upgrading to the same Kubernetes version ([#315](https://github.com/kubermatic/kubeone/pull/315))
* Allow the custom VPC to be used with the example AWS Terraform scripts and switch to the T3 instances ([#306](https://github.com/kubermatic/kubeone/pull/306))

# [v0.4.0](https://github.com/kubermatic/kubeone/releases/tag/v0.4.0) - 2019-03-21

## Action Required

* In [#264](https://github.com/kubermatic/kubeone/pull/264) the environment variable names were changed to match names used by Terraform. In order for KubeOne to correctly fetch credentials you need to use the following variables:
  * `DIGITALOCEAN_TOKEN` instead of `DO_TOKEN`
  * `OS_USERNAME` instead of `OS_USER_NAME`
  * `HCLOUD_TOKEN` instead of `HZ_TOKEN`
* The `cloud-config` file is now required for OpenStack clusters. Validation will fail if the `cloud-config` is not provided. See the [OpenStack quickstart](https://github.com/kubermatic/kubeone/blob/v0.4.0/docs/quickstart-openstack.md) for details how the `cloud-config` file should look like.
* Ark integration is removed from KubeOne in [#265](https://github.com/kubermatic/kubeone/pull/265). You need to remove the `backups` section from the existing configuration files. If you wish to deploy Ark on new clusters you have to do it manually.

## Added

* Add support for enabling DynamicAuditing ([#261](https://github.com/kubermatic/kubeone/pull/261))

## Changed

* Deploy machine-controller v1.0.7 on the new clusters ([#247](https://github.com/kubermatic/kubeone/pull/247))
* Automatically download the Kubeconfing file after install ([#248](https://github.com/kubermatic/kubeone/pull/248))
* Default `--destory-workers` to `true` for the `kubeone reset` command ([#252](https://github.com/kubermatic/kubeone/pull/252))
* Improve OpenStack Terraform scripts ([#253](https://github.com/kubermatic/kubeone/pull/253))
* The `cloud-config` file for OpenStack is required ([#253](https://github.com/kubermatic/kubeone/pull/253))
* The environment variable names are changed to match names used by Terraform ([#264](https://github.com/kubermatic/kubeone/pull/264))

## Removed

* Option for deploying Ark is removed ([#265](https://github.com/kubermatic/kubeone/pull/265))

# [v0.3.0](https://github.com/kubermatic/kubeone/releases/tag/v0.3.0) - 2019-03-08

## Action Required

* If you're using `provider.Name` `external` to configure control plane nodes to work with external CCM you need to:
    * Set `provider.Name` to name of the cloud provider you're using or to `none` (see [`config.yaml.dist`](https://github.com/kubermatic/kubeone/blob/v0.3.0/config.yaml.dist) for supported values),
    * Set `provider.External` to `true`.
    * **Note: using external CCM is not supported and is currently not working as expected.**
* If you're using `provider.Name` `none`:
    * If you're deploying `machine-controller` you need to set `MachineController.Provider` to name of the cloud provider you're using for worker nodes,
    * Otherwise you need to set `MachineController.Deploy` to `false`, so validation passes.

## Added

* Add support for cluster upgrades ([#206](https://github.com/kubermatic/kubeone/pull/206), [#211](https://github.com/kubermatic/kubeone/pull/211), [#214](https://github.com/kubermatic/kubeone/pull/214))
* Add support for enabling PodSecurityPolicy ([#218](https://github.com/kubermatic/kubeone/pull/218))
* Add `kubeone version` command ([#221](https://github.com/kubermatic/kubeone/pull/221))
* Add initial support for OpenStack worker nodes ([#209](https://github.com/kubermatic/kubeone/pull/209))

## Changed

* Fix `none` provider bugs and ensure correct behavior ([#227](https://github.com/kubermatic/kubeone/pull/227))
* External Cloud Controller Manager is now enabled by setting `Provider.External` to `true` instead of setting `Provider.Name` to `external` ([#230](https://github.com/kubermatic/kubeone/pull/230), [#237](https://github.com/kubermatic/kubeone/pull/237))
* Set correct Internal address on nodes with 2+ network interfaces ([#240](https://github.com/kubermatic/kubeone/pull/240))
* Deploy `machine-controller` on uninitialized nodes ([#241](https://github.com/kubermatic/kubeone/pull/241))

# [v0.2.0](https://github.com/kubermatic/kubeone/releases/tag/v0.2.0) - 2019-02-15

## Added

* Add support for workers creation on DigitalOcean ([#153](https://github.com/kubermatic/kubeone/pull/153))
* Add support for CentOS ([#154](https://github.com/kubermatic/kubeone/pull/154))
* Add support for external cloud controller managers ([#161](https://github.com/kubermatic/kubeone/pull/161))
* Add Terraform scripts for OpenStack ([#170](https://github.com/kubermatic/kubeone/pull/170))
* Add support for configuring proxy for Docker daemon, `apt-get`/`yum` and `curl` ([#182](https://github.com/kubermatic/kubeone/pull/182))

## Changed

* AWS credentials can be obtained from the profile file ([#156](https://github.com/kubermatic/kubeone/pull/156))
* Fix Etcd init on clusters with more than one network interface ([#160](https://github.com/kubermatic/kubeone/pull/160))
* Allow `provider.Name` to be specified for providers without external Cloud Controller Manager ([#161](https://github.com/kubermatic/kubeone/pull/161))
* Refactor KubeOne CLI ([#177](https://github.com/kubermatic/kubeone/pull/177))
    * Global flags are parsed correctly regardless of their placement in the command
    * `--verbose` and `--tfjson` are global flags
    * The main package is moved to the root of project (`github.com/kubermatic/kubeone`)
* Deploy `machine-controller` v1.0.4 instead of v0.10.0 to fix CVE-2019-5736 ([#191](https://github.com/kubermatic/kubeone/pull/191))
* Deploy Docker 18.09.2 on Ubuntu to fix CVE-2019-5736 ([#193](https://github.com/kubermatic/kubeone/pull/193))

## Removed

* Remove support for Kubernetes 1.12 ([#184](https://github.com/kubermatic/kubeone/pull/184))
* Remove API fields related to Etcd ([#194](https://github.com/kubermatic/kubeone/pull/194))

