provider "digitalocean" {}

locals {
  kube_cluster_tag = "kubernetes-cluster:${var.cluster_name}"
}

resource "digitalocean_tag" "kube_cluster_tag" {
  name = "${local.kube_cluster_tag}"
}

resource "digitalocean_ssh_key" "deployer" {
  name       = "${var.cluster_name}-deployer-key"
  public_key = "${file("${var.ssh_public_key_file}")}"
}

resource "digitalocean_droplet" "control_plane" {
  count = "${var.control_plane_count}"
  name  = "${var.cluster_name}-control-plane-${count.index + 1}"

  tags = [
    "${local.kube_cluster_tag}",
  ]

  image  = "${var.droplet_image}"
  region = "${var.region}"
  size   = "${var.droplet_size}"

  private_networking = "${var.droplet_private_networking}"
  monitoring         = "${var.droplet_monitoring}"
  ipv6               = "${var.droplet_ipv6}"

  ssh_keys = [
    "${digitalocean_ssh_key.deployer.id}",
  ]
}

resource "digitalocean_loadbalancer" "control_plane" {
  name   = "${var.cluster_name}-load-balancer"
  region = "${var.region}"

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

  droplet_tag = "${local.kube_cluster_tag}"
}
