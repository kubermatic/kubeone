output "kubeone_api" {
  value = {
    endpoint = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      hostnames            = "${aws_instance.control_plane.*.private_dns}"
      private_address      = "${aws_instance.control_plane.*.private_ip}"
      public_address       = "${aws_instance.control_plane.*.public_ip}"
      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_port             = "${var.ssh_port}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
      ssh_user             = "ubuntu"
    }
  }
}

output "kubeone_worker" {
  value = {
    aws = {
      availability_zones     = ["${data.aws_availability_zones.available.names}"]
      iam_instance_profile   = "${aws_iam_instance_profile.profile.name}"
      region                 = "${var.aws_region}"
      subnet_id              = "${data.aws_subnet_ids.default.ids[0]}"
      vpc_id                 = "${aws_default_vpc.default.id}"
      vpc_security_group_ids = ["${aws_security_group.common.id}"]
      instance_type          = "${var.worker_instance_type}"
      disk_size              = "${var.worker_disk_size}"
      disk_type              = "${var.worker_disk_type}"
    }
  }
}
