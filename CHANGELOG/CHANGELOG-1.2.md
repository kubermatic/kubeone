# [v1.2.3](https://github.com/kubermatic/kubeone/releases/tag/v1.2.3) - 2021-06-14

## Changed

### Bug Fixes

* Pass the `-node-external-cloud-provider` flag to the machine-controller-webhook. This fixes the issue with the worker nodes not using the external CCM on the clusters with the external CCM enabled ([#1380](https://github.com/kubermatic/kubeone/pull/1380))
* Disable `repo_gpgcheck` for the Kubernetes yum repository. This fixes the cluster provisioning and upgrading failures for CentOS/RHEL caused by yum failing to install Kubernetes packages ([#1304](https://github.com/kubermatic/kubeone/pull/1304))


# [v1.2.2](https://github.com/kubermatic/kubeone/releases/tag/v1.2.2) - 2021-06-11

## Changed

### Bug Fixes

* Fix AWS config for terraform 0.15 ([#1372](https://github.com/kubermatic/kubeone/pull/1372))
  * AWS terraform config now works under terraform 0.15+ (including 1.0)
* Update machinecontroller to v1.30.0 ([#1370](https://github.com/kubermatic/kubeone/pull/1370))
  * machinecontroller to v1.30.0 relaxes docker / containerd version constraints
* Relax docker/containerd version constraints ([#1371](https://github.com/kubermatic/kubeone/pull/1371))

# [v1.2.1](https://github.com/kubermatic/kubeone/releases/tag/v1.2.1) - 2021-03-23

**Check out the changelog for the [v1.2.1 release](https://github.com/kubermatic/kubeone/releases/tag/v1.2.1) for more information about what changes were introduced in the 1.2 release.**

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
