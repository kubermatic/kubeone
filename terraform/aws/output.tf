output "kubeone_api" {
  value = {
    endpoint = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      cloud_provider       = "aws"
      private_address      = "${aws_instance.control_plane.*.private_ip}"
      public_address       = "${aws_instance.control_plane.*.public_ip}"
      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_port             = "${var.ssh_port}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
      ssh_user             = "ubuntu"
    }
  }
}

output "kubeone_workers" {
  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    fra1-a = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${local.az_a}"
      instanceProfile  = "${aws_iam_instance_profile.profile.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${aws_default_vpc.default.id}"
      subnetId         = "${data.aws_subnet.az_a.id}"
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
    }

    fra1-b = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${local.az_b}"
      instanceProfile  = "${aws_iam_instance_profile.profile.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${aws_default_vpc.default.id}"
      subnetId         = "${data.aws_subnet.az_b.id}"
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
    }

    fra1-c = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${local.az_c}"
      instanceProfile  = "${aws_iam_instance_profile.profile.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${aws_default_vpc.default.id}"
      subnetId         = "${data.aws_subnet.az_c.id}"
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
    }
  }
}
