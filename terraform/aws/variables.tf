variable "cluster_name" {
  default     = "kubeone"
  description = "profix for cloud resources"
}

variable "aws_region" {
  default     = "eu-central-1"
  description = "AWS region to speak to"
}

variable "ssh_key_file" {
  description = "SSH key name"
}

variable "control_plane_count" {
  default     = 3
  description = "Number of instances"
}

variable "control_plane_type" {
  default     = "t3.medium"
  description = "AWS instance type"
}

variable "control_plane_volume_size" {
  default     = 100
  description = "Size of the EBS volume, in Gb"
}
