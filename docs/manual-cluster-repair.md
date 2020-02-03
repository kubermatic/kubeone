# Manual cluster repair
When one of the control-plane VM instances fails (i.e. lost at cloud-provider),
it's necessary to replace it, as fast as possible to avoid losing etcd quorum
and blocking all `kube-apiserver` operations.

This guide will demonstrate how to restore your cluster to the normal state
(i.e. have 3 healthy etcd instances running).

Note: **lost at cloud-provider**: means that VM is ether malfunctions or
terminated forever. In ether case the VM instance should be deleted and
recreated, for later use by KubeOne.

## One dead control-plane instance
It can happen that cloud-provider VM instance, that hosts control-plane node,
will fail and lost forever.

In this document we will cover the case when 1 (out of 3 in total) control-plane
instances is lost.

### Remove dead etcd member
Even when one etcd member is physically (and abruptly) removed, etcd-ring still
hopes it might return back online. Unfortunately this is not our case and we
need to let etcd-ring know that dead etcd member is gone forever (i.e. remove
dead etcd member from the known peers).

Exec into the shell of the alive etcd container:
```bash
kubectl -n kube-system exec -it etcd-<ALIVE-HOSTNAME> sh
```

Setup client TLS authentication:
```bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=/etc/kubernetes/pki/etcd/ca.crt
export ETCDCTL_CERT=/etc/kubernetes/pki/etcd/healthcheck-client.crt
export ETCDCTL_KEY=/etc/kubernetes/pki/etcd/healthcheck-client.key
```

Retrieve currently known members list (example output):
```bash
etcdctl member list
```

Example output:
```bash
2ce40012b4b4e4e6, started, ip-172-31-153-216.eu-west-3.compute.internal, https://172.31.153.216:2380, https://172.31.153.216:2379, false
2e39cf93b81fb7ed, started, ip-172-31-153-246.eu-west-3.compute.internal, https://172.31.153.246:2380, https://172.31.153.246:2379, false
6713c8f2e74fb553, started, ip-172-31-153-235.eu-west-3.compute.internal, https://172.31.153.235:2380, https://172.31.153.235:2379, false
```

We are going to need an ID of the dead etcd member. To locate it, please check
(by IP for example, or hostnames) your online instances list and find one from
the list above, that is not online anymore.

Locate and delete dead etcd member:
```bash
etcdctl member remove 6713c8f2e74fb553
Member 6713c8f2e74fb553 removed from cluster 4ec111e0dee094c3
```

Exit the etcd pod shell.

### Recreate failed instance
Assuming you've used Terraform to provision your cloud infrastructure, use
`terraform apply` to restore cloud infrastructure to the declared state, i.e. 3
control-plane instances for HA clusters.

From your local machine:
```bash
terraform apply
```

Save the terraform JSON output to the file:
```bash
terraform output -json > tfout.json
```

### What is Leader instance
In KubeOne the Leader instance is the first instance from the `Hosts` (or
`kubeone_hosts` in Terraform output) list. This instance will be used to "init"
the cluster.

If Leader instance has failed, we can reorder `Hosts` list (or `kubeone_hosts`
in Terraform output), and place a healthy instance as first one. 

Let's review Terraform case.

Locate the `Leader` node.
```bash
jq '.kubeone_hosts.value.control_plane.hostnames' < tfout.json
[
  "ip-172-31-153-222.eu-west-3.compute.internal",
  "ip-172-31-153-228.eu-west-3.compute.internal",
  "ip-172-31-153-246.eu-west-3.compute.internal"
]

jq '.kubeone_hosts.value.control_plane.private_address' < tfout.json
[
  "172.31.153.222",
  "172.31.153.228",
  "172.31.153.246"
]
```

In case when the first instance from those lists has failed, edit JSON file
directly, and reorder it.

Example how to reorder:
```bash
jq '.kubeone_hosts.value.control_plane.hostnames' < tfout.json
[
  "ip-172-31-153-228.eu-west-3.compute.internal",
  "ip-172-31-153-222.eu-west-3.compute.internal",
  "ip-172-31-153-246.eu-west-3.compute.internal"
]

jq '.kubeone_hosts.value.control_plane.private_address' < tfout.json
[
  "172.31.153.228",
  "172.31.153.222",
  "172.31.153.246"
]
```

### Join new control-plane node back to the cluster
`kubeone install` will install kubernetes dependencies to the freshly created
instance and join it back to the cluster as one of the control plane nodes.

```bash
kubeone install -v config.yaml -t tfout.json
```

## Multiple dead control-plane instances
In this case, you'd need to recreate cluster from the backup of etcd.
