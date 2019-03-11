resource "hcloud_ssh_key" "kubeone" {
  name       = "kubeone-${var.cluster_name}"
  public_key = "${file("${var.ssh_public_key_file}")}"
}

resource "hcloud_server" "control_plane" {
  count       = "${var.control_plane_count}"
  name        = "${var.cluster_name}-control-plane-${count.index +1}"
  server_type = "cx31"
  image       = "centos-7"
  location    = "hel1"

  ssh_keys = [
    "${hcloud_ssh_key.kubeone.id}",
  ]

  labels = "${map("kubeone_cluster_name", "${var.cluster_name}")}"
}
