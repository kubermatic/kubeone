# Equinix Metal Quickstart Terraform configs

The Equinix Metal Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.5/guides/using-terraform-configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/v1.5/examples/ha-load-balancing/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_metal"></a> [metal](#requirement\_metal) | ~> 3.3.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_metal"></a> [metal](#provider\_metal) | ~> 3.3.0 |
| <a name="provider_null"></a> [null](#provider\_null) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [metal_device.control_plane](https://registry.terraform.io/providers/equinix/metal/latest/docs/resources/device) | resource |
| [metal_device.lb](https://registry.terraform.io/providers/equinix/metal/latest/docs/resources/device) | resource |
| [metal_ssh_key.deployer](https://registry.terraform.io/providers/equinix/metal/latest/docs/resources/ssh_key) | resource |
| [null_resource.lb_config](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_operating_system"></a> [control\_plane\_operating\_system](#input\_control\_plane\_operating\_system) | Image to use for control plane provisioning | `string` | `"ubuntu_18_04"` | no |
| <a name="input_device_type"></a> [device\_type](#input\_device\_type) | type (size) of the device | `string` | `"m3.small.x86"` | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | Number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_lb_device_type"></a> [lb\_device\_type](#input\_lb\_device\_type) | type (size) of the load balancer device | `string` | `"m3.small.x86"` | no |
| <a name="input_lb_operating_system"></a> [lb\_operating\_system](#input\_lb\_operating\_system) | Image to use for loadbalancer provisioning | `string` | `"ubuntu_18_04"` | no |
| <a name="input_metro"></a> [metro](#input\_metro) | Metro area for cluster | `string` | `"AM"` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | project ID | `string` | n/a | yes |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `"root"` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
