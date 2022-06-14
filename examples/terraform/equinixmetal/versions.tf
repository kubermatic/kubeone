terraform {
  required_version = ">= 1.0.0"
  required_providers {
    metal = {
      source  = "equinix/metal"
      version = "~> 3.2.0"
    }
  }
}
