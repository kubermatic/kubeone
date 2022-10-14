# [v1.3.5](https://github.com/kubermatic/kubeone/releases/tag/v1.3.5) - 2022-04-26

## Attention Needed

This patch releases updates etcd to v3.5.3 which includes a fix for the data inconsistency issues reported earlier (https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ). To upgrade etcd for an existing cluster, you need to [force upgrade the cluster as described here](https://docs.kubermatic.com/kubeone/main/guides/etcd-corruption/#enabling-etcd-corruption-checks). If you're running Kubernetes 1.22 or newer, we strongly recommend upgrading etcd **as soon as possible**.

## Updated

- Upgrade machine-controller to v1.37.3 ([#1984](https://github.com/kubermatic/kubeone/pull/1984))
  - This fixes an issue where the machine-controller would not wait for the volumeAttachments deletion before deleting the node.
- Deploy etcd v3.5.3 for clusters running Kubernetes 1.22 or newer. etcd v3.5.3 includes a fix for [the data inconsistency issues announced by the etcd maintainers](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ. To upgrade etcd) for an existing cluster, you need to [force upgrade the cluster as described here](https://docs.kubermatic.com/kubeone/v1.4/guides/etcd_corruption/#enabling-etcd-corruption-checks) ([#1953](https://github.com/kubermatic/kubeone/pull/1953))

# [v1.3.4](https://github.com/kubermatic/kubeone/releases/tag/v1.3.4) - 2022-04-05

## Attention Needed

This patch release enables the etcd corruption checks on every etcd member that is running etcd 3.5 (which applies to all Kubernetes 1.22+ clusters). This change is a [recommendation from the etcd maintainers](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ) due to issues in etcd 3.5 that can cause data consistency issues. The changes in this patch release will prevent corrupted etcd members from joining or staying in the etcd ring.

## Changed

* Enable the etcd integrity checks (on startup and every 4 hours) for Kubernetes 1.22+ clusters. See [the official etcd announcement for more details](https://groups.google.com/a/kubernetes.io/g/dev/c/B7gJs88XtQc/m/rSgNOzV2BwAJ). ([#1928](https://github.com/kubermatic/kubeone/pull/1928))
* Validate Kubernetes version against supported versions constraints. The minimum supported version is 1.19, and the maximum supported version is 1.22 ([#1817](https://github.com/kubermatic/kubeone/pull/1817))
* Fix AMI filter in Terraform configs for AWS to always use `x86_64` images ([#1692](https://github.com/kubermatic/kubeone/pull/1692))

# [v1.3.3](https://github.com/kubermatic/kubeone/releases/tag/v1.3.3) - 2021-12-16

## Changed

### Fixed

* Allow pods with the seccomp profile defined to get scheduled if the PodSecurityPolicy (PSP) feature is enabled ([#1687](https://github.com/kubermatic/kubeone/pull/1687))
* Fix the image loader script to support KubeOne 1.3+ and Kubernetes 1.22+ ([#1672](https://github.com/kubermatic/kubeone/pull/1672))
* The `kubeone config images` command now shows images for the latest Kubernetes version (instead of for the oldest) ([#1672](https://github.com/kubermatic/kubeone/pull/1672))
* Add a new `--kubernetes-version` flag to the `kubeone config images` command ([#1672](https://github.com/kubermatic/kubeone/pull/1672))
  * This flag is used to filter images for a particular Kubernetes version. The flag cannot be used along with the KubeOneCluster manifest (`--manifest` flag)

### Addons

* Deploy default StorageClass for GCP clusters if the `default-storage-class` addon is enabled ([#1639](https://github.com/kubermatic/kubeone/pull/1639))

### Updated

* Update machine-controller to v1.37.2 ([#1654](https://github.com/kubermatic/kubeone/pull/1654))
  * machine-controller is now using Ubuntu 20.04 instead of 18.04 by default for all newly-created Machines on AWS, Azure, DO, GCE, Hetzner, Openstack, and Equinix Metal
  * This release defaults the provisioning utility for Flatcar machines on AWS to cloud-init (previously ignition). Ignition is currently not working on AWS because of the user data limit
  * If you have the provisioning utility explicitly set to Ignition, you'll not be able to provision new Flatcar machines on AWS. In that case, manually changing the provisioning utility to cloud-init is required

# [v1.3.2](https://github.com/kubermatic/kubeone/releases/tag/v1.3.2) - 2021-11-18

## Changed

### General

* Create MachineDeployments only for newly-provisioned clusters ([#1628](https://github.com/kubermatic/kubeone/pull/1628))
* Show warning about LBs on CCM migration for OpenStack clusters ([#1628](https://github.com/kubermatic/kubeone/pull/1628))

### Fixed

* Force drain nodes to remove standalone pods ([#1628](https://github.com/kubermatic/kubeone/pull/1628))
* Check for minor version when choosing kubeadm API version ([#1628](https://github.com/kubermatic/kubeone/pull/1628))
* Provide --cluster-name flag to the OpenStack external CCM (read PR description for more details) ([#1632](https://github.com/kubermatic/kubeone/pull/1623))
* Enable ip_tables related kernel modules and disable nm-cloud-setup tool on AWS for RHEL machines ([#1616](https://github.com/kubermatic/kubeone/pull/1616))
* Properly pass machine-controllers args ([#1596](https://github.com/kubermatic/kubeone/pull/1596))
  * This fixes the issue causing machine-controller and machine-controller-webhook deployments to run with incorrect flags
  * If you created your cluster with KubeOne 1.2 or older, and already upgraded to KubeOne 1.3, we recommend running kubeone apply again to properly reconcile machine-controller deployments
* Edit SELinux config file only if file exists ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* Fix `yum versionlock delete containerd.io` error ([#1602](https://github.com/kubermatic/kubeone/pull/1602))
* Ensure containerd/docker be upgraded automatically when running kubeone apply ([#1590](https://github.com/kubermatic/kubeone/pull/1590))

### Addons

* Add new "required" addons template function ([#1624](https://github.com/kubermatic/kubeone/pull/1624))
* Replace critical-pod annotation with priorityClassName ([#1628](https://github.com/kubermatic/kubeone/pull/1628))
* Update Hetzner Cloud Controller Manager to v1.12.0 ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* Default image in the cluster-autoscaler addon and allow the image to be overridden using addon parameters ([#1553](https://github.com/kubermatic/kubeone/pull/1553))
* Minor improvements to OpenStack CCM and CSI addons. OpenStack CSI controller can now be scheduled on control plane nodes ([#1536](https://github.com/kubermatic/kubeone/pull/1536))

### Terraform Configs

* OpenStack: Open NodePorts by default ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* GCE: Open NodePorts by default ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* Azure: Open NodePorts by default ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* Azure: Default VM type is changed to Standard_F2 ([#1592](https://github.com/kubermatic/kubeone/pull/1592))
* Add additional Availability Set used for worker nodes to Terraform configs for Azure ([#1562](https://github.com/kubermatic/kubeone/pull/1562))
  * Make sure to check the [production recommendations for Azure clusters](https://docs.kubermatic.com/kubeone/v1.3/cheat_sheets/production_recommendations/#azure) for more information about how this additional availability set is used
* Fix keepalived script in Terraform configs for vSphere to assume yes when updating repos ([#1538](https://github.com/kubermatic/kubeone/pull/1538))

## Removed

* Remove Ansible examples ([#1634](https://github.com/kubermatic/kubeone/pull/1634))

# [v1.3.1](https://github.com/kubermatic/kubeone/releases/tag/v1.3.1) - unreleased

**The v1.3.1 release has never been released due to an issue with the release process. Please check the [v1.3.2 release](https://github.com/kubermatic/kubeone/releases/tag/v1.3.2) instead.**

# [v1.3.0](https://github.com/kubermatic/kubeone/releases/tag/v1.3.0) - 2021-09-15

## Attention Needed

Check out the [Upgrading from 1.2 to 1.3 tutorial](https://docs.kubermatic.com/kubeone/v1.3/tutorials/upgrading/upgrading_from_1.2_to_1.3/) for more details about the breaking changes and how to mitigate them.

### Breaking changes / Action Required

* Increase the minimum Kubernetes version to v1.19.0. If you have Kubernetes clusters running v1.18 or older, you need to use an older KubeOne release to upgrade them to v1.19, and then upgrade to KubeOne 1.3.
* Increase the minimum Terraform version to 1.0.0.
* Remove support for Debian and RHEL 7 clusters. If you have Debian clusters, we recommend migrating to another operating system, for example Ubuntu. If you have RHEL 7 clusters, you should consider migrating to RHEL 8 which is supported.
* Automatically deploy CSI plugins for Hetzner, OpenStack, and vSphere clusters using external cloud provider. If you already have the CSI plugin deployed, you need to make sure that your CSI plugin deployment is compatible with the KubeOne CSI plugin addon.
* The `kubeone reset` command requires an explicit confirmation like the `apply` command starting with this release. The command can be automatically approved by using the `--auto-approve` flag.

### Deprecations

* KubeOne Addons can now be organized into subdirectories. It currently remains possible to put addons in the root of the addons directory, however, this is option is considered as deprecated as of this release. We highly recommend all users to reorganize their addons into subdirectories, where each subdirectory is for YAML manifests related to one addon.
* We're deprecating support for CentOS 8 because it's reaching [End-of-Life (EOL) on December 31, 2021](https://www.centos.org/centos-linux-eol/). CentOS 7 remains supported by KubeOne for now.

## Known Issues

* It's currently **not** possible to provision or upgrade to Kubernetes 1.22 for clusters running on vSphere. This is because vSphere CCM and CSI don't support Kubernetes 1.22. We'll introduce Kubernetes 1.22 support for vSphere as soon as new CCM and CSI releases with support for Kubernetes 1.22 are out.
* Newly-provisioned Kubernetes 1.22 clusters or clusters upgraded from Kubernetes 1.21 to 1.22 using KubeOne 1.3.0-alpha.1  use a metrics-server version incompatible with Kubernetes 1.22. This might cause issues with deleting Namespaces that manifests by the Namespace being stuck in the Terminating state. This can be fixed by upgrading KubeOne to v1.3.0-rc.0 or newer and running `kubeone apply`.
* The new Addons API requires the addons directory path (`.addons.path`) to be provided and the directory must exist (it can be empty), even if only embedded addons are used. If the path is not provided, it'll default to `./addons`.

## Added

### API

* Implement the Addons API used to manage addons deployed by KubeOne. The new Addons API can be used to deploy the addons embedded in the KubeOne binary. Currently available addons are: `backups-restic`, `cluster-autoscaler`, `default-storage-class`, and `unattended-upgrades` ([#1462](https://github.com/kubermatic/kubeone/pull/1462), [#1486](https://github.com/kubermatic/kubeone/pull/1486))
  * More information about the new API can be found in the [Addons documentation](https://docs.kubermatic.com/kubeone/v1.3/guides/addons/) or by running `kubeone config print --full`.
* Add support for specifying a custom Root CA bundle ([#1316](https://github.com/kubermatic/kubeone/pull/1316))
* Add new kube-proxy configuration API ([#1420](https://github.com/kubermatic/kubeone/pull/1420))
  * This API allows users to switch kube-proxy to IPVS mode, and configure IPVS properties such as strict ARP and scheduler
  * The default kube-proxy mode remains iptables

### Features

* Docker to containerd automated migration ([#1362](https://github.com/kubermatic/kubeone/pull/1362))
  * Check out the [Migrating to containerd document](https://docs.kubermatic.com/kubeone/v1.3/guides/containerd_migration/) for more details about this features, including how to use it.
* Add containerd support for Flatcar clusters ([#1340](https://github.com/kubermatic/kubeone/pull/1340))
* Add support for Kubernetes 1.22 ([#1447](https://github.com/kubermatic/kubeone/pull/1447), [#1456](https://github.com/kubermatic/kubeone/pull/1456))
* Add Cinder CSI plugin. The plugin is deployed by default for OpenStack clusters using the external cloud provider ([#1465](https://github.com/kubermatic/kubeone/pull/1465))
  * Check out the Attention Needed section of the changelog for more information.
* Add vSphere CSI plugin. The CSI plugin is deployed automatically if `.cloudProvider.csiConfig` is provided and `.cloudProvider.external` is enabled ([#1484](https://github.com/kubermatic/kubeone/pull/1484))
  * More information about the CSI plugin configuration can be found in the [vSphere CSI docs](https://vsphere-csi-driver.sigs.k8s.io/driver-deployment/installation.html#create_csi_vsphereconf)
  * Check out the Attention Needed section of the changelog for more information.
* Add Hetzner CSI plugin ([#1418](https://github.com/kubermatic/kubeone/pull/1418))
  * Check out the Attention Needed section of the changelog for more information.
* Implement the CCM/CSI migration for OpenStack and vSphere ([#1468](https://github.com/kubermatic/kubeone/pull/1468), [#1469](https://github.com/kubermatic/kubeone/pull/1469), [#1472](https://github.com/kubermatic/kubeone/pull/1472), [#1482](https://github.com/kubermatic/kubeone/pull/1482), [#1487](https://github.com/kubermatic/kubeone/pull/1487), [#1494](https://github.com/kubermatic/kubeone/pull/1494))
  * Check out the [CCM/CSI migration document](https://docs.kubermatic.com/kubeone/v1.3/guides/ccm_csi_migration/) for more details about this features, including how to use it.
* Add support for Encryption Providers ([#1241](https://github.com/kubermatic/kubeone/pull/1241), [#1320](https://github.com/kubermatic/kubeone/pull/1320))
  * Check out the [Enabling Kubernetes Encryption Providers document](https://docs.kubermatic.com/kubeone/v1.3/guides/encryption_providers/) for more details about this features, including how to use it.
* Add a new `kubeone config images list` subcommand to list images used by KubeOne and kubeadm ([#1334](https://github.com/kubermatic/kubeone/pull/1334))
* Automatically renew Kubernetes certificates when running `kubeone apply` if they're supposed to expire in less than 90 days ([#1300](https://github.com/kubermatic/kubeone/pull/1300))
* Add support for running Kubernetes clusters on Amazon Linux 2 ([#1339](https://github.com/kubermatic/kubeone/pull/1339))
* Use the kubeadm v1beta3 API for all Kubernetes 1.22+ clusters ([#1457](https://github.com/kubermatic/kubeone/pull/1457))

### Addons

* Implement a mechanism for embedding YAML addons into KubeOne binary. The embedded addons can be enabled or overridden using the Addons API ([#1387](https://github.com/kubermatic/kubeone/pull/1387))
* Support organizing addons into subdirectories ([#1364](https://github.com/kubermatic/kubeone/pull/1364))
* Add a new optional embedded addon `default-storage-class` used to deploy default StorageClass for AWS, Azure, GCP, OpenStack, vSphere, or Hetzner clusters ([#1488](https://github.com/kubermatic/kubeone/pull/1488))
* Add a new KubeOne addon for handling unattended upgrades of the operating system ([#1291](https://github.com/kubermatic/kubeone/pull/1291))

## Changed

### General

* Increase the minimum Kubernetes version to v1.19.0. If you have Kubernetes clusters running v1.18 or older, you need to use an older KubeOne release to upgrade them to v1.19, and then upgrade to KubeOne 1.3.

### CLI

* The `kubeone reset` command requires an explicit confirmation like the `apply` command starting with this release. The command can be automatically approved by using the `--auto-approve` flag
* Improve the `kubeone reset` output to include more information about the target cluster ([#1474](https://github.com/kubermatic/kubeone/pull/1474))

### Fixed

* Make `kubeone apply` skip already provisioned static worker nodes ([#1485](https://github.com/kubermatic/kubeone/pull/1485))
* Extend restart API server script to handle failing `crictl logs` due to missing symlink. This fixes the issue with `kubeone apply` failing to restart the API server containers when provisioning or upgrading the cluster ([#1448](https://github.com/kubermatic/kubeone/pull/1448))
* Fix subsequent apply failures if CABundle is enabled ([#1404](https://github.com/kubermatic/kubeone/pull/1404))
* Fix NPE when migrating to containerd ([#1499](https://github.com/kubermatic/kubeone/pull/1499))
* Fix adding second container to the machine-controller-webhook Deployment ([#1433](https://github.com/kubermatic/kubeone/pull/1433))
* Fix missing ClusterRole rule for cluster-autoscaler ([#1331](https://github.com/kubermatic/kubeone/pull/1331))
* Fix missing confirmation for reset ([#1251](https://github.com/kubermatic/kubeone/pull/1251))
* Fix kubeone reset error when trying to list Machines ([#1416](https://github.com/kubermatic/kubeone/pull/1416))
* Ignore preexisting static manifests kubeadm preflight error ([#1335](https://github.com/kubermatic/kubeone/pull/1335))

### Updated

* Upgrade Terraform to 1.0.0. The minimum Terraform version as of this KubeOne release is v1.0.0. ([#1368](https://github.com/kubermatic/kubeone/pull/1368), [#1376](https://github.com/kubermatic/kubeone/pull/1376))
* Use latest available (wildcard) docker and containerd version ([#1358](https://github.com/kubermatic/kubeone/pull/1358))
* Update machine-controller to v1.35.2 ([#1489](https://github.com/kubermatic/kubeone/pull/1489))
* Update metrics-server to v0.5.0. This fixes support for Kubernetes 1.22 clusters ([#1483](https://github.com/kubermatic/kubeone/pull/1483))
  * The metrics-server now uses serving certificates signed by the Kubernetes CA instead of the self-signed certificates.
* OpenStack CCM version now depends on the Kubernetes version ([#1465](https://github.com/kubermatic/kubeone/pull/1465))
* vSphere CCM (CPI) version now depends on the Kubernetes version ([#1489](https://github.com/kubermatic/kubeone/pull/1489))
  * Kubernetes 1.22+ clusters are currently unsupported on vSphere (see Known Issues for more details)
* Update Hetzner CCM to v1.9.1 ([#1428](https://github.com/kubermatic/kubeone/pull/1428))
  * Add `HCLOUD_LOAD_BALANCERS_USE_PRIVATE_IP=true` to the environment if the network is configured
* Update Hetzner CSI driver to v1.6.0 ([#1491](https://github.com/kubermatic/kubeone/pull/1491))
* Update DigitalOcean CCM to v0.1.33 ([#1429](https://github.com/kubermatic/kubeone/pull/1429))
* Upgrade machine-controller addon apiextensions to v1 API ([#1423](https://github.com/kubermatic/kubeone/pull/1423))
* Update Go to 1.16.7 ([#1441](https://github.com/kubermatic/kubeone/pull/1441))

### Addons

* Replace the Canal CNI Go template with an embedded addon ([#1405](https://github.com/kubermatic/kubeone/pull/1405))
* Replace the WeaveNet Go template with an embedded addon ([#1407](https://github.com/kubermatic/kubeone/pull/1407))
* Replace the NodeLocalDNS template with an addon ([#1392](https://github.com/kubermatic/kubeone/pull/1392))
* Replace the metrics-server CCM Go template with an embedded addon ([#1411](https://github.com/kubermatic/kubeone/pull/1411))
* Replace the machine-controller Go template with an embedded addon ([#1412](https://github.com/kubermatic/kubeone/pull/1412))
* Replace the DigitalOcean CCM Go template with an embedded addon ([#1396](https://github.com/kubermatic/kubeone/pull/1396))
* Replace the Hetzner CCM Go template with an embedded addon ([#1397](https://github.com/kubermatic/kubeone/pull/1397))
* Replace the Packet CCM Go template with an embedded addon ([#1401](https://github.com/kubermatic/kubeone/pull/1401))
* Replace the OpenStack CCM Go template with an embedded addon ([#1402](https://github.com/kubermatic/kubeone/pull/1402))
* Replace the vSphere CCM Go template with an embedded addon ([#1410](https://github.com/kubermatic/kubeone/pull/1410))
* Upgrade calico-vxlan CNI plugin addon to v3.19.1 ([#1403](https://github.com/kubermatic/kubeone/pull/1403))

### Terraform Configs

* Inherit the firmware settings from the template VM in the Terraform configs for vSphere ([#1445](https://github.com/kubermatic/kubeone/pull/1445))

## Removed

* Remove CSIMigration and CSIMigrationComplete fields from the API ([#1473](https://github.com/kubermatic/kubeone/pull/1473))
  * Those two fields were non-functional since they were added, so this change shouldn't affect users.
  * If you have any of those those two fields set in the KubeOneCluster manifest, make sure to remove them or otherwise the validation will fail.
* Remove CNI patching ([#1386](https://github.com/kubermatic/kubeone/pull/1386))

# [v1.3.0-rc.0](https://github.com/kubermatic/kubeone/releases/tag/v1.3.0-rc.0) - 2021-09-06

## Attention Needed

* [**BREAKING/ACTION REQUIRED**] Increase the minimum Kubernetes version to v1.19.0 ([#1466](https://github.com/kubermatic/kubeone/pull/1466))
  * If you have Kubernetes clusters running v1.18 or older, you need to use an older KubeOne release to upgrade them to v1.19, and then upgrade to KubeOne 1.3.
  * Check out the [Compatibility guide](https://docs.kubermatic.com/kubeone/main/architecture/compatibility/) for more information about supported Kubernetes versions for each KubeOne release.
* [**BREAKING/ACTION REQUIRED**] Add support for CSI plugins for clusters using external cloud provider (i.e. `.cloudProvider.external` is enabled)
  * The Cinder CSI plugin is deployed by default for OpenStack clusters ([#1465](https://github.com/kubermatic/kubeone/pull/1465))
  * The Hetzner CSI plugin is deployed by default for Hetzner clusters
  * The vSphere CSI plugin is deployed by default if the CSI plugin configuration is provided via newly-added  `cloudProvider.csiConfig` field
    * More information about the CSI plugin configuration can be found in the vSphere CSI docs: https://vsphere-csi-driver.sigs.k8s.io/driver-deployment/installation.html#create_csi_vsphereconf
    * Note: the vSphere CSI plugin requires vSphere version 6.7U3.
  * The default StorageClass is **not** deployed by default. It can be deployed via new Addons API by enabling the `default-storage-class` addon, or manually.
  * **ACTION REQUIRED**: If you already have the CSI plugin deployed, you need to make sure that your CSI plugin deployment is compatible with the KubeOne CSI plugin addon.
    * You can find the CSI addons in the `addons` directory: https://github.com/kubermatic/kubeone/tree/main/addons
    * If your CSI plugin deployment is incompatible with the KubeOne CSI addon, you can resolve it in one of the following ways:
      * Delete your CSI deployment and let KubeOne install the CSI driver for you. **Note**: you'll **not** be able to mount volumes until you don't install the CSI driver again.
      * Override the appropriate CSI addon with your deployment manifest. With this way, KubeOne will install the CSI plugin using your manifests. To do this, you need to:
        * Enable addons in the KubeOneCluster manifest (`.addons.enable`) and provide the path to addons directory (`.addons.path`, for example: `./addons`)
        * Create a subdirectory in the addons directory named same as the CSI addon used by KubeOne, for example `./addons/csi-openstack-cinder` or `./addons/csi-vsphere` (see https://github.com/kubermatic/kubeone/tree/main/addons for addon names)
        * Put your CSI deployment manifests in the newly created subdirectory

## Known Issues

* It's currently **not** possible to provision or upgrade to Kubernetes 1.22 for clusters running on vSphere. This is because vSphere CCM and CSI don't support Kubernetes 1.22. We'll introduce Kubernetes 1.22 support for vSphere as soon as new CCM and CSI releases with support for Kubernetes 1.22 are out.
* Clusters provisioned with Kubernetes 1.22 or upgraded from 1.21 to 1.22 using KubeOne 1.3.0-alpha.1 use a metrics-server version incompatible with Kubernetes 1.22. This might cause issues with deleting Namespaces that manifests by the Namespace being stuck in the Terminating state. This can be fixed by upgrading the metrics-server by running `kubeone apply`.
* The new Addons API requires the addons directory path (`.addons.path`) to be provided and the directory must exist (it can be empty), even if only embedded addons are used. If the path is not provided, it'll default to `./addons`.

## Added

### Features

* Implement the Addons API used to manage addons deployed by KubeOne ([#1462](https://github.com/kubermatic/kubeone/pull/1462), [#1486](https://github.com/kubermatic/kubeone/pull/1486))
  * The new Addons API can be used to deploy the addons embedded in the KubeOne binary.
  * Currently available addons are: `backups-restic`, `default-storage-class`, and `unattended-upgrades`.
  * More information about the new API can be found by running `kubeone config print --full`.
* [**BREAKING/ACTION REQUIRED**] Add support for the Cinder CSI plugin ([#1465](https://github.com/kubermatic/kubeone/pull/1465))
  * The plugin is deployed by default for OpenStack clusters using the external cloud provider.
  * Check out the Attention Needed section of the changelog for more information.
* Add support for the vSphere CSI plugin ([#1484](https://github.com/kubermatic/kubeone/pull/1484))
  * Deploying the CSI plugin requires providing the CSI configuration using a newly added `.cloudProvider.csiConfig` field
    * More information about the CSI plugin configuration can be found in the vSphere CSI docs: https://vsphere-csi-driver.sigs.k8s.io/driver-deployment/installation.html#create_csi_vsphereconf
  * The CSI plugin is deployed automatically if `.cloudProvider.csiConfig` is provided and `.cloudProvider.external` is enabled
  * Check out the Attention Needed section of the changelog for more information.
* Implement the CCM/CSI migration for OpenStack and vSphere ([#1468](https://github.com/kubermatic/kubeone/pull/1468), [#1469](https://github.com/kubermatic/kubeone/pull/1469), [#1472](https://github.com/kubermatic/kubeone/pull/1472), [#1482](https://github.com/kubermatic/kubeone/pull/1482), [#1487](https://github.com/kubermatic/kubeone/pull/1487), [#1494](https://github.com/kubermatic/kubeone/pull/1494))
  * The CCM/CSI migration is used to migrate clusters running in-tree cloud provider (i.e. with `.cloudProvider.external` set to `false`) to the external CCM (cloud-controller-manager) and CSI plugin.
  * The migration is implemented with the `kubeone migrate to-ccm-csi` command.
  * The CCM/CSI migration for vSphere is currently experimental and not tested.
  * More information about how the CCM/CSI migration works can be found by running `kubeone migrate to-ccm-csi --help`.

### Addons

* Add a new optional embedded addon `default-storage-class` used to deploy default StorageClass for AWS, Azure, GCP, OpenStack, vSphere, or Hetzner clusters ([#1488](https://github.com/kubermatic/kubeone/pull/1488))

## Changed

### General

* [**BREAKING/ACTION REQUIRED**] Increase the minimum Kubernetes version to v1.19.0 ([#1466](https://github.com/kubermatic/kubeone/pull/1466))
  * If you have Kubernetes clusters running v1.18 or older, you need to use an older KubeOne release to upgrade them to v1.19, and then upgrade to KubeOne 1.3.
  * Check out the [Compatibility guide](https://docs.kubermatic.com/kubeone/main/architecture/compatibility/) for more information about supported Kubernetes versions for each KubeOne release.
* Improve the `kubeone reset` output to include more information about the target cluster ([#1474](https://github.com/kubermatic/kubeone/pull/1474))

### Fixed

* Make `kubeone apply` skip already provisioned static worker nodes ([#1485](https://github.com/kubermatic/kubeone/pull/1485))
* Fix NPE when migrating to containerd ([#1499](https://github.com/kubermatic/kubeone/pull/1499))

### Updated

* OpenStack CCM version now depends on the Kubernetes version ([#1465](https://github.com/kubermatic/kubeone/pull/1465))
  * Kubernetes 1.19 clusters use OpenStack CCM v1.19.2
  * Kubernetes 1.20 clusters use OpenStack CCM v1.20.2
  * Kubernetes 1.21 clusters use OpenStack CCM v1.21.0
  * Kubernetes 1.22+ clusters use OpenStack CCM v1.22.0
* vSphere CCM (CPI) version now depends on the Kubernetes version ([#1489](https://github.com/kubermatic/kubeone/pull/1489))
  * Kubernetes 1.19 clusters use vSphere CPI v1.19.0
  * Kubernetes 1.20 clusters use vSphere CPI v1.20.0
  * Kubernetes 1.21 clusters use vSphere CPI v1.21.0
  * Kubernetes 1.22+ clusters are currently unsupported on vSphere (see Known Issues for more details)
* Update metrics-server to v0.5.0 ([#1483](https://github.com/kubermatic/kubeone/pull/1483))
  * This fixes support for Kubernetes 1.22 clusters.
  * The metrics-server now uses serving certificates signed by the Kubernetes CA instead of the self-signed certificates.
* Update machine-controller to v1.35.2 ([#1489](https://github.com/kubermatic/kubeone/pull/1489))
* Update Hetzner CSI driver to v1.6.0 ([#1491](https://github.com/kubermatic/kubeone/pull/1491))

## Removed

* Remove CSIMigration and CSIMigrationComplete fields from the API ([#1473](https://github.com/kubermatic/kubeone/pull/1473))
  * Those two fields were non-functional since they were added, so this change shouldn't affect users.
  * If you have any of those those two fields set in the KubeOneCluster manifest, make sure to remove them or otherwise the validation will fail.

# [v1.3.0-alpha.1](https://github.com/kubermatic/kubeone/releases/tag/v1.3.0-alpha.1) - 2021-08-18

## Known Issues

* Clusters provisioned with Kubernetes 1.22 or upgraded from 1.21 to 1.22 using KubeOne 1.3.0-alpha.1 use a metrics-server version incompatible with Kubernetes 1.22. This might cause issues with deleting Namespaces that manifests by the Namespace being stuck in the Terminating state. This can be fixed by upgrading to KubeOne 1.3.0-rc.0 and running `kubeone apply`.

## Added

* Add support for Kubernetes 1.22 ([#1447](https://github.com/kubermatic/kubeone/pull/1447), [#1456](https://github.com/kubermatic/kubeone/pull/1456))
* Add support for the kubeadm v1beta3 API. The kubeadm v1beta3 API is used for all Kubernetes 1.22+ clusters. ([#1457](https://github.com/kubermatic/kubeone/pull/1457))

## Changed

### Fixed

* Fix adding second container to the machine-controller-webhook Deployment ([#1433](https://github.com/kubermatic/kubeone/pull/1433))
* Extend restart API server script to handle failing `crictl logs` due to missing symlink. This fixes the issue with `kubeone apply` failing to restart the API server containers when provisioning or upgrading the cluster ([#1448](https://github.com/kubermatic/kubeone/pull/1448))

### Updated

* Update Go to 1.16.7 ([#1441](https://github.com/kubermatic/kubeone/pull/1441))
* Update machine-controller to v1.35.1 ([#1440](https://github.com/kubermatic/kubeone/pull/1440))
* Update Hetzner CCM to v1.9.1 ([#1428](https://github.com/kubermatic/kubeone/pull/1428))
  * Add `HCLOUD_LOAD_BALANCERS_USE_PRIVATE_IP=true` to the environment if the network is configured
* Update DigitalOcean CCM to v0.1.33 ([#1429](https://github.com/kubermatic/kubeone/pull/1429))

### Terraform Configs

* Inherit the firmware settings from the template VM in the Terraform configs for vSphere ([#1445](https://github.com/kubermatic/kubeone/pull/1445))

# [v1.3.0-alpha.0](https://github.com/kubermatic/kubeone/releases/tag/v1.3.0-alpha.0) - 2021-07-21

## Attention Needed

* [**BREAKING/ACTION REQUIRED**] The `kubeone reset` command requires an explicit confirmation like the `apply` command starting with this release
  * Running the `reset` command requires typing `yes` to confirm the intention to unprovision/reset the cluster
  * The command can be automatically approved by using the `--auto-approve` flag
* [**BREAKING/ACTION REQUIRED**] Upgrade Terraform to 1.0.0. The minimum Terraform version as of this KubeOne release is v1.0.0. ([#1368](https://github.com/kubermatic/kubeone/pull/1368))
* [**BREAKING/ACTION REQUIRED**] Use AdmissionRegistration v1 API for machine-controller-webhook. The minimum supported Kubernetes version is now 1.16. ([#1290](https://github.com/kubermatic/kubeone/pull/1290))
  * Since AdmissionRegistration v1 got introduced in Kubernetes 1.16, the minimum Kubernetes version that can be managed by KubeOne is now 1.16. If you're running the Kubernetes clusters running 1.15 or older, please use the older release of KubeOne to upgrade those clusters
* KubeOne Addons can now be organized into subdirectories. It currently remains possible to put addons in the root of the addons directory, however, this is option is considered as deprecated as of this release. We highly recommend all users to reorganize their addons into subdirectories, where each subdirectory is for YAML manifests related to one addon.

## Added

### API

* Add new kube-proxy configuration API ([#1420](https://github.com/kubermatic/kubeone/pull/1420))
  * This API allows users to switch kube-proxy to IPVS mode, and configure IPVS properties such as strict ARP and scheduler
  * The default kube-proxy mode remains iptables
* Add support for Encryption Providers ([#1241](https://github.com/kubermatic/kubeone/pull/1241), [#1320](https://github.com/kubermatic/kubeone/pull/1320))
* Add support for specifying a custom Root CA bundle ([#1316](https://github.com/kubermatic/kubeone/pull/1316))

### Features

* Docker to containerd automated migration ([#1362](https://github.com/kubermatic/kubeone/pull/1362))
* Automatically renew Kubernetes certificates when running `kubeone apply` if they're supposed to expire in less than 90 days ([#1300](https://github.com/kubermatic/kubeone/pull/1300))
* Ignore preexisting static manifests kubeadm preflight error ([#1335](https://github.com/kubermatic/kubeone/pull/1335))
* Add a new `kubeone config images list` subcommand to list images used by KubeOne and kubeadm ([#1334](https://github.com/kubermatic/kubeone/pull/1334))
* Add containerd support for Flatcar clusters ([#1340](https://github.com/kubermatic/kubeone/pull/1340))
* Add support for running Kubernetes clusters on Amazon Linux 2 ([#1339](https://github.com/kubermatic/kubeone/pull/1339))

### Addons

* Implement a mechanism for embedding YAML addons into KubeOne binary ([#1387](https://github.com/kubermatic/kubeone/pull/1387))
* Support organizing addons into subdirectories ([#1364](https://github.com/kubermatic/kubeone/pull/1364))
* Add a new KubeOne addon for handling unattended upgrades of the operating system ([#1291](https://github.com/kubermatic/kubeone/pull/1291))
* Add a new KubeOne addon for deploying the Hetzner CSI plugin ([#1418](https://github.com/kubermatic/kubeone/pull/1418))

## Changed

### CLI

* [**BREAKING/ACTION REQUIRED**] The `kubeone reset` command requires an explicit confirmation like the `apply` command starting with this release
  * Running the `reset` command requires typing `yes` to confirm the intention to unprovision/reset the cluster
  * The command can be automatically approved by using the `--auto-approve` flag

### Bug Fixes

* Fix missing ClusterRole rule for cluster autoscaler ([#1331](https://github.com/kubermatic/kubeone/pull/1331))
* Fix missing confirmation for reset ([#1251](https://github.com/kubermatic/kubeone/pull/1251))
* Remove CNI patching ([#1386](https://github.com/kubermatic/kubeone/pull/1386))
* Fix subsequent apply failures if CABundle is enabled ([#1404](https://github.com/kubermatic/kubeone/pull/1404))
* Fix kubeone reset error when trying to list Machines ([#1416](https://github.com/kubermatic/kubeone/pull/1416))

### Updated

* [**BREAKING/ACTION REQUIRED**] Upgrade Terraform to 1.0.0. The minimum Terraform version as of this KubeOne release is v1.0.0. ([#1368](https://github.com/kubermatic/kubeone/pull/1368), [#1376](https://github.com/kubermatic/kubeone/pull/1376))
* Use latest available (wildcard) docker and containerd version ([#1358](https://github.com/kubermatic/kubeone/pull/1358))
* Upgrade machinecontroller to v1.33.0 ([#1391](https://github.com/kubermatic/kubeone/pull/1391))
* Upgrade machine-controller addon apiextensions to v1 API ([#1423](https://github.com/kubermatic/kubeone/pull/1423))
* Upgrade calico-vxlan CNI plugin addon to v3.19.1 ([#1403](https://github.com/kubermatic/kubeone/pull/1403))
* Update Go to 1.16.1 ([#1267](https://github.com/kubermatic/kubeone/pull/1267))

### Addons

* Replace the Canal CNI Go template with an embedded addon ([#1405](https://github.com/kubermatic/kubeone/pull/1405))
* Replace the WeaveNet Go template with an embedded addon ([#1407](https://github.com/kubermatic/kubeone/pull/1407))
* Replace the NodeLocalDNS template with an addon ([#1392](https://github.com/kubermatic/kubeone/pull/1392))
* Replace the metrics-server CCM Go template with an embedded addon ([#1411](https://github.com/kubermatic/kubeone/pull/1411))
* Replace the machine-controller Go template with an embedded addon ([#1412](https://github.com/kubermatic/kubeone/pull/1412))
* Replace the DigitalOcean CCM Go template with an embedded addon ([#1396](https://github.com/kubermatic/kubeone/pull/1396))
* Replace the Hetzner CCM Go template with an embedded addon ([#1397](https://github.com/kubermatic/kubeone/pull/1397))
* Replace the Packet CCM Go template with an embedded addon ([#1401](https://github.com/kubermatic/kubeone/pull/1401))
* Replace the OpenStack CCM Go template with an embedded addon ([#1402](https://github.com/kubermatic/kubeone/pull/1402))
* Replace the vSphere CCM Go template with an embedded addon ([#1410](https://github.com/kubermatic/kubeone/pull/1410))
