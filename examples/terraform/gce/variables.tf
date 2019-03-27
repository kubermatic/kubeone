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
  description = "prefix for cloud resources"
}

variable "project" {
  description = "Project to be used for all resources"
}

variable "region" {
  default     = "europe-west3"
  description = "GCP region to speak to"
}

variable "control_plane_count" {
  default     = 3
  description = "Number of instances"
}

variable "control_plane_type" {
  default     = "n1-standard-1"
  description = "GCE instance type"
}

variable "control_plane_volume_size" {
  default     = 100
  description = "Size of the boot volume, in GB"
}

variable "control_plane_image_family" {
  default     = "ubuntu-1804-lts"
  description = "Image family to use for provisioning instances"
}

variable "control_plane_image_project" {
  default     = "ubuntu-os-cloud"
  description = "Project of the image to use for provisioning instances"
}

variable "workers_type" {
  default     = "n1-standard-1"
  description = "GCE instance type"
}

variable "workers_volume_size" {
  default     = 100
  description = "Size of the boot volume, in GB"
}


variable "cluster_network_cidr" {
  default     = "10.240.0.0/24"
  description = "Cluster network subnet cidr"
}

variable "ssh_port" {
  description = "SSH port"
  default     = 22
}

variable "ssh_public_key_file" {
  default     = "~/.ssh/id_rsa.pub"
  description = "SSH public key file"
}

variable "ssh_username" {
  default     = "kubeadmin"
  description = "Username to provision with the ssh_public_key_file"
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
}
