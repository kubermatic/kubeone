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
  default     = "root"
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

# Provider specific settings

variable "project" {
  description = "Project to be used for all resources"
  type        = string
}

variable "region" {
  default     = "europe-west3"
  description = "GCP region to speak to"
  type        = string
}

variable "control_plane_target_pool_members_count" {
  default = 3
  type    = number
}

variable "control_plane_type" {
  default     = "n1-standard-2"
  description = "GCE instance type"
  type        = string
}

variable "control_plane_volume_size" {
  default     = 100
  description = "Size of the boot volume, in GB"
  type        = number
}

variable "control_plane_image_family" {
  default     = "ubuntu-2004-lts"
  description = "Image family to use for provisioning instances"
  type        = string
}

variable "control_plane_image_project" {
  default     = "ubuntu-os-cloud"
  description = "Project of the image to use for provisioning instances"
  type        = string
}

variable "workers_type" {
  default     = "n1-standard-2"
  description = "GCE instance type"
  type        = string
}

variable "cluster_network_cidr" {
  default     = "10.240.0.0/24"
  description = "Cluster network subnet cidr"
  type        = string
}
