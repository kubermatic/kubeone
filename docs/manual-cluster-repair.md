# Manual Cluster Repair

## Overview
When one of the control-plane VM instances fails (i.e. VM has failed at
cloud-provider), it's necessary to replace the failed instance with the new one,
as fast as possible to avoid losing etcd quorum and blocking all
`kube-apiserver` operations.

This guide will demonstrate how to restore your cluster to the normal state
(i.e. to have 3 healthy kube-apiservers with etcd instances running).

## Terminology
* _etcd ring_: a group of etcd instances forming a single etcd cluster.
* _etcd member_: a known peer instance (running on the control-plane nodes) of
  etcd inside the etcd ring.
* _leader instance_: a VM instance, where cluster PKI get generated and first
  control-plane components are launched at cluster initialization time.

## Non-goals (Out of the Scope)
* General cluster troubleshooting
* Cluster recovery from the backup
* Cluster migration

## Goals
* Replace missing control-plane node
* Restore etcd ring healthy state (i.e. odd number of healthy etcd members)
* Restore all other control-plane components

## Symptoms
* A control-plane Node has disappeared from the `kubectl get node --selector node-role.kubernetes.io/master` output
* A control-plane VM instance has grave but unknown issues (i.e. hardware
  issues) but it's still in running state
* A control-plane VM instance is in terminated state

## The Recovery Plan
* Remove the malfunctioning VM instance
* Remove the former (now absent) etcd member from the known etcd peers list
* Create a fresh VM replacement
* Join new VM to the cluster as control-plane node

## Remove the malfunctioning VM instance
If the VM instance is not in the appropriate healthy state (i.e. underlying
hardware issues), and/or is unresponsive (for a myriad of reasons), it's often
easier to replace it when trying to fix. So please go on and delete (in cloud
console) the malfunctioning instance if there is still one in the running state.

## Remove the former etcd member from the known etcd peers
Even when one etcd member is physically (and abruptly) removed, etcd ring still
hopes it might come back online at a later time. Unfortunately this is not our
case and we need to let etcd ring know that dead etcd member is gone forever
(i.e. remove dead etcd member from the known peers list).

### Nodes
First of all, check your Nodes
```bash
kubectl get node --selector node-role.kubernetes.io/master -o wide
```

Failed control-plane node will be displayed as NotReady or even absent from the
output (running Cloud Controller Manager will remove Node object eventually).

### etcd
Even when a control-plane node is absent, there are still other alive nodes,
that contain healthy etcd ring members. Exec into the shell of one of those
alive etcd containers:
```bash
kubectl -n kube-system exec -it etcd-<ALIVE-HOSTNAME> sh
```

Setup client TLS authentication in order to be able to communicate with etcd:
```bash
export ETCDCTL_API=3
export ETCDCTL_CACERT=/etc/kubernetes/pki/etcd/ca.crt
export ETCDCTL_CERT=/etc/kubernetes/pki/etcd/healthcheck-client.crt
export ETCDCTL_KEY=/etc/kubernetes/pki/etcd/healthcheck-client.key
```

Retrieve currently known members list:
```bash
etcdctl member list
```

Example output:
```bash
2ce40012b4b4e4e6, started, ip-172-31-153-216.eu-west-3.compute.internal, https://172.31.153.216:2380, https://172.31.153.216:2379, false
2e39cf93b81fb7ed, started, ip-172-31-153-246.eu-west-3.compute.internal, https://172.31.153.246:2380, https://172.31.153.246:2379, false
6713c8f2e74fb553, started, ip-172-31-153-235.eu-west-3.compute.internal, https://172.31.153.235:2380, https://172.31.153.235:2379, false
```

By comparing the Nodes list with etcd members list (hostnames, IPs) we can find
the ID of the missing etcd member (dead etcd member would be missing from the
Nodes list, or will be in NotReady state).

For example it's found that there is not control-plane Node with IP
`172.31.153.235`. It means etcd member ID `6713c8f2e74fb553` is the one we are
looking for to remove.

To remove dead etcd member:
```bash
etcdctl member remove 6713c8f2e74fb553
Member 6713c8f2e74fb553 removed from cluster 4ec111e0dee094c3
```

Now members list should display only 2 members.
```bash
etcdctl member list
```

Example output:
```bash
2ce40012b4b4e4e6, started, ip-172-31-153-216.eu-west-3.compute.internal, https://172.31.153.216:2380, https://172.31.153.216:2379, false
2e39cf93b81fb7ed, started, ip-172-31-153-246.eu-west-3.compute.internal, https://172.31.153.246:2380, https://172.31.153.246:2379, false
```

Exit the shell in etcd pod.

## Create a fresh VM replacement
Assuming you've used Terraform to provision your cloud infrastructure, use
`terraform apply` to restore cloud infrastructure to the declared state, i.e. 3
control-plane instances for HA clusters.

From your local machine:
```bash
terraform apply
```

The result should be: 3 running control-plane VM instances. Two existing and currently members
of the cluster, and the fresh one which will be joined to the cluster as
replacement for failed VM.

## Join new VM to the cluster as control-plane node

### WARNING: It is super important to appoint one of the existing and healthy control-plane nodes as a Leader.
Otherwise, new PKI will be generated and spread across your control-plane
instances, effectively disrupting control-plane kubernetes components.

### What is Leader instance
In KubeOne the Leader instance is the first instance from the `Hosts` in KubeOne
config file (or `kubeone_hosts` in Terraform output) list. This instance will be
used to "init" the cluster initially (generate PKI, start the first etcd
instance, launch different kubernetes control-plane components).

By default, the first `Host` instance defined in KubeOne config (or
`kubeone_hosts` in Terraform output) will be a Leader Host.

It's possible to choose which instance will be a Leader using config or
Terraform output. Please keep in mind, there can be only 1 Leader Host.

#### No Terraform
Example config
```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo-cluster
hosts:
- privateAddress: '172.18.0.1'
  ...
  isLeader: true
```

#### Terraform
Or in terraform `output.tf` file:
```
output "kubeone_hosts" {
  value = {
    control_plane = {
      private_address      = aws_instance.control_plane.*.private_ip
      ...
      leader_ip            = aws_instance.control_plane.0.private_ip
```

It's necessary to rerun `terraform apply` in order to incorporate new changes to
the Terraform state.

### KubeOne Install
`kubeone install` will install kubernetes dependencies to the freshly created
VM instance and join it back to the cluster as one of the control-plane nodes.

#### Terraform
Save terraform output to the JSON file:
```bash
terraform output -json > tfout.json
```

Use new terraform outputs:
```bash
kubeone install -v config.yaml -t tfout.json
```

#### No Terraform
```bash
kubeone install -v config.yaml
```
