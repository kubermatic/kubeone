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

output "vm" {
  value = {
    image = data.openstack_images_image_v2.image.name
  }
}

output "networking" {
  value = {
    networkID = openstack_networking_network_v2.network.id
    # the subnet ID (not its name) is required by KubeOne to add the control
    # plane VMs to the load balancer pool
    subnetID      = openstack_networking_subnet_v2.subnet.id
    securityGroup = openstack_networking_secgroup_v2.securitygroup.name
  }
}

output "bastion" {
  value = {
    address = openstack_networking_floatingip_v2.bastion.address
    port    = var.bastion_port
    user    = var.bastion_user
  }
}

output "loadbalancer" {
  value = {
    name = openstack_lb_loadbalancer_v2.kube_apiserver.name
  }
}
