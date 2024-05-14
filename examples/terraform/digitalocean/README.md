# DigitalOcean Quickstart Terraform configs

The DigitalOcean Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_digitalocean"></a> [digitalocean](#requirement\_digitalocean) | ~> 2.9.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_digitalocean"></a> [digitalocean](#provider\_digitalocean) | ~> 2.9.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [digitalocean_droplet.control_plane](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/droplet) | resource |
| [digitalocean_loadbalancer.control_plane](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/loadbalancer) | resource |
| [digitalocean_ssh_key.deployer](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/ssh_key) | resource |
| [digitalocean_tag.kube_cluster_tag](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/tag) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_droplet_image"></a> [control\_plane\_droplet\_image](#input\_control\_plane\_droplet\_image) | Image to use for provisioning control plane droplets | `string` | `""` | no |
| <a name="input_control_plane_size"></a> [control\_plane\_size](#input\_control\_plane\_size) | Size of control plane nodes | `string` | `"s-2vcpu-4gb"` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of control plane instances | `number` | `3` | no |
| <a name="input_disable_kubeapi_loadbalancer"></a> [disable\_kubeapi\_loadbalancer](#input\_disable\_kubeapi\_loadbalancer) | E2E tests specific variable to disable usage of any loadbalancer in front of kubeapi-server | `bool` | `false` | no |
| <a name="input_image_references"></a> [image\_references](#input\_image\_references) | map with images | <pre>map(object({<br>    image_name   = string<br>    ssh_username = string<br>    worker_os    = string<br>  }))</pre> | <pre>{<br>  "centos": {<br>    "image_name": "centos-7-x64",<br>    "ssh_username": "root",<br>    "worker_os": "centos"<br>  },<br>  "rockylinux": {<br>    "image_name": "rockylinux-8-x64",<br>    "ssh_username": "root",<br>    "worker_os": "rockylinux"<br>  },<br>  "ubuntu": {<br>    "image_name": "ubuntu-22-04-x64",<br>    "ssh_username": "root",<br>    "worker_os": "ubuntu"<br>  }<br>}</pre> | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | Number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_os"></a> [os](#input\_os) | Operating System to use in image filtering and MachineDeployment | `string` | `"ubuntu"` | no |
| <a name="input_region"></a> [region](#input\_region) | Region to speak to | `string` | `"fra1"` | no |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `""` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `""` | no |
| <a name="input_worker_size"></a> [worker\_size](#input\_worker\_size) | Size of worker nodes | `string` | `"s-2vcpu-4gb"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
