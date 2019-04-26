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
}

variable "facility" {
  default     = "ams1"
  description = "Facility (datacenter)"
}

variable "control_plane_count" {
  default     = 3
  description = "Number of master instances"
}

variable "ssh_public_key_file" {
  description = "SSH public key file"
  default     = "~/.ssh/id_rsa.pub"
}

variable "ssh_port" {
  default     = 22
  description = "SSH port to be used to provision instances"
}

variable "ssh_username" {
  default     = "root"
  description = "SSH user, used only in output"
}

variable "ssh_private_key_file" {
  description = "SSH private key file used to access instances"
  default     = ""
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
}

variable "operating_system" {
  default     = "ubuntu_18_04"
  description = "Image to use for provisioning device"
}

variable "device_type" {
  default     = "t1.small.x86"
  description = "type (size) of the device"
}

variable "project_id" {
  description = "project ID"
}
