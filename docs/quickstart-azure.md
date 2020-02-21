# How To Install Kubernetes On Azure cloud Using KubeOne

In this quick start we're going to show how to get started with KubeOne on
Azure. We'll cover how to create the needed infrastructure using our example
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
your Azure cluster.

For the terraform reference please take a look at [Azure provider docs][3]

The following environment variables should be set:

```bash
export ARM_CLIENT_ID=<your client id>
export ARM_CLIENT_SECRET=<your client secret id>
export ARM_TENANT_ID=<your tenant id>
export ARM_SUBSCRIPTION_ID=<your subscribtion id>
```

**Note:** The credentials are deployed to the cluster to be used by
`machine-controller` for creating worker nodes.

## Creating Infrastructure

KubeOne is based on the Bring-Your-Own-Infra approach, which means that you have
to provide machines and needed resources yourself. To make this task easier we
are providing Terraform scripts that you can use to get started. You're free to
use your own scripts or any other preferred approach.

The Terraform scripts for Azure are located in the
[`./examples/terraform/azure`][4] directory.

**Note:** KubeOne comes with Terraform integration that is capable of reading
information about the infrastructure from Terraform output. If you decide not to
use our Terraform scripts but want to use Terraform integration, make sure
variable names in the output match variable names used by KubeOne.
Alternatively, if you decide not to use Terraform, you can provide needed
information about the infrastructure manually in the KubeOne configuration file.

First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/azure
```

Before we can use Terraform to create the infrastructure for us Terraform needs
to download the Azure plugin and setup it's environment. This is done by
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
* `location` (optional) - Azure datacenter, default westeurope
* `worker_vm_size` (optional) - VM Size for worker machines, default Standard_B2s

The `terraform.tfvars` file can look like:

```
cluster_name   = "demo"
worker_vm_size = "Standard_D4s_v3"
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

Infrastructure provisioning takes around 5-10 minutes.

**Note:** To obtain IP addresses (which are a bit delayed) of the VMs, it's
required to run:

```bash
terraform refresh
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
see [Azure example manifest][8] for available options.

For Azure you also need to provide a `cloud-config` file containing credentials,
so Azure Cloud Controller Manager works as expected. Make sure to replace sample
values with real values. For example, to create a cluster with Kubernetes
`1.16.1`, save the following to `config.yaml`:

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
versions:
  kubernetes: '1.16.1'
cloudProvider:
  name: 'azure'
  cloudConfig: |
    {
      "tenantId": "<AZURE TENANT ID>",
      "subscriptionId": "<AZURE SUBSCIBTION ID>",
      "aadClientId": "<AZURE CLIENT ID>",
      "aadClientSecret": "<AZURE CLIENT SECRET>",
      "resourceGroup": "<SOME RESOURCE GROUP>",
      "location": "westeurope",
      "subnetName": "<SOME SUBNET NAME>",
      "routeTableName": "",
      "securityGroupName": "<SOME SECURITY GROUP>",
      "vnetName": "<SOME VIRTUAL NETWORK>",
      "primaryAvailabilitySetName": "<SOME AVAILABILITY SET NAME>",
      "useInstanceMetadata": true,
      "useManagedIdentityExtension": false,
      "userAssignedIdentityID": ""
    }
```

Finally, we're going to install Kubernetes by using the `install` command and
providing the configuration file and the Terraform output:

```bash
kubeone install config.yaml --tfjson <DIR-WITH-tfstate-FILE>
```

**Note:** `--tfjson` accepts a file as well as a directory containing the
terraform state file. To pass a file, generate the JSON output by running the
following and use it as the value for the `--tfjson` flag:
```bash
terraform output -json > tf.json
```

Alternatively, if the terraform state file is in the current working directory
 `--tfjson .` can be used as well.

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

Worker nodes are managed by the machine-controller. It creates initially only one and can be
scaled up and down (including to 0) using the Kubernetes API. To do so you first got to retrieve
the `machinedeployments` by

```bash
kubectl get machinedeployments -n kube-system
```

The names of the `machinedeployments` are generated. You can scale the workers in those via

```bash
kubectl --namespace kube-system scale machinedeployment/<MACHINE-DEPLOYMENT-NAME> --replicas=3
```

**Note:** The `kubectl scale` command is not working as expected with `kubectl` 1.15,
returning an error such as:

```
The machinedeployments "<MACHINE-DEPLOYMENT-NAME>" is invalid: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
```

For a workaround, please follow the steps described in the [issue 593][scale_issue] or upgrade to `kubectl` 1.16 or newer.

## Deleting The Cluster

Before deleting a cluster you should clean up all MachineDeployments, so all
worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson <DIR-WITH-tfstate-FILE>
```

This command will wait for all worker nodes to be gone. Once it's done you can
proceed and destroy the Azure infrastructure using Terraform:

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
[3]: https://www.terraform.io/docs/providers/azurerm/index.html#argument-reference
[4]: https://github.com/kubermatic/kubeone/tree/master/examples/terraform/azure
[6]: https://github.com/kubermatic/kubeone/blob/master/examples/terraform/azure/variables.tf
[7]: https://github.com/kubermatic/machine-controller
[8]: https://github.com/kubermatic/machine-controller/blob/master/examples/azure-machinedeployment.yaml
[9]: https://github.com/kubermatic/kubeone/tree/master/docs
[scale_issue]: https://github.com/kubermatic/kubeone/issues/593#issuecomment-513282468
