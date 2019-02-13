# kubeone upgrade

**Author**: Artiom Diomin (@kron4eg)  
**Status**: Draft proposal

## Motivation and Background
The point of `kubeone` project since the beginning was to provide full cluster
life-cycle experience, which should include ability to upgrade cluster over
time.

## Goals
* orchestrate a minimally viable way to upgrade kubernetes control plane
  components

## Non Goals
Following topics are subjects for future proposals:
* "addons manager"
    * upgrading whatsoever custom deployments kubeone created during `install`
* before upgrade backups of currently running configuration and etcd snapshots

## Implementation
`kubeadm` makes the process of upgrade quite simple. Before proceed to actual
upgrade kubeone need to grab some info about cluster in question:

### Pre-flight checks
* grab nodes info from API server
    * versions
    * external/internal IPs
    * node labels
* grab previously saved `kubeone config` from configmap
* make sure cluster in healthy before next steps
    * 3/3 of hosts are accessible / initialized
    * 3/3 of nodes are ready
    * 3/3 of nodes versions <= requested version
    * 3/3 of nodes pass the kubernetes version skew policy
    * 0/3 of nodes have `kubeone.io/upgrade-in-process` label (overridden by --force)

### `kubeone.io/upgrading-in-process` label
This node label is a fail-safe mechanism. It signify that node is being upgraded
by a `kubeone`. It's a kind of lock, to lock concurrent upgrades, and also to
interrupt upgrade attempt in case if previous had failed.

In case of `kubeone upgrade` failure, label will signify to consequent `kubeone
upgrade` that something is broken. Kubeone operator would need to fix problem
manually and remove that label from the node.

### Upgrade commit
loop over nodes nodes do:
* upgrade kubeadm binary
* label as `kubeone.io/upgrade-in-process`
* run `kubeadm upgrade apply` on "leader node"
* run `kubeadm upgrade node experimental-control-plane` on "other nodes"
* upgrade/restart kubelet
* wait for etcd to settle after restart (watch pod to became ready, which means
  Running 1/1)
* unlabel `kubeone.io/upgrade-in-process`

Once done, update MachineDeployment to upgrade workers as well.

## Tasks & effort
* build pre-flight checks
* scripts to support system packages upgrades
* scripts to support `kubeadm` invocations
* upgrade MachineDeployments if any
* build new `kubeone upgrade` CLI command
