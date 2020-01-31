# Manual cluster repair
When one of the control-plane nodes fails, it's necessary to restore it back as
fast as possible to avoid loosing ETCD quorum and blocking all kubeapi-server
operations.

This is a guide on how to restore you cluster back to the normal state.

## One dead control-plane instance
If you are running KubeOne cluster with machine-controller enabled, it will
automatically recreate your failed worker nodes. But it can happen that
cloud-instance with control-plane node fails.

In this document we will cover the case when 1 (out of 3 in total) control-plane
node is lost.

### Remove dead ETCD member
Even when one ETCD member is physically (and abruptly) removed, ETCD-ring still
hopes it might return back online. Unfortunately this is not our case and we
need to let ETCD-ring know that dead ETCD member is dead forever.

Exec into the shell of the alive ETCD container:
```bash
kubectl -n kube-system exec -it ETCD-<ALIVE-HOSTNAME> sh
```

Setup client TLS authentication:
```bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=/etc/kubernetes/pki/ETCD/ca.crt
export ETCDCTL_CERT=/etc/kubernetes/pki/ETCD/healthcheck-client.crt
export ETCDCTL_KEY=/etc/kubernetes/pki/ETCD/healthcheck-client.key
```

Retrieve currently known members list (example output):
```bash
etcdctl member list
2ce40012b4b4e4e6, started, ip-172-31-153-216.eu-west-3.compute.internal, https://172.31.153.216:2380, https://172.31.153.216:2379, false
2e39cf93b81fb7ed, started, ip-172-31-153-246.eu-west-3.compute.internal, https://172.31.153.246:2380, https://172.31.153.246:2379, false
6713c8f2e74fb553, started, ip-172-31-153-235.eu-west-3.compute.internal, https://172.31.153.235:2380, https://172.31.153.235:2379, false
```

We are going to need an ID of the dead ETCD member. To locate it, please check
(by IP for example, or hostnames) your online instances list and find one from
the list above, that is not online anymore.

Locate and delete dead ETCD member:
```bash
etcdctl member remove 6713c8f2e74fb553
Member 6713c8f2e74fb553 removed from cluster 4ec111e0dee094c3
```

Exit the ETCD pod shell.

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

In case when first instance from those lists has failed, edit JSON file directy,
and reorder it.

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

## Mulpiple dead control-plane instances
In this case, you'd need to recreate cluster from the backup of ETCD.
