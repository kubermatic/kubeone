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

provider "digitalocean" {
}

locals {
  cluster_name     = random_pet.cluster_name.id
  kube_cluster_tag = "kubernetes-cluster:${local.cluster_name}"
}

resource "random_pet" "cluster_name" {
  prefix = var.cluster_name
  length = 1
}

resource "digitalocean_tag" "kube_cluster_tag" {
  name = local.kube_cluster_tag
}

resource "digitalocean_ssh_key" "deployer" {
  name       = "${local.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

resource "digitalocean_droplet" "control_plane" {
  count = 3
  name  = "${local.cluster_name}-cp-${count.index + 1}"

  tags = [
    local.kube_cluster_tag,
    "kubeone",
  ]

  image              = var.control_plane_droplet_image
  region             = var.region
  size               = var.control_plane_size
  private_networking = true
  monitoring         = false
  ipv6               = false

  ssh_keys = [
    digitalocean_ssh_key.deployer.id,
  ]
}

resource "digitalocean_loadbalancer" "control_plane" {
  name   = "${local.cluster_name}-lb"
  region = var.region

  forwarding_rule {
    entry_port     = 6443
    entry_protocol = "tcp"

    target_port     = 6443
    target_protocol = "tcp"
  }

  healthcheck {
    port     = 6443
    protocol = "tcp"
  }

  droplet_tag = local.kube_cluster_tag
}

