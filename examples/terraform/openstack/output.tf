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

output "kubeone_api" {
  value = {
    endpoint = "${openstack_networking_floatingip_v2.lb.address}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      cloud_provider       = "openstack"
      private_address      = "${openstack_compute_instance_v2.control_plane.*.access_ip_v4}"
      public_address       = "${openstack_networking_floatingip_v2.control_plane.*.address}"
      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_port             = "${var.ssh_port}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
      ssh_user             = "ubuntu"
    }
  }
}

output "kubeone_workers" {
  value = {
    nodes1 = {
      image            = "${var.image}"
      instanceProfile  = "${var.worker_flavor}"
      securityGroupIDs = ["${openstack_networking_secgroup_v2.securitygroup.id}"]
      floatingIPPool   = "${var.external_network_name}"
      network          = "${openstack_networking_network_v2.network.name}"
      subnet           = "${openstack_networking_subnet_v2.subnet.name}"
      operatingSystem  = "ubuntu"
      replicas         = 1
    }
  }
}
