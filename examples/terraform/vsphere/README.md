# vSphere Quickstart Terraform configs for Debian-based operating system

The vSphere Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

These Terraform configs are supposed to be used with Debian-based operating
systems that support vApp (e.g. Ubuntu). For more information on how to prepare
a template VM to be used with these configs, check out our [Ubuntu Template VM]
guide.

We also provide Terraform configs for [CentOS-based operating systems](../vsphere_centos)
and [Flatcar Linux](../vsphere_flatcar).

[Ubuntu Template VM]: https://docs.kubermatic.com/kubeone/v1.8/guides/vsphere-template-vm/ubuntu/

## Required environment variables

* `VSPHERE_USER`
* `VSPHERE_PASSWORD`
* `VSPHERE_SERVER`
* `VSPHERE_ALLOW_UNVERIFIED_SSL`

## How to prepare a template

See <https://github.com/kubermatic/machine-controller/blob/main/docs/vsphere.md>

## Networking

KubeOne supports both IPv4 only and IPv4+IPv6 dual-stack clusters. With the given terraform configuration it is assumed that DHCP is supported and enabled on the network. If you want to use static IPs, you need to modify the terraform configuration accordingly.

Based on IP assignment in terraform, we assume that the first IP would be IPv4 and the second IP would be IPv6. As for IP assignment, the first address or the default address is the first IPv4 address that is reachable through the default gateway configured on the machine, then the first reachable IPv6 address, and then the first general discovered address if neither exist.

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/v1.8/examples/ha-load-balancing/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_vsphere"></a> [vsphere](#requirement\_vsphere) | ~> 2.1.1 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_null"></a> [null](#provider\_null) | n/a |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |
| <a name="provider_vsphere"></a> [vsphere](#provider\_vsphere) | ~> 2.1.1 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [null_resource.keepalived_config](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |
| [random_string.keepalived_auth_pass](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/string) | resource |
| [vsphere_compute_cluster_vm_anti_affinity_rule.vm_anti_affinity_rule](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/resources/compute_cluster_vm_anti_affinity_rule) | resource |
| [vsphere_virtual_machine.control_plane](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/resources/virtual_machine) | resource |
| [vsphere_compute_cluster.cluster](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/compute_cluster) | data source |
| [vsphere_datacenter.dc](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/datacenter) | data source |
| [vsphere_datastore.datastore](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/datastore) | data source |
| [vsphere_network.network](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/network) | data source |
| [vsphere_resource_pool.pool](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/resource_pool) | data source |
| [vsphere_virtual_machine.template](https://registry.terraform.io/providers/hashicorp/vsphere/latest/docs/data-sources/virtual_machine) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allow_insecure"></a> [allow\_insecure](#input\_allow\_insecure) | allow insecure https connection to vCenter | `bool` | `false` | no |
| <a name="input_api_vip"></a> [api\_vip](#input\_api\_vip) | virtual IP address for Kubernetes API | `string` | `""` | no |
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_bastion_host"></a> [bastion\_host](#input\_bastion\_host) | ssh jumphost (bastion) hostname | `string` | `""` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_bastion_port"></a> [bastion\_port](#input\_bastion\_port) | ssh jumphost (bastion) port | `number` | `22` | no |
| <a name="input_bastion_username"></a> [bastion\_username](#input\_bastion\_username) | ssh jumphost (bastion) username | `string` | `""` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_compute_cluster_name"></a> [compute\_cluster\_name](#input\_compute\_cluster\_name) | internal vSphere cluster name | `string` | `"cl-1"` | no |
| <a name="input_control_plane_memory"></a> [control\_plane\_memory](#input\_control\_plane\_memory) | memory size of each control plane node in MB | `number` | `2048` | no |
| <a name="input_control_plane_num_cpus"></a> [control\_plane\_num\_cpus](#input\_control\_plane\_num\_cpus) | number of cpus of each control plane node | `number` | `2` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of VMs | `number` | `3` | no |
| <a name="input_datastore_cluster_name"></a> [datastore\_cluster\_name](#input\_datastore\_cluster\_name) | datastore cluster name | `string` | `""` | no |
| <a name="input_datastore_name"></a> [datastore\_name](#input\_datastore\_name) | datastore name | `string` | `"datastore1"` | no |
| <a name="input_dc_name"></a> [dc\_name](#input\_dc\_name) | datacenter name | `string` | `"dc-1"` | no |
| <a name="input_disk_size"></a> [disk\_size](#input\_disk\_size) | disk size | `number` | `50` | no |
| <a name="input_folder_name"></a> [folder\_name](#input\_folder\_name) | folder name | `string` | `"kubeone"` | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified, the default value will be added by machine-controller addon. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | Number of replicas per MachineDeployment | `number` | `2` | no |
| <a name="input_ip_family"></a> [ip\_family](#input\_ip\_family) | IPFamily of the cluster. Defaults to IPv4. | `string` | `"IPv4"` | no |
| <a name="input_is_vsphere_enterprise_plus_license"></a> [is\_vsphere\_enterprise\_plus\_license](#input\_is\_vsphere\_enterprise\_plus\_license) | toggle on/off based on your vsphere enterprise license | `bool` | `true` | no |
| <a name="input_network_name"></a> [network\_name](#input\_network\_name) | network name | `string` | `"public"` | no |
| <a name="input_resource_pool_name"></a> [resource\_pool\_name](#input\_resource\_pool\_name) | cluster resource pool name | `string` | `""` | no |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `"root"` | no |
| <a name="input_template_name"></a> [template\_name](#input\_template\_name) | template name | `string` | `"ubuntu-22.04"` | no |
| <a name="input_vrrp_interface"></a> [vrrp\_interface](#input\_vrrp\_interface) | network interface for API virtual IP | `string` | `"ens192"` | no |
| <a name="input_vrrp_router_id"></a> [vrrp\_router\_id](#input\_vrrp\_router\_id) | vrrp router id for API virtual IP. Must be unique in used subnet | `number` | `42` | no |
| <a name="input_worker_disk"></a> [worker\_disk](#input\_worker\_disk) | disk size of each worker node in GB | `number` | `10` | no |
| <a name="input_worker_memory"></a> [worker\_memory](#input\_worker\_memory) | memory size of each worker node in MB | `number` | `2048` | no |
| <a name="input_worker_num_cpus"></a> [worker\_num\_cpus](#input\_worker\_num\_cpus) | number of cpus of each workers node | `number` | `2` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines | `string` | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
