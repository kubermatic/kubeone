output "kubeone_api" {
  value = {
    endpoint = "${openstack_networking_floatingip_v2.lb.fixed_ip}"
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

      operatingSystem = "ubuntu"
      replicas        = 1
    }
  }
}
