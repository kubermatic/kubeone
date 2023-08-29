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

resource "openstack_networking_port_v2" "control_plane" {
  count = var.control_plane_vm_count
  name  = "${var.cluster_name}-control_plane-${count.index}"

  admin_state_up     = "true"
  network_id         = openstack_networking_network_v2.network.id
  security_group_ids = [openstack_networking_secgroup_v2.securitygroup.id]

  fixed_ip {
    subnet_id = openstack_networking_subnet_v2.subnet.id
  }
}

resource "openstack_compute_instance_v2" "control_plane" {
  count = var.control_plane_vm_count
  name  = "${var.cluster_name}-cp-${count.index}"

  image_name      = data.openstack_images_image_v2.image.name
  flavor_name     = var.control_plane_flavor
  key_pair        = openstack_compute_keypair_v2.deployer.name
  security_groups = [openstack_networking_secgroup_v2.securitygroup.name]

  network {
    port = element(openstack_networking_port_v2.control_plane[*].id, count.index)
  }
}

