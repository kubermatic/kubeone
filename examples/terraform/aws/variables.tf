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
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$", var.cluster_name))
    error_message = "Value of cluster_name should be lowercase and can only contain alphanumeric characters and hyphens(-)."
  }
}

variable "apiserver_alternative_names" {
  description = "subject alternative names for the API Server signing cert."
  default     = []
  type        = list(string)
}

variable "worker_os" {
  description = "OS to run on worker machines, default to var.os"

  # valid choices are:
  # * ubuntu
  # * centos
  # * flatcar
  # * rhel
  # * amzn2
  default = ""
  type    = string
}

variable "ssh_public_key_file" {
  description = "SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
  type        = string
}

variable "ssh_port" {
  description = "SSH port to be used to provision instances"
  default     = 22
  type        = number
}

variable "ssh_username" {
  description = "SSH user, used only in output"
  default     = "ubuntu"
  type        = string
}

variable "ssh_private_key_file" {
  description = "SSH private key file used to access instances"
  default     = ""
  type        = string
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
  type        = string
}

variable "bastion_port" {
  description = "Bastion SSH port"
  default     = 22
  type        = number
}

variable "bastion_user" {
  description = "Bastion SSH username"
  default     = "ubuntu"
  type        = string
}

# Provider specific settings

variable "aws_region" {
  default     = "eu-west-3"
  description = "AWS region to speak to"
  type        = string

}

variable "vpc_id" {
  default     = "default"
  description = "VPC to use ('default' for default VPC)"
  type        = string
}

variable "control_plane_type" {
  default     = "t3.medium"
  description = "AWS instance type"
}

variable "control_plane_volume_size" {
  default     = 100
  description = "Size of the EBS volume, in Gb"
  type        = number
}

variable "worker_type" {
  default     = "t3.medium"
  description = "instance type for workers"
  type        = string
}

variable "bastion_type" {
  default     = "t3.nano"
  description = "instance type for bastion"
  type        = string
}

variable "os" {
  description = "Operating System to use in AMI filtering and MachineDeployment"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rhel
  # * flatcar
  # * amzn2
  default = "ubuntu"
  type    = string
}

variable "ami" {
  description = "AMI ID, use it to fixate control-plane AMI in order to avoid force-recreation it at later times"
  default     = ""
  type        = string
}

variable "ami_filters" {
  description = "map with AMI filters"
  type = map(object({
    owners     = list(string)
    image_name = list(string)
    osp_name   = string
  }))
  default = {
    ubuntu = {
      owners     = ["099720109477"] # Canonical
      image_name = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
      osp_name   = "osp-ubuntu"
    }

    centos = {
      owners     = ["792107900819"] # RockyLinux
      image_name = ["Rocky-8-ec2-*.x86_64"]
      osp_name   = "osp-centos"
    }

    flatcar = {
      owners     = ["075585003325"] # Kinvolk
      image_name = ["Flatcar-stable-*-hvm"]
      osp_name   = "osp-flatcar"
    }

    rhel = {
      owners     = ["309956199498"] # Red Hat
      image_name = ["RHEL-8*_HVM-*-x86_64-*"]
      osp_name   = "osp-rhel"
    }

    amzn2 = {
      owners     = ["137112412989"] # Amazon
      image_name = ["amzn2-ami-hvm-2.0.*-x86_64-gp2"]
      osp_name   = "osp-amzn2"
    }
  }
}

variable "subnets_cidr" {
  default     = 24
  description = "CIDR mask bits per subnet"
  type        = number
}

variable "internal_api_lb" {
  default     = false
  description = "make kubernetes API loadbalancer internal (reachible only from inside the VPC)"
  type        = bool
}

variable "initial_machinedeployment_replicas" {
  default     = 1
  description = "number of replicas per MachineDeployment"
  type        = number

}

variable "static_workers_count" {
  description = "number of static workers"
  default     = 0
  type        = number
}

variable "initial_machinedeployment_spotinstances" {
  description = "use spot instances for initial machine-deployment"
  default     = false
  type        = bool
}

variable "worker_deploy_ssh_key" {
  description = "add provided ssh public key to MachineDeployments"
  default     = true
  type        = bool
}

variable "control_plane_vm_count" {
  description = "Number of control plane instances"
  default     = 3
  type        = number
}
