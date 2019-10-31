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
  cluster_name     = random_pet.cluster_name.id
  kube_cluster_tag = "kubernetes.io/cluster/${local.cluster_name}"
  ami              = var.ami == "" ? data.aws_ami.ubuntu.id : var.ami
  zoneA            = data.aws_availability_zones.available.names[0]
  zoneB            = data.aws_availability_zones.available.names[1]
  zoneC            = data.aws_availability_zones.available.names[2]
  subnets = {
    "${local.zoneA}" = aws_subnet.public[0].id
    "${local.zoneB}" = aws_subnet.public[1].id
    "${local.zoneC}" = aws_subnet.public[2].id
  }
}

data "aws_availability_zones" "available" {}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_vpc" "selected" {
  id = var.vpc_id == "default" ? aws_default_vpc.default.id : var.vpc_id
}

resource "random_integer" "cidr_block" {
  min = 0
  max = var.subnet_total - 1
}

resource "random_pet" "cluster_name" {
  prefix = var.cluster_name
  length = 1
}

resource "aws_default_vpc" "default" {}

resource "aws_subnet" "public" {
  count                   = 3
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  vpc_id                  = data.aws_vpc.selected.id

  cidr_block = cidrsubnet(
    data.aws_vpc.selected.cidr_block,
    var.subnet_mask,
    (random_integer.cidr_block.result + count.index) % var.subnet_total,
  )

  tags = map(
    "Name", "${local.cluster_name}_${data.aws_availability_zones.available.names[count.index]}",
    "Cluster", local.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_key_pair" "deployer" {
  key_name   = "${local.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

resource "aws_security_group" "common" {
  name        = "${local.cluster_name}-common"
  description = "cluster common rules"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    local.kube_cluster_tag, "shared",
  )

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
  name        = "${local.cluster_name}-control_planes"
  description = "cluster control_planes"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    local.kube_cluster_tag, "shared",
  )

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "nodeports" {
  name        = "${local.cluster_name}-nodeports"
  description = "nodeport whitelist"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    local.kube_cluster_tag, "shared",
  )

  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "role" {
  name = "${local.cluster_name}-host"

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
  name = "${local.cluster_name}-host"
  role = aws_iam_role.role.name
}

resource "aws_iam_role_policy" "policy" {
  name = "${local.cluster_name}-host"
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

resource "aws_lb" "control_plane" {
  name               = "${local.cluster_name}-api-lb"
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public.*.id

  tags = map(
    "Cluster", local.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_lb_target_group" "control_plane_api" {
  name     = "${local.cluster_name}-api"
  port     = 6443
  protocol = "TCP"
  vpc_id   = data.aws_vpc.selected.id
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
  count                  = 3
  instance_type          = var.control_plane_type
  iam_instance_profile   = aws_iam_instance_profile.profile.name
  ami                    = local.ami
  key_name               = aws_key_pair.deployer.key_name
  vpc_security_group_ids = [aws_security_group.common.id, aws_security_group.control_plane.id, aws_security_group.nodeports.id]
  availability_zone      = data.aws_availability_zones.available.names[count.index]
  subnet_id              = local.subnets[data.aws_availability_zones.available.names[count.index]]
  ebs_optimized          = true

  root_block_device {
    volume_type = "gp2"
    volume_size = var.control_plane_volume_size
  }

  tags = map(
    "Name", "${local.cluster_name}-cp-${count.index + 1}",
    local.kube_cluster_tag, "shared",
  )
}
