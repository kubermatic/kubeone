# How To Install Kubernetes On Hetzner Cluster Using KubeOne

In this quick start we're going to show how to get started with KubeOne on Hetzner. We'll cover how to create the needed infrastructure using our example Terraform scripts and then install Kubernetes. Finally, we're going to show how to destroy the cluster along with the infrastructure.

As a result, you'll get Kubernetes 1.14.1 High-Available (HA) clusters with three control plane nodes and three worker nodes.

### Prerequisites

To follow this quick start, you'll need:

* `kubeone` v0.7.0 or newer installed, which can be done by following the `Installing KubeOne` section of [the README](https://github.com/kubermatic/kubeone/blob/master/README.md),
* `terraform` v0.11 installed. The binaries for `terraform` can be found on the [Terraform website](https://www.terraform.io/downloads.html)

**Note:** Due to breaking changes made in Terraform v0.12, it's currently not possible to use example Terraform scripts with Terraform v0.12.

## Setting Up Credentials

In order for Terraform to successfully create the infrastructure and for KubeOne to install Kubernetes and create worker nodes you need a Hetzner API Token.

Once you have the API token you need to set the `HCLOUD_TOKEN` environment variable containing the token:

```bash
export HCLOUD_TOKEN=...
```

**Note:** The API token is deployed to the cluster to be used by `machine-controller` for creating worker nodes.

## Creating Infrastructure

KubeOne is based on the Bring-Your-Own-Infra approach, which means that you have to provide machines and needed resources yourself. To make this task easier we are providing Terraform scripts that you can use to get started. You're free to use your own scripts or any other preferred approach.

The Terraform scripts for Hetzner are located in the [`./examples/terraform/hetzner`](https://github.com/kubermatic/kubeone/tree/master/examples/terraform/hetzner) directory.

**Note:** KubeOne comes with Terraform integration that is capable of reading information about the infrastructure from Terraform output. If you decide not to use our Terraform scripts but want to use Terraform integration, make sure variable names in the output match variable names used by KubeOne. Alternatively, if you decide not to use Terraform, you can provide needed information about the infrastructure manually in the KubeOne configuration file.

First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/hetzner
```

Before we can use Terraform to create the infrastructure for us, Terraform needs to download the Hetzner plugin and setup it's environment. This is done by running the `init` command:

```bash
terraform init
```

**Note:** You need to run this command only the first time before using scripts.

You may want to configure the provisioning process by setting variables defining the cluster name, control plane count and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](https://github.com/kubermatic/kubeone/blob/master/examples/terraform/hetzner/variables.tf) file. You should consider setting `cluster_name` which is a prefix for cloud resources and required.

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"
```

Now that you configured Terraform you can use the `plan` command to see what changes will be made:

```bash
terraform plan
```

Finally, if you agree with changes you can proceed and provision the infrastructure:

```bash
terraform apply
```

Shortly after you'll be asked to enter `yes` to confirm your intention to provision the infrastructure.

Infrastructure provisioning takes around 5 minutes. Once it's done you need to create a Terraform state file that is parsed by KubeOne:

```bash
terraform output -json > tf.json
```

## Installing Kubernetes

Now that you have infrastructure you can proceed with installing Kubernetes
using KubeOne.

Before you start you'll need a configuration file that defines how Kubernetes
will be installed, e.g. what version will be used and what features will be
enabled. For the configuration file reference run `kubeone config print --full`.

To get started you can use the following configuration. It'll install Kubernetes 1.14.1, create 3 worker nodes and deploy the [external cloud controller manager](https://github.com/hetznercloud/hcloud-cloud-controller-manager). The external cloud controller manager takes care of providing correct information about the nodes. KubeOne automatically populates all needed information about worker nodes from the [Terraform output](https://github.com/kubermatic/kubeone/blob/a874fd5913ca2a86c3b8136982c2a00e835c2f62/examples/terraform/hetzner/output.tf#L26-L36). Alternatively, you can set those information manually. As KubeOne is using [Kubermatic `machine-controller`](https://github.com/kubermatic/machine-controller) for creating worker nodes see [Hetzner example manifest](https://github.com/kubermatic/machine-controller/blob/master/examples/hetzner-machinedeployment.yaml) for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo
versions:
  kubernetes: '1.14.1'
cloudProvider:
  name: 'hetzner'
  external: true
```

Finally, we're going to install Kubernetes by using the `install` command and providing the configuration file:

```bash
kubeone install config.yaml --tfjson tf.json
```

The installation process takes some time, usually 5-10 minutes. The output should look like the following one:

```
time="11:59:19 UTC" level=info msg="Installing prerequisites…"
time="11:59:20 UTC" level=info msg="Determine operating system…" node=157.230.114.40
time="11:59:20 UTC" level=info msg="Determine operating system…" node=157.230.114.39
time="11:59:20 UTC" level=info msg="Determine operating system…" node=157.230.114.42
time="11:59:21 UTC" level=info msg="Determine hostname…" node=157.230.114.40
time="11:59:21 UTC" level=info msg="Creating environment file…" node=157.230.114.40
time="11:59:21 UTC" level=info msg="Installing kubeadm…" node=157.230.114.40 os=ubuntu
time="11:59:21 UTC" level=info msg="Determine hostname…" node=157.230.114.39
time="11:59:21 UTC" level=info msg="Creating environment file…" node=157.230.114.39
time="11:59:21 UTC" level=info msg="Installing kubeadm…" node=157.230.114.39 os=ubuntu
time="11:59:22 UTC" level=info msg="Determine hostname…" node=157.230.114.42
time="11:59:22 UTC" level=info msg="Creating environment file…" node=157.230.114.42
time="11:59:22 UTC" level=info msg="Installing kubeadm…" node=157.230.114.42 os=ubuntu
time="11:59:59 UTC" level=info msg="Deploying configuration files…" node=157.230.114.39 os=ubuntu
time="12:00:03 UTC" level=info msg="Deploying configuration files…" node=157.230.114.42 os=ubuntu
time="12:00:04 UTC" level=info msg="Deploying configuration files…" node=157.230.114.40 os=ubuntu
time="12:00:05 UTC" level=info msg="Generating kubeadm config file…"
time="12:00:06 UTC" level=info msg="Configuring certs and etcd on first controller…"
time="12:00:06 UTC" level=info msg="Ensuring Certificates…" node=157.230.114.39
time="12:00:09 UTC" level=info msg="Generating PKI…"
time="12:00:09 UTC" level=info msg="Running kubeadm…" node=157.230.114.39
time="12:00:09 UTC" level=info msg="Downloading PKI files…" node=157.230.114.39
time="12:00:10 UTC" level=info msg="Creating local backup…" node=157.230.114.39
time="12:00:10 UTC" level=info msg="Deploying PKI…"
time="12:00:10 UTC" level=info msg="Uploading files…" node=157.230.114.42
time="12:00:10 UTC" level=info msg="Uploading files…" node=157.230.114.40
time="12:00:13 UTC" level=info msg="Configuring certs and etcd on consecutive controller…"
time="12:00:13 UTC" level=info msg="Ensuring Certificates…" node=157.230.114.40
time="12:00:13 UTC" level=info msg="Ensuring Certificates…" node=157.230.114.42
time="12:00:15 UTC" level=info msg="Initializing Kubernetes on leader…"
time="12:00:15 UTC" level=info msg="Running kubeadm…" node=157.230.114.39
time="12:01:47 UTC" level=info msg="Joining controlplane node…"
time="12:03:01 UTC" level=info msg="Copying Kubeconfig to home directory…" node=157.230.114.39
time="12:03:01 UTC" level=info msg="Copying Kubeconfig to home directory…" node=157.230.114.40
time="12:03:01 UTC" level=info msg="Copying Kubeconfig to home directory…" node=157.230.114.42
time="12:03:03 UTC" level=info msg="Building Kubernetes clientset…"
time="12:03:04 UTC" level=info msg="Applying canal CNI plugin…"
time="12:03:06 UTC" level=info msg="Installing machine-controller…"
time="12:03:28 UTC" level=info msg="Installing machine-controller webhooks…"
time="12:03:28 UTC" level=info msg="Waiting for machine-controller to come up…"
time="12:04:08 UTC" level=info msg="Creating worker machines…"
time="12:04:10 UTC" level=info msg="Skipping Ark deployment because no backup provider was configured."
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

## Deleting The Cluster

Before deleting a cluster you should clean up all MachineDeployments, so all worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson tf.json
```

This command will wait for all worker nodes to be gone. Once it's done you can proceed and destroy the Hetzner infrastructure using Terraform:

```bash
terraform destroy
```

You'll be asked to enter `yes` to confirm your intention to destroy the cluster.

Congratulations! You're now running Kubernetes 1.14.1 HA cluster with three control plane nodes and three worker nodes. If you want to learn more about KubeOne and its features, such as [upgrades](upgrading_cluster.md), make sure to check our [documentation](https://github.com/kubermatic/kubeone/tree/master/docs).
