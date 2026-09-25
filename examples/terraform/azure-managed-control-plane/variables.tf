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

# Provider specific settings

variable "location" {
  description = "Azure datacenter to use"
  default     = "westeurope"
  type        = string
}

variable "os" {
  description = "Operating System to use for finding the image reference used in controlPlane.nodeSets"

  # valid choices are:
  # * ubuntu
  # * centos
  # * rockylinux
  # * rhel
  # * flatcar
  default = "ubuntu"
  type    = string
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
      # See https://documentation.ubuntu.com/azure/en/latest/azure-how-to/instances/find-ubuntu-images/
      image = {
        publisher = "Canonical"
        offer     = "ubuntu-24_04-lts"
        sku       = "server-gen1"
        version   = "latest"
      }
      plan         = []
      ssh_username = "ubuntu"
      worker_os    = "ubuntu"
    }

    flatcar = {
      image = {
        publisher = "kinvolk"
        offer     = "flatcar-container-linux-corevm-amd64"
        sku       = "stable"
        version   = "4593.2.2"
      }
      plan         = []
      ssh_username = "core"
      worker_os    = "flatcar"
    }

    rhel = {
      image = {
        publisher = "RedHat"
        offer     = "rhel-byos"
        sku       = "rhel-lvm95"
        version   = "9.5.2024112215"
      }
      plan = [{
        name      = "rhel-lvm95"
        publisher = "redhat"
        product   = "rhel-byos"
      }]
      ssh_username = "rhel-user"
      worker_os    = "rhel"
    }

    rockylinux = {
      image = {
        publisher = "resf"
        offer     = "rockylinux-x86_64"
        sku       = "9-base"
        version   = "9.6.20250531"
      }
      plan = [{
        name      = "9-base"
        publisher = "resf"
        product   = "rockylinux-x86_64"
      }]
      ssh_username = "rocky"
      worker_os    = "rockylinux"
    }
  }
}
