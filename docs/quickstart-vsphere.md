# How To Install Kubernetes On vSphere Cluster Using KubeOne

In this quick start we're going to show how to get started with KubeOne on
vSphere. We'll cover how to create the needed infrastructure using our example
Terraform scripts and then install Kubernetes. Finally, we're going to show how
to destroy the cluster along with the infrastructure.

As a result, you'll get Kubernetes 1.16.1 High-Available (HA) clusters with
three control plane nodes and two worker nodes.

### Prerequisites

To follow this quick start, you'll need:

* `kubeone` v0.10.0 or newer installed, which can be done by following the `Installing KubeOne` section of [the README](https://github.com/kubermatic/kubeone/blob/master/README.md),
* `terraform` v0.12.0 or later installed. Older releases are not compatible. The binaries for `terraform` can be found on the [Terraform website](https://www.terraform.io/downloads.html)

## Setting Up Credentials

In order for Terraform to successfully create the infrastructure and for KubeOne
to install Kubernetes and create worker nodes you need to setup credentials for
your vSphere cluster.

For the terraform reference please take a look at [vSphere provider docs][3]

The following environment variables should be set:

```bash
export VSPHERE_ALLOW_UNVERIFIED_SSL=false
export VSPHERE_SERVER=<YOUR VCENTER ENDPOINT>
export VSPHERE_USER=<USER>
export VSPHERE_PASSWORD=<PASSWORD>
```

**Note:** The credentials are deployed to the cluster to be used by
`machine-controller` for creating worker nodes.

## Creating Infrastructure

KubeOne is based on the Bring-Your-Own-Infra approach, which means that you have
to provide machines and needed resources yourself. To make this task easier we
are providing Terraform scripts that you can use to get started. You're free to
use your own scripts or any other preferred approach.

The Terraform scripts for vSphere are located in the
[`./examples/terraform/vsphere`][4] directory.

**Note:** KubeOne comes with Terraform integration that is capable of reading
information about the infrastructure from Terraform output. If you decide not to
use our Terraform scripts but want to use Terraform integration, make sure
variable names in the output match variable names used by KubeOne.
Alternatively, if you decide not to use Terraform, you can provide needed
information about the infrastructure manually in the KubeOne configuration file.

**Note:** As vSphere does not have Load Balancers as a Service (LBaaS), the
example Terraform scripts will create an instance for a Load Balancer and setup
it using [GoBetween][5]. This setup may not be appropriate for the production
usage, but it allows us to provide better HA experience in an easy to consume
manner.

First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/vsphere
```

Before we can use Terraform to create the infrastructure for us Terraform needs
to download the vSphere plugin and setup it's environment. This is done by
running the `init` command:

```bash
terraform init
```

**Note:** You need to run this command only the first time before using scripts.

You may want to configure the provisioning process by setting variables defining
the cluster name, image to be used, instance size and similar. The easiest way
is to create the `terraform.tfvars` file and store variables there. This file is
automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the
[`variables.tf`][6] file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `dc_name` (optional, default = "dc-1") - datacenter name
* `compute_cluster_name` (optional, default = "cl-1") - internal vSphere cluster name
* `datastore_name` (optional, default = "datastore1") - vSphere datastore name
* `network_name` (optional, default = "public") - vSphere network name
* `template_name` (optional, default = "ubuntu-18.04") - vSphere template name to clone VMs from

The `terraform.tfvars` file can look like:

```
cluster_name   = "demo"
datastore_name = "exsi-nas"
network_name   = "NAT Network"
template_name  = "kubeone-ubuntu-18.04"
ssh_username   = "ubuntu"
```

Now that you configured Terraform you can use the `plan` command to see what
changes will be made:

```bash
terraform plan
```

Finally, if you agree with changes you can proceed and provision the
infrastructure:

```bash
terraform apply
```

Shortly after you'll be asked to enter `yes` to confirm your intention to
provision the infrastructure.

Infrastructure provisioning takes around 5 minutes. Once it's done you need to
create a Terraform state file that is parsed by KubeOne:

```bash
terraform output -json > tf.json
```

## Installing Kubernetes

Now that you have infrastructure you can proceed with installing Kubernetes
using KubeOne.

Before you start you'll need a configuration file that defines how Kubernetes
will be installed, e.g. what version will be used and what features will be
enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes
1.16.1 and create one worker node. KubeOne automatically populates information
about template, VM size and networking settings for worker nodes from the
Terraform output. Alternatively, you can set those information manually. As
KubeOne is using [Kubermatic `machine-controller`][7] for creating worker nodes,
see [vSphere example manifest][8] for available options.

For vSphere you also need to provide a `cloud-config` file containing
credentials, so vSphere Cloud Controller Manager works as expected. Make sure
to replace sample values with real values.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
versions:
  kubernetes: '1.16.1'
cloudProvider:
  name: 'vsphere'
  cloudConfig: |
    [Global]
    user = "<USERNAME>"
    password = "<PASSWORD>"
    port = "443"
    insecure-flag = "0"

    [VirtualCenter "1.1.1.1"]

    [Workspace]
    server = "1.1.1.1"
    datacenter = "dc-1"
    default-datastore="exsi-nas"
    resourcepool-path="kubeone"
    folder = "kubeone"

    [Disk]
    scsicontrollertype = pvscsi

    [Network]
    public-network = "NAT Network"  
```

Finally, we're going to install Kubernetes by using the `install` command and
providing the configuration file and the Terraform output:

```bash
kubeone install config.yaml --tfjson tf.json
```

The installation process takes some time, usually 5-10 minutes. The output
should look like the following one:

```
$ kubeone install config.yaml -t tf.json
INFO[13:15:31 EEST] Installing prerequisites…
INFO[13:15:32 EEST] Determine operating system…                   node=192.168.11.142
INFO[13:15:33 EEST] Determine operating system…                   node=192.168.11.139
INFO[13:15:34 EEST] Determine hostname…                           node=192.168.11.142
INFO[13:15:34 EEST] Creating environment file…                    node=192.168.11.142
INFO[13:15:34 EEST] Installing kubeadm…                           node=192.168.11.142 os=ubuntu
INFO[13:15:34 EEST] Determine operating system…                   node=192.168.11.140
INFO[13:15:36 EEST] Determine hostname…                           node=192.168.11.139
INFO[13:15:36 EEST] Creating environment file…                    node=192.168.11.139
INFO[13:15:36 EEST] Installing kubeadm…                           node=192.168.11.139 os=ubuntu
INFO[13:15:36 EEST] Determine hostname…                           node=192.168.11.140
INFO[13:15:36 EEST] Creating environment file…                    node=192.168.11.140
INFO[13:15:37 EEST] Installing kubeadm…                           node=192.168.11.140 os=ubuntu
INFO[13:16:45 EEST] Deploying configuration files…                node=192.168.11.139 os=ubuntu
INFO[13:16:45 EEST] Deploying configuration files…                node=192.168.11.140 os=ubuntu
INFO[13:17:03 EEST] Deploying configuration files…                node=192.168.11.142 os=ubuntu
INFO[13:17:04 EEST] Generating kubeadm config file…
INFO[13:17:06 EEST] Configuring certs and etcd on first controller…
INFO[13:17:06 EEST] Ensuring Certificates…                        node=192.168.11.139
INFO[13:17:14 EEST] Downloading PKI files…                        node=192.168.11.139
INFO[13:17:16 EEST] Creating local backup…                        node=192.168.11.139
INFO[13:17:16 EEST] Deploying PKI…
INFO[13:17:16 EEST] Uploading files…                              node=192.168.11.142
INFO[13:17:16 EEST] Uploading files…                              node=192.168.11.140
INFO[13:17:21 EEST] Configuring certs and etcd on consecutive controller…
INFO[13:17:21 EEST] Ensuring Certificates…                        node=192.168.11.142
INFO[13:17:21 EEST] Ensuring Certificates…                        node=192.168.11.140
INFO[13:17:27 EEST] Initializing Kubernetes on leader…
INFO[13:17:27 EEST] Running kubeadm…                              node=192.168.11.139
INFO[13:18:45 EEST] Joining controlplane node…
INFO[13:18:45 EEST] Waiting 30s to ensure main control plane components are up…  node=192.168.11.140
INFO[13:20:27 EEST] Waiting 30s to ensure main control plane components are up…  node=192.168.11.142
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.140
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.139
INFO[13:22:03 EEST] Copying Kubeconfig to home directory…         node=192.168.11.142
INFO[13:22:10 EEST] Building Kubernetes clientset…
INFO[13:22:16 EEST] Creating credentials secret…
INFO[13:22:16 EEST] Applying canal CNI plugin…
INFO[13:22:21 EEST] Installing machine-controller…
INFO[13:22:27 EEST] Installing machine-controller webhooks…
INFO[13:22:30 EEST] Waiting for machine-controller to come up…
INFO[13:23:15 EEST] Creating worker machines…
```

KubeOne automatically downloads the Kubeconfig file for the cluster. It's named
as **\<cluster_name>-kubeconfig**, where **\<cluster_name>** is the name from
your configuration. You can use it with kubectl such as

```bash
kubectl --kubeconfig=<cluster_name>-kubeconfig
```

or export the `KUBECONFIG` variable environment variable:
```bash
export KUBECONFIG=$PWD/<cluster_name>-kubeconfig
```

## Scaling Worker Nodes

As worker nodes are managed by machine-controller, they can be scaled up and down
(including to 0) using Kubernetes API.

```bash
kubectl --namespace kube-system scale machinedeployment/pool1-deployment --replicas=3
```

**Note:** The `kubectl scale` command is not working as expected with `kubectl` 1.15,
returning an error such as:
```
The machinedeployments "pool1" is invalid: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
```
For a workaround, please follow the steps described in the [issue 593][scale_issue].

## Deleting The Cluster

Before deleting a cluster you should clean up all MachineDeployments, so all
worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson tf.json
```

This command will wait for all worker nodes to be gone. Once it's done you can
proceed and destroy the vSphere infrastructure using Terraform:

```bash
terraform destroy
```

You'll be asked to enter `yes` to confirm your intention to destroy the cluster.

Congratulations! You're now running Kubernetes 1.16.1 HA cluster with three
control plane nodes and two worker nodes. If you want to learn more about
KubeOne and its features, such as [upgrades](upgrading_cluster.md), make sure to
check our [documentation][9].

[1]: https://github.com/kubermatic/kubeone/blob/master/README.md
[2]: https://www.terraform.io/downloads.html
[3]: https://www.terraform.io/docs/providers/vsphere/index.html#argument-reference
[4]: https://github.com/kubermatic/kubeone/tree/master/examples/terraform/vsphere
[5]: https://github.com/yyyar/gobetween
[6]: https://github.com/kubermatic/kubeone/blob/master/examples/terraform/vsphere/variables.tf
[7]: https://github.com/kubermatic/machine-controller
[8]: https://github.com/kubermatic/machine-controller/blob/master/examples/vsphere-machinedeployment.yaml
[9]: https://github.com/kubermatic/kubeone/tree/master/docs
[scale_issue]: https://github.com/kubermatic/kubeone/issues/593#issuecomment-513282468
