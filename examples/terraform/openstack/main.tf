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

provider "openstack" {
}

data "openstack_networking_network_v2" "external_network" {
  name     = var.external_network_name
  external = true
}

data "openstack_images_image_v2" "image" {
  name        = var.image
  most_recent = true
  properties  = var.image_properties_query
}

resource "openstack_compute_keypair_v2" "deployer" {
  name       = "${var.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
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
  admin_state_up      = "true"
  external_network_id = data.openstack_networking_network_v2.external_network.id
}

resource "openstack_networking_router_interface_v2" "router_subnet_link" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.subnet.id
}

resource "openstack_networking_secgroup_v2" "securitygroup" {
  name        = "${var.cluster_name}-cluster"
  description = "Security group for the Kubeone Kubernetes cluster ${var.cluster_name}"
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

resource "openstack_compute_instance_v2" "control_plane" {
  count = 3
  name  = "${var.cluster_name}-cp-${count.index}"

  image_name      = data.openstack_images_image_v2.image.name
  flavor_name     = var.control_plane_flavor
  key_pair        = openstack_compute_keypair_v2.deployer.name
  security_groups = [openstack_networking_secgroup_v2.securitygroup.name]

  network {
    port = element(openstack_networking_port_v2.control_plane.*.id, count.index)
  }
}

resource "openstack_compute_instance_v2" "lb" {
  name       = "${var.cluster_name}-lb"
  image_name = data.openstack_images_image_v2.image.name

  flavor_name     = var.lb_flavor
  key_pair        = openstack_compute_keypair_v2.deployer.name
  security_groups = [openstack_networking_secgroup_v2.securitygroup.name]

  network {
    port = openstack_networking_port_v2.lb.id
  }

  connection {
    type = "ssh"
    host = openstack_networking_floatingip_v2.lb.address
    user = var.ssh_username
  }

  provisioner "remote-exec" {
    script = "gobetween.sh"
  }
}

resource "openstack_networking_port_v2" "control_plane" {
  count = 3
  name  = "${var.cluster_name}-control_plane-${count.index}"

  admin_state_up     = "true"
  network_id         = openstack_networking_network_v2.network.id
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.subnet.id
  }
}

resource "openstack_networking_port_v2" "lb" {
  name = "${var.cluster_name}-lb"

  admin_state_up     = "true"
  network_id         = openstack_networking_network_v2.network.id
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.subnet.id
  }
}

resource "openstack_networking_floatingip_v2" "lb" {
  pool = var.external_network_name
}

resource "openstack_networking_floatingip_associate_v2" "lb" {
  floating_ip = openstack_networking_floatingip_v2.lb.address
  port_id     = openstack_networking_port_v2.lb.id
}

locals {
  rendered_lb_config = templatefile("./etc_gobetween.tpl", {
    lb_targets = openstack_compute_instance_v2.control_plane.*.access_ip_v4,
  })
}

resource "null_resource" "lb_config" {
  triggers = {
    cluster_instance_ids = join(",", openstack_compute_instance_v2.control_plane.*.id)
    config               = local.rendered_lb_config
  }

  depends_on = [
    openstack_compute_instance_v2.lb
  ]

  connection {
    host = openstack_networking_floatingip_v2.lb.address
    user = var.ssh_username
  }

  provisioner "file" {
    content     = local.rendered_lb_config
    destination = "/tmp/gobetween.toml"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/gobetween.toml /etc/gobetween.toml",
      "sudo systemctl restart gobetween",
    ]
  }
}

