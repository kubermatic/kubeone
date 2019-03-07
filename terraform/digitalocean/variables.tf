variable "cluster_name" {
  description = "Name of the cluster"
}

variable "region" {
  default     = "fra1"
  description = "Region to speak to"
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

variable "ssh_private_key_file" {
  description = "SSH private key file used to access instances"
  default     = ""
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
}

variable "droplet_image" {
  default     = "ubuntu-18-04-x64"
  description = "Image to use for provisioning droplet"
}

variable "droplet_size" {
  default     = "s-2vcpu-4gb"
  description = "Size of Droplets"
}

variable "droplet_private_networking" {
  default     = true
  description = "Enable Private Networking on Droplets (recommended)"
}

variable "droplet_monitoring" {
  default     = false
  description = "Enable advance Droplet metrics"
}

variable "droplet_ipv6" {
  default     = false
  description = "Enable IPv6"
}
