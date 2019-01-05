variable "cluster_name" {
  description = "prefix for cloud resources"
}

variable "control_plane_count" {
  default     = 3
  description = "Number of instances"
}
