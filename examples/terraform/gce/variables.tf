variable "project" {
  description = "Project to be used for all resources"
  type        = string
  default = "ps-workspace"
}

variable "region" {
  default     = "europe-west3"
  description = "GCP region to speak to"
  type        = string
}

variable "ssh_key_files" {
  description = "SSH public key file"
  default = {
    public : "~/.ssh/ammar.lakis@gmail.com.pub"
    private : "~/.ssh/ammar.lakis@gmail.com"
  }
  type = object({
    public : string
    private : string
  })
}

variable "cluster_specs" {
  type = object({
    name : string
    network_cidr : string
  })
  default = {
    network_cidr = "10.240.0.0/24"
    name         = "mycluster"
  }
}

variable "control_plane_specs" {
  type = object({
    instance_count: number
    instance_type : string
    volume_size : number
    image_family : string
    image_project : string
  })

  default = {
    instance_count: 1
    instance_type : "n1-standard-2"
    volume_size : 100
    image_family : "ubuntu-1804-lts"
    image_project : "ubuntu-os-cloud"
  }
}

variable "static_workers_specs" {
  type = object({
    instance_count: number
    instance_type : string
    volume_size : number
    image_family : string
    image_project : string
  })

  default = {
    instance_count: 0
    instance_type : "n1-standard-2"
    volume_size : 100
    image_family : "ubuntu-1804-lts"
    image_project : "ubuntu-os-cloud"
  }
}
