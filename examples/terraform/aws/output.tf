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
      vpcId            = "${local.vpc_id}"
      subnetId         = "${data.aws_subnet.az_a.id}"
      instanceType     = "t2.medium"
      diskSize         = 50
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 1
      operatingSystem  = "ubuntu"
    }

    fra1-b = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${local.az_b}"
      instanceProfile  = "${aws_iam_instance_profile.profile.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${local.vpc_id}"
      subnetId         = "${data.aws_subnet.az_b.id}"
      instanceType     = "t2.medium"
      diskSize         = 50
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 1
      operatingSystem  = "ubuntu"
    }

    fra1-c = {
      region           = "${var.aws_region}"
      ami              = "${data.aws_ami.ubuntu.id}"
      availabilityZone = "${local.az_c}"
      instanceProfile  = "${aws_iam_instance_profile.profile.name}"
      securityGroupIDs = ["${aws_security_group.common.id}"]
      vpcId            = "${local.vpc_id}"
      subnetId         = "${data.aws_subnet.az_c.id}"
      instanceType     = "t2.medium"
      diskSize         = 50
      sshPublicKeys    = ["${aws_key_pair.deployer.public_key}"]
      replicas         = 1
      operatingSystem  = "ubuntu"
    }
  }
}
