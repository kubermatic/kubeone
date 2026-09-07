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

variable "name" {
  description = "common name for the resources"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$", var.name))
    error_message = "Value of name should be lowercase and can only contain alphanumeric characters and hyphens(-)."
  }
}

variable "ssh_port" {
  description = "SSH port to be used to provision instances"
  default     = 22
  type        = number
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

variable "public_cidr" {
  description = "CIDR for public subnet"
  type        = string
}

variable "os" {
  description = "Operating System to use in AMI filtering and MachineDeployment"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rhel
  # * flatcar
  # * rockylinux
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
    owners       = list(string)
    image_name   = list(string)
    ssh_username = string
    worker_os    = string
  }))
  default = {
    ubuntu = {
      owners       = ["099720109477"] # Canonical
      image_name   = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
      ssh_username = "ubuntu"
      worker_os    = "ubuntu"
    }

    centos = {
      owners       = ["125523088429"]
      image_name   = ["CentOS Linux 7 x86_64*"]
      ssh_username = "centos"
      worker_os    = "centos"
    }

    flatcar = {
      owners       = ["075585003325"] # Kinvolk
      image_name   = ["Flatcar-stable-*-hvm"]
      ssh_username = "core"
      worker_os    = "flatcar"
    }

    rhel = {
      owners       = ["309956199498"] # Red Hat
      image_name   = ["RHEL-9*_HVM-*-x86_64-*"]
      ssh_username = "ec2-user"
      worker_os    = "rhel"
    }

    rockylinux = {
      owners       = ["792107900819"] # RockyLinux
      image_name   = ["Rocky-9-EC2-*.x86_64"]
      ssh_username = "rocky"
      worker_os    = "rockylinux"
    }
  }
}
