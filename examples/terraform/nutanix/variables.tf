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

variable "control_plane_vm_count" {
  description = "number of control plane instances"
  default     = 3
  type        = number
}

# Provider specific settings

variable "nutanix_cluster_name" {
  description = "Name of the Nutanix Cluster which will be used for this Kubernetes cluster"
  type        = string
}

variable "project_name" {
  description = "Name of the Nutanix Project"
  type        = string
}

variable "subnet_name" {
  description = "Name of the subnet"
  type        = string
}

variable "image_name" {
  description = "Image to be used for instances (control plane, bastion/LB, workers)"
  type        = string
}

variable "control_plane_vcpus" {
  default     = 2
  description = "Number of vCPUs per socket for control plane nodes"
  type        = number
}

variable "control_plane_sockets" {
  default     = 1
  description = "Number of sockets for control plane nodes"
  type        = number
}

variable "control_plane_memory_size" {
  default     = 4096
  description = "Memory size, in Mib, for control plane nodes"
  type        = number
}

variable "control_plane_disk_size" {
  default     = 102400
  description = "Disk size size, in Mib, for control plane nodes"
  type        = number
}

variable "bastion_vcpus" {
  default     = 1
  description = "Number of vCPUs per socket for bastion/LB node"
  type        = number
}

variable "bastion_sockets" {
  default     = 1
  description = "Number of sockets for bastion/LB node"
  type        = number
}

variable "bastion_memory_size" {
  default     = 4096
  description = "Memory size, in Mib, for bastion/LB node"
  type        = number
}

variable "bastion_disk_size" {
  default     = 102400
  description = "Disk size, in Mib, for bastion/LB node"
  type        = number
}

variable "worker_vcpus" {
  default     = 2
  description = "Number of vCPUs per socket for worker nodes"
  type        = number
}

variable "worker_sockets" {
  default     = 1
  description = "Number of sockets for worker nodes"
  type        = number
}

variable "worker_memory_size" {
  default     = 4096
  description = "Memory size, in Mib, for worker nodes"
  type        = number
}

variable "worker_disk_size" {
  default     = 50
  description = "Disk size size, in Gb, for worker nodes"
  type        = number
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

variable "initial_machinedeployment_operating_system_profile" {
  default     = ""
  type        = string
  description = <<EOF
Name of operating system profile for MachineDeployments, only applicable if operating-system-manager addon is enabled.
If not specified, the default value will be added by machine-controller addon.
EOF
}
