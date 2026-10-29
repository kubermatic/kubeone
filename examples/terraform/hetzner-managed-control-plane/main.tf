/*
Copyright 2026 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

provider "hcloud" {
  # to avoid rate limiting issues
  poll_interval = "10s"
}

locals {
  base_network_mask = parseint(split("/", var.base_network_cidr)[1], 10)
  subnet_newbits    = var.subnet_mask - local.base_network_mask
  subnet_netnum     = pow(2, local.subnet_newbits) - 1
  ip_range = var.ip_range != "" ? var.ip_range : cidrsubnet(
    var.base_network_cidr,
    local.subnet_newbits,
    random_integer.random_subnet_netnum.result,
  )
}

resource "random_integer" "random_subnet_netnum" {
  min = 0
  max = local.subnet_netnum
}

############################### NETWORKING SETUP ###############################

# KubeOne looks the network up by its name (cloudProvider.hetzner.networkID)
# and attaches the API load balancer to it
resource "hcloud_network" "net" {
  name     = var.name
  ip_range = local.ip_range

  labels = {
    "kubeone_cluster_name" = var.name
  }
}

resource "hcloud_network_subnet" "kubeone" {
  network_id   = hcloud_network.net.id
  type         = "server"
  network_zone = var.network_zone
  ip_range     = local.ip_range
}

################################### FIREWALL ###################################

# Applies to the control plane servers created by KubeOne and to the workers,
# both are labeled with kubeone_cluster_name. Hetzner firewalls filter only the
# public interfaces, the private network and the load balancer are not affected.
resource "hcloud_firewall" "cluster" {
  name = "${var.name}-fw"

  labels = {
    "kubeone_cluster_name" = var.name
  }

  apply_to {
    label_selector = "kubeone_cluster_name=${var.name}"
  }

  rule {
    description = "allow ICMP"
    direction   = "in"
    protocol    = "icmp"
    source_ips = [
      "0.0.0.0/0",
    ]
  }

  rule {
    description = "allow all TCP inside cluster"
    direction   = "in"
    protocol    = "tcp"
    port        = "any"
    source_ips = [
      hcloud_network.net.ip_range,
    ]
  }

  rule {
    description = "allow all UDP inside cluster"
    direction   = "in"
    protocol    = "udp"
    port        = "any"
    source_ips = [
      hcloud_network.net.ip_range,
    ]
  }

  rule {
    description = "allow SSH from any"
    direction   = "in"
    protocol    = "tcp"
    port        = var.ssh_port
    source_ips = [
      "0.0.0.0/0",
    ]
  }

  rule {
    description = "allow NodePorts from any"
    direction   = "in"
    protocol    = "tcp"
    port        = "30000-32767"
    source_ips = [
      "0.0.0.0/0",
    ]
  }
}
