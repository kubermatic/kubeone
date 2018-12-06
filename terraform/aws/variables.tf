variable "cluster_name" {
  description = "prefix for cloud resources"
}

variable "aws_region" {
  default     = "eu-central-1"
  description = "AWS region to speak to"
}

variable "ssh_public_key_file" {
  description = "SSH public key file"
}

variable "ssh_private_key_file" {
  description = "SSH private key file, only specify in absence of SSH agent"
  default     = ""
}

variable "ssh_agent_socket" {
  description = "SSH Agent socket, default to grab from $SSH_AUTH_SOCK"
  default     = "env:SSH_AUTH_SOCK"
}

variable "ssh_port" {
  description = "SSH port"
  default     = 22
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

variable "s3_access_key" {
  default     = "env:BACKUP_AWS_ACCESS_KEY_ID"
  description = "AWS Access Key used to access S3 bucket for storing backups"
}

variable "s3_secret_access_key" {
  default     = "env:BACKUP_AWS_SECRET_ACCESS_KEY"
  description = "AWS Secret Access Key used to access S3 bucket for storing backups"
}