output "kubeone_api" {
  value = {
    endpoint = "${digitalocean_loadbalancer.control_plane.ip}"
  }
}

output "kubeone_hosts" {
  value = {
    control_plane = {
      public_address  = "${digitalocean_droplet.control_plane.*.ipv4_address}"
      private_address = "${digitalocean_droplet.control_plane.*.ipv4_address_private}"
      ssh_user        = "root"
      ssh_port        = "${var.ssh_port}"

      ssh_agent_socket     = "${var.ssh_agent_socket}"
      ssh_private_key_file = "${var.ssh_private_key_file}"
    }
  }
}
