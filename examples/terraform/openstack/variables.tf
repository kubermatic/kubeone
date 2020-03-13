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
  description = "OS to run on worker machines"

  # valid choices are:
  # * ubuntu
  # * centos
  # * coreos
  default = "ubuntu"
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

variable "control_plane_flavor" {
  default     = "m1.small"
  description = "OpenStack instance flavor for the control plane nodes"
}

variable "worker_flavor" {
  default     = "m1.small"
  description = "OpenStack instance flavor for the worker nodes"
}

variable "lb_flavor" {
  default     = "m1.tiny"
  description = "OpenStack instance flavor for the LoadBalancer node"
}

variable "image" {
  default     = ""
  description = "image name to use"
}

variable "image_properties_query" {
  default = {
    os_distro  = "ubuntu"
    os_version = "18.04"
  }
  description = "in absense of var.image, this will be used to query API for the image"
}

variable "subnet_cidr" {
  default     = "192.168.1.0/24"
  description = "OpenStack subnet cidr"
}

variable "external_network_name" {
  description = "OpenStack external network name"
}

variable "subnet_dns_servers" {
  type    = list(string)
  default = ["8.8.8.8", "8.8.4.4"]
}

