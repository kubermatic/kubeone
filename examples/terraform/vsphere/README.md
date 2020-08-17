# vSphere Quickstart Terraform configs

The vSphere Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the following
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

## Required environment variables

* `VSPHERE_USER`
* `VSPHERE_PASSWORD`
* `VSPHERE_SERVER`
* `VSPHERE_ALLOW_UNVERIFIED_SSL`

## How to prepare a template

See https://github.com/kubermatic/machine-controller/blob/master/docs/vsphere.md

## Kubernetes API Server Load Balancing

See the [Terraform loadbalancers in examples document][docs-tf-loadbalancer].

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/master/infrastructure/terraform_configs/
[docs-tf-loadbalancer]: https://docs.kubermatic.com/kubeone/master/advanced/example_loadbalancer/

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_name | Name of the cluster | string | n/a | yes |
| compute\_cluster\_name | internal vSphere cluster name | string | `"cl-1"` | no |
| datastore\_name | datastore name | string | `"datastore1"` | no |
| dc\_name | datacenter name | string | `"dc-1"` | no |
| disk\_size | disk size | string | `"50"` | no |
| network\_name | network name | string | `"public"` | no |
| ssh\_agent\_socket | SSH Agent socket, default to grab from $SSH_AUTH_SOCK | string | `"env:SSH_AUTH_SOCK"` | no |
| ssh\_port | SSH port to be used to provision instances | string | `"22"` | no |
| ssh\_private\_key\_file | SSH private key file used to access instances | string | `""` | no |
| ssh\_public\_key\_file | SSH public key file | string | `"~/.ssh/id_rsa.pub"` | no |
| ssh\_username | SSH user, used only in output | string | `"root"` | no |
| template\_name | template name | string | `"ubuntu-18.04"` | no |
| worker\_os | OS to run on worker machines | string | `"ubuntu"` | no |

## Outputs

| Name | Description |
|------|-------------|
| kubeone\_api | kube-apiserver LB endpoint |
| kubeone\_hosts | Control plane endpoints to SSH to |
| kubeone\_workers | Workers definitions, that will be transformed into MachineDeployment object |

