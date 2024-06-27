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

# VMware Cloud Director provider configuration
variable "vcd_org_name" {
  description = "Organization name for the VMware Cloud Director setup"
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

variable "allow_insecure" {
  description = "allow insecure https connection to VMware Cloud Director API"
  default     = false
  type        = bool
}

variable "logging" {
  description = "Enable logging of VMware Cloud Director API activities into go-vcloud-director.log"
  default     = false
  type        = bool
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

variable "kubeapi_hostname" {
  description = "DNS name for the kube-apiserver"
  default     = ""
  type        = string
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

variable "ssh_hosts_keys" {
  default     = null
  description = "A list of SSH hosts public keys to verify"
  type        = list(string)
}

variable "bastion_host_key" {
  description = "Bastion SSH host public key"
  default     = null
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

variable "control_plane_vm_count" {
  description = "number of control plane instances"
  default     = 3
  type        = number
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
  default     = 25600 # 24 GiB
  type        = number
}

variable "control_plane_disk_storage_profile" {
  description = "Name of storage profile to use for disks"
  default     = ""
  type        = string
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

variable "gateway_ip" {
  description = "Gateway IP for the routed network"
  default     = "192.168.1.1"
  type        = string
}

variable "dhcp_start_address" {
  description = "Starting address for the DHCP IP Pool range"
  default     = "192.168.1.2"
  type        = string

  validation {
    condition     = can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.dhcp_start_address))
    error_message = "Invalid DHCP start address."
  }
}

variable "dhcp_end_address" {
  description = "Last address for the DHCP IP Pool range"
  default     = "192.168.1.50"
  type        = string
  validation {
    condition     = can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.dhcp_end_address))
    error_message = "Invalid DHCP end address."
  }
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

variable "external_network_name" {
  description = "Name of the external network to be used to send traffic to the external networks. Defaults to edge gateway's default external network."
  default     = ""
  type        = string
}

variable "external_network_ip" {
  default = ""
  type    = string
  validation {
    condition     = length(var.external_network_ip) == 0 || can(regex("^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", var.external_network_ip))
    error_message = "Invalid extenral network IP provided."
  }
  description = <<EOF
IP address to which source addresses (the virtual machines) on outbound packets are translated to when they send traffic to the external network.
Defaults to default external network IP for the edge gateway.
EOF
}

variable "initial_machinedeployment_replicas" {
  default     = 2
  description = "number of replicas per MachineDeployment"
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

variable "worker_os" {
  description = "OS to run on worker machines"

  # valid choices are:
  # * ubuntu
  # * flatcar
  default = "ubuntu"
  type    = string
  validation {
    condition     = can(regex("^ubuntu$|^flatcar$", var.worker_os))
    error_message = "Unsupported OS specified for worker machines."
  }
}
variable "worker_memory" {
  description = "Memory size of each worker VM in MB"
  default     = 4096
  type        = number
}

variable "worker_cpus" {
  description = "Number of CPUs for the worker VMs"
  default     = 2
  type        = number
}

variable "worker_cpu_cores" {
  description = "Number of cores per socket for the worker VMs"
  default     = 1
  type        = number
}

variable "worker_disk_size_gb" {
  description = "Disk size for worker VMs in GB"
  default     = 25
  type        = number
}

variable "worker_disk_storage_profile" {
  description = "Name of storage profile to use for worker VMs attached disks"
  default     = ""
  type        = string
}

variable "initial_machinedeployment_operating_system_profile" {
  default     = ""
  type        = string
  description = <<EOF
Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.
If not specified, the default value will be added by machine-controller addon.
EOF
}

variable "enable_bastion_host" {
  description = "Enable bastion host"
  default     = false
  type        = bool
}

variable "bastion_host_memory" {
  description = "Memory size of the bastion host in MB"
  default     = 2048
  type        = number
}

variable "bastion_host_cpus" {
  description = "Number of CPUs for the bastion host VM"
  default     = 1
  type        = number
}

variable "bastion_host_cpu_cores" {
  description = "Number of cores per socket for the bastion host VM"
  default     = 1
  type        = number
}

variable "bastion_host_ssh_port" {
  description = "SSH port to be used for the bastion host"
  default     = 22
  type        = number
}