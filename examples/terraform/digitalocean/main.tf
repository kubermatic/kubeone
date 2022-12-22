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
  kube_cluster_tag            = "kubernetes-cluster:${var.cluster_name}"
  control_plane_droplet_image = var.control_plane_droplet_image == "" ? var.image_references[var.os].image_name : var.control_plane_droplet_image
  worker_os                   = var.worker_os == "" ? var.image_references[var.os].worker_os : var.worker_os
  ssh_username                = var.ssh_username == "" ? var.image_references[var.os].ssh_username : var.ssh_username
  kubeapi_endpoint            = var.disable_kubeapi_loadbalancer ? digitalocean_droplet.control_plane.0.ipv4_address_private : digitalocean_loadbalancer.control_plane.0.ip
  loadbalancer_count          = var.disable_kubeapi_loadbalancer ? 0 : 1

  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas
}

resource "digitalocean_tag" "kube_cluster_tag" {
  name = local.kube_cluster_tag
}

resource "digitalocean_ssh_key" "deployer" {
  name       = "${var.cluster_name}-deployer-key"
  public_key = file(var.ssh_public_key_file)
}

resource "digitalocean_droplet" "control_plane" {
  count = var.control_plane_vm_count
  name  = "${var.cluster_name}-control-plane-${count.index + 1}"

  tags = [
    local.kube_cluster_tag,
    "kubeone",
  ]

  image              = local.control_plane_droplet_image
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
  count = local.loadbalancer_count

  name   = "${var.cluster_name}-lb"
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

