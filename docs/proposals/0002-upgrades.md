# kubeone upgrade

**Author**: Artiom Diomin (@kron4eg)  
**Status**: Draft proposal

## Motivation and Background
The point of `kubeone` project since the beginning was to provide full cluster
life-cycle experience, which should include ability to upgrade cluster over
time.

## Implementation
`kubeadm` makes the process of upgrade quite simple. Before proceed to actual
upgrade kubeone need to grab some info about cluster in question:

### Reconciliation
* grab nodes info from API server
    * versions
    * external/internal IPs
    * node labels
* grab previously saved `kubeone config` from configmap
* make sure cluster in healthy before next steps
    * 3/3 of hosts are accessible / initialized
    * 3/3 of nodes are ready
    * 3/3 of nodes versions <= requested version
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
* cordon/drain node
* run `kubeadm upgrade`
* upgrade/restart kubelet
* wait for etcd to settle after restart (watch pod to became ready, which means
  Running 1/1)
* uncordon node
* unlabel `kubeone.io/upgrade-in-process`

## Tasks & effort
* build intel gathering process
* build drain process (using eviction API)
* build new `kubeone upgrade` CLI command
