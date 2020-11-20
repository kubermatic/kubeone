/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

variable "cluster_name" {
  description = "Name of the cluster"
}

variable "worker_os" {
  description = "OS to run on worker machines, default to var.os"

  # valid choices are:
  # * ubuntu
  # * centos
  # * flatcar
  # * rhel
  default = ""
}

variable "ssh_public_key_file" {
  description = "SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
}

variable "ssh_port" {
  description = "SSH port to be used to provision instances"
  default     = 22
}

variable "ssh_username" {
  description = "SSH user, used only in output"
  default     = "ubuntu"
}

variable "ssh_private_key_file" {
  description = "SSH private key file used to access instances"
  default     = ""
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
}

variable "bastion_port" {
  description = "Bastion SSH port"
  default     = 22
}

variable "bastion_user" {
  description = "Bastion SSH username"
  default     = "ubuntu"
}

# Provider specific settings

variable "aws_region" {
  default     = "eu-west-3"
  description = "AWS region to speak to"
}

variable "vpc_id" {
  default     = "default"
  description = "VPC to use ('default' for default VPC)"
}

variable "control_plane_type" {
  default     = "t3.medium"
  description = "AWS instance type"
}

variable "control_plane_volume_size" {
  default     = 100
  description = "Size of the EBS volume, in Gb"
}

variable "worker_type" {
  default     = "t3.medium"
  description = "instance type for workers"
}

variable "bastion_type" {
  default     = "t3.nano"
  description = "instance type for bastion"
}

variable "os" {
  description = "Operating System to use in AMI filtering and MachineDeployment"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rhel
  # * flatcar
  default = "ubuntu"
}

variable "ami" {
  description = "AMI ID, use it to fixate control-plane AMI in order to avoid force-recreation it at later times"
  default     = ""
}

variable "ami_filters" {
  description = "map with AMI filters"
  default = {
    ubuntu = {
      owners     = ["099720109477"] # Canonical
      image_name = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
    }

    centos = {
      owners     = ["125523088429"] # CentOS
      image_name = ["CentOS 8.2.* x86_64"]
    }

    flatcar = {
      owners     = ["075585003325"] # Kinvolk
      image_name = ["Flatcar-stable-*-hvm"]
    }

    rhel = {
      owners     = ["309956199498"] # Red Hat
      image_name = ["RHEL-8*_HVM-*-x86_64-*"]
    }

    amazon_linux2 = {
      owners     = ["137112412989"] # Amazon
      image_name = ["amzn2-ami-hvm-2.0.*-x86_64-gp2"]
    }
  }
}

variable "subnets_cidr" {
  default     = 24
  description = "CIDR mask bits per subnet"
}

variable "internal_api_lb" {
  default     = false
  description = "make kubernetes API loadbalancer internal (reachible only from inside the VPC)"
}

variable "open_nodeports" {
  default     = false
  description = "open NodePorts flag"
}

variable "initial_machinedeployment_replicas" {
  default     = 1
  description = "number of replicas per MachineDeployment"
}

variable "static_workers_count" {
  description = "number of static workers"
  default     = 0
}
