# Upgrading Kubernetes Cluster Using KubeOne

The cluster is upgraded using the `kubeone upgrade` command invoked with the KubeOne configuration file and optionally Terraform state file.

## Scope of The Upgrade Process

KubeOne takes care of upgrading `kubeadm` and `kubelet` binaries, running `kubeadm upgrade` on all control plane nodes, and optionally upgrading all MachineDeployments objects to the desired Kubernetes version. Upgrades are done in-place, i.e. KubeOne connects to nodes over SSH and runs commands needed to upgrade the node.

**Note:** KubeOne currently doesn't take care of upgrading the Canal CNI plugin and `machine-controller`.

## Prerequisites

KubeOne is doing a set of preflight checks to ensure all prerequisites are satisfied. The following checks are done by KubeOne:

* Docker, Kubelet and Kubeadm are installed,
* information about nodes from the API matches what we have in the KubeOne configuration,
* all nodes are healthy,
* the [Kubernetes version skew policy](https://kubernetes.io/docs/setup/version-skew-policy/) is satisfied.

Once the upgrade process starts for a node, KubeOne applies the `kubeone.io/upgrade-in-progress` label on the node object. This label is used as a lock mechanism, so if upgrade fails or it's already in progress you can't start it again.

It's recommended to backup your cluster before running the upgrade process. You can do it using [Velero](https://github.com/heptio/velero) or any other tool of your choice.

Before running upgrade please ensure that your KubeOne version supports upgrading to the desired Kubernetes version. Check the [Kubernetes Versions Compatibility](https://github.com/kubermatic/kubeone#kubernetes-versions-compatibility) part of the KubeOne's README for more details on supported Kubernetes versions for each KubeOne release. You can check what KubeOne version you're running using the `kubeone version` command.

## Running Upgrades

You need to update the KubeOne configuration file to use the desired Kubernetes version by changing the `versions.Kubernetes` field. KubeOne supports upgrading to the newer minor or patch release. It is possible to upgrade only to the next minor release or to any patch release as long as the minor version is same or the next one.

Everything you need to do is to run the `upgrade` command:

```bash
kubeone upgrade config.yaml
```

If you want to use the Terraform state to source information about the infrastructure, use:

```bash
kubeone upgrade config.yaml --tfjson tf.json
```

KubeOne first runs the preflight checks as described in the prerequisites section and then upgrades control plane nodes one by one. The upgrade process will take some time, usually 5-10 minutes. You should see output such as:

```
time="11:18:38 UTC" level=info msg="Building Kubernetes clientset…"
time="11:18:40 UTC" level=info msg="Determine hostname…"
time="11:18:41 UTC" level=info msg="Determine operating system…"
time="11:18:41 UTC" level=info msg="Checking are all prerequisites installed…" node=35.181.4.44
time="11:18:41 UTC" level=info msg="Checking are all prerequisites installed…" node=35.180.55.129
time="11:18:41 UTC" level=info msg="Checking are all prerequisites installed…" node=35.180.86.102
time="11:18:41 UTC" level=info msg="Running preflight checks…"
time="11:18:41 UTC" level=info msg="Verifying are all nodes running…"
time="11:18:41 UTC" level=info msg="Verifying are correct labels set on nodes…"
time="11:18:41 UTC" level=info msg="Verifying do all node IP addresses match with our state…"
time="11:18:41 UTC" level=info msg="Verifying is it possible to upgrade to the desired version…"
time="11:18:41 UTC" level=info msg="Labeling leader control plane…" node=35.181.4.44
time="11:18:41 UTC" level=info msg="Upgrading kubeadm on leader control plane…" node=35.181.4.44
time="11:18:49 UTC" level=info msg="Generating kubeadm config …" node=35.181.4.44
time="11:18:49 UTC" level=info msg="Uploading kubeadm config to leader control plane node…" node=35.181.4.44
time="11:18:49 UTC" level=info msg="Running 'kubeadm upgrade' on leader control plane node…" node=35.181.4.44
time="11:19:13 UTC" level=info msg="Upgrading kubelet on leader control plane…" node=35.181.4.44
time="11:19:24 UTC" level=info msg="Unlabeling leader control plane…" node=35.181.4.44
time="11:19:24 UTC" level=info msg="Labeling follower control plane…" node=35.180.55.129
time="11:19:24 UTC" level=info msg="Upgrading kubeadm on follower control plane…" node=35.180.55.129
time="11:19:32 UTC" level=info msg="Running 'kubeadm upgrade' on the follower control plane node…" node=35.180.55.129
time="11:19:43 UTC" level=info msg="Upgrading kubelet…" node=35.180.55.129
time="11:19:53 UTC" level=info msg="Unlabeling follower control plane…" node=35.180.55.129
time="11:19:53 UTC" level=info msg="Labeling follower control plane…" node=35.180.86.102
time="11:19:53 UTC" level=info msg="Upgrading kubeadm on follower control plane…" node=35.180.86.102
time="11:20:00 UTC" level=info msg="Running 'kubeadm upgrade' on the follower control plane node…" node=35.180.86.102
time="11:20:08 UTC" level=info msg="Upgrading kubelet…" node=35.180.86.102
time="11:20:19 UTC" level=info msg="Unlabeling follower control plane…" node=35.180.86.102
time="11:20:19 UTC" level=info msg="Upgrade MachineDeployments…"
```

**Note:** By default KubeOne does **not** update the MachineDeployment objects. If you want to update them run the `upgrade` command with the `--upgrade-machine-deployments` flag. This updates all MachineDeployment objects regardless of what's specified in the KubeOne configuration.

If the upgrade process fails, it's recommended to continue manually and resolve errors. In this case the `kubeone.io/upgrade-in-progress` label will prevent you from running KubeOne again but you can ignore it using the `--force` flag.

Optionally, you can now manually upgrade other cluster components such as `machine-controller` or Canal CNI plugin.
