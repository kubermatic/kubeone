# Equinix Metal Quickstart Terraform configs

The Equinix Metal Quickstart Terraform configs can be used to create the needed
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
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_operating_system"></a> [control\_plane\_operating\_system](#input\_control\_plane\_operating\_system) | Image to use for control plane provisioning | `string` | `""` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of control plane instances | `number` | `3` | no |
| <a name="input_device_type"></a> [device\_type](#input\_device\_type) | type (size) of the device | `string` | `"m3.small.x86"` | no |
| <a name="input_image_references"></a> [image\_references](#input\_image\_references) | map with images | <pre>map(object({<br>    image_name   = string<br>    ssh_username = string<br>    worker_os    = string<br>  }))</pre> | <pre>{<br>  "centos": {<br>    "image_name": "centos_7",<br>    "ssh_username": "root",<br>    "worker_os": "centos"<br>  },<br>  "flatcar": {<br>    "image_name": "flatcar_stable",<br>    "ssh_username": "core",<br>    "worker_os": "flatcar"<br>  },<br>  "rockylinux": {<br>    "image_name": "rocky_8",<br>    "ssh_username": "root",<br>    "worker_os": "rockylinux"<br>  },<br>  "ubuntu": {<br>    "image_name": "ubuntu_22_04",<br>    "ssh_username": "root",<br>    "worker_os": "ubuntu"<br>  }<br>}</pre> | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | Number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_lb_device_type"></a> [lb\_device\_type](#input\_lb\_device\_type) | type (size) of the load balancer device | `string` | `"m3.small.x86"` | no |
| <a name="input_lb_operating_system"></a> [lb\_operating\_system](#input\_lb\_operating\_system) | Image to use for loadbalancer provisioning | `string` | `"ubuntu_22_04"` | no |
| <a name="input_metro"></a> [metro](#input\_metro) | Metro area for cluster | `string` | `"AM"` | no |
| <a name="input_os"></a> [os](#input\_os) | Operating System to use in image filtering and MachineDeployment | `string` | `"ubuntu"` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | project ID | `string` | n/a | yes |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `""` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
