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

data "openstack_images_image_v2" "image" {
  name        = var.image
  most_recent = true
  properties  = var.image_properties_query
}

resource "openstack_compute_keypair_v2" "deployer" {
  name       = "${var.name}-deployer-key"
  public_key = file(pathexpand(var.ssh_public_key_file))
}

# Networking

data "openstack_networking_network_v2" "external_network" {
  name     = var.external_network_name
  external = true
}

resource "openstack_networking_network_v2" "network" {
  name           = "${var.name}-cluster"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "subnet" {
  name            = "${var.name}-cluster"
  network_id      = openstack_networking_network_v2.network.id
  cidr            = var.subnet_cidr
  ip_version      = 4
  dns_nameservers = var.subnet_dns_servers
}

resource "openstack_networking_router_v2" "router" {
  name                = "${var.name}-cluster"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external_network.id
}

resource "openstack_networking_router_interface_v2" "router_subnet_link" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.subnet.id
}

# Security group of all the nodes, the bastion and the load balancer

resource "openstack_networking_secgroup_v2" "securitygroup" {
  name        = "${var.name}-cluster"
  description = "Security group for the Kubeone Kubernetes cluster ${var.name}"
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_allow_internal_ipv4" {
  description       = "Allow security group internal IPv4 traffic"
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_group_id   = openstack_networking_secgroup_v2.securitygroup.id
  security_group_id = openstack_networking_secgroup_v2.securitygroup.id
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_ssh" {
  description       = "Allow SSH"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.securitygroup.id
}

resource "openstack_networking_secgroup_rule_v2" "nodeports" {
  description       = "Allow NodePorts"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 30000
  port_range_max    = 32767
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.securitygroup.id
}

resource "openstack_networking_secgroup_rule_v2" "secgroup_apiserver" {
  description       = "Allow kube-apiserver"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.securitygroup.id
}

# Bastion, the control plane VMs created by KubeOne have no floating IPs

resource "openstack_networking_port_v2" "bastion" {
  name               = "${var.name}-bastion"
  admin_state_up     = "true"
  network_id         = openstack_networking_network_v2.network.id
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.subnet.id
  }
}

resource "openstack_compute_instance_v2" "bastion" {
  name            = "${var.name}-bastion"
  image_name      = data.openstack_images_image_v2.image.name
  flavor_name     = var.bastion_flavor
  key_pair        = openstack_compute_keypair_v2.deployer.name
  security_groups = [openstack_networking_secgroup_v2.securitygroup.name]
  config_drive    = var.config_drive

  network {
    port = openstack_networking_port_v2.bastion.id
  }
}

resource "openstack_networking_floatingip_v2" "bastion" {
  pool = var.external_network_name
}

resource "openstack_networking_floatingip_associate_v2" "bastion" {
  floating_ip = openstack_networking_floatingip_v2.bastion.address
  port_id     = openstack_networking_port_v2.bastion.id

  depends_on = [openstack_networking_router_interface_v2.router_subnet_link]
}

# kube-apiserver load balancer, KubeOne adds the control plane VMs to the pool

resource "openstack_lb_loadbalancer_v2" "kube_apiserver" {
  name               = "${var.name}-kube-apiserver"
  admin_state_up     = true
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]
  vip_network_id     = openstack_networking_network_v2.network.id
  vip_subnet_id      = openstack_networking_subnet_v2.subnet.id
}

resource "openstack_lb_pool_v2" "kube_apiservers" {
  name            = "${var.name}-kube-apiservers"
  protocol        = "TCP"
  lb_method       = "ROUND_ROBIN"
  loadbalancer_id = openstack_lb_loadbalancer_v2.kube_apiserver.id
}

resource "openstack_lb_listener_v2" "kube_apiserver" {
  name            = "${var.name}-kube-apiserver"
  protocol        = "TCP"
  protocol_port   = 6443
  admin_state_up  = true
  default_pool_id = openstack_lb_pool_v2.kube_apiservers.id
  loadbalancer_id = openstack_lb_loadbalancer_v2.kube_apiserver.id
}

resource "openstack_lb_monitor_v2" "lb_monitor_tcp" {
  name        = "${var.name}-kube-apiserver"
  pool_id     = openstack_lb_pool_v2.kube_apiservers.id
  type        = "TCP"
  delay       = 10
  timeout     = 5
  max_retries = 5
}
