# Unattended Upgrades

This addon will automate upgrading system packages of the distro of your choice.

## Requirements

Since KubeOne 1.3+ we automatically label control-plane nodes with
`v1.kubeone.io/operating-system` and worker nodes with
`v1.machine-controller.kubermatic.io/operating-system` and use those labels as
nodeAffinity in this addon manifests.

## What's included

This addon provides bunch of DaemonSets and operators:

* **Debian/Ubuntu**
  DaemonSet that will install `unattended-upgrades`
* **RHEL/CentOS/Rocky Linux/Amazon Linux 2**
  DaemonSet that will install and configure `yum-cron`/`dnf-automatic`
* **Flatcar Linux**
  [Flatcar Linux Update Operator](https://github.com/kinvolk/flatcar-linux-update-operator)
  (only deployed if the cluster has at least one Flatcar Linux node, see
  [Flatcar Linux Update Operator](#flatcar-linux-update-operator))

## Deployment instructions

In `kubeone.yaml` config add:
```yaml
addons:
  enable: true
  addons:
    - name: "unattended-upgrades"
```

## Flatcar Linux Update Operator

The Flatcar Linux Update Operator is deployed only if at least one node of the
cluster runs Flatcar Linux. KubeOne looks at the control plane hosts and the
static workers from the KubeOne manifest, at the `dynamicWorkers` from the
KubeOne manifest and at the MachineDeployments that exist in the cluster. If
none of them is Flatcar Linux, none of the operator resources (namespace, RBAC,
Deployment and DaemonSet) are rendered.

If you add Flatcar Linux nodes later, the operator is deployed on the next
`kubeone apply`.

The result can be overridden with the `flatcarNodesPresent` addon parameter:

```yaml
addons:
  enable: true
  addons:
    - name: "unattended-upgrades"
      params:
        flatcarNodesPresent: "true"
```

On clusters where the operator was deployed by an earlier KubeOne version, the
namespace and the workloads in it are removed on the next `kubeone apply`. The
cluster-scoped `flatcar-linux-update-agent` and `flatcar-linux-update-operator`
ClusterRoles and ClusterRoleBindings are not pruned by `kubectl apply --prune`
and can be removed manually.

## Information about permissions

Since daemonSets provided by this addon are making changes on the nodes
themselves they require elevated permissions like full root access to the host
machine.

## Development

### Generate Flatcar Linux Update Operator YAML
```shell
make update
```

The surrounding template condition makes sure the operator is only deployed on
clusters with Flatcar Linux nodes.
