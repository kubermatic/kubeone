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

variable "cluster_network_cidr" {
  default     = "10.240.0.0/24"
  description = "Cluster network subnet cidr"
}

variable "ssh_port" {
  description = "SSH port"
  default     = 22
}
