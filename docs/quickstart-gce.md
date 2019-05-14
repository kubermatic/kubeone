# How To Install Kubernetes On GCE Cluster Using KubeOne

In this quick start we're going to show how to get started with KubeOne on GCE.
We'll cover how to create the needed infrastructure using our example terraform
configuration and then install Kubernetes. Finally, we're going to show how to
destroy the cluster along with the infrastructure.

As a result, you'll get Kubernetes 1.14.1 High-Available (HA) clusters with
three control plane nodes and two worker nodes.

### Prerequisites

To follow this quick start, you'll need:

* `kubeone` v0.6.2 or newer installed, which can be done by following the `Installing KubeOne`
  section of [the
  README](https://github.com/kubermatic/kubeone/blob/master/README.md),
* `terraform` installed. The binaries for `terraform` can be found on the
  [Terraform website](https://www.terraform.io/downloads.html)

## Setting Up Credentials

In order for Terraform to successfully create the infrastructure and for KubeOne
to install Kubernetes and create worker nodes you need an [Service
Account](https://cloud.google.com/iam/docs/creating-managing-service-accounts)
with the appropriate permissions.

Once you have the service account you need to set `GOOGLE_CREDENTIALS`
environment variable:

```bash
export GOOGLE_CREDENTIALS=$(cat path/to/your_service_account.json)
```

**Note:** The credentials are also deployed to the cluster to be used by
`machine-controller` for creating worker nodes.

## Creating Infrastructure

KubeOne is based on the Bring-Your-Own-Infra approach, which means that you have
to provide machines and needed resources yourself. To make this task easier we
are providing Terraform scripts that you can use to get started. You're free to
use your own scripts or any other preferred approach.

The example terraform configuration for GCE is located in the
[`./examples/terraform/gce`](https://github.com/kubermatic/kubeone/tree/master/examples/terraform/gce)
directory.

**Note:** KubeOne comes with Terraform integration that is capable of reading
information about the infrastructure from Terraform output. If you decide not to
use our Terraform scripts but want to use Terraform integration, make sure
variable names in the output match variable names used by KubeOne.
Alternatively, if you decide not to use Terraform, you can provide needed
information about the infrastructure manually in the KubeOne configuration file.

First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/gce
```

Before we can use Terraform to create the infrastructure for us, Terraform needs
to download the AWS plugin and setup it's environment. This is done by running
the `init` command:

```bash
terraform init
```

**Note:** You need to run this command only the first time before using scripts.

You may want to configure the provisioning process by setting variables defining
the cluster name, AWS region, instances size and similar. The easiest way is to
create the `terraform.tfvars` file and store variables there. This file is
automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the
[`variables.tf`](https://github.com/kubermatic/kubeone/blob/master/examples/terraform/gce/variables.tf)
file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `project` (required) — GCP Project ID
* `region` (default: europe-west3)
* `ssh_public_key_file` (default: `~/.ssh/id_rsa.pub`) - path to your SSH public
  key that's deployed on instances
* `control_plane_type` (default: n1-standard-1) - note that you should have at
  least 2 GB RAM and 2 CPUs for Kubernetes to work properly

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"

project = "kubeone-demo-project"

region = "europe-west1"
```

Now that you configured Terraform you can use the `plan` command to see what
changes will be made:

```bash
terraform plan
```

Finally, if you agree with changes you can proceed and provision the
infrastructure:

```bash
terraform apply control_plane_target_pool_members_count=1
```

`control_plane_target_pool_members_count` is needed in order to bootstrap
control plane. Once install is done it's recommended to include all control
plane VMs into the LB (will be covered a bit later in this document).

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
1.14.1 and create 3 worker nodes. KubeOne automatically populates information
about worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/ec8bf305446ac22529e9683fd4ce3c9abf753d1e/examples/terraform/gce/output.tf#L41-L81).
Alternatively, you can set those information manually. As KubeOne is using
[Kubermatic
`machine-controller`](https://github.com/kubermatic/machine-controller) for
creating worker nodes, see [GCE example
manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/gce-machinedeployment.yaml)
for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.1'
cloudProvider:
  name: 'gce'
```

Finally, we're going to install Kubernetes by using the `install` command and
providing the configuration file and the Terraform output:

```bash
kubeone install config.yaml --tfjson tf.json
```

The installation process takes some time, usually 5-10 minutes. The output
should look like the following one:

```
INFO[17:24:41 EET] Installing prerequisites…
INFO[17:24:42 EET] Determine operating system…                   node=35.198.117.209
INFO[17:24:42 EET] Determine operating system…                   node=35.246.186.88
INFO[17:24:42 EET] Determine operating system…                   node=35.198.129.205
INFO[17:24:42 EET] Determine hostname…                           node=35.198.117.209
INFO[17:24:42 EET] Creating environment file…                    node=35.198.117.209
INFO[17:24:42 EET] Installing kubeadm…                           node=35.198.117.209 os=ubuntu
INFO[17:24:43 EET] Deploying configuration files…                node=35.198.117.209 os=ubuntu
INFO[17:24:43 EET] Determine hostname…                           node=35.246.186.88
INFO[17:24:43 EET] Creating environment file…                    node=35.246.186.88
INFO[17:24:43 EET] Installing kubeadm…                           node=35.246.186.88 os=ubuntu
INFO[17:24:43 EET] Determine hostname…                           node=35.198.129.205
INFO[17:24:43 EET] Deploying configuration files…                node=35.246.186.88 os=ubuntu
INFO[17:24:43 EET] Creating environment file…                    node=35.198.129.205
INFO[17:24:43 EET] Installing kubeadm…                           node=35.198.129.205 os=ubuntu
INFO[17:24:43 EET] Deploying configuration files…                node=35.198.129.205 os=ubuntu
INFO[17:24:44 EET] Generating kubeadm config file…
INFO[17:24:45 EET] Configuring certs and etcd on first controller…
INFO[17:24:45 EET] Ensuring Certificates…                        node=35.246.186.88
INFO[17:24:47 EET] Downloading PKI files…                        node=35.246.186.88
INFO[17:24:49 EET] Creating local backup…                        node=35.246.186.88
INFO[17:24:49 EET] Deploying PKI…
INFO[17:24:49 EET] Uploading files…                              node=35.198.117.209
INFO[17:24:49 EET] Uploading files…                              node=35.198.129.205
INFO[17:24:52 EET] Configuring certs and etcd on consecutive controller…
INFO[17:24:52 EET] Ensuring Certificates…                        node=35.198.117.209
INFO[17:24:52 EET] Ensuring Certificates…                        node=35.198.129.205
INFO[17:24:54 EET] Initializing Kubernetes on leader…
INFO[17:24:54 EET] Running kubeadm…                              node=35.246.186.88
INFO[17:25:09 EET] Joining controlplane node…
INFO[17:26:36 EET] Copying Kubeconfig to home directory…         node=35.198.117.209
INFO[17:26:36 EET] Copying Kubeconfig to home directory…         node=35.246.186.88
INFO[17:26:36 EET] Copying Kubeconfig to home directory…         node=35.198.129.205
INFO[17:26:37 EET] Building Kubernetes clientset…
INFO[17:26:39 EET] Applying canal CNI plugin…
INFO[17:26:43 EET] Installing machine-controller…
INFO[17:26:46 EET] Installing machine-controller webhooks…
INFO[17:26:47 EET] Waiting for machine-controller to come up…
INFO[17:27:12 EET] Creating worker machines…
```

Once it's finished in order in include 2 other control plane VMs into the LB:
```bash
terraform apply
```

KubeOne automatically downloads the Kubeconfig file for the cluster. It's named
as `cluster-name-kubeconfig`. You can use it with kubectl such as `kubectl
--kubeconfig cluster-name-kubeconfig` or export the `KUBECONFIG` variable
environment variable:
```bash
export KUBECONFIG=$PWD/cluster-name-kubeconfig
```

## Deleting The Cluster

Before deleting a cluster you should clean up all MachineDeployments, so all
worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson tf.json
```

This command will wait for all worker nodes to be gone. Once it's done you can
proceed and destroy the AWS infrastructure using Terraform:

```bash
terraform destroy
```

You'll be asked to enter `yes` to confirm your intention to destroy the cluster.

Congratulations! You're now running Kubernetes 1.14.1 HA cluster with three
control plane nodes and three worker nodes. If you want to learn more about
KubeOne and its features, such as [upgrades](upgrading_cluster.md), make sure to
check our
[documentation](https://github.com/kubermatic/kubeone/tree/master/docs).
