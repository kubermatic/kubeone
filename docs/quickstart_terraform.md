# How To manage the infrastructure with Terraform

In this quick start we're going to show how to use our provided terraform plans to setup infrastructure that may be UserGuidefor KubeOne

## Plan selection and configuration

<details><summary>AWS</summary><p>

The Terraform scripts for AWS are located in the [examples/terraform/aws`](../examples/terraform/aws) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/aws
```

You may want to configure the provisioning process by setting variables defining the cluster name, AWS region, instances size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/aws/variables.tf) file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `aws_region` (default: eu-west-3)
* `ssh_public_key_file` (default: `~/.ssh/id_rsa.pub`) - path to your SSH public key that's deployed on instances
* `control_plane_type` (default: t3.medium) - note that you should have at least 2 GB RAM and 2 CPUs for Kubernetes to work properly

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"

aws_region = "us-east-1"
```

</p></details><details><summary>Azure</summary><p>

The Terraform scripts for Azure are located in the [examples/terraform/azure](../examples/terraform/azure) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/azure
```

You may want to configure the provisioning process by setting variables defining the cluster name, image to be used, instance size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/azure/variables.tf) file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `location` (optional) - Azure datacenter, default westeurope
* `worker_vm_size` (optional) - VM Size for worker machines, default Standard_B2s

The `terraform.tfvars` file can look like:

```
cluster_name   = "demo"
worker_vm_size = "Standard_D4s_v3"
```

</p></details><details><summary>Digital Ocean</summary><p>

The Terraform scripts for DigitalOcean are located in the [examples/terraform/digitalocean`](../examples/terraform/digitalocean) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/digitalocean
```

You may want to configure the provisioning process by setting variables defining the cluster name, Droplets region, size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/digitalocean/variables.tf) file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `region` (default: fra1)
* `ssh_public_key_file` (default: `~/.ssh/id_rsa.pub`) - path to your SSH public key that's deployed on instances
* `droplet_size` (default: s-2vcpu-4gb) - note that you should have at least 2 GB RAM and 2 CPUs for Kubernetes to work properly

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"
region = "fra1"
```

</p></details><details><summary>GCE</summary><p>

The example terraform configuration for GCE is located in the [examples/terraform/gce`](../examples/terraform/gce) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/gce
```

You may want to configure the provisioning process by setting variables defining the cluster name, AWS region, instances size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/gce/variables.tf) file. You should consider setting:

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

Now that you configured Terraform you can use the `plan` command to see what changes will be made:

</p></details><details><summary>Hetzner</summary><p>

The Terraform scripts for Hetzner are located in the [examples/terraform/hetzner](../examples/terraform/hetzner) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/hetzner
```

You may want to configure the provisioning process by setting variables defining the cluster name, control plane count and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/hetzner/variables.tf) file. You should consider setting `cluster_name` which is a prefix for cloud resources and required.

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"
```

</p></details><details><summary>OpenStack</summary><p>

**Note:** As not all OpenStack providers have Load Balancers as a Service (LBaaS), the example Terraform scripts will create an instance for a Load Balancer and setup it using [GoBetween](https://github.com/yyyar/gobetween). This setup may not be appropriate for the production usage, but it allows us to provide better HA experience in an easy to consume manner.

The Terraform scripts for OpenStack are located in the [examples/terraform/openstack](../examples/terraform/openstack) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/openstack
```

You may want to configure the provisioning process by setting variables defining the cluster name, image to be used, instance size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/openstack/variables.tf) file. You should consider setting:

* `cluster_name` (required) - prefix for cloud resources
* `image` - (default: Ubuntu 18.04 LTS) image to be used for provisioning instances
* `ssh_public_key_file` (default: `~/.ssh/id_rsa.pub`) - path to your SSH public key that's deployed on instances
* `control_plane_flavor` (default: m1.small) - instance size of control plane nodes
* `worker_flavor` (default: m1.small) - instance size of worker nodes

The `terraform.tfvars` file can look like:

```
cluster_name = "demo"

ssh_public_key_file = "~/.ssh/openstack_rsa.pub"
```

</p></details><details><summary>Packet</summary><p>

**Note:** As Packet doesn't have Load Balancers as a Service (LBaaS), the example Terraform scripts will create an instance for a Load Balancer and setup it using [GoBetween](https://github.com/yyyar/gobetween). This setup may not be appropriate for the production usage, but it allows us to provide better HA experience in an easy to consume manner.


The Terraform scripts for Packet are located in the [examples/terraform/packet](../examples/terraform/packet) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd examples/terraform/packet
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/packet/variables.tf) file. You should consider setting:

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

</p></details><details><summary>vSphere</summary><p>

The Terraform scripts for vSphere are located in the [examples/terraform/vsphere](../examples/terraform/vsphere) directory. First, we need to switch to the directory with Terraform scripts:

```bash
cd ./examples/terraform/vsphere
```

You may want to configure the provisioning process by setting variables defining the cluster name, image to be used, instance size and similar. The easiest way is to create the `terraform.tfvars` file and store variables there. This file is automatically read by Terraform.

```bash
nano terraform.tfvars
```

For the list of available settings along with their names please see the [`variables.tf`](../examples/terraform/vsphere/variables.tf) file. You should consider setting:

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

</p></details>

## Plan usage

### Initialization

Before we can use Terraform to create the infrastructure for us Terraform needs to download the vSphere plugin and setup it's environment. This is done by running the `init` command:

```bash
terraform init
```

**Note:** You need to run this command only the first time before using scripts.

### Creation

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


### Destruction

Once it's done you can proceed and destroy the vSphere infrastructure using Terraform:

```bash
terraform destroy
```

You'll be asked to enter `yes` to confirm your intention to destroy the cluster.
