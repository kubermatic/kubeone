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

resource "openstack_lb_loadbalancer_v2" "kube_apiserver" {
  name               = "${var.cluster_name}-kube-apiserver"
  admin_state_up     = true
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]
  vip_network_id     = openstack_networking_network_v2.network.id
  vip_subnet_id      = openstack_networking_subnet_v2.subnet.id
}

resource "openstack_lb_pool_v2" "kube_apiservers" {
  name            = "${var.cluster_name}-kube-apiservers"
  protocol        = "TCP"
  lb_method       = "ROUND_ROBIN"
  loadbalancer_id = openstack_lb_loadbalancer_v2.kube_apiserver.id
}

resource "openstack_lb_listener_v2" "kube_apiserver" {
  name            = "${var.cluster_name}-kube-apiserver"
  protocol        = "TCP"
  protocol_port   = 6443
  admin_state_up  = true
  default_pool_id = openstack_lb_pool_v2.kube_apiservers.id
  loadbalancer_id = openstack_lb_loadbalancer_v2.kube_apiserver.id
}

resource "openstack_lb_monitor_v2" "lb_monitor_tcp" {
  name        = "${var.cluster_name}-kube-apiserver"
  pool_id     = openstack_lb_pool_v2.kube_apiservers.id
  type        = "TCP"
  delay       = 30
  timeout     = 10
  max_retries = 5
}

resource "openstack_lb_member_v2" "kube_apiserver" {
  count         = length(openstack_compute_instance_v2.control_plane)
  name          = "${var.cluster_name}-kube_apiserver-${openstack_compute_instance_v2.control_plane[count.index].access_ip_v4}"
  pool_id       = openstack_lb_pool_v2.kube_apiservers.id
  address       = openstack_compute_instance_v2.control_plane[count.index].access_ip_v4
  protocol_port = 6443
}

resource "openstack_networking_floatingip_v2" "kube_apiserver" {
  pool = var.external_network_name
}

resource "openstack_networking_floatingip_associate_v2" "kube_apiserver" {
  floating_ip = openstack_networking_floatingip_v2.kube_apiserver.address
  port_id     = openstack_lb_loadbalancer_v2.kube_apiserver.vip_port_id
}
