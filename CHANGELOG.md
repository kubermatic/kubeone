# Changelog

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

