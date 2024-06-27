# VMware Cloud Director Quickstart Terraform Configs

The VMware Cloud Director  Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/

## Setup

In this setup, we assume that a dedicated org VDC has been created. It's connected to an external network using an edge gateway. NSX-V is enabled in the infrastructure since the sample configs only support NSX-V, for now.

The kube-apiserver will be assigned the private IP address of the first control plane VM.

### Credentials

Following environment variables or terraform variables can be used to authenticate with the provider:

| Environment Variable | Terraform Variable |
|------|---------|
| VCD_USER | vcd.user |
| VCD_PASSWORD | vcd.password |
| VCD_ORG | vcd.org |
| VCD_URL | vcd.url |

To use access token instead of username and password:


| Environment Variable | Terraform Variable |
|------|---------|
| VCD_AUTH_TYPE | vcd.auth_type |
| VCD_API_TOKEN | vcd.api_token |

#### References

- <https://registry.terraform.io/providers/vmware/vcd/latest/docs#connecting-as-sys-admin-with-default-org-and-vdc>
- <https://registry.terraform.io/providers/vmware/vcd/latest/docs#argument-reference>
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_vcd"></a> [vcd](#requirement\_vcd) | 3.8.2 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_vcd"></a> [vcd](#provider\_vcd) | 3.8.2 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [vcd_network_routed.network](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/network_routed) | resource |
| [vcd_nsxv_dnat.rule_ssh_bastion](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/nsxv_dnat) | resource |
| [vcd_nsxv_firewall_rule.rule_internet](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/nsxv_firewall_rule) | resource |
| [vcd_nsxv_firewall_rule.rule_ssh_bastion](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/nsxv_firewall_rule) | resource |
| [vcd_nsxv_snat.rule_internal](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/nsxv_snat) | resource |
| [vcd_nsxv_snat.rule_internet](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/nsxv_snat) | resource |
| [vcd_vapp.cluster](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/vapp) | resource |
| [vcd_vapp_org_network.network](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/vapp_org_network) | resource |
| [vcd_vapp_vm.bastion](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/vapp_vm) | resource |
| [vcd_vapp_vm.control_plane](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/resources/vapp_vm) | resource |
| [vcd_catalog.catalog](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/data-sources/catalog) | data source |
| [vcd_catalog_vapp_template.vapp_template](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/data-sources/catalog_vapp_template) | data source |
| [vcd_edgegateway.edge_gateway](https://registry.terraform.io/providers/vmware/vcd/3.8.2/docs/data-sources/edgegateway) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allow_insecure"></a> [allow\_insecure](#input\_allow\_insecure) | allow insecure https connection to VMware Cloud Director API | `bool` | `false` | no |
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | Subject alternative names for the API Server signing certificate | `list(string)` | `[]` | no |
| <a name="input_bastion_host_cpu_cores"></a> [bastion\_host\_cpu\_cores](#input\_bastion\_host\_cpu\_cores) | Number of cores per socket for the bastion host VM | `number` | `1` | no |
| <a name="input_bastion_host_cpus"></a> [bastion\_host\_cpus](#input\_bastion\_host\_cpus) | Number of CPUs for the bastion host VM | `number` | `1` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_bastion_host_memory"></a> [bastion\_host\_memory](#input\_bastion\_host\_memory) | Memory size of the bastion host in MB | `number` | `2048` | no |
| <a name="input_bastion_host_ssh_port"></a> [bastion\_host\_ssh\_port](#input\_bastion\_host\_ssh\_port) | SSH port to be used for the bastion host | `number` | `22` | no |
| <a name="input_catalog_name"></a> [catalog\_name](#input\_catalog\_name) | Name of catalog that contains vApp templates | `string` | n/a | yes |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_cpu_cores"></a> [control\_plane\_cpu\_cores](#input\_control\_plane\_cpu\_cores) | Number of cores per socket for the control plane VMs | `number` | `1` | no |
| <a name="input_control_plane_cpus"></a> [control\_plane\_cpus](#input\_control\_plane\_cpus) | Number of CPUs for the control plane VMs | `number` | `2` | no |
| <a name="input_control_plane_disk_size"></a> [control\_plane\_disk\_size](#input\_control\_plane\_disk\_size) | Disk size in MB | `number` | `25600` | no |
| <a name="input_control_plane_disk_storage_profile"></a> [control\_plane\_disk\_storage\_profile](#input\_control\_plane\_disk\_storage\_profile) | Name of storage profile to use for disks | `string` | `""` | no |
| <a name="input_control_plane_memory"></a> [control\_plane\_memory](#input\_control\_plane\_memory) | Memory size of each control plane node in MB | `number` | `4096` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of control plane instances | `number` | `3` | no |
| <a name="input_dhcp_end_address"></a> [dhcp\_end\_address](#input\_dhcp\_end\_address) | Last address for the DHCP IP Pool range | `string` | `"192.168.1.50"` | no |
| <a name="input_dhcp_start_address"></a> [dhcp\_start\_address](#input\_dhcp\_start\_address) | Starting address for the DHCP IP Pool range | `string` | `"192.168.1.2"` | no |
| <a name="input_enable_bastion_host"></a> [enable\_bastion\_host](#input\_enable\_bastion\_host) | Enable bastion host | `bool` | `false` | no |
| <a name="input_external_network_ip"></a> [external\_network\_ip](#input\_external\_network\_ip) | IP address to which source addresses (the virtual machines) on outbound packets are translated to when they send traffic to the external network.<br>Defaults to default external network IP for the edge gateway. | `string` | `""` | no |
| <a name="input_external_network_name"></a> [external\_network\_name](#input\_external\_network\_name) | Name of the external network to be used to send traffic to the external networks. Defaults to edge gateway's default external network. | `string` | `""` | no |
| <a name="input_gateway_ip"></a> [gateway\_ip](#input\_gateway\_ip) | Gateway IP for the routed network | `string` | `"192.168.1.1"` | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_kubeapi_hostname"></a> [kubeapi\_hostname](#input\_kubeapi\_hostname) | DNS name for the kube-apiserver | `string` | `""` | no |
| <a name="input_logging"></a> [logging](#input\_logging) | Enable logging of VMware Cloud Director API activities into go-vcloud-director.log | `bool` | `false` | no |
| <a name="input_network_dns_server_1"></a> [network\_dns\_server\_1](#input\_network\_dns\_server\_1) | Primary DNS server for the routed network | `string` | `""` | no |
| <a name="input_network_dns_server_2"></a> [network\_dns\_server\_2](#input\_network\_dns\_server\_2) | Secondary DNS server for the routed network. | `string` | `""` | no |
| <a name="input_network_interface_type"></a> [network\_interface\_type](#input\_network\_interface\_type) | Type of interface for the routed network | `string` | `"internal"` | no |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `"ubuntu"` | no |
| <a name="input_template_name"></a> [template\_name](#input\_template\_name) | Name of the vApp template to use | `string` | n/a | yes |
| <a name="input_vcd_edge_gateway_name"></a> [vcd\_edge\_gateway\_name](#input\_vcd\_edge\_gateway\_name) | Name of the Edge Gateway | `string` | n/a | yes |
| <a name="input_vcd_org_name"></a> [vcd\_org\_name](#input\_vcd\_org\_name) | Organization name for the VMware Cloud Director setup | `string` | n/a | yes |
| <a name="input_vcd_vdc_name"></a> [vcd\_vdc\_name](#input\_vcd\_vdc\_name) | Virtual datacenter name | `string` | n/a | yes |
| <a name="input_worker_cpu_cores"></a> [worker\_cpu\_cores](#input\_worker\_cpu\_cores) | Number of cores per socket for the worker VMs | `number` | `1` | no |
| <a name="input_worker_cpus"></a> [worker\_cpus](#input\_worker\_cpus) | Number of CPUs for the worker VMs | `number` | `2` | no |
| <a name="input_worker_disk_size_gb"></a> [worker\_disk\_size\_gb](#input\_worker\_disk\_size\_gb) | Disk size for worker VMs in GB | `number` | `25` | no |
| <a name="input_worker_disk_storage_profile"></a> [worker\_disk\_storage\_profile](#input\_worker\_disk\_storage\_profile) | Name of storage profile to use for worker VMs attached disks | `string` | `""` | no |
| <a name="input_worker_memory"></a> [worker\_memory](#input\_worker\_memory) | Memory size of each worker VM in MB | `number` | `4096` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
| <a name="output_ssh_commands"></a> [ssh\_commands](#output\_ssh\_commands) | n/a |
