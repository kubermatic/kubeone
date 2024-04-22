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
  description = "Operating System to use for finding image reference and in MachineDeployment"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rockylinux
  # * rhel
  # * flatcar
  default = "ubuntu"
  type    = string
}

variable "worker_os" {
  description = "OS to run on worker machines"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rockylinux
  # * rhel
  # * flatcar
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

# Provider specific settings

variable "location" {
  description = "Azure datacenter to use"
  default     = "westeurope"
  type        = string
}

variable "image_references" {
  description = "map with image references used for control plane"
  type = map(object({
    image = object({
      publisher = string
      offer     = string
      sku       = string
      version   = string
    })
    plan = list(object({
      name      = string
      publisher = string
      product   = string
    }))
    ssh_username = string
    worker_os    = string
  }))
  default = {
    ubuntu = {
      image = {
        publisher = "Canonical"
        offer     = "0001-com-ubuntu-server-jammy"
        sku       = "22_04-lts"
        version   = "latest"
      }
      plan         = []
      ssh_username = "ubuntu"
      worker_os    = "ubuntu"
    }

    centos = {
      image = {
        publisher = "OpenLogic"
        offer     = "CentOS"
        sku       = "7_9"
        version   = "latest"
      }
      plan         = []
      ssh_username = "centos"
      worker_os    = "centos"
    }

    flatcar = {
      image = {
        publisher = "kinvolk"
        offer     = "flatcar-container-linux"
        sku       = "stable"
        version   = "3815.2.0"
      }
      plan = [{
        name      = "stable"
        publisher = "kinvolk"
        product   = "flatcar-container-linux"
      }]
      ssh_username = "core"
      worker_os    = "flatcar"
    }

    rhel = {
      image = {
        publisher = "RedHat"
        offer     = "rhel-byos"
        sku       = "rhel-lvm85"
        version   = "8.5.20220316"
      }
      plan = [{
        name      = "rhel-lvm85"
        publisher = "redhat"
        product   = "rhel-byos"
      }]
      ssh_username = "rhel-user"
      worker_os    = "rhel"
    }

    rockylinux = {
      image = {
        publisher = "procomputers"
        offer     = "rocky-linux-8-5"
        sku       = "rocky-linux-8-5"
        version   = "8.5.20211118"
      }
      plan = [{
        name      = "rocky-linux-8-5"
        publisher = "procomputers"
        product   = "rocky-linux-8-5"
      }]
      ssh_username = "rocky"
      worker_os    = "rockylinux"
    }
  }
}

variable "control_plane_vm_size" {
  description = "VM Size for control plane machines"
  default     = "Standard_F2"
  type        = string
}

variable "worker_vm_size" {
  description = "VM Size for worker machines"
  default     = "Standard_F2"
  type        = string
}

variable "control_plane_vm_count" {
  description = "Number of control plane instances"
  default     = 3
  type        = number
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

# RHEL subscription
variable "rhsm_username" {
  description = "RHSM username"
  default     = ""
  type        = string
  sensitive   = true
}

variable "rhsm_password" {
  description = "RHSM password"
  default     = ""
  type        = string
  sensitive   = true
}

variable "rhsm_offline_token" {
  description = "RHSM offline token"
  default     = ""
  type        = string
  sensitive   = true
}

variable "disable_auto_update" {
  description = "Disable automatic flatcar updates (and reboot)"
  type        = bool
  default     = false
}
