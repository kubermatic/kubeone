/*
Copyright 2022 The KubeOne Authors.

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

# Configure the VMware vCloud Director Provider
provider "vcd" {
  /*
  See https://registry.terraform.io/providers/vmware/vcd/latest/docs#argument-reference
  for config options reference
  */
  org = var.vcd_org_name
  vdc = var.vcd_vdc_name
}

locals {
  external_network = element([for net in data.vcd_edgegateway.edge_gateway.external_network : net.name if tolist(net.subnet)[0].use_for_default_route], 0)
  public_ip        = var.external_network_ip == "" ? data.vcd_edgegateway.edge_gateway.default_external_network_ip : var.external_network_ip
}

# Existing edge gateway in VDC
data "vcd_edgegateway" "edge_gateway" {
  name = var.vcd_edge_gateway_name
}

# Routed network that will be connected to the edge gateway
resource "vcd_network_routed" "network" {
  name        = "${var.cluster_name}-routed-network"
  description = "Routed network for ${var.cluster_name} vApp"

  edge_gateway = data.vcd_edgegateway.edge_gateway.name

  interface_type = var.network_interface_type

  gateway = cidrhost(var.network_subnet, 1)

  dhcp_pool {
    start_address = cidrhost(var.network_subnet, 2)
    end_address   = cidrhost(var.network_subnet, 50)
  }

  dns1 = var.network_dns_server_1
  dns2 = var.network_dns_server_2
}

# Dedicated vApp for cluster resources; vms, disks, network, etc.
resource "vcd_vapp" "cluster" {
  name        = var.cluster_name
  description = "vApp for ${var.vcd_vdc_name} cluster"

  metadata = {
    provisioner  = "Kubeone"
    cluster_name = "${var.cluster_name}"
    type         = "Kubernetes Cluster"
  }

  depends_on = [vcd_network_routed.network]
}

# Connect the dedicated routed network to vApp
resource "vcd_vapp_org_network" "network" {
  vapp_name = var.cluster_name

  org_network_name = vcd_network_routed.network.name

  depends_on = [vcd_vapp.cluster, vcd_network_routed.network]
}

# Create VMs for control plane
resource "vcd_vapp_vm" "control_plane" {
  count         = 3
  vapp_name     = vcd_vapp.cluster.name
  name          = "${var.cluster_name}-master-${count.index + 1}"
  computer_name = "${var.cluster_name}-master-${count.index + 1}"

  metadata = {
    provisioner  = "Kubeone"
    cluster_name = "${var.cluster_name}"
    role         = "master"
  }

  guest_properties = {
    "instance-id" = "${var.cluster_name}-master-${count.index + 1}"
    "hostname"    = "${var.cluster_name}-master-${count.index + 1}"
    "public-keys" = file(var.ssh_public_key_file)
  }

  catalog_name  = var.catalog_name
  template_name = var.template_name

  # resource allocation for the VM
  memory                 = var.control_plane_memory
  cpus                   = var.control_plane_cpus
  cpu_cores              = var.control_plane_cpu_cores
  cpu_hot_add_enabled    = false
  memory_hot_add_enabled = false

  # Wait upto 5 minutes for IP addresses to be assigned
  network_dhcp_wait_seconds = 300

  network {
    type               = "org"
    name               = vcd_vapp_org_network.network.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  override_template_disk {
    bus_type        = "paravirtual"
    size_in_mb      = var.control_plane_disk_size
    bus_number      = 0
    unit_number     = 0
    storage_profile = var.control_plane_disk_storage_profile
  }

  depends_on = [vcd_vapp_org_network.network]
}

# Create VMs for bastion host
resource "vcd_vapp_vm" "bastion" {
  vapp_name     = vcd_vapp.cluster.name
  name          = "${var.cluster_name}-bastion"
  computer_name = "${var.cluster_name}-bastion"

  metadata = {
    provisioner  = "Kubeone"
    cluster_name = "${var.cluster_name}"
    role         = "bastion"
  }

  guest_properties = {
    "instance-id" = "${var.cluster_name}-bastion"
    "hostname"    = "${var.cluster_name}-bastion"
    "public-keys" = file(var.ssh_public_key_file)
  }

  catalog_name  = var.catalog_name
  template_name = var.template_name

  # resource allocation for the VM
  memory                 = 2048
  cpus                   = 1
  cpu_cores              = 1
  cpu_hot_add_enabled    = false
  memory_hot_add_enabled = false

  # Wait upto 5 minutes for IP addresses to be assigned
  network_dhcp_wait_seconds = 300

  network {
    type               = "org"
    name               = vcd_vapp_org_network.network.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  depends_on = [vcd_vapp_org_network.network]
}

#################################### NAT and Firewall rules ####################################

# Create the firewall rule to access the Internet
resource "vcd_nsxv_firewall_rule" "rule_internet" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name
  name         = "${var.cluster_name}-firewall-rule-internet"

  action = "accept"

  source {
    org_networks = [vcd_network_routed.network.name]
  }

  destination {
    ip_addresses = []
  }

  service {
    protocol = "any"
  }
}

# Create SNAT rule to access the Internet
resource "vcd_nsxv_snat" "rule_internet" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name
  network_type = "ext"
  network_name = local.external_network

  original_address   = var.network_subnet
  translated_address = local.public_ip
}

# Create the firewall rule to allow SSH from the Internet
resource "vcd_nsxv_firewall_rule" "rule_ssh_bastion" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name
  name         = "${var.cluster_name}-firewall-rule-ssh"

  action = "accept"

  source {
    ip_addresses = ["any"]
  }

  destination {
    ip_addresses = [local.public_ip]
  }

  service {
    protocol = "tcp"
    port     = 22
  }
}

# Create DNAT rule to allow SSH from the Internet
resource "vcd_nsxv_dnat" "rule_ssh_bastion" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name
  network_type = "ext"
  network_name = local.external_network

  original_address = local.public_ip
  original_port    = 22

  translated_address = vcd_vapp_vm.bastion.network[0].ip
  translated_port    = 22
  protocol           = "tcp"
}

#################################### Loadbalancer ####################################

resource "vcd_lb_app_profile" "app_profile" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name

  name = "${var.cluster_name}-app-profile"
  type = "tcp"
}

resource "vcd_lb_service_monitor" "lb_monitor" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name

  name        = "${var.cluster_name}-control-plane-monitor"
  interval    = 15
  timeout     = 20
  max_retries = 3
  type        = "http"
  method      = "GET"
  url         = "/healthz"
}

resource "vcd_lb_server_pool" "control_plane" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name

  name                = "${var.cluster_name}-control-plane"
  algorithm           = "round-robin"
  enable_transparency = "true"

  monitor_id = vcd_lb_service_monitor.lb_monitor.id

  dynamic "member" {
    for_each = vcd_vapp_vm.control_plane
    content {
      condition    = "enabled"
      name         = member.value.name
      ip_address   = member.value.network[0].ip
      port         = 6443
      monitor_port = 6443
      weight       = 1
    }
  }
}

resource "vcd_lb_virtual_server" "lb" {
  edge_gateway = data.vcd_edgegateway.edge_gateway.name

  name           = "${var.cluster_name}-control-plane"
  ip_address     = local.public_ip
  protocol       = "tcp"
  port           = 6443
  app_profile_id = vcd_lb_app_profile.app_profile.id
  server_pool_id = vcd_lb_server_pool.control_plane.id
}

# Create the firewall rule to allow access to API server
resource "vcd_nsxv_firewall_rule" "rule_kube_apiserver" {
  count        = var.expose_kube_apiserver ? 1 : 0
  edge_gateway = data.vcd_edgegateway.edge_gateway.name
  name         = "${var.cluster_name}-firewall-rule-kube-apiserver"

  action = "accept"

  source {
    ip_addresses = ["any"]
  }

  destination {
    ip_addresses = [local.public_ip]
  }

  service {
    protocol = "tcp"
    port     = 6443
  }
}
