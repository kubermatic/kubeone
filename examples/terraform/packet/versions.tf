terraform {
  required_version = ">= 1.0.0"
  required_providers {
    packet = {
      source  = "packethost/packet"
      version = "~> 3.2.1"
    }
  }
}
