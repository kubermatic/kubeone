# VMware Cloud Director Quickstart Terraform Configs

The VMware Cloud Director  Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/master/guides/using_terraform_configs/

## Setup

In this setup, we assume that a dedicated org VDC has been created. It's connected to an external network using an edge gateway. NSX-V is enabled in the infrastructure since the sample configs only support NSX-V, for now.

The kube-apiserver will be assigned the private IP address of the first control plane VM.

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_vcd"></a> [vcd](#requirement\_vcd) | ~> 3.6.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_vcd"></a> [vcd](#provider\_vcd) | 3.6.0 |

### Credentials

Following environment variables or terraform variables can be used to authenticate with the provider:

| Environment Variable | Terraform Variable |
|------|---------|
| VCD_USER | vcd.user |
| VCD_PASSWORD | vcd.user |
| VCD_ORG | vcd.org |
| VCD_URL | vcd.url |

#### References

- <https://registry.terraform.io/providers/vmware/vcd/latest/docs#connecting-as-sys-admin-with-default-org-and-vdc>
- <https://registry.terraform.io/providers/vmware/vcd/latest/docs#argument-reference>

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| vcd\_org\_name | Name of the vcd organization | string | n/a | yes |
| vcd\_vdc\_name | Name of the virutal datacenter | string | n/a | yes |
| vcd\_edge\_gateway\_name | Name of the edge gateway defined for the VDC | string | n/a | yes |
| catalog\_name | Name of catalog that contains vApp templates | string | n/a | yes |
| template\_name | Name of the vApp template to use for master VMs | string | n/a | yes |
| external\_network\_name | Name of the network used for external connectivity, defaults to edge gateway's default external network | string | n/a | no |
| external\_network\_ip | SNAT address to allows outbound traffic, defaults to edge gateway's default external network IP | string | n/a | no |
| control\_plane\_memory | Memory size of each control plane node in MB | number | `4096` | no |
| control\_plane\_cpus | Number of CPUs for the control plane VMs | number | `2` | no |
| control\_plane\_cpu\_cores | Number of cores per socket for the control plane VMs | number | `1` | no |
| control\_plane\_disk\_size | Disk size for control plane VMs in MB | number | `25600` | no |
| control\_plane\_disk\_storage_profile | Name of storage profile to use for disks | string | `""` | no |
| network\_interface\_type | Type of interface for the routed network | string | `internal` | no |
| gateway\_ip | Gateway IP for the routed network | string | `192.168.1.1` | no |
| dhcp\_start\_address | Starting address for the DHCP IP Pool range | string | `192.168.1.2` | no |
| dhcp\_end\_address | Last address for the DHCP IP Pool range | string | `192.168.1.50` | no |
| network\_dns\_server\_1 | Primary DNS server for the routed network | string | `""` | no |
| network\_dns\_server\_2 | Secondary DNS server for the routed network | string | `""` | no |
| apiserver\_alternative\_names | Subject alternative names for the API Server signing cert. | list(string) | `[]` | no |
| kubeapi\_hostname | DNS name for the kube-apiserver. | string | `""` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"ubuntu"` | no |
| worker\_os | OS to run on worker machines | string | `ubuntu` | no |
| worker\_memory | Number of replicas per MachineDeployment | number | `1` | no |
| worker\_cpus | Number of CPUs for the worker VMs | number | `2` | no |
| worker\_cpu\_cores | Number of cores per socket for the worker VMs | number | `1` | no |
| worker\_disk\_size | Disk size for worker VMs in MB | number | `25600` | no |
| worker\_disk\_storage\_profile | Name of storage profile to use for worker VMs attached disks | string | `""` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver endpoint, virutal IP of first control plane VM |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
