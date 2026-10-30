/*
Copyright 2026 The KubeOne Authors.

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
  description = "common name for the resources, use the cluster name to match the default load balancer name of KubeOne"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$", var.name))
    error_message = "Value of name should be lowercase and can only contain alphanumeric characters and hyphens(-)."
  }
}

variable "image" {
  description = "image name to use"
  default     = ""
  type        = string
}

variable "image_properties_query" {
  description = "in absence of var.image, this will be used to query API for the image"
  default = {
    os_distro  = "ubuntu"
    os_version = "24.04"
  }
  type = map(any)
}

# Network Variables
variable "subnet_cidr" {
  default     = "192.168.1.0/24"
  description = "OpenStack subnet cidr"
  type        = string

  validation {
    condition     = can(cidrhost(var.subnet_cidr, 32))
    error_message = "Must be valid IPv4 CIDR."
  }
}

variable "external_network_name" {
  description = "OpenStack external network name"
  type        = string
}

variable "subnet_dns_servers" {
  type    = list(string)
  default = ["8.8.8.8", "8.8.4.4"]

  validation {
    condition = alltrue([
      for a in var.subnet_dns_servers : can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", a))
    ])
    error_message = "All elements must be a valid IPv4 address."
  }
}

# Bastion Variables
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

variable "bastion_flavor" {
  description = "OpenStack instance flavor for the bastion host"
  default     = "m1.tiny"
  type        = string
}

# SSH Variables
variable "ssh_public_key_file" {
  description = "SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
  type        = string
}

variable "config_drive" {
  description = "If enabled, metadata for machines is stored on a configuration drive instead of the metadata service."
  default     = false
  type        = bool
}
