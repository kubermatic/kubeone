variable "cluster_name" {
  description = "prefix for cloud resources"
}

variable "control_plane_count" {
  default     = 3
  description = "Number of instances"
}

variable "ssh_public_key_file" {
  default     = "~/.ssh/id_rsa.pub"
  description = "SSH public key file"
}
