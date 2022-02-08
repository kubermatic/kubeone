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
  description = "OS to run on worker machines"

  # valid choices are:
  # * ubuntu
  # * centos
  default = "ubuntu"
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

variable "control_plane_flavor" {
  default     = "m1.small"
  description = "OpenStack instance flavor for the control plane nodes"
  type        = string
}

variable "worker_flavor" {
  default     = "m1.small"
  description = "OpenStack instance flavor for the worker nodes"
  type        = string
}

variable "lb_flavor" {
  default     = "m1.tiny"
  description = "OpenStack instance flavor for the LoadBalancer node"
  type        = string
}

variable "image" {
  default     = ""
  description = "image name to use"
  type        = string
}

variable "image_properties_query" {
  default = {
    os_distro  = "ubuntu"
    os_version = "20.04"
  }
  description = "in absense of var.image, this will be used to query API for the image"
  type        = map(any)
}

variable "subnet_cidr" {
  default     = "192.168.1.0/24"
  description = "OpenStack subnet cidr"
  type        = string
}

variable "external_network_name" {
  description = "OpenStack external network name"
  type        = string
}

variable "subnet_dns_servers" {
  type    = list(string)
  default = ["8.8.8.8", "8.8.4.4"]
}
