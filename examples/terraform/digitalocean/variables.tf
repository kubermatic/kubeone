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

variable "os" {
  description = "Operating System to use in image filtering and MachineDeployment"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rockylinux
  default = "ubuntu"
  type    = string
}


variable "worker_os" {
  description = "OS to run on worker machines"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rockylinux
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
  default     = ""
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

variable "disable_kubeapi_loadbalancer" {
  type        = bool
  default     = false
  description = "E2E tests specific variable to disable usage of any loadbalancer in front of kubeapi-server"
}

variable "control_plane_vm_count" {
  description = "number of control plane instances"
  default     = 3
  type        = number
}

# Provider specific settings

variable "image_references" {
  description = "map with images"
  type = map(object({
    image_name   = string
    ssh_username = string
    worker_os    = string
  }))
  default = {
    ubuntu = {
      image_name   = "ubuntu-22-04-x64"
      ssh_username = "root"
      worker_os    = "ubuntu"
    }

    centos = {
      image_name   = "centos-7-x64"
      ssh_username = "root"
      worker_os    = "centos"
    }

    rockylinux = {
      image_name   = "rockylinux-8-x64"
      ssh_username = "root"
      worker_os    = "rockylinux"
    }
  }
}

variable "region" {
  description = "Region to speak to"
  default     = "fra1"
  type        = string
}

variable "control_plane_droplet_image" {
  description = "Image to use for provisioning control plane droplets"
  default     = ""
  type        = string
}

variable "control_plane_size" {
  description = "Size of control plane nodes"
  default     = "s-2vcpu-4gb"
  type        = string
}

variable "worker_size" {
  description = "Size of worker nodes"
  default     = "s-2vcpu-4gb"
  type        = string
}

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
