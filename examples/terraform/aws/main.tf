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
  kube_cluster_tag = "kubernetes.io/cluster/${var.cluster_name}"
  ami              = var.ami == "" ? data.aws_ami.ami.id : var.ami
  zoneA            = data.aws_availability_zones.available.names[0]
  zoneB            = data.aws_availability_zones.available.names[1]
  zoneC            = data.aws_availability_zones.available.names[2]
  vpc_mask         = parseint(split("/", data.aws_vpc.selected.cidr_block)[1], 10)
  subnet_total     = pow(2, var.subnets_cidr - local.vpc_mask)
  subnet_newbits   = var.subnets_cidr - (32 - local.vpc_mask)
  worker_os        = var.worker_os == "" ? var.os : var.worker_os

  subnets = {
    (local.zoneA) = length(aws_subnet.public.*.id) > 0 ? aws_subnet.public[0].id : ""
    (local.zoneB) = length(aws_subnet.public.*.id) > 0 ? aws_subnet.public[1].id : ""
    (local.zoneC) = length(aws_subnet.public.*.id) > 0 ? aws_subnet.public[2].id : ""
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
}

data "aws_vpc" "selected" {
  id = var.vpc_id == "default" ? aws_default_vpc.default.id : var.vpc_id
}

data "aws_internet_gateway" "default" {
  filter {
    name   = "attachment.vpc-id"
    values = [data.aws_vpc.selected.id]
  }
}

resource "aws_default_vpc" "default" {}

resource "random_integer" "cidr_block" {
  min = 0
  max = local.subnet_total - 1
}

############################### NETWORKING SETUP ###############################

resource "aws_subnet" "public" {
  count                   = 3
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  vpc_id                  = data.aws_vpc.selected.id

  cidr_block = cidrsubnet(
    data.aws_vpc.selected.cidr_block,
    local.subnet_newbits,
    (random_integer.cidr_block.result + count.index) % local.subnet_total,
  )

  tags = map(
    "Name", "${var.cluster_name}-${data.aws_availability_zones.available.names[count.index]}",
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

################################### FIREWALL ###################################

resource "aws_security_group" "common" {
  name        = "${var.cluster_name}-common"
  description = "cluster common rules"
  vpc_id      = data.aws_vpc.selected.id

  tags = map(
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_security_group_rule" "ingress_self_allow_all" {
  type              = "ingress"
  security_group_id = aws_security_group.common.id

  description = "allow all incomming traffic from members of this group"
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
  count             = var.open_nodeports ? 1 : 0
  type              = "ingress"
  security_group_id = aws_security_group.common.id

  description = "open nodeports"
  from_port   = 30000
  to_port     = 32767
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group" "elb" {
  name        = "${var.cluster_name}-api-lb"
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

  tags = map(
    "Cluster", var.cluster_name,
  )
}

resource "aws_security_group" "ssh" {
  name        = "${var.cluster_name}-ssh"
  description = "ssh access"
  vpc_id      = data.aws_vpc.selected.id

  ingress {
    description = "allow incomming SSH"
    from_port   = var.ssh_port
    to_port     = var.ssh_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = map(
    "Cluster", var.cluster_name,
  )
}

################################## KUBE-API LB #################################

resource "aws_elb" "control_plane" {
  name            = "${var.cluster_name}-api-lb"
  internal        = var.internal_api_lb
  subnets         = aws_subnet.public.*.id
  security_groups = [aws_security_group.elb.id, aws_security_group.common.id]
  instances       = aws_instance.control_plane.*.id
  idle_timeout    = 600

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "HTTPS:6443/healthz"
    interval            = 30
  }

  tags = map(
    "Cluster", var.cluster_name,
    local.kube_cluster_tag, "shared",
  )
}

#################################### SSH KEY ###################################
resource "aws_key_pair" "deployer" {
  key_name   = "${var.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

##################################### IAM ######################################
resource "aws_iam_role" "role" {
  name = "${var.cluster_name}-host"

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
  name = "${var.cluster_name}-host"
  role = aws_iam_role.role.name
}

resource "aws_iam_role_policy" "policy" {
  name = "${var.cluster_name}-host"
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

############################ CONTROL PLANE INSTANCES ###########################

resource "aws_instance" "control_plane" {
  count                  = 3
  instance_type          = var.control_plane_type
  iam_instance_profile   = aws_iam_instance_profile.profile.name
  ami                    = local.ami
  key_name               = aws_key_pair.deployer.key_name
  vpc_security_group_ids = [aws_security_group.common.id]
  availability_zone      = data.aws_availability_zones.available.names[count.index]
  subnet_id              = local.subnets[data.aws_availability_zones.available.names[count.index]]
  ebs_optimized          = true

  root_block_device {
    volume_type = "gp2"
    volume_size = var.control_plane_volume_size
  }

  tags = map(
    "Name", "${var.cluster_name}-cp-${count.index + 1}",
    local.kube_cluster_tag, "shared",
  )
}

resource "aws_instance" "static_workers1" {
  count                  = var.static_workers_count
  instance_type          = var.worker_type
  iam_instance_profile   = aws_iam_instance_profile.profile.name
  ami                    = local.ami
  key_name               = aws_key_pair.deployer.key_name
  vpc_security_group_ids = [aws_security_group.common.id]
  availability_zone      = data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]
  subnet_id              = local.subnets[data.aws_availability_zones.available.names[count.index % length(data.aws_availability_zones.available.names)]]
  ebs_optimized          = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 50
  }

  tags = map(
    "Name", "${var.cluster_name}-workers1-${count.index + 1}",
    local.kube_cluster_tag, "shared",
  )
}

#################################### BASTION ###################################

resource "aws_instance" "bastion" {
  instance_type               = var.bastion_type
  ami                         = local.ami
  key_name                    = aws_key_pair.deployer.key_name
  vpc_security_group_ids      = [aws_security_group.common.id, aws_security_group.ssh.id]
  availability_zone           = data.aws_availability_zones.available.names[0]
  subnet_id                   = local.subnets[local.zoneA]
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 100
  }

  tags = map(
    "Cluster", var.cluster_name,
    "Name", "${var.cluster_name}-bastion",
    local.kube_cluster_tag, "shared",
  )
}
