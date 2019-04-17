output "kubeone_api" {
  value = {
    endpoint = "${aws_lb.control_plane.dns_name}"
  }
}

output "kubeone_bastion" {
  value = "${aws_instance.bastion.0.public_ip}"
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name         = "${var.cluster_name}"
      cloud_provider       = "aws"
      private_address      = "${aws_instance.control_plane.*.private_ip}"
      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_port             = "${var.ssh_port}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
      ssh_user             = "${var.ssh_username}"
    }
  }
}

output "kubeone_workers" {
  value = {
    # following outputs will be parsed by kubeone and automatically merged into
    # corresponding (by name) worker definition
    pool-a = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${data.aws_availability_zones.available.names[0]}"
      instanceProfile  = "${aws_iam_instance_profile.workers.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${data.aws_vpc.selected.id}"
      subnetId         = "${aws_subnet.private.0.id}"
      instanceType     = "${var.worker_type}"
      diskSize         = 100
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 3
      operatingSystem  = "ubuntu"
    }

    pool-b = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${data.aws_availability_zones.available.names[1]}"
      instanceProfile  = "${aws_iam_instance_profile.workers.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${data.aws_vpc.selected.id}"
      subnetId         = "${aws_subnet.private.1.id}"
      instanceType     = "${var.worker_type}"
      diskSize         = 100
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 3
      operatingSystem  = "ubuntu"
    }

    pool-c = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${data.aws_availability_zones.available.names[2]}"
      instanceProfile  = "${aws_iam_instance_profile.workers.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${data.aws_vpc.selected.id}"
      subnetId         = "${aws_subnet.private.2.id}"
      instanceType     = "${var.worker_type}"
      diskSize         = 100
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 3
      operatingSystem  = "ubuntu"
    }
  }
}
