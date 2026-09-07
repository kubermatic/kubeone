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
  kube_cluster_tag = "kubernetes.io/cluster/${var.name}"
  ami              = var.ami == "" ? data.aws_ami.ami.id : var.ami
  zoneA            = data.aws_availability_zones.available.names[0]
  subnets = {
    (local.zoneA) = length(aws_subnet.public[*].id) > 0 ? aws_subnet.public[0].id : ""
  }
}

################################# DATA SOURCES #################################

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "ami" {
  most_recent = true
  owners      = var.ami_filters[var.os].owners

  filter {
    name   = "name"
    values = var.ami_filters[var.os].image_name
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id == "default" ? aws_default_vpc.default.id : var.vpc_id
}

resource "aws_default_vpc" "default" {}

############################### NETWORKING SETUP ###############################

resource "aws_subnet" "public" {
  count                   = 1
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  vpc_id                  = data.aws_vpc.selected.id

  cidr_block = var.public_cidr

  tags = tomap({
    "Name"                   = "${var.name}-${data.aws_availability_zones.available.names[count.index]}",
    "Cluster"                = var.name,
    (local.kube_cluster_tag) = "shared",
    "keep"                   = "true",
  })
}

################################### FIREWALL ###################################

resource "aws_security_group" "common" {
  name        = "${var.name}-common"
  description = "cluster common rules"
  vpc_id      = data.aws_vpc.selected.id

  tags = tomap({
    "Cluster"                = var.name,
    (local.kube_cluster_tag) = "shared",
    "keep"                   = "true",
  })
}

resource "aws_security_group_rule" "ingress_self_allow_all" {
  type              = "ingress"
  security_group_id = aws_security_group.common.id

  description = "allow all incoming traffic from members of this group"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  self        = true
}

resource "aws_security_group_rule" "egress_allow_all" {
  type              = "egress"
  security_group_id = aws_security_group.common.id

  description = "allow all outgoing traffic"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "nodeports" {
  type              = "ingress"
  security_group_id = aws_security_group.common.id

  description = "open nodeports"
  from_port   = 30000
  to_port     = 32767
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group" "elb" {
  name        = "${var.name}-api-lb"
  description = "kube-api firewall"
  vpc_id      = data.aws_vpc.selected.id

  egress {
    description = "allow all outgoing traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "allow anyone to connect to tcp/6443"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = tomap({
    "Cluster" = var.name,
    "keep"    = "true",
  })
}

resource "aws_security_group" "ssh" {
  name        = "${var.name}-ssh"
  description = "ssh access"
  vpc_id      = data.aws_vpc.selected.id

  ingress {
    description = "allow incoming SSH"
    from_port   = var.ssh_port
    to_port     = var.ssh_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = tomap({
    "Cluster" = var.name,
    "keep"    = "true",
  })
}

##################################### IAM ######################################
resource "aws_iam_role" "role" {
  name = "${var.name}-host"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "ec2.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "profile" {
  name = "${var.name}-host"
  role = aws_iam_role.role.name
}

resource "aws_iam_role_policy" "policy" {
  name = "${var.name}-host"
  role = aws_iam_role.role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = ["ec2:*"],
        Resource = ["*"]
      },
      {
        Effect   = "Allow",
        Action   = ["elasticloadbalancing:*"],
        Resource = ["*"]
      }
    ]
  })
}
