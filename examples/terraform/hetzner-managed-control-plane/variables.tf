/*
Copyright 2026 The KubeOne Authors.

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

variable "name" {
  description = "common name for the resources, should be the name of the cluster"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$", var.name))
    error_message = "Value of name should be lowercase and can only contain alphanumeric characters and hyphens(-)."
  }
}

variable "ssh_port" {
  description = "SSH port to be used to provision instances"
  default     = 22
  type        = number
}

# Provider specific settings

variable "ip_range" {
  default     = ""
  description = "ip range to use for private network"
  type        = string
}

variable "base_network_cidr" {
  default     = "10.100.0.0/16"
  description = "base cidr, resulting cidr is randomly generated depending on provided subnet_mask"
  type        = string
}

variable "subnet_mask" {
  default     = 24
  description = "subnet mask to use for generating cidr for a private network"
  type        = number
}

variable "network_zone" {
  default     = "eu-central"
  description = "network zone to use for private network, the servers and the load balancer must be in a location of this zone"
  type        = string
}
