terraform {
  required_version = ">= 1.0.0"
  required_providers {
    packet = {
      source  = "hashicorp/vsphere"
      version = "~> 2.0.1"
    }
  }
}
