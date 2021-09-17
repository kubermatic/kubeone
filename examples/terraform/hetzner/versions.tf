terraform {
  required_version = ">= 1.0.0"
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.31.0"
    }
  }
}
