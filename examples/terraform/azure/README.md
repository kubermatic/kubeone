# Azure Quickstart Terraform configs

The Azure Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | ~> 3.10.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) | ~> 3.10.0 |
| <a name="provider_time"></a> [time](#provider\_time) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [azurerm_availability_set.avset](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/availability_set) | resource |
| [azurerm_lb.lb](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/lb) | resource |
| [azurerm_lb_backend_address_pool.backend_pool](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/lb_backend_address_pool) | resource |
| [azurerm_lb_probe.lb_probe](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/lb_probe) | resource |
| [azurerm_lb_rule.lb_rule](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/lb_rule) | resource |
| [azurerm_network_interface.control_plane](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_interface) | resource |
| [azurerm_network_interface_backend_address_pool_association.control_plane](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_interface_backend_address_pool_association) | resource |
| [azurerm_network_security_group.sg](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_security_group) | resource |
| [azurerm_public_ip.control_plane](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/public_ip) | resource |
| [azurerm_public_ip.lbip](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/public_ip) | resource |
| [azurerm_resource_group.rg](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/resource_group) | resource |
| [azurerm_route_table.rt](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/route_table) | resource |
| [azurerm_subnet.subnet](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet_network_security_group_association.sg_subnet](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet_network_security_group_association) | resource |
| [azurerm_virtual_machine.control_plane](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_machine) | resource |
| [azurerm_virtual_network.vpc](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network) | resource |
| [time_sleep.wait_30_seconds](https://registry.terraform.io/providers/hashicorp/time/latest/docs/resources/sleep) | resource |
| [azurerm_public_ip.control_plane](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/public_ip) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | Number of control plane instances | `number` | `3` | no |
| <a name="input_control_plane_vm_size"></a> [control\_plane\_vm\_size](#input\_control\_plane\_vm\_size) | VM Size for control plane machines | `string` | `"Standard_F2"` | no |
| <a name="input_disable_auto_update"></a> [disable\_auto\_update](#input\_disable\_auto\_update) | Disable automatic flatcar updates (and reboot) | `bool` | `false` | no |
| <a name="input_disable_kubeapi_loadbalancer"></a> [disable\_kubeapi\_loadbalancer](#input\_disable\_kubeapi\_loadbalancer) | E2E tests specific variable to disable usage of any loadbalancer in front of kubeapi-server | `bool` | `false` | no |
| <a name="input_image_references"></a> [image\_references](#input\_image\_references) | map with image references used for control plane | <pre>map(object({<br>    image = object({<br>      publisher = string<br>      offer     = string<br>      sku       = string<br>      version   = string<br>    })<br>    plan = list(object({<br>      name      = string<br>      publisher = string<br>      product   = string<br>    }))<br>    ssh_username = string<br>    worker_os    = string<br>  }))</pre> | <pre>{<br>  "centos": {<br>    "image": {<br>      "offer": "CentOS",<br>      "publisher": "OpenLogic",<br>      "sku": "7_9",<br>      "version": "latest"<br>    },<br>    "plan": [],<br>    "ssh_username": "centos",<br>    "worker_os": "centos"<br>  },<br>  "flatcar": {<br>    "image": {<br>      "offer": "flatcar-container-linux",<br>      "publisher": "kinvolk",<br>      "sku": "stable",<br>      "version": "3815.2.0"<br>    },<br>    "plan": [<br>      {<br>        "name": "stable",<br>        "product": "flatcar-container-linux",<br>        "publisher": "kinvolk"<br>      }<br>    ],<br>    "ssh_username": "core",<br>    "worker_os": "flatcar"<br>  },<br>  "rhel": {<br>    "image": {<br>      "offer": "rhel-byos",<br>      "publisher": "RedHat",<br>      "sku": "rhel-lvm85",<br>      "version": "8.5.20220316"<br>    },<br>    "plan": [<br>      {<br>        "name": "rhel-lvm85",<br>        "product": "rhel-byos",<br>        "publisher": "redhat"<br>      }<br>    ],<br>    "ssh_username": "rhel-user",<br>    "worker_os": "rhel"<br>  },<br>  "rockylinux": {<br>    "image": {<br>      "offer": "rocky-linux-8-5",<br>      "publisher": "procomputers",<br>      "sku": "rocky-linux-8-5",<br>      "version": "8.5.20211118"<br>    },<br>    "plan": [<br>      {<br>        "name": "rocky-linux-8-5",<br>        "product": "rocky-linux-8-5",<br>        "publisher": "procomputers"<br>      }<br>    ],<br>    "ssh_username": "rocky",<br>    "worker_os": "rockylinux"<br>  },<br>  "ubuntu": {<br>    "image": {<br>      "offer": "0001-com-ubuntu-server-jammy",<br>      "publisher": "Canonical",<br>      "sku": "22_04-lts",<br>      "version": "latest"<br>    },<br>    "plan": [],<br>    "ssh_username": "ubuntu",<br>    "worker_os": "ubuntu"<br>  }<br>}</pre> | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | Number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_location"></a> [location](#input\_location) | Azure datacenter to use | `string` | `"westeurope"` | no |
| <a name="input_os"></a> [os](#input\_os) | Operating System to use for finding image reference and in MachineDeployment | `string` | `"ubuntu"` | no |
| <a name="input_rhsm_offline_token"></a> [rhsm\_offline\_token](#input\_rhsm\_offline\_token) | RHSM offline token | `string` | `""` | no |
| <a name="input_rhsm_password"></a> [rhsm\_password](#input\_rhsm\_password) | RHSM password | `string` | `""` | no |
| <a name="input_rhsm_username"></a> [rhsm\_username](#input\_rhsm\_username) | RHSM username | `string` | `""` | no |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `""` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `""` | no |
| <a name="input_worker_vm_size"></a> [worker\_vm\_size](#input\_worker\_vm\_size) | VM Size for worker machines | `string` | `"Standard_F2"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
