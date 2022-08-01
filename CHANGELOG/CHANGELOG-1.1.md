# [v1.1.0](https://github.com/kubermatic/kubeone/releases/tag/v1.1.0) - 2020-11-13

**Changelog since v1.0.5.**

## Attention Needed

* [**Breaking**] Use Ubuntu 20.04 (Focal) in the example Hetzner Terraform config ([#1102](https://github.com/kubermatic/kubeone/pull/1102))
  * It's **highly recommended** to bind the image by setting `var.image` to the image you're currently using to prevent the instances from being recreated the next time you run Terraform!

## Added

### General

* Implement the OverwriteRegistry functionality ([#1145](https://github.com/kubermatic/kubeone/pull/1145))
  * This PR adds a new top-level API field `registryConfiguration` which controls how images used for components deployed by KubeOne and kubeadm are pulled from an image registry.
  * The `registryConfiguration.overwriteRegistry` field specifies a custom Docker registry to be used instead of the default one.
  * The `registryConfiguration.insecureRegistry` field configures Docker to consider the registry specified in `registryConfiguration.overwriteRegistry` as an insecure registry.
  * For example, if `registryConfiguration.overwriteRegistry` is set to `127.0.0.1:5000`, image called `k8s.gcr.io/kube-apiserver:v1.19.3` would become `127.0.0.1:5000/kube-apiserver:v1.19.3`.
  * Setting `registryConfiguration.overwriteRegistry` applies to all images deployed by KubeOne and kubeadm, including addons deployed by KubeOne.
  * Setting `registryConfiguration.overwriteRegistry` applies to worker nodes managed by machine-controller and KubeOne as well.
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
  * The `registryConfiguration.overwriteRegistry` field can be used to specify a custom Docker registry to be used instead of the default one.
  * For example, if `registryConfiguration.overwriteRegistry` is set to `127.0.0.1:5000`, image called `k8s.gcr.io/kube-apiserver:v1.19.3` would become `127.0.0.1:5000/kube-apiserver:v1.19.3`.
  * Setting `registryConfiguration.overwriteRegistry` applies to all images deployed by KubeOne and kubeadm, including addons deployed by KubeOne.
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
