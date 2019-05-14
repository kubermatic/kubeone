# How To Install Kubernetes On Packet Cluster Using KubeOne

In this quick start we're going to show how to get started with KubeOne on
Packet. We'll cover how to create the needed infrastructure using our
example Terraform scripts and then install Kubernetes. Finally, we're going to
show how to destroy the cluster along with the infrastructure.

As a result, you'll get Kubernetes 1.14.1 High-Available (HA) clusters with
three control plane nodes and one worker node (which can be easily scaled).

### Prerequisites

To follow this quick start, you'll need:

* `kubeone` v0.6.2 or newer installed, which can be done by following the
  `Installing KubeOne` section of [the README][main_readme],
* `terraform` installed. The binaries for `terraform` can be found on the
  [Terraform website][terraform_website]

## Setting Up Credentials

In order for Terraform to successfully create the infrastructure and for KubeOne
to install Kubernetes and create worker nodes you need an API Access Token. You
can refer to [the official documentation][packet_support_docs] for guidelines
for generating the token.

Once you have the API access token you need to set the `PACKET_AUTH_TOKEN` and
`PACKET_PROJECT_ID` environment variables:

```bash
export PACKET_AUTH_TOKEN=<api key>
export PACKET_PROJECT_ID=<project UUID>
```

**Note:** The API access token is deployed to the cluster to be used by
`machine-controller` for creating worker nodes.

## Creating Infrastructure

KubeOne is based on the Bring-Your-Own-Infra approach, which means that you have
to provide machines and needed resources yourself. To make this task easier we
are providing Terraform scripts that you can use to get started. You're free to
use your own scripts or any other preferred approach.

The Terraform scripts for Packet are located in the
[`./examples/terraform/packet`][packet_terraform] directory.

**Note:** KubeOne comes with Terraform integration that is capable of reading
information about the infrastructure from Terraform output. If you decide not to
use our Terraform scripts but want to use Terraform integration, make sure
variable names in the output match variable names used by KubeOne.
Alternatively, if you decide not to use Terraform, you can provide needed
information about the infrastructure manually in the KubeOne configuration file.

**Note:** As Packet doesn't have Load Balancers as a Service (LBaaS), the example
Terraform scripts will create an instance for a Load Balancer and setup it using
[GoBetween](https://github.com/yyyar/gobetween). This setup may not be appropriate
for the production usage, but it allows us to provide better HA experience in an
easy to consume manner.

First, we need to switch to the directory with Terraform scripts:

```bash
cd examples/terraform/packet
```

Before we can use Terraform to create the infrastructure for us, Terraform needs
to download the Packet provider and setup it's environment. This is done by
running the `init` command:

```bash
terraform init
```

**Note:** You need to run this command only the first time before using scripts.

You may want to configure the provisioning process by setting variables defining
the cluster name, device type, facility and similar. The easiest way is to
create the `terraform.tfvars` file and store variables there. This file is
automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the
[`variables.tf`][packet_variables] file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `project_id` (required) - packet project UUID
* `ssh_public_key_file` (default: `~/.ssh/id_rsa.pub`) - path to your SSH public
  key that's deployed on instances
* `device_type` (default: t1.small.x86) — type of the instance

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"
project_id = "<PROJECT-UUID>"
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

Infrastructure provisioning takes around 3 minutes. Once it's done you need to
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
1.14.1, create 1 worker node and deploy the 
[external cloud controller manager][packet_ccm]. The external cloud controller
manager takes care of providing correct information about nodes from the Packet
API. KubeOne automatically populates information about the worker nodes from the
[Terraform output][packet_tf_output].

Alternatively, you can set those information manually. As KubeOne is using
Kubermatic [`machine-controller`][machine_controller] for creating worker nodes,
see [Packet example manifest][packet_mc_example] for available options.

```yaml
apiVersion: kubeone.io/v1alpha1
kind: KubeOneCluster
name: demo

versions:
  kubernetes: "1.14.1"

cloudProvider:
  name: "packet"
  external: true

clusterNetwork:
  podSubnet: "192.168.0.0/16"
  serviceSubnet: "172.16.0.0/12"
```

**Note:** It's important to provide custom `clusterNetwork` settings in order to
avoid colliding with private Packet network (which is `10.0.0.0/8`).

Finally, we're going to install Kubernetes by using the `install` command and
providing the configuration file and the Terraform output:

```bash
kubeone install config.yaml --tfjson tf.json
```

The installation process takes some time, usually 5-10 minutes. The output
should look like the following one:

```
$ kubeone install config.yaml -t tf.json
INFO[14:19:46 EEST] Installing prerequisites…
INFO[14:19:47 EEST] Determine operating system…                   node=147.75.80.241
INFO[14:19:47 EEST] Determine hostname…                           node=147.75.80.241
INFO[14:19:47 EEST] Creating environment file…                    node=147.75.80.241
INFO[14:19:47 EEST] Determine operating system…                   node=147.75.81.119
INFO[14:19:48 EEST] Installing kubeadm…                           node=147.75.80.241 os=ubuntu
INFO[14:19:48 EEST] Determine hostname…                           node=147.75.81.119
INFO[14:19:48 EEST] Creating environment file…                    node=147.75.81.119
INFO[14:19:48 EEST] Determine operating system…                   node=147.75.84.57
INFO[14:19:48 EEST] Installing kubeadm…                           node=147.75.81.119 os=ubuntu
INFO[14:19:49 EEST] Determine hostname…                           node=147.75.84.57
INFO[14:19:49 EEST] Creating environment file…                    node=147.75.84.57
INFO[14:19:49 EEST] Installing kubeadm…                           node=147.75.84.57 os=ubuntu
INFO[14:20:36 EEST] Deploying configuration files…                node=147.75.80.241 os=ubuntu
INFO[14:20:38 EEST] Deploying configuration files…                node=147.75.81.119 os=ubuntu
INFO[14:20:40 EEST] Deploying configuration files…                node=147.75.84.57 os=ubuntu
INFO[14:20:41 EEST] Generating kubeadm config file…
INFO[14:20:42 EEST] Configuring certs and etcd on first controller…
INFO[14:20:42 EEST] Ensuring Certificates…                        node=147.75.80.241
INFO[14:20:54 EEST] Downloading PKI files…                        node=147.75.80.241
INFO[14:20:56 EEST] Creating local backup…                        node=147.75.80.241
INFO[14:20:56 EEST] Deploying PKI…
INFO[14:20:56 EEST] Uploading files…                              node=147.75.81.119
INFO[14:20:56 EEST] Uploading files…                              node=147.75.84.57
INFO[14:21:01 EEST] Configuring certs and etcd on consecutive controller…
INFO[14:21:01 EEST] Ensuring Certificates…                        node=147.75.81.119
INFO[14:21:01 EEST] Ensuring Certificates…                        node=147.75.84.57
INFO[14:21:11 EEST] Initializing Kubernetes on leader…
INFO[14:21:11 EEST] Running kubeadm…                              node=147.75.80.241
INFO[14:22:29 EEST] Joining controlplane node…
INFO[14:22:29 EEST] Waiting 30s to ensure main control plane components are up…  node=147.75.84.57
INFO[14:24:22 EEST] Waiting 30s to ensure main control plane components are up…  node=147.75.81.119
INFO[14:26:21 EEST] Copying Kubeconfig to home directory…         node=147.75.81.119
INFO[14:26:21 EEST] Copying Kubeconfig to home directory…         node=147.75.84.57
INFO[14:26:21 EEST] Copying Kubeconfig to home directory…         node=147.75.80.241
INFO[14:26:22 EEST] Building Kubernetes clientset…
INFO[14:26:26 EEST] Creating credentials secret…
INFO[14:26:26 EEST] Ensure external CCM is up to date
INFO[14:26:27 EEST] Patching coreDNS with uninitialized toleration…
INFO[14:26:27 EEST] Applying canal CNI plugin…
INFO[14:26:31 EEST] Installing machine-controller…
INFO[14:26:35 EEST] Installing machine-controller webhooks…
INFO[14:26:37 EEST] Waiting for machine-controller to come up…
INFO[14:27:17 EEST] Creating worker machines…
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

As worker nodes are managed by MachineController, they can be scaled up and down
(including to 0) using Kubernetes API.

```bash
kubectl --namespace kube-system scale machinedeployment/pool1-deployment --replicas=3
```

## Deleting The Cluster

Before deleting a cluster you should clean up all MachineDeployments, so all
worker nodes are deleted. You can do it with the `kubeone reset` command:

```bash
kubeone reset config.yaml --tfjson tf.json
```

This command will wait for all worker nodes to be gone. Once it's done you can
proceed and destroy the Packet infrastructure using Terraform:

```bash
terraform destroy
```

You'll be asked to enter `yes` to confirm your intention to destroy the cluster.

Congratulations! You're now running Kubernetes 1.14.1 HA cluster with three
control plane nodes and one worker node. If you want to learn more about KubeOne
and its features, such as [upgrades](upgrading_cluster.md), make sure to check
our [documentation][kubeone_docs].

[main_readme]: https://github.com/kubermatic/kubeone/blob/master/README.md
[terraform_website]: https://www.terraform.io/downloads.html
[packet_support_docs]: https://support.packet.com/kb/articles/api-integrations
[packet_terraform]: https://github.com/kubermatic/kubeone/tree/master/examples/terraform/packet
[packet_variables]: https://github.com/kubermatic/kubeone/blob/master/examples/terraform/packet/variables.tf
[packet_ccm]: https://github.com/packethost/packet-ccm
[packet_tf_output]: https://github.com/kubermatic/kubeone/blob/789509f54b3a4aed7b15cd8b27b2e5bb2a4fa6c1/examples/terraform/packet/output.tf
[machine_controller]: https://github.com/kubermatic/machine-controller
[packet_mc_example]: https://github.com/kubermatic/machine-controller/blob/master/examples/packet-machinedeployment.yaml
[kubeone_docs]: https://github.com/kubermatic/kubeone/tree/master/docs
