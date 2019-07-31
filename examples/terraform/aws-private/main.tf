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

resource "aws_default_vpc" "default" {
}

locals {
  kube_cluster_tag = "kubernetes.io/cluster/${var.cluster_name}"
  ami              = var.ami == "" ? data.aws_ami.ubuntu.id : var.ami
  zoneA            = data.aws_availability_zones.available.names[0]
  zoneB            = data.aws_availability_zones.available.names[1]
  zoneC            = data.aws_availability_zones.available.names[2]

  subnets = {
    public = {
      "${local.zoneA}" = aws_subnet.public[0].id
      "${local.zoneB}" = aws_subnet.public[1].id
      "${local.zoneC}" = aws_subnet.public[2].id
    }
    private = {
      "${local.zoneA}" = aws_subnet.private[0].id
      "${local.zoneB}" = aws_subnet.private[1].id
      "${local.zoneC}" = aws_subnet.private[2].id
    }
  }
}

################################# DATA SOURCES #################################
data "aws_vpc" "selected" {
  id = var.vpc_id == "default" ? aws_default_vpc.default.id : var.vpc_id
}

data "aws_internet_gateway" "default" {
  filter {
    name   = "attachment.vpc-id"
    values = [data.aws_vpc.selected.id]
  }
}

data "aws_availability_zones" "available" {}

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

############################### NETWORKING SETUP ###############################

resource "aws_eip" "nat" {
  count = length(data.aws_availability_zones.available.names)
  vpc   = true
}

resource "aws_subnet" "public" {
  count             = length(data.aws_availability_zones.available.names)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block = cidrsubnet(
    data.aws_vpc.selected.cidr_block,
    var.subnet_netmask_bits,
    var.subnet_offset + count.index,
  )
  map_public_ip_on_launch = true
  vpc_id                  = data.aws_vpc.selected.id

  tags = map(
    "Name", "${var.cluster_name}_public_${data.aws_availability_zones.available.names[count.index]}",
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_subnet" "private" {
  count             = length(data.aws_availability_zones.available.names)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block = cidrsubnet(
    data.aws_vpc.selected.cidr_block,
    var.subnet_netmask_bits,
    var.subnet_offset + length(aws_subnet.public.*.id) + count.index,
  )
  map_public_ip_on_launch = false
  vpc_id                  = data.aws_vpc.selected.id

  tags = map(
    "Name", "${var.cluster_name}_private_${data.aws_availability_zones.available.names[count.index]}",
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_nat_gateway" "main" {
  count         = length(data.aws_availability_zones.available.names)
  allocation_id = element(aws_eip.nat.*.id, count.index)
  subnet_id     = element(aws_subnet.public.*.id, count.index)
}

resource "aws_route_table" "public" {
  vpc_id = data.aws_vpc.selected.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = data.aws_internet_gateway.default.id
  }

  tags = map(local.kube_cluster_tag, "shared")
}

resource "aws_route_table" "private" {
  count  = length(data.aws_availability_zones.available.names)
  vpc_id = data.aws_vpc.selected.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = element(aws_nat_gateway.main.*.id, count.index)
  }

  tags = map(local.kube_cluster_tag, "shared")
}

resource "aws_route_table_association" "public" {
  count          = length(data.aws_availability_zones.available.names)
  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count          = length(data.aws_availability_zones.available.names)
  subnet_id      = element(aws_subnet.private.*.id, count.index)
  route_table_id = element(aws_route_table.private.*.id, count.index)
}

################################ LOAD BALLANCER ################################

resource "aws_lb" "control_plane" {
  name               = "${var.cluster_name}-api-lb"
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public.*.id

  tags = map(
    "Name", "${var.cluster_name}-control_plane",
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_lb_target_group" "control_plane_api" {
  name     = "${var.cluster_name}-api"
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

############################### SECURITY GROUPS ################################
resource "aws_security_group" "common" {
  name        = "${var.cluster_name}-common"
  description = "cluster common rules"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    "Name", "${var.cluster_name}-common",
    "Cluster", local.kube_cluster_tag,
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
  name        = "${var.cluster_name}-control_planes"
  description = "cluster control_planes"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    "Name", "${var.cluster_name}-control_plane",
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

##################################### IAM ######################################
resource "aws_iam_role" "control_plane" {
  name = "${var.cluster_name}-host"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
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

resource "aws_iam_instance_profile" "control_plane" {
  name = "${var.cluster_name}-control-plane"
  role = aws_iam_role.control_plane.name
}

resource "aws_iam_role_policy" "control_plane" {
  name = "${var.cluster_name}-control-plane"
  role = aws_iam_role.control_plane.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow",
        Action   = ["ec2:*"],
        Resource = ["*"],
      },
      {
        Effect   = "Allow",
        Action   = ["elasticloadbalancing:*"],
        Resource = ["*"],
      }
    ]
  })
}

resource "aws_iam_role" "workers" {
  name = "${var.cluster_name}-workers"
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

resource "aws_iam_instance_profile" "workers" {
  name = "${var.cluster_name}-workers"
  role = aws_iam_role.workers.name
}

resource "aws_iam_role_policy" "workers" {
  name = "${var.cluster_name}-workers"
  role = aws_iam_role.workers.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "ec2:Describe*"
        ],
        Effect   = "Allow",
        Resource = "*"
      }
    ]
  })
}

################################## SSH ACCESS ##################################

resource "aws_key_pair" "deployer" {
  key_name   = "${var.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

################################## INSTANCES ###################################

resource "aws_instance" "control_plane" {
  count = 3

  tags = map(
    "Cluster", var.cluster_name,
    "Name", "${var.cluster_name}-control_plane-${count.index + 1}",
    local.kube_cluster_tag, "shared",
  )

  instance_type               = var.control_plane_type
  iam_instance_profile        = aws_iam_instance_profile.control_plane.name
  ami                         = local.ami
  key_name                    = aws_key_pair.deployer.key_name
  vpc_security_group_ids      = [aws_security_group.common.id, aws_security_group.control_plane.id]
  availability_zone           = data.aws_availability_zones.available.names[count.index]
  subnet_id                   = local.subnets["private"][data.aws_availability_zones.available.names[count.index]]
  associate_public_ip_address = false
  ebs_optimized               = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 100
  }
}

resource "aws_instance" "bastion" {
  tags = map(
    "Cluster", var.cluster_name,
    "Name", "${var.cluster_name}-bastion",
    local.kube_cluster_tag, "shared",
  )

  instance_type               = "t3.nano"
  ami                         = local.ami
  key_name                    = aws_key_pair.deployer.key_name
  vpc_security_group_ids      = [aws_security_group.common.id]
  availability_zone           = data.aws_availability_zones.available.names[0]
  subnet_id                   = local.subnets["public"][local.zoneA]
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 100
  }
}

