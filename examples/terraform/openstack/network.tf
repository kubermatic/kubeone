/*
Copyright 2019 The KubeOne Authors.

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

data "openstack_networking_network_v2" "external_network" {
  name     = var.external_network_name
  external = true
}

resource "openstack_networking_network_v2" "network" {
  name           = "${var.cluster_name}-cluster"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "subnet" {
  name            = "${var.cluster_name}-cluster"
  network_id      = openstack_networking_network_v2.network.id
  cidr            = var.subnet_cidr
  ip_version      = 4
  dns_nameservers = var.subnet_dns_servers
}

resource "openstack_networking_router_v2" "router" {
  name                = "${var.cluster_name}-cluster"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external_network.id
}

resource "openstack_networking_router_interface_v2" "router_subnet_link" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.subnet.id
}
