provider "google" {
  credentials = "${file("~/credentials.json")}"
  region      = "${var.region}"
  project     = "${var.project}"
}

locals {
  zones_count   = "${length(data.google_compute_zones.available.names)}"
}

data "google_compute_zones" "available" {}

data "google_compute_image" "control_plane_image" {
  family  = "${var.control_plane_image_family}"
  project = "${var.control_plane_image_project}"
}

resource "google_compute_network" "network" {
  name                    = "${var.cluster_name}"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.cluster_name}-subnet"
  network       = "${google_compute_network.network.self_link}"
  ip_cidr_range = "${var.cluster_network_cidr}"
}

resource "google_compute_firewall" "common" {
  name    = "${var.cluster_name}-common"
  network = "${google_compute_network.network.self_link}"

  allow {
    protocol  = "tcp"
    ports     = ["${var.ssh_port}"]    
  }

  source_ranges = [
    "0.0.0.0/0"
  ]
}

resource "google_compute_firewall" "control_plane" {
  name    = "${var.cluster_name}-control-plane"
  network = "${google_compute_network.network.self_link}"

  allow {
    protocol = "tcp"
    ports    = ["6443"]
  }

  source_ranges = [
    "0.0.0.0/0"
  ]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.cluster_name}-internal"
  network = "${google_compute_network.network.self_link}"

  allow {
    protocol  = "tcp"
    ports     = ["0-65535"]
  }
  allow {
    protocol  = "udp"
    ports     = ["0-65535"]
  }
  allow {
    protocol  = "icmp"   
  }

  source_ranges = [
    "${var.cluster_network_cidr}"
  ]
}

resource "google_compute_address" "lb_ip" {
  name = "${var.cluster_name}-lb-ip"
}

resource "google_compute_health_check" "control_plane" {
  name = "${var.cluster_name}-control-plane-health"

  timeout_sec         = 5
  check_interval_sec  = 3

  https_health_check {
    port          = "6443"
    request_path  = "/healthz"
  }
}

resource "google_compute_target_pool" "control_plane_pool" {
  name  = "${var.cluster_name}-pool"
    
  #instances = [
  # TODO: expand this with all instances  
  #]

  health_checks = [
    "${google_compute_health_check.control_plane.self_link}"
  ]
}

resource "google_compute_forwarding_rule" "control_plane" {
  name       = "${var.cluster_name}-forwarding"
  target     = "${google_compute_target_pool.control_plane_pool.self_link}"
  port_range = "6443"
  ip_address = "${google_compute_address.lb_ip.self_link}"
}

resource "google_compute_instance" "control_plane" {
  count = "${var.control_plane_count}"

  name          = "${var.cluster_name}-control-plane-${count.index+1}"
  machine_type  = "${var.control_plane_type}"
  zone          = "${data.google_compute_zones.available.names[count.index % local.zones_count]}"
  
  boot_disk {
    initialize_params {
      size  = "${var.control_plane_volume_size}"
      image = "${data.google_compute_image.control_plane_image.self_link}"      
    }
  }

  network_interface {
    network = "${google_compute_network.network.self_link}"
  }  

  tags = ["kubernetes.io/cluster/${var.cluster_name}: shared"]

  service_account {
    scopes = ["compute-rw", "storage-ro"]
  }
}
