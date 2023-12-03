terraform {
  required_version = ">= 1.0.0"

  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.52.0"
    }
  }
}

provider "openstack" {
  use_octavia = true
}
