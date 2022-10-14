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
  * More information about migrating to the new API and what has been changed can be found in the [API migration document](https://docs.kubermatic.com/kubeone/main/advanced/api_migration/)
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
  * More details about how to use the apply command can be found in the [Cluster reconciliation (apply) document](https://docs.kubermatic.com/kubeone/main/advanced/cluster_reconciliation/)
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
  * Check out [the documentation](https://docs.kubermatic.com/kubeone/main/workers/static_workers/) to learn more about static worker nodes
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
  * More information about how to use this addon can be found on the [docs website](https://docs.kubermatic.com/kubeone/main/using_kubeone/calico-vxlan-addon/)
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
  * More details about how to use the apply command can be found in the [Cluster reconciliation (apply) document](https://docs.kubermatic.com/kubeone/main/using_kubeone/cluster_reconciliation/)
* Implement the `kubeone config machinedeployments` command ([#966](https://github.com/kubermatic/kubeone/pull/966))
  * The new command is used to generate a YAML manifest containing all MachineDeployment objects defined in the KubeOne configuration manifest and Terraform output
  * The generated manifest can be used with kubectl if you want to create and modify MachineDeployments once the cluster is created
* Add support for CentOS 8 ([#981](https://github.com/kubermatic/kubeone/pull/981))
* Add the Calico VXLAN addon ([#972](https://github.com/kubermatic/kubeone/pull/972))
  * More information about how to use this addon can be found on the [docs website](https://docs.kubermatic.com/kubeone/main/using_kubeone/calico-vxlan-addon/)

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

* See known issues for the [v1.0.0-beta.1 release](https://github.com/kubermatic/kubeone/blob/main/CHANGELOG.md#known-issues) for more details.

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
