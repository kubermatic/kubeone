# AWS Quickstart Terraform configs

The AWS Quickstart Terraform configs can be used to create the needed
infrastructure for a Kubernetes HA cluster. Check out the
[Creating Infrastructure guide][docs-infrastructure] to learn more about how to
use the configs and how to provision a Kubernetes cluster using KubeOne.

[docs-infrastructure]: https://docs.kubermatic.com/kubeone/v1.8/guides/using-terraform-configs/

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 4.1.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 4.1.0 |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_default_vpc.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_vpc) | resource |
| [aws_elb.control_plane](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elb) | resource |
| [aws_iam_instance_profile.profile](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_instance_profile) | resource |
| [aws_iam_role.role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_instance.bastion](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance) | resource |
| [aws_instance.control_plane](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance) | resource |
| [aws_instance.static_workers1](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance) | resource |
| [aws_key_pair.deployer](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/key_pair) | resource |
| [aws_security_group.common](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.elb](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.ssh](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group_rule.egress_allow_all](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.ingress_self_allow_all](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.nodeports](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_subnet.public](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [random_integer.cidr_block](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/integer) | resource |
| [aws_ami.ami](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ami) | data source |
| [aws_availability_zones.available](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/availability_zones) | data source |
| [aws_internet_gateway.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/internet_gateway) | data source |
| [aws_vpc.selected](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/vpc) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_ami"></a> [ami](#input\_ami) | AMI ID, use it to fixate control-plane AMI in order to avoid force-recreation it at later times | `string` | `""` | no |
| <a name="input_ami_filters"></a> [ami\_filters](#input\_ami\_filters) | map with AMI filters | <pre>map(object({<br>    owners       = list(string)<br>    image_name   = list(string)<br>    ssh_username = string<br>    worker_os    = string<br>  }))</pre> | <pre>{<br>  "amzn": {<br>    "image_name": [<br>      "amzn2-ami-hvm-2.0.*-x86_64-gp2"<br>    ],<br>    "owners": [<br>      "137112412989"<br>    ],<br>    "ssh_username": "ec2-user",<br>    "worker_os": "amzn2"<br>  },<br>  "centos": {<br>    "image_name": [<br>      "CentOS Linux 7 x86_64*"<br>    ],<br>    "owners": [<br>      "125523088429"<br>    ],<br>    "ssh_username": "centos",<br>    "worker_os": "centos"<br>  },<br>  "flatcar": {<br>    "image_name": [<br>      "Flatcar-stable-*-hvm"<br>    ],<br>    "owners": [<br>      "075585003325"<br>    ],<br>    "ssh_username": "core",<br>    "worker_os": "flatcar"<br>  },<br>  "rhel": {<br>    "image_name": [<br>      "RHEL-8*_HVM-*-x86_64-*"<br>    ],<br>    "owners": [<br>      "309956199498"<br>    ],<br>    "ssh_username": "ec2-user",<br>    "worker_os": "rhel"<br>  },<br>  "rockylinux": {<br>    "image_name": [<br>      "Rocky-8-EC2-*.x86_64"<br>    ],<br>    "owners": [<br>      "792107900819"<br>    ],<br>    "ssh_username": "rocky",<br>    "worker_os": "rockylinux"<br>  },<br>  "ubuntu": {<br>    "image_name": [<br>      "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"<br>    ],<br>    "owners": [<br>      "099720109477"<br>    ],<br>    "ssh_username": "ubuntu",<br>    "worker_os": "ubuntu"<br>  }<br>}</pre> | no |
| <a name="input_apiserver_alternative_names"></a> [apiserver\_alternative\_names](#input\_apiserver\_alternative\_names) | subject alternative names for the API Server signing cert. | `list(string)` | `[]` | no |
| <a name="input_aws_region"></a> [aws\_region](#input\_aws\_region) | AWS region to speak to | `string` | `"eu-west-3"` | no |
| <a name="input_bastion_host_key"></a> [bastion\_host\_key](#input\_bastion\_host\_key) | Bastion SSH host public key | `string` | `null` | no |
| <a name="input_bastion_port"></a> [bastion\_port](#input\_bastion\_port) | Bastion SSH port | `number` | `22` | no |
| <a name="input_bastion_type"></a> [bastion\_type](#input\_bastion\_type) | instance type for bastion | `string` | `"t3.nano"` | no |
| <a name="input_bastion_user"></a> [bastion\_user](#input\_bastion\_user) | Bastion SSH username | `string` | `""` | no |
| <a name="input_cluster_autoscaler_max_replicas"></a> [cluster\_autoscaler\_max\_replicas](#input\_cluster\_autoscaler\_max\_replicas) | maximum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_autoscaler_min_replicas"></a> [cluster\_autoscaler\_min\_replicas](#input\_cluster\_autoscaler\_min\_replicas) | minimum number of replicas per MachineDeployment (requires cluster-autoscaler) | `number` | `0` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster | `string` | n/a | yes |
| <a name="input_control_plane_http_put_max_hops"></a> [control\_plane\_http\_put\_max\_hops](#input\_control\_plane\_http\_put\_max\_hops) | HTTP PUT response hop limit for the instance metadata requests | `number` | `3` | no |
| <a name="input_control_plane_labels"></a> [control\_plane\_labels](#input\_control\_plane\_labels) | custom labels to add (and remove) to control plane | `map(string)` | <pre>{<br>  "custom-label-to-add": "custom-value",<br>  "custom-label-to-remove-": ""<br>}</pre> | no |
| <a name="input_control_plane_type"></a> [control\_plane\_type](#input\_control\_plane\_type) | AWS instance type | `string` | `"t3.medium"` | no |
| <a name="input_control_plane_vm_count"></a> [control\_plane\_vm\_count](#input\_control\_plane\_vm\_count) | number of control plane instances | `number` | `3` | no |
| <a name="input_control_plane_volume_size"></a> [control\_plane\_volume\_size](#input\_control\_plane\_volume\_size) | Size of the EBS volume, in Gb | `number` | `100` | no |
| <a name="input_disable_kubeapi_loadbalancer"></a> [disable\_kubeapi\_loadbalancer](#input\_disable\_kubeapi\_loadbalancer) | E2E tests specific variable to disable usage of any loadbalancer in front of kubeapi-server | `bool` | `false` | no |
| <a name="input_initial_machinedeployment_operating_system_profile"></a> [initial\_machinedeployment\_operating\_system\_profile](#input\_initial\_machinedeployment\_operating\_system\_profile) | Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.<br>If not specified default is used based on the OS specified for workers. | `string` | `""` | no |
| <a name="input_initial_machinedeployment_replicas"></a> [initial\_machinedeployment\_replicas](#input\_initial\_machinedeployment\_replicas) | number of replicas per MachineDeployment | `number` | `1` | no |
| <a name="input_initial_machinedeployment_spotinstances_max_price"></a> [initial\_machinedeployment\_spotinstances\_max\_price](#input\_initial\_machinedeployment\_spotinstances\_max\_price) | used to specify max spot instance price for initial machine-deployment | `number` | `0` | no |
| <a name="input_internal_api_lb"></a> [internal\_api\_lb](#input\_internal\_api\_lb) | make kubernetes API loadbalancer internal (reachible only from inside the VPC) | `bool` | `false` | no |
| <a name="input_os"></a> [os](#input\_os) | Operating System to use in AMI filtering and MachineDeployment | `string` | `"ubuntu"` | no |
| <a name="input_provisioning_utility"></a> [provisioning\_utility](#input\_provisioning\_utility) | provisioning utility to be used for Flatcar worker nodes | `string` | `""` | no |
| <a name="input_ssh_agent_socket"></a> [ssh\_agent\_socket](#input\_ssh\_agent\_socket) | SSH Agent socket, default to grab from $SSH\_AUTH\_SOCK | `string` | `"env:SSH_AUTH_SOCK"` | no |
| <a name="input_ssh_hosts_keys"></a> [ssh\_hosts\_keys](#input\_ssh\_hosts\_keys) | A list of SSH hosts public keys to verify | `list(string)` | `null` | no |
| <a name="input_ssh_port"></a> [ssh\_port](#input\_ssh\_port) | SSH port to be used to provision instances | `number` | `22` | no |
| <a name="input_ssh_private_key_file"></a> [ssh\_private\_key\_file](#input\_ssh\_private\_key\_file) | SSH private key file used to access instances | `string` | `""` | no |
| <a name="input_ssh_public_key_file"></a> [ssh\_public\_key\_file](#input\_ssh\_public\_key\_file) | SSH public key file | `string` | `"~/.ssh/id_rsa.pub"` | no |
| <a name="input_ssh_username"></a> [ssh\_username](#input\_ssh\_username) | SSH user, used only in output | `string` | `""` | no |
| <a name="input_static_workers_count"></a> [static\_workers\_count](#input\_static\_workers\_count) | number of static workers | `number` | `0` | no |
| <a name="input_static_workers_http_put_max_hops"></a> [static\_workers\_http\_put\_max\_hops](#input\_static\_workers\_http\_put\_max\_hops) | HTTP PUT response hop limit for the instance metadata requests | `number` | `3` | no |
| <a name="input_subnets_cidr"></a> [subnets\_cidr](#input\_subnets\_cidr) | CIDR mask bits per subnet | `number` | `24` | no |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | VPC to use ('default' for default VPC) | `string` | `"default"` | no |
| <a name="input_worker_deploy_ssh_key"></a> [worker\_deploy\_ssh\_key](#input\_worker\_deploy\_ssh\_key) | add provided ssh public key to MachineDeployments | `bool` | `true` | no |
| <a name="input_worker_os"></a> [worker\_os](#input\_worker\_os) | OS to run on worker machines, default to var.os | `string` | `""` | no |
| <a name="input_worker_type"></a> [worker\_type](#input\_worker\_type) | instance type for workers | `string` | `"t3.medium"` | no |
| <a name="input_worker_volume_size"></a> [worker\_volume\_size](#input\_worker\_volume\_size) | Size of the EBS volume, in Gb | `number` | `50` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_kubeone_api"></a> [kubeone\_api](#output\_kubeone\_api) | kube-apiserver LB endpoint |
| <a name="output_kubeone_hosts"></a> [kubeone\_hosts](#output\_kubeone\_hosts) | Control plane endpoints to SSH to |
| <a name="output_kubeone_static_workers"></a> [kubeone\_static\_workers](#output\_kubeone\_static\_workers) | Static worker config |
| <a name="output_kubeone_workers"></a> [kubeone\_workers](#output\_kubeone\_workers) | Workers definitions, that will be transformed into MachineDeployment object |
| <a name="output_ssh_commands"></a> [ssh\_commands](#output\_ssh\_commands) | n/a |
