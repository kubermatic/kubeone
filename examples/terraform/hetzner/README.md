# Hetzner Quickstart Terraform scripts

The Hetzner Quickstart Terraform scripts can be used to create the needed infrastructure for a Kubernetes HA cluster.
Check out the following [Hetzner getting started walkthrough][hetzner-quickstart] to learn more about how to use the
scripts and how to provision a Kubernetes cluster using KubeOne.

## Kubernetes APIserver LoadBalancing
See https://docs.kubermatic.com/kubeone/master/using_kubeone/example_loadbalancer/

[hetzner-quickstart]: https://docs.kubermatic.com/kubeone/master/getting_started/hetzner/

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | prefix for cloud resources | string | n/a | yes |
| control\_plane\_type |  | string | `"cx21"` | no |
| datacenter |  | string | `"fsn1"` | no |
| image |  | string | `"ubuntu-18.04"` | no |
| lb\_type |  | string | `"cx11"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |
| worker\_type |  | string | `"cx21"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |
