# KubeOneCluster API

**Author:** Marko Mudrinić ([@xmudrii](https://github.com/xmudrii))  
**Status:** Draft | Review | **Final**  
**Created:** 2019-04-09  
**Last updated:** 2019-04-19

## Abstract

The KubeOne API should be refactored to follow the Kubernetes API conventions and eventually adopt recommendations
for creating the Cluster APIs from the upstream Cluster-API project. Following the Kubernetes conventions will allow
us to reuse the mechanisms for handling tasks such as the API versioning.

It's important to ensure the KubeOne API is stable and that we can provide the backwards compatibility promise
before reaching beta. In case of any API change, it should be possible to use KubeOne with the old API version
for some time, while migrating to the new API version should be as easy as possible.

## Goals

* Build the KubeOne Cluster API to replace the API we're currently using.
* Ensure the new API follows the Kubernetes API conventions.
* Implement the versioning support for the new KubeOne API.
* Follow the recommendations for creating the Cluster APIs from the Cluster-API project.

## Non-goals

* Adopt the upstream Cluster-API.
  * Since the Cluster-API project is in the alpha phase and that some features we need may be missing,
  initially we'll build a downstream API for KubeOne while following the recommendations from the upstream API.
* Store the manifest used to provision the cluster on the target cluster in form of a Secret or a CustomResource.
  * See [the issue #173](https://github.com/kubermatic/kubeone/issues/173) for more details about this feature.
* Provide a mechanism for automatic migration from the old KubeOne API to the new API.
  * The steps needed to be done in order to migrate to the new API will be documented.

## Implementation

The new API should try to match the old one as much as possible. If there is space to improve the user experience,
those improvements should be done now.

The API group is going to be called `kubeone.io` and the object will be called `KubeOneCluster`. The first API
version is going to be `v1alpha1`.

The following changes **will be made** in order to improve the user experience:

* `Network` will be renamed to `ClusterNetwork`
  * Currently the `Network` structure is used to configure Kubernetes networking (pod and services CIDR and
  node port range). Therefore, it makes more sense to name it `ClusterNetwork`, so it's clear it's about the cluster
  and not individual machines.
  * A new field for service domain name will be added (`ServiceDomain`).
* The `Name` field will be removed from the Spec. Instead, we'll use the object name as the cluster name.

The following changes **should be considered** in order to improve the user experience:

* Rename `APIServerConfig` to `APIEndpoint`
  * We have a `Features` structure that is used to configure various cluster features, while the `APIServerConfig`
  only contains the load balancer IP address.
  * At some point we should consider moving it from the Spec to the Status.
* Refactor `Features` to be consistent
  * All config structures inside the `Features` structure should at least have a field for enabling the feature.
  * Currently, `PodSecurityPolicy` and `DynamicAuditing` have the `enable` field, while the `MetricsServer` has the
  the `disable` field. This can lead to confusion and therefore should be made consistent.
  * To make this possible, all fields in the `Features` structure will be turned into pointers.

### API versioning

Each Kubernetes API has the `internal` API, that is used in code, and the versioned API (e.g. `v1alpha1` API), that is consumed
by the end user.

Kubernetes automatically generates code that converts any supported versioned API to the internal API. In some cases,
the conversion can't be done automatically (e.g. some fields are added or removed from the internal API) and in such cases
it's possible to define custom logic for conversion based on the generated conversion methods.

This approach is successfully used by `kubeadm`. For example, [the following code](https://github.com/kubernetes/kubernetes/blob/release-1.12/cmd/kubeadm/app/apis/kubeadm/v1alpha3/conversion.go) was used to convert the `v1alpha3` API to the internal API.

## Backwards Compatibility

The backwards compatibility policy will be based on the Kubernetes API conventions,
which are described in the following documents:

* [The Kubernetes API](https://kubernetes.io/docs/concepts/overview/kubernetes-api/)
* [Kubernetes Deprecation Policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)

The API should be divided into the following levels:

* Alpha (`v1alpha`):
  * Unstable APIs that are **not** recommended to be used in the production.
  * Can be changed at any time without any prior notice.
  * The API can be entirely dropped at any time.
* Beta (`v1beta`):
  * Pre-release API that can be safely used by operators.
  * Include features that may contain bugs and/or are not properly tested.
  * The API falls under the deprecation policy.
* Stable (`v1`):
  * The API is stable
  * The features provided by the API are stable and well-tested.
  * The API is supported for many upcoming releases.

The following rules must be satisfied:

* When an API element is removed or changed, the API version must be incremented.
  * This includes changing the behavior that's covered by the API element.
* API objects must be able to round-trip between API versions without the information loss.
  * The object written as v1 and then read as v2 and converted back to v1 should be identical to the original v1 object.
  * The representations between v1 and v2 may differ, but the system should be able to convert them in both ways.
  * Exceptions are the API objects that don't exist in a given version.
    * In this situation, this should be migrated by adding an annotation for example.
* An API version may not be deprecated at least until a new beta API is not released to replace the old API.

The APIs must be supported for at least:

* Alpha APIs can be deprecated at any time
* Beta APIs must be supported for at least 2 versions or 3 months (whichever is longer)
* Stable APIs must be supported for at least 5 versions or 6 months (whichever is longer)
