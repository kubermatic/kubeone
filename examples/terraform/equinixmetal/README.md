# Equinix Metal Quickstart Terraform configs

The Equinix Metal Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/master/guides/using_terraform_configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/master/examples/ha_load_balancing/

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| control\_plane\_operating\_system | Image to use for control plane provisioning | string | `"ubuntu_18_04"` | no |
| device\_type | type (size) of the device | string | `"t1.small.x86"` | no |
| facility | Facility (datacenter) | string | `"ams1"` | no |
| lb\_operating\_system | Image to use for loadbalancer provisioning | string | `"ubuntu_18_04"` | no |
| project\_id | project ID | string | n/a | yes |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |
