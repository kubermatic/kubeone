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

# Cluster Variables
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

variable "image" {
  description = "image name to use"
  default     = ""
  type        = string
}

variable "image_properties_query" {
  description = "in absence of var.image, this will be used to query API for the image"
  default = {
    os_distro  = "ubuntu"
    os_version = "22.04"
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

# Controlplane Variables
variable "control_plane_vm_count" {
  description = "number of control plane instances"
  default     = 3
  type        = number
}

variable "control_plane_flavor" {
  description = "OpenStack instance flavor for the control plane nodes"
  default     = "m1.small"
  type        = string
}

# Worker Node Variables
variable "worker_os" {
  description = "OS to run on worker machines"
  default     = "ubuntu"
  type        = string
}

variable "worker_flavor" {
  description = "OpenStack instance flavor for the worker nodes"
  default     = "m1.small"
  type        = string
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

variable "bastion_host_key" {
  description = "Bastion SSH host public key"
  default     = null
  type        = string
}

variable "bastion_flavor" {
  description = "OpenStack instance flavor for the LoadBalancer node"
  default     = "m1.tiny"
  type        = string
}

# SSH Variables
variable "ssh_private_key_file" {
  description = "SSH private key file used to access instances"
  default     = ""
  type        = string
}

variable "ssh_public_key_file" {
  description = "SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
  type        = string
}

variable "ssh_hosts_keys" {
  description = "A list of SSH hosts public keys to verify"
  default     = null
  type        = list(string)
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

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
  type        = string
}

# Machine Deployment Variables
variable "initial_machinedeployment_replicas" {
  description = "Number of replicas per MachineDeployment"
  default     = 2
  type        = number
}

variable "cluster_autoscaler_min_replicas" {
  default     = 0
  description = "minimum number of replicas per MachineDeployment (requires cluster-autoscaler)"
  type        = number
}

variable "cluster_autoscaler_max_replicas" {
  default     = 0
  description = "maximum number of replicas per MachineDeployment (requires cluster-autoscaler)"
  type        = number
}

variable "initial_machinedeployment_operating_system_profile" {
  default     = ""
  type        = string
  description = <<EOF
Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.
If not specified, the default value will be added by machine-controller addon.
EOF
}
