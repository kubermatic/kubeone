# Cluster Autoscaler

[This addon][addon] deploys and configures the
[Kubernetes Cluster Autoscaler][autoscaler]. The Cluster Autoscaler is
a component that automatically adjusts the size of a Kubernetes Cluster so that
all pods have a place to run and there are no unneeded nodes.

## Prerequisites

* The worker nodes need to be managed by
  [Kubermatic machine-controller][machine-controller]
  * We recommend checking the [Concepts][docs-concepts] document to learn more
    about how Cluster-API and Kubermatic machine-controller work
* Cluster running Kubernetes v1.18 or newer is recommended

## How It Works?

This addon configures the Cluster Autoscaler to use the Cluster-API provider.
The cluster is autoscaled by increasing/decreasing the number of replicas on
the chosen [MachineDeployment object][docs-machinedeployment].

Once the MachineDeployment is scaled, Kubermatic machine-controller creates a
new instance and joins it a cluster (if the cluster is scaled up), or deletes
one of the existing instances (if the cluster is scaled down).
The MachineDeployment object for scaling is chosen randomly from a set of
MachineDeployments that have autoscaling enabled.

The cluster is automatically scaled up/down when one of the following
conditions is satisfied:

* there are pods that failed to run in the cluster due to insufficient
  resources
* there are nodes in the cluster that have been underutilized for an extended
  period of time (10 minutes by default) and their pods can be placed on other
  existing nodes

## Comparison to Autoscaling Groups (ASGs)

The Kubermatic machine-controller is responsible for creating instances,
joining them a cluster, and deleting them once the appropriate Machine object
is deleted. It works directly with instances, i.e. ASGs are **not** used.

The advantage over using ASGs (or other similar mechanisms) is that all worker
nodes are defined and managed using Kubernetes objects. You can use kubectl or
the Kubernetes API directly to:

* create new worker nodes, modify, or delete existing ones
* check the health status of worker nodes
* use rolling updates to upgrade and/or modifying existing worker nodes

KubeOne uses Kubermatic machine-controller by default for managing worker
nodes.

## Supported Kubernetes Versions

The Cluster-API provider for Cluster Autoscaler has been implemented in the
Cluster Autoscaler version 1.18. The Cluster Autoscaler team
[recommends][recommended-autoscaler-versions] matching the minor version of the
Kubernetes cluster with the minor version of the Cluster Autoscaler. This means
that it's recommended to use this addon only on clusters running Kubernetes
1.18 or newer.

**Note:** The addon might work on older Kubernetes clusters as well, however,
it has not been tested.

## Choosing MachineDeployment objects for Autoscaling

The Cluster Autoscaler only considers MachineDeployment with the valid
annotations. The annotations are used to control the minimum and maximum number
of replicas per MachineDeployment:

* `cluster.k8s.io/cluster-api-autoscaler-node-group-min-size` - the minimum
  number of replicas (must be greater than zero)
* `cluster.k8s.io/cluster-api-autoscaler-node-group-max-size` - the maximum
  number of replicas

**Note:** You don't need to apply those annotations to all MachineDeployment
objects. They should be applied only on MachineDeployments that should be
considered by Cluster Autoscaler.

The annotations can be applied to MachineDeployments once the cluster is
provisioned and worker nodes are created.

Run the following kubectl command to inspect available MachineDeployments:

```bash
kubectl get machinedeployments -n kube-system
```

Run the following commands to annotate the MachineDeployment object. Make sure
to replace the MachineDeployment name and minimum/maximum size with the
appropriate values.

```bash
kubectl annotate machinedeployment -n kube-system <machinedeployment-name> cluster.k8s.io/cluster-api-autoscaler-node-group-min-size="<min-size>"
kubectl annotate machinedeployment -n kube-system <machinedeployment-name> cluster.k8s.io/cluster-api-autoscaler-node-group-max-size="<max-size>"
```

## Using The Addon

You need to replace the following values with the actual ones:

* `AUTOSCALER_VERSION` needs to be replaced with the appropriate Cluster
  Autoscaler version
  * The minor versions of Cluster Autoscaler and Kubernetes cluster should
    match, as per
    [Cluster Autoscaler recommendations][recommended-autoscaler-versions]
  * You can find the available Cluster Autoscaler versions by searching for
    Cluster Autoscaler in the
    [autoscaler GitHub repository][autoscaler-releases]

You can find more information about deploying addons in the
[Addons document][using-addons].

[addon]: ./cluster-autoscaler.yaml
[autoscaler]: https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler
[machine-controller]: https://github.com/kubermatic/machine-controller
[docs-concepts]: https://docs.kubermatic.com/kubeone/main/architecture/concepts/
[docs-machinedeployment]: https://docs.kubermatic.com/kubeone/main/architecture/concepts/#machinedeployments
[recommended-autoscaler-versions]: https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler#releases
[autoscaler-releases]: https://github.com/kubernetes/autoscaler/releases
[using-addons]: https://docs.kubermatic.com/kubeone/main/guides/addons/
