# Nutanix Quickstart Terraform configs

The Nutanix Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/v1.8/examples/ha-load-balancing/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_nutanix"></a> [nutanix](#requirement\_nutanix) | 1.2.2 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_null"></a> [null](#provider\_null) | n/a |
| <a name="provider_nutanix"></a> [nutanix](#provider\_nutanix) | 1.2.2 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [null_resource.lb_config](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |
| [nutanix_category_key.category_key](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/resources/category_key) | resource |
| [nutanix_category_value.category_value](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/resources/category_value) | resource |
| [nutanix_virtual_machine.control_plane](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/resources/virtual_machine) | resource |
| [nutanix_virtual_machine.lb](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/resources/virtual_machine) | resource |
| [nutanix_cluster.cluster](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/data-sources/cluster) | data source |
| [nutanix_image.image](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/data-sources/image) | data source |
| [nutanix_project.project](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/data-sources/project) | data source |
| [nutanix_subnet.subnet](https://registry.terraform.io/providers/nutanix/nutanix/1.2.2/docs/data-sources/subnet) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_bastion_disk_size"></a> [bastion\_disk\_size](#input\_bastion\_disk\_size) | Disk size, in Mib, for bastion/LB node | `number` | `102400` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_bastion_memory_size"></a> [bastion\_memory\_size](#input\_bastion\_memory\_size) | Memory size, in Mib, for bastion/LB node | `number` | `4096` | no |
| <a name="input_bastion_port"></a> [bastion\_port](#input\_bastion\_port) | Bastion SSH port | `number` | `22` | no |
| <a name="input_bastion_sockets"></a> [bastion\_sockets](#input\_bastion\_sockets) | Number of sockets for bastion/LB node | `number` | `1` | no |
| <a name="input_bastion_user"></a> [bastion\_user](#input\_bastion\_user) | Bastion SSH username | `string` | `"ubuntu"` | no |
| <a name="input_bastion_vcpus"></a> [bastion\_vcpus](#input\_bastion\_vcpus) | Number of vCPUs per socket for bastion/LB node | `number` | `1` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_disk_size"></a> [control\_plane\_disk\_size](#input\_control\_plane\_disk\_size) | Disk size size, in Mib, for control plane nodes | `number` | `102400` | no |
| <a name="input_control_plane_memory_size"></a> [control\_plane\_memory\_size](#input\_control\_plane\_memory\_size) | Memory size, in Mib, for control plane nodes | `number` | `4096` | no |
| <a name="input_control_plane_sockets"></a> [control\_plane\_sockets](#input\_control\_plane\_sockets) | Number of sockets for control plane nodes | `number` | `1` | no |
| <a name="input_control_plane_vcpus"></a> [control\_plane\_vcpus](#input\_control\_plane\_vcpus) | Number of vCPUs per socket for control plane nodes | `number` | `2` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of control plane instances | `number` | `3` | no |
| <a name="input_image_name"></a> [image\_name](#input\_image\_name) | Image to be used for instances (control plane, bastion/LB, workers) | `string` | n/a | yes |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_nutanix_cluster_name"></a> [nutanix\_cluster\_name](#input\_nutanix\_cluster\_name) | Name of the Nutanix Cluster which will be used for this Kubernetes cluster | `string` | n/a | yes |
| <a name="input_project_name"></a> [project\_name](#input\_project\_name) | Name of the Nutanix Project | `string` | n/a | yes |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `"ubuntu"` | no |
| <a name="input_subnet_name"></a> [subnet\_name](#input\_subnet\_name) | Name of the subnet | `string` | n/a | yes |
| <a name="input_worker_disk_size"></a> [worker\_disk\_size](#input\_worker\_disk\_size) | Disk size size, in Gb, for worker nodes | `number` | `50` | no |
| <a name="input_worker_memory_size"></a> [worker\_memory\_size](#input\_worker\_memory\_size) | Memory size, in Mib, for worker nodes | `number` | `4096` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines, default to var.os | `string` | `"ubuntu"` | no |
| <a name="input_worker_sockets"></a> [worker\_sockets](#input\_worker\_sockets) | Number of sockets for worker nodes | `number` | `1` | no |
| <a name="input_worker_vcpus"></a> [worker\_vcpus](#input\_worker\_vcpus) | Number of vCPUs per socket for worker nodes | `number` | `2` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
| <a name="output_ssh_commands"></a> [ssh\_commands](#output\_ssh\_commands) | n/a |
