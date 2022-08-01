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
* Default `--destroy-workers` to `true` for the `kubeone reset` command ([#252](https://github.com/kubermatic/kubeone/pull/252))
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
