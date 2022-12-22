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

provider "google" {
  region  = var.region
  project = var.project
}

locals {
  zones_count        = length(data.google_compute_zones.available.names)
  zone_first         = data.google_compute_zones.available.names[0]
  kubeapi_endpoint   = var.disable_kubeapi_loadbalancer ? google_compute_instance.control_plane.0.network_interface.0.network_ip : google_compute_address.lb_ip.0.address
  loadbalancer_count = var.disable_kubeapi_loadbalancer ? 0 : 1

  cluster_autoscaler_min_replicas = var.cluster_autoscaler_min_replicas > 0 ? var.cluster_autoscaler_min_replicas : var.initial_machinedeployment_replicas
  cluster_autoscaler_max_replicas = var.cluster_autoscaler_max_replicas > 0 ? var.cluster_autoscaler_max_replicas : var.initial_machinedeployment_replicas
}

data "google_compute_zones" "available" {
}

data "google_compute_image" "control_plane_image" {
  family  = var.control_plane_image_family
  project = var.control_plane_image_project
}

data "google_compute_network" "network" {
  name = "default"
}

data "google_compute_subnetwork" "subnet" {
  name   = "default"
  region = var.region
}

resource "google_compute_firewall" "common" {
  name    = "${var.cluster_name}-common"
  network = data.google_compute_network.network.self_link

  allow {
    protocol = "tcp"
    ports    = [var.ssh_port]
  }

  source_ranges = [
    "0.0.0.0/0",
  ]
}

resource "google_compute_firewall" "control_plane" {
  name    = "${var.cluster_name}-control-plane"
  network = data.google_compute_network.network.self_link

  allow {
    protocol = "tcp"
    ports    = ["6443"]
  }

  source_ranges = [
    "0.0.0.0/0",
  ]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.cluster_name}-internal"
  network = data.google_compute_network.network.self_link

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "icmp"
  }

  source_ranges = [
    data.google_compute_subnetwork.subnet.ip_cidr_range,
  ]
}

resource "google_compute_firewall" "nodeports" {
  name    = "${var.cluster_name}-nodeports"
  network = data.google_compute_network.network.self_link

  allow {
    protocol = "tcp"
    ports    = ["30000-32767"]
  }

  source_ranges = [
    "0.0.0.0/0",
  ]
}


resource "google_compute_address" "lb_ip" {
  count = local.loadbalancer_count

  name = "${var.cluster_name}-lb-ip"
}

resource "google_compute_http_health_check" "control_plane" {
  name = "${var.cluster_name}-control-plane-health"

  port         = 10256
  request_path = "/healthz"

  timeout_sec        = 3
  check_interval_sec = 5
}

resource "google_compute_target_pool" "control_plane_pool" {
  name = "${var.cluster_name}-control-plane"

  instances = slice(
    google_compute_instance.control_plane.*.self_link,
    0,
    var.control_plane_target_pool_members_count,
  )

  health_checks = [
    google_compute_http_health_check.control_plane.self_link,
  ]
}

resource "google_compute_forwarding_rule" "control_plane" {
  count = local.loadbalancer_count

  name       = "${var.cluster_name}-apiserver"
  target     = google_compute_target_pool.control_plane_pool.self_link
  port_range = "6443-6443"
  ip_address = google_compute_address.lb_ip.0.address
}

resource "google_compute_instance" "control_plane" {
  count = var.control_plane_vm_count

  name         = "${var.cluster_name}-control-plane-${count.index + 1}"
  machine_type = var.control_plane_type
  zone         = data.google_compute_zones.available.names[count.index % local.zones_count]

  # Changing the machine_type, min_cpu_platform, or service_account on an
  # instance requires stopping it. To acknowledge this,
  # allow_stopping_for_update = true is required
  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      size  = var.control_plane_volume_size
      image = data.google_compute_image.control_plane_image.self_link
    }
  }

  network_interface {
    subnetwork = data.google_compute_subnetwork.subnet.self_link

    access_config {
      nat_ip = ""
    }
  }

  metadata = {
    sshKeys = "${var.ssh_username}:${file(var.ssh_public_key_file)}"
  }

  # https://cloud.google.com/sdk/gcloud/reference/alpha/compute/instances/set-scopes#--scopes
  # listing of possible scopes
  service_account {
    scopes = [
      "compute-rw",
      "logging-write",
      "monitoring-write",
      "service-control",
      "service-management",
      "storage-ro",
    ]
  }
}

