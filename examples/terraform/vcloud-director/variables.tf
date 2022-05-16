/*
Copyright 2022 The KubeOne Authors.

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

# vCloud Director provider configuration
variable "vcd_org_name" {
  description = "Organization name for the vCloud Director setup"
  type        = string
}

variable "vcd_vdc_name" {
  description = "Virtual datacenter name"
  type        = string
}

variable "vcd_edge_gateway_name" {
  description = "Name of the Edge Gateway"
  type        = string
}

# Cluster specific configuration
variable "cluster_name" {
  description = "Name of the cluster"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$", var.cluster_name))
    error_message = "Value of cluster_name should be lowercase and can only contain alphanumeric characters and hyphens(-)."
  }
}

variable "apiserver_alternative_names" {
  description = "Subject alternative names for the API Server signing certificate"
  default     = []
  type        = list(string)
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

variable "catalog_name" {
  description = "Name of catalog that contains vApp templates"
  type        = string
}

variable "template_name" {
  description = "Name of the vApp template to use"
  type        = string
}

variable "control_plane_memory" {
  description = "Memory size of each control plane node in MB"
  default     = 4096
  type        = number
}

variable "control_plane_cpus" {
  description = "Number of CPUs for the control plane VMs"
  default     = 2
  type        = number
}

variable "control_plane_cpu_cores" {
  description = "Number of cores per socket for the control plane VMs"
  default     = 1
  type        = number
}

variable "control_plane_disk_size" {
  description = "Disk size in MB"
  default     = 51200 # 50 GiB
  type        = number
}

variable "control_plane_disk_storage_profile" {
  description = "Name of storage profile to use for disks"
  default     = "Intermediate"
  type        = string
}

variable "external_network_ip" {
  type    = string
  default = ""
  validation {
    condition     = length(var.external_network_ip) == 0 || can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.external_network_ip))
    error_message = "Invalid DNS server provided."
  }
  description = <<EOF
External network IP for bastion host and loadbalancer, allows outbound traffic from the edge gateway.
This should be a public IP from the IP pool assigned to the edge gateway that is connected to the vApp network using a routed network.
Defaults to default/primary external address for the edge gateway.
EOF
}

variable "network_interface_type" {
  description = "Type of interface for the routed network"
  # For NSX-T internal is the only supported value
  default = "internal"
  type    = string
  validation {
    condition     = can(regex("^internal$|^subinterface$|^distributed$", var.network_interface_type))
    error_message = "Invalid network interface type."
  }
}

variable "network_subnet" {
  description = "Subnet for the routed network specified using CIDR notation"
  default     = "192.168.1.0/24"
  type        = string
}

variable "network_dns_server_1" {
  description = "Primary DNS server for the routed network"
  default     = ""
  type        = string
  validation {
    condition     = length(var.network_dns_server_1) == 0 || can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.network_dns_server_1))
    error_message = "Invalid DNS server provided."
  }
}

variable "network_dns_server_2" {
  description = "Secondary DNS server for the routed network."
  default     = ""
  type        = string
  validation {
    condition     = length(var.network_dns_server_2) == 0 || can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.network_dns_server_2))
    error_message = "Invalid DNS server provided."
  }
}