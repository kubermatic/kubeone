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

provider "aws" {
  region = var.aws_region
}

locals {
  az_count         = length(data.aws_availability_zones.available.names)
  az_a             = "${var.aws_region}a"
  az_b             = "${var.aws_region}b"
  az_c             = "${var.aws_region}c"
  kube_cluster_tag = "kubernetes.io/cluster/${var.cluster_name}"
  vpc_id           = var.vpc_id == "default" ? aws_default_vpc.default.id : var.vpc_id

  ami = var.ami == "" ? data.aws_ami.ubuntu.id : var.ami
}

data "aws_availability_zones" "available" {
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_subnet_ids" "default" {
  vpc_id = local.vpc_id
}

data "aws_subnet" "az_a" {
  availability_zone = local.az_a
  vpc_id            = local.vpc_id
}

data "aws_subnet" "az_b" {
  availability_zone = local.az_b
  vpc_id            = local.vpc_id
}

data "aws_subnet" "az_c" {
  availability_zone = local.az_c
  vpc_id            = local.vpc_id
}

locals {
  all_subnets = [data.aws_subnet.az_a.id, data.aws_subnet.az_b.id, data.aws_subnet.az_c.id]
}

resource "aws_default_vpc" "default" {
}

resource "aws_key_pair" "deployer" {
  key_name   = "${var.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

resource "aws_security_group" "common" {
  name        = "${var.cluster_name}-common"
  description = "cluster common rules"
  vpc_id      = local.vpc_id

  tags = map(local.kube_cluster_tag, "shared")

  ingress {
    from_port   = var.ssh_port
    to_port     = var.ssh_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "control_plane" {
  name        = "${var.cluster_name}-control_planes"
  description = "cluster control_planes"
  vpc_id      = local.vpc_id

  tags = map(local.kube_cluster_tag, "shared")

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "role" {
  name = "${var.cluster_name}-host"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_iam_instance_profile" "profile" {
  name = "${var.cluster_name}-host"
  role = aws_iam_role.role.name
}

resource "aws_iam_role_policy" "policy" {
  name = "${var.cluster_name}-host"
  role = aws_iam_role.role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ec2:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["elasticloadbalancing:*"],
      "Resource": ["*"]
    }
  ]
}
EOF

}

resource "aws_lb" "control_plane" {
  name               = "${var.cluster_name}-api-lb"
  internal           = false
  load_balancer_type = "network"
  subnets            = local.all_subnets

  tags = map("Cluster", var.cluster_name, local.kube_cluster_tag, "shared")
}

resource "aws_lb_target_group" "control_plane_api" {
  name     = "${var.cluster_name}-api"
  port     = 6443
  protocol = "TCP"
  vpc_id   = local.vpc_id
}

resource "aws_lb_listener" "control_plane_api" {
  load_balancer_arn = aws_lb.control_plane.arn
  port              = 6443
  protocol          = "TCP"

  default_action {
    target_group_arn = aws_lb_target_group.control_plane_api.arn
    type             = "forward"
  }
}

resource "aws_lb_target_group_attachment" "control_plane_api" {
  count            = 3
  target_group_arn = aws_lb_target_group.control_plane_api.arn
  target_id        = element(aws_instance.control_plane.*.id, count.index)
  port             = 6443
}

resource "aws_instance" "control_plane" {
  count = 3

  tags = map("Name", "${var.cluster_name}-control_plane-${count.index + 1}", local.kube_cluster_tag, "shared")

  instance_type          = var.control_plane_type
  iam_instance_profile   = aws_iam_instance_profile.profile.name
  ami                    = local.ami
  key_name               = aws_key_pair.deployer.key_name
  vpc_security_group_ids = [aws_security_group.common.id, aws_security_group.control_plane.id]
  availability_zone      = data.aws_availability_zones.available.names[count.index % local.az_count]
  subnet_id              = local.all_subnets[count.index % local.az_count]

  ebs_optimized = true

  root_block_device {
    volume_type = "gp2"
    volume_size = var.control_plane_volume_size
  }
}
