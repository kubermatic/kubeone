output "kubeone_hosts" {
  value = {
    control_plane = {
      cluster_name   = "${var.cluster_name}"
      public_address = "${hcloud_server.control_plane.*.ipv4_address}"
    }
  }
}
